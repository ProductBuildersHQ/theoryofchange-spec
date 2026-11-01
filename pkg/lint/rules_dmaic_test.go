package lint

import (
	"fmt"
	"slices"
	"testing"

	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
)

// dmaicDoc is the chain document with its pathway marked for DMAIC and one
// weekly indicator of the pathway's outcome that is measured from a system.
func dmaicDoc() *impact.Model {
	d := chainDoc()
	d.Pathways[0].Projections = []impact.Projection{impact.ProjectionDMAIC}
	addIndicator(d, "ind-weekly", "oc-s1", func(i *impact.Indicator) {
		i.Frequency, i.DataSource = "P1W", "Learning platform"
	})
	return d
}

// Each case seeds one defect, checks the rule reports it, repairs it, and checks
// the rule falls quiet.
func TestDMAICRulesFindSeededDefectsAndAcceptRepairs(t *testing.T) {
	survey := func(respondents int) func(*impact.Indicator) {
		return func(i *impact.Indicator) {
			i.Frequency = "P1W"
			i.Collection = &impact.Collection{Mode: impact.ModeSurvey, System: "Survey tool", RespondentsPerYear: respondents}
		}
	}
	tests := []struct {
		rule, element string
		defect, fix   func(d *impact.Model)
	}{
		{"IMP-DMA-001", "ind-weekly",
			func(d *impact.Model) { indicator(d, "ind-weekly").Frequency = "" },
			func(d *impact.Model) { indicator(d, "ind-weekly").Frequency = "P1W" }},
		{"IMP-DMA-001", "ind-weekly",
			func(d *impact.Model) { indicator(d, "ind-weekly").DataSource = "" },
			func(d *impact.Model) {
				indicator(d, "ind-weekly").Collection = &impact.Collection{Mode: impact.ModePassive, System: "Learning platform"}
			}},
		{"IMP-DMA-002", "ind-weekly",
			func(d *impact.Model) { indicator(d, "ind-weekly").Frequency = "P1M" },
			func(d *impact.Model) { indicator(d, "ind-weekly").Frequency = "P2W" }},
		{"IMP-DMA-002", "ind-weekly",
			func(d *impact.Model) { indicator(d, "ind-weekly").Frequency = "P3M" },
			func(d *impact.Model) { d.Pathways[0].Projections = nil }},
		{"IMP-DMA-003", "ind-nps",
			func(d *impact.Model) { addIndicator(d, "ind-nps", "oc-s1", survey(0)) },
			func(d *impact.Model) { indicator(d, "ind-nps").Collection.RespondentsPerYear = 5200 }},
		// 780 a year, weekly, is 15 per measurement; monthly it is 65.
		{"IMP-DMA-003", "ind-nps",
			func(d *impact.Model) { addIndicator(d, "ind-nps", "oc-s1", survey(780)) },
			func(d *impact.Model) { indicator(d, "ind-nps").Frequency = "P1M" }},
		// A sample of half the people asked halves the responses.
		{"IMP-DMA-003", "ind-nps",
			func(d *impact.Model) {
				addIndicator(d, "ind-nps", "oc-s1", survey(2600))
				indicator(d, "ind-nps").Collection.SampleRate = 0.5
			},
			func(d *impact.Model) { indicator(d, "ind-nps").Collection.SampleRate = 0 }},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%s/%d", tt.rule, i), func(t *testing.T) {
			d := dmaicDoc()
			tt.defect(d)
			if !slices.Contains(fires(t, d, tt.rule), tt.element) {
				t.Fatalf("%s did not report %q after the defect; reported %v", tt.rule, tt.element, fires(t, d, tt.rule))
			}
			tt.fix(d)
			if slices.Contains(fires(t, d, tt.rule), tt.element) {
				t.Errorf("%s still reports %q after the repair", tt.rule, tt.element)
			}
		})
	}
}

// A pathway that is not marked for DMAIC is not held to DMAIC's cadence, even
// when its indicators are measured once a year.
func TestDMAICRulesAreQuietForUnmarkedPathways(t *testing.T) {
	d := chainDoc()
	addIndicator(d, "ind-yearly", "oc-s1", func(i *impact.Indicator) {
		i.Frequency = "P1Y"
		i.Collection = &impact.Collection{Mode: impact.ModeSurvey, RespondentsPerYear: 12}
	})
	for _, rule := range []string{"IMP-DMA-001", "IMP-DMA-002", "IMP-DMA-003"} {
		if got := fires(t, d, rule); len(got) != 0 {
			t.Errorf("%s reported %v on a pathway not marked for DMAIC", rule, got)
		}
	}
}

func TestPerYear(t *testing.T) {
	tests := []struct {
		iso  string
		want float64
		ok   bool
	}{
		{"P1Y", 1, true},
		{"P1M", 12, true},
		{"P3M", 4, true},
		{"P1W", 52.18, true},
		{"P2W", 26.09, true},
		{"P1D", 365.24, true},
		{"PT12H", 730.49, true},
		{"", 0, false},
		{"P0D", 0, false},
		{"monthly", 0, false},
	}
	for _, tt := range tests {
		got, ok := perYear(tt.iso)
		if ok != tt.ok || (ok && fmt.Sprintf("%.2f", got) != fmt.Sprintf("%.2f", tt.want)) {
			t.Errorf("perYear(%q) = %.2f, %v; want %.2f, %v", tt.iso, got, ok, tt.want, tt.ok)
		}
	}
}
