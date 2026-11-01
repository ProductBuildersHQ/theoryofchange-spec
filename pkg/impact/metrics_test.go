package impact

import (
	"slices"
	"testing"
)

// metricsDoc has two pathways, an intervention, and one outcome at each level
// the tiers care about.
func metricsDoc() *Model {
	return &Model{
		Outcomes: []Outcome{
			{ID: "oc-out", Title: "Output", Level: LevelOutput},
			{ID: "oc-st", Title: "Short", Level: LevelShortTerm},
			{ID: "oc-lt", Title: "Long", Level: LevelLongTerm},
			{ID: "oc-imp", Title: "Impact", Level: LevelImpact},
		},
		Interventions: []Intervention{{ID: "int-a", Title: "Activity", Targets: []string{"oc-out"}}},
		Assumptions:   []Assumption{{ID: "asm-a", Statement: "Holds"}},
		Pathways: []Pathway{
			{ID: "pw-1", Title: "One", OutcomeIDs: []string{"oc-out", "oc-st"}, InterventionIDs: []string{"int-a"}},
			{ID: "pw-2", Title: "Two", OutcomeIDs: []string{"oc-lt"}},
			{ID: "pw-gone", Title: "Rejected", OutcomeIDs: []string{"oc-st"}, Provenance: &Provenance{Basis: BasisRejected}},
		},
	}
}

func TestIndicatorTier(t *testing.T) {
	d := metricsDoc()
	for measures, want := range map[string]Tier{
		"int-a":   TierActivity,
		"oc-out":  TierOutput,
		"oc-st":   TierOutcome,
		"oc-lt":   TierOutcome,
		"oc-imp":  TierImpact,
		"asm-a":   TierAssumption,
		"pw-1":    TierPathway,
		"missing": "",
	} {
		if got := d.IndicatorTier(Indicator{ID: "ind", Measures: measures}); got != want {
			t.Errorf("measuring %q: tier = %q, want %q", measures, got, want)
		}
	}
}

func TestIndicatorPathways(t *testing.T) {
	d := metricsDoc()
	tests := []struct {
		name string
		ind  Indicator
		want []string
	}{
		{"by the outcome it measures", Indicator{Measures: "oc-out"}, []string{"pw-1"}},
		{"by the intervention it measures", Indicator{Measures: "int-a"}, []string{"pw-1"}},
		{"by measuring the pathway itself", Indicator{Measures: "pw-2"}, []string{"pw-2"}},
		{"by naming the pathway", Indicator{Measures: "oc-imp", PathwayIDs: []string{"pw-2"}}, []string{"pw-2"}},
		{"naming and measuring add up", Indicator{Measures: "oc-out", PathwayIDs: []string{"pw-2"}}, []string{"pw-1", "pw-2"}},
		{"a repeated name is one pathway", Indicator{Measures: "oc-out", PathwayIDs: []string{"pw-1", "pw-1"}}, []string{"pw-1"}},
		{"a rejected pathway is never served", Indicator{Measures: "oc-st", PathwayIDs: []string{"pw-gone"}}, []string{"pw-1"}},
		{"a pathway that does not exist is not served", Indicator{Measures: "oc-imp", PathwayIDs: []string{"pw-nope"}}, nil},
		{"an assumption belongs to no pathway", Indicator{Measures: "asm-a"}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := d.IndicatorPathways(tt.ind)
			if !slices.Equal(got, tt.want) {
				t.Errorf("IndicatorPathways = %v, want %v", got, tt.want)
			}
			for _, id := range []string{"pw-1", "pw-2", "pw-gone", "pw-nope"} {
				if d.IndicatorServes(tt.ind, id) != slices.Contains(tt.want, id) {
					t.Errorf("IndicatorServes(%q) disagrees with IndicatorPathways %v", id, got)
				}
			}
		})
	}
}

func TestAllTiersAreDistinct(t *testing.T) {
	seen := map[Tier]bool{}
	for _, tier := range AllTiers() {
		if tier == "" || seen[tier] {
			t.Errorf("tier %q is empty or repeated", tier)
		}
		seen[tier] = true
	}
}
