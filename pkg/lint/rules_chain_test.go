package lint

import (
	"fmt"
	"slices"
	"testing"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

// chainDoc is the clean base with one program pathway that owns the short-term
// outcomes, an audience, and indicators of the kinds the rules tell apart.
func chainDoc() *toc.Document {
	d := clean()
	d.Pathways = []toc.Pathway{{
		ID: "pw-1", Title: "Program", Kind: toc.PathwayProgram,
		OutcomeIDs: []string{"oc-s1", "oc-a"}, InterventionIDs: []string{"int-1"},
	}}
	d.Audiences = []toc.Audience{{ID: "aud-1", Name: "Board"}}
	return d
}

func ptr[T any](v T) *T { return &v }

// Each case seeds one defect, checks the rule reports it, repairs it, and checks
// the rule falls quiet. These rules check the chain from what is done to what
// changes, the numbers each audience is told, and the North Star.
func TestChainAndHeadlineRulesFindSeededDefectsAndAcceptRepairs(t *testing.T) {
	headline := func(d *toc.Document, ids ...string) {
		d.Audiences[0].Headline = nil
		for _, id := range ids {
			d.Audiences[0].Headline = append(d.Audiences[0].Headline, toc.HeadlineMetric{IndicatorID: id})
		}
	}
	forAudience := func(i *toc.Indicator) { i.AudienceIDs = []string{"aud-1"} }
	focus := func(d *toc.Document, f toc.Focus) { d.Pathways[0].Focus = &f }

	tests := []struct {
		rule, element string
		defect, fix   func(d *toc.Document)
	}{
		// ---- from activity to impact
		{"TOC-INT-003", "int-1", func(d *toc.Document) {},
			func(d *toc.Document) { d.Interventions[0].Outputs = []string{"Mixers held"} }},
		{"TOC-INT-003", "int-1", func(d *toc.Document) {},
			func(d *toc.Document) {
				d.Outcomes = append(d.Outcomes, toc.Outcome{ID: "oc-out", Title: "Mixers held", Level: toc.LevelOutput})
				d.Interventions[0].Targets = append(d.Interventions[0].Targets, "oc-out")
			}},
		{"TOC-INT-004", "int-3", func(d *toc.Document) {
			d.Outcomes = append(d.Outcomes, toc.Outcome{ID: "oc-dead", Title: "A dead end", Level: toc.LevelShortTerm})
			d.Interventions = append(d.Interventions, toc.Intervention{ID: "int-3", Title: "Do it", Targets: []string{"oc-dead"}})
		}, func(d *toc.Document) {
			d.Links = append(d.Links, toc.CausalLink{ID: "ln-dead-a", From: "oc-dead", To: "oc-a", Rationale: "Because it helps.", AssumptionIDs: []string{"asm-1"}})
		}},
		{"TOC-INT-005", "ln-s1-a", func(d *toc.Document) {
			for i := range d.Links {
				if d.Links[i].ID == "ln-s1-a" {
					d.Links[i].InterventionIDs = []string{"int-2"} // int-2 targets oc-s2, neither end of this link
				}
			}
		}, func(d *toc.Document) {
			for i := range d.Links {
				if d.Links[i].ID == "ln-s1-a" {
					d.Links[i].InterventionIDs = []string{"int-1"}
				}
			}
		}},
		{"TOC-INT-006", "int-1", func(d *toc.Document) {},
			func(d *toc.Document) { addIndicator(d, "ind-did", "int-1") }},
		{"TOC-INT-007", "int-1", func(d *toc.Document) {},
			func(d *toc.Document) { d.Interventions[0].Resources = "Two staff and a room" }},

		// ---- the numbers each audience is told
		{"TOC-AUD-002", "aud-1", func(d *toc.Document) { outcome(d, "oc-impact").AudienceIDs = []string{"aud-1"} },
			func(d *toc.Document) { addIndicator(d, "ind-n", "oc-s1", forAudience) }},
		{"TOC-HLN-001", "aud-1", func(d *toc.Document) {
			var ids []string
			for i := 0; i < maxHeadline+1; i++ {
				id := fmt.Sprintf("ind-h%d", i)
				addIndicator(d, id, "oc-s1", forAudience)
				ids = append(ids, id)
			}
			headline(d, ids...)
		}, func(d *toc.Document) { d.Audiences[0].Headline = d.Audiences[0].Headline[:maxHeadline] }},
		{"TOC-HLN-002", "aud-1", func(d *toc.Document) {
			addIndicator(d, "ind-h", "oc-s1")
			headline(d, "ind-h")
		}, func(d *toc.Document) { forAudience(indicator(d, "ind-h")) }},
		{"TOC-HLN-003", "aud-1", func(d *toc.Document) {
			addIndicator(d, "ind-h", "oc-s1", forAudience)
			headline(d, "ind-h")
		}, func(d *toc.Document) { indicator(d, "ind-h").DataAvailability = toc.DataInternalExisting }},
		{"TOC-HLN-004", "aud-1", func(d *toc.Document) {
			for i := 0; i < curateAbove+1; i++ {
				addIndicator(d, fmt.Sprintf("ind-r%d", i), "oc-s1", forAudience)
			}
		}, func(d *toc.Document) { headline(d, "ind-r0", "ind-r1") }},

		// ---- the North Star
		{"TOC-NSM-001", "pw-1", func(d *toc.Document) {
			d.Outcomes = append(d.Outcomes, toc.Outcome{ID: "oc-out", Title: "Mixers held", Level: toc.LevelOutput})
			d.Pathways[0].OutcomeIDs = append(d.Pathways[0].OutcomeIDs, "oc-out")
			addIndicator(d, "ind-ns", "oc-out")
			focus(d, toc.Focus{NorthStarID: "ind-ns"})
		}, func(d *toc.Document) {
			addIndicator(d, "ind-ns2", "oc-s1")
			d.Pathways[0].Focus.NorthStarID = "ind-ns2"
		}},
		{"TOC-NSM-001", "pw-1", func(d *toc.Document) {
			addIndicator(d, "ind-ns", "int-1")
			focus(d, toc.Focus{NorthStarID: "ind-ns"})
		}, func(d *toc.Document) {
			addIndicator(d, "ind-ns2", "oc-s1")
			d.Pathways[0].Focus.NorthStarID = "ind-ns2"
		}},
		{"TOC-NSM-002", "pw-1", func(d *toc.Document) {
			d.Ceiling = &toc.Ceiling{Level: toc.LevelMediumTerm}
			addIndicator(d, "ind-ns", "oc-impact")
			d.Pathways[0].OutcomeIDs = append(d.Pathways[0].OutcomeIDs, "oc-impact")
			focus(d, toc.Focus{NorthStarID: "ind-ns"})
		}, func(d *toc.Document) {
			addIndicator(d, "ind-ns2", "oc-a")
			d.Pathways[0].Focus.NorthStarID = "ind-ns2"
		}},
		{"TOC-NSM-003", "pw-1", func(d *toc.Document) {
			addIndicator(d, "ind-ns", "oc-s1")
			focus(d, toc.Focus{NorthStarID: "ind-ns"})
		}, func(d *toc.Document) {
			ns := indicator(d, "ind-ns")
			ns.Baseline = &toc.Baseline{Value: ptr(10.0), Date: "2026-01-01"}
			ns.Targets = []toc.Target{{Value: ptr(20.0), Date: "2027-01-01"}}
		}},
		{"TOC-NSM-004", "pw-1", func(d *toc.Document) {
			addIndicator(d, "ind-ns", "oc-s1")
			focus(d, toc.Focus{NorthStarID: "ind-ns"})
		}, func(d *toc.Document) {
			addIndicator(d, "ind-g", "oc-a")
			d.Pathways[0].Focus.GuardrailIDs = []string{"ind-g"}
		}},
		{"TOC-NSM-005", "pw-1", func(d *toc.Document) {
			addIndicator(d, "ind-ns", "oc-s2") // oc-s2 is not listed by pw-1
			focus(d, toc.Focus{NorthStarID: "ind-ns"})
		}, func(d *toc.Document) { d.Pathways[0].OutcomeIDs = append(d.Pathways[0].OutcomeIDs, "oc-s2") }},
		{"TOC-NSM-006", "pw-1", func(d *toc.Document) {
			addIndicator(d, "ind-ns", "oc-s1")
			focus(d, toc.Focus{NorthStarID: "ind-ns", GuardrailIDs: []string{"ind-ns"}})
		}, func(d *toc.Document) {
			addIndicator(d, "ind-g", "oc-a")
			d.Pathways[0].Focus.GuardrailIDs = []string{"ind-g"}
		}},
		{"TOC-NSM-007", "pw-1", func(d *toc.Document) {},
			func(d *toc.Document) {
				addIndicator(d, "ind-ns", "oc-s1")
				focus(d, toc.Focus{NorthStarID: "ind-ns"})
			}},

		// ---- framing
		{"TOC-CTX-001", "", func(d *toc.Document) {},
			func(d *toc.Document) { d.Context.Mission = "Why we exist." }},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%s/%s/%d", tt.rule, tt.element, i), func(t *testing.T) {
			d := chainDoc()
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

// A document that uses none of the new fields is not asked for them: the
// North Star, guardrail, and headline rules are about things it does not have.
func TestNewRulesAreQuietWhenTheFieldsAreAbsent(t *testing.T) {
	d := clean()
	for _, rule := range []string{"TOC-HLN-001", "TOC-HLN-002", "TOC-HLN-003", "TOC-HLN-004",
		"TOC-NSM-001", "TOC-NSM-002", "TOC-NSM-003", "TOC-NSM-004", "TOC-NSM-005", "TOC-NSM-006", "TOC-NSM-007", "TOC-AUD-002"} {
		if got := fires(t, d, rule); len(got) != 0 {
			t.Errorf("%s reported %v on a document with no audiences, pathways, or focus", rule, got)
		}
	}
}
