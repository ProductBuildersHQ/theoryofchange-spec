package validate

import (
	"testing"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

func TestMetricAndIndicatorIssues(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(d *toc.Document)
		code   string
		path   string
	}{
		{"metric without name", func(d *toc.Document) { d.Metrics[0].Name = "" }, "VAL-MET-001", "/metrics/0/name"},
		{"metric without definition", func(d *toc.Document) { d.Metrics[0].Definition = " " }, "VAL-MET-001", "/metrics/0/definition"},
		{"metric without formula", func(d *toc.Document) { d.Metrics[0].Formula = "" }, "VAL-MET-001", "/metrics/0/formula"},
		{"metric aggregation invalid", func(d *toc.Document) { d.Metrics[0].Aggregation = "median" }, "VAL-ENUM-001", "/metrics/0/aggregation"},
		{"metric layer invalid", func(d *toc.Document) { d.Metrics[0].Layer = "global" }, "VAL-ENUM-001", "/metrics/0/layer"},
		{"weighted mean without weight", func(d *toc.Document) { d.Metrics[1].WeightMetricID = "" }, "VAL-MET-003", "/metrics/1/weightMetricId"},
		{"weight on a non-weighted metric", func(d *toc.Document) { d.Metrics[0].WeightMetricID = "met-provider-share" }, "VAL-MET-004", "/metrics/0/weightMetricId"},
		{"metric weights itself", func(d *toc.Document) { d.Metrics[1].WeightMetricID = "met-provider-share" }, "VAL-MET-004", "/metrics/1/weightMetricId"},
		{"weight metric unknown", func(d *toc.Document) { d.Metrics[1].WeightMetricID = "met-ghost" }, "VAL-REF-001", "/metrics/1/weightMetricId"},
		{"glossary term without term", func(d *toc.Document) { d.Glossary[0].Term = "" }, "VAL-GLO-001", "/glossary/0/term"},
		{"glossary term without definition", func(d *toc.Document) { d.Glossary[0].Definition = "" }, "VAL-GLO-001", "/glossary/0/definition"},
		{"glossary related unknown", func(d *toc.Document) { d.Glossary[0].RelatedIDs = []string{"ghost"} }, "VAL-REF-001", "/glossary/0/relatedIds/0"},
		{"indicator metric unknown", func(d *toc.Document) { d.Indicators[0].MetricID = "met-ghost" }, "VAL-REF-001", "/indicators/0/metricId"},
		{"indicator metric is an outcome", func(d *toc.Document) { d.Indicators[0].MetricID = "oc-idp-support" }, "VAL-REF-002", "/indicators/0/metricId"},
		{"indicator pathway unknown", func(d *toc.Document) { d.Indicators[0].PathwayIDs = []string{"pw-ghost"} }, "VAL-REF-001", "/indicators/0/pathwayIds/0"},
		{"indicator claim invalid", func(d *toc.Document) { d.Indicators[0].Claim = "proof" }, "VAL-ENUM-001", "/indicators/0/claim"},
		{"rollup source unknown", func(d *toc.Document) { d.Indicators[5].RollupFrom[0] = "ind-ghost" }, "VAL-REF-001", "/indicators/5/rollupFrom/0"},
		{"rollup without a metric", func(d *toc.Document) { d.Indicators[5].MetricID = "" }, "VAL-MET-005", "/indicators/5/rollupFrom"},
		{"rollup of a non-aggregating metric", func(d *toc.Document) { d.Metrics[0].Aggregation = toc.AggregationNone }, "VAL-MET-002", "/indicators/5/rollupFrom"},
		{"rollup source uses another metric", func(d *toc.Document) { d.Indicators[4].MetricID = "met-provider-share" }, "VAL-MET-002", "/indicators/5/rollupFrom/1"},
		{"rollup from itself", func(d *toc.Document) { d.Indicators[5].RollupFrom = []string{"ind-idp-total"} }, "VAL-MET-006", "/indicators/5/rollupFrom/0"},
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
	if (toc.Indicator{}).EffectiveClaim() != toc.ClaimContribution {
		t.Fatal("an unset claim must default to contribution")
	}
	if (toc.Indicator{Claim: toc.ClaimAttribution}).EffectiveClaim() != toc.ClaimAttribution {
		t.Fatal("an explicit claim must be kept")
	}
}
