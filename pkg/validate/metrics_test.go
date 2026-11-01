package validate

import (
	"testing"

	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
)

func TestMetricAndIndicatorIssues(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(d *impact.Model)
		code   string
		path   string
	}{
		{"metric without name", func(d *impact.Model) { d.Metrics[0].Name = "" }, "VAL-MET-001", "/metrics/0/name"},
		{"metric without definition", func(d *impact.Model) { d.Metrics[0].Definition = " " }, "VAL-MET-001", "/metrics/0/definition"},
		{"metric without formula", func(d *impact.Model) { d.Metrics[0].Formula = "" }, "VAL-MET-001", "/metrics/0/formula"},
		{"metric aggregation invalid", func(d *impact.Model) { d.Metrics[0].Aggregation = "geometric_mean" }, "VAL-ENUM-001", "/metrics/0/aggregation"},
		{"metric layer invalid", func(d *impact.Model) { d.Metrics[0].Layer = "global" }, "VAL-ENUM-001", "/metrics/0/layer"},
		{"weighted mean without weight", func(d *impact.Model) { d.Metrics[1].WeightMetricID = "" }, "VAL-MET-003", "/metrics/1/weightMetricId"},
		{"weight on a non-weighted metric", func(d *impact.Model) { d.Metrics[0].WeightMetricID = "met-provider-share" }, "VAL-MET-004", "/metrics/0/weightMetricId"},
		{"metric weights itself", func(d *impact.Model) { d.Metrics[1].WeightMetricID = "met-provider-share" }, "VAL-MET-004", "/metrics/1/weightMetricId"},
		{"weight metric unknown", func(d *impact.Model) { d.Metrics[1].WeightMetricID = "met-ghost" }, "VAL-REF-001", "/metrics/1/weightMetricId"},
		{"glossary term without term", func(d *impact.Model) { d.Glossary[0].Term = "" }, "VAL-GLO-001", "/glossary/0/term"},
		{"glossary term without definition", func(d *impact.Model) { d.Glossary[0].Definition = "" }, "VAL-GLO-001", "/glossary/0/definition"},
		{"glossary related unknown", func(d *impact.Model) { d.Glossary[0].RelatedIDs = []string{"ghost"} }, "VAL-REF-001", "/glossary/0/relatedIds/0"},
		{"indicator metric unknown", func(d *impact.Model) { d.Indicators[0].MetricID = "met-ghost" }, "VAL-REF-001", "/indicators/0/metricId"},
		{"indicator metric is an outcome", func(d *impact.Model) { d.Indicators[0].MetricID = "oc-idp-support" }, "VAL-REF-002", "/indicators/0/metricId"},
		{"indicator pathway unknown", func(d *impact.Model) { d.Indicators[0].PathwayIDs = []string{"pw-ghost"} }, "VAL-REF-001", "/indicators/0/pathwayIds/0"},
		{"indicator claim invalid", func(d *impact.Model) { d.Indicators[0].Claim = "proof" }, "VAL-ENUM-001", "/indicators/0/claim"},
		{"rollup source unknown", func(d *impact.Model) { d.Indicators[5].RollupFrom[0] = "ind-ghost" }, "VAL-REF-001", "/indicators/5/rollupFrom/0"},
		{"rollup without a metric", func(d *impact.Model) { d.Indicators[5].MetricID = "" }, "VAL-MET-005", "/indicators/5/rollupFrom"},
		{"rollup of a non-aggregating metric", func(d *impact.Model) { d.Metrics[0].Aggregation = impact.AggregationNone }, "VAL-MET-002", "/indicators/5/rollupFrom"},
		{"rollup source uses another metric", func(d *impact.Model) { d.Indicators[4].MetricID = "met-provider-share" }, "VAL-MET-002", "/indicators/5/rollupFrom/1"},
		{"rollup from itself", func(d *impact.Model) { d.Indicators[5].RollupFrom = []string{"ind-idp-total"} }, "VAL-MET-006", "/indicators/5/rollupFrom/0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := fullDoc(t)
			if got := doc.Indicators[5].ID; got != "ind-idp-total" {
				t.Fatalf("test setup: Indicators[5] is %q", got)
			}
			tt.mutate(doc)
			for _, i := range Document(doc) {
				if i.Code == tt.code && i.Path == tt.path {
					return
				}
			}
			t.Fatalf("expected %s at %s; got:%s", tt.code, tt.path, describe(Document(doc)))
		})
	}
}

func TestEffectiveClaimDefaultsToContribution(t *testing.T) {
	if (impact.Indicator{}).EffectiveClaim() != impact.ClaimContribution {
		t.Fatal("an unset claim must default to contribution")
	}
	if (impact.Indicator{Claim: impact.ClaimAttribution}).EffectiveClaim() != impact.ClaimAttribution {
		t.Fatal("an explicit claim must be kept")
	}
}

func TestIndicatorsMayMeasureAPathway(t *testing.T) {
	// A network-wide roll-up has no single outcome to measure; a pathway is
	// an honest subject for it.
	doc := fullDoc(t)
	doc.Indicators[5].Measures = "pw-supply"
	if issues := Document(doc); len(issues) != 0 {
		t.Fatalf("measuring a pathway should be valid:%s", describe(issues))
	}
}

func TestNewEnumValuesAreAccepted(t *testing.T) {
	doc := fullDoc(t)
	doc.Metrics[0].Aggregation = impact.AggregationMedian
	doc.Metrics[1].Aggregation = impact.AggregationMode
	doc.Metrics[1].WeightMetricID = ""
	doc.Actors[0].Type = impact.ActorGovernment
	doc.ExternalFactors[0].Kind = impact.FactorOpportunity
	// The roll-up cannot use an aggregation the library defines as median? It can:
	// median is an aggregating metric, so a roll-up from it remains valid.
	if issues := Document(doc); len(issues) != 0 {
		t.Fatalf("new enum values rejected:%s", describe(issues))
	}
}

func TestNestedGroupMustTargetTheSameOutcome(t *testing.T) {
	doc := fullDoc(t)
	doc.Groups = append(doc.Groups, impact.PreconditionGroup{
		ID: "grp-outer", Target: "oc-enterprise-require",
		Members: []string{"grp-supply", "oc-reference-impl"},
		Logic:   impact.Logic{Type: impact.LogicAny},
	})
	if issues := Document(doc); len(issues) != 0 {
		t.Fatalf("a nested group with the same target is valid:%s", describe(issues))
	}
	doc.Groups[len(doc.Groups)-1].Target = "oc-safer-agents"
	for _, i := range Document(doc) {
		if i.Code == "VAL-GRP-004" && i.Path == "/preconditionGroups/1/members/0" {
			return
		}
	}
	t.Fatalf("expected VAL-GRP-004; got:%s", describe(Document(doc)))
}
