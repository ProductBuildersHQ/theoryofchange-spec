package lint

import (
	"slices"
	"testing"

	core "github.com/grokify/prism-core"

	"github.com/ProductBuildersHQ/impact-model/pkg/graph"
	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
)

// addIndicator appends an indicator to the document and lets the caller
// adjust it.
func addIndicator(d *impact.Model, id, measures string, adjust ...func(*impact.Indicator)) {
	ind := impact.Indicator{ID: id, Measures: measures, Name: id}
	for _, f := range adjust {
		f(&ind)
	}
	d.Indicators = append(d.Indicators, ind)
}

func selected(i *impact.Indicator) { i.Selection = &impact.Selection{Status: impact.SelectionSelected} }

func outcome(d *impact.Model, id string) *impact.Outcome {
	for i := range d.Outcomes {
		if d.Outcomes[i].ID == id {
			return &d.Outcomes[i]
		}
	}
	panic("no outcome " + id)
}

func indicator(d *impact.Model, id string) *impact.Indicator {
	for i := range d.Indicators {
		if d.Indicators[i].ID == id {
			return &d.Indicators[i]
		}
	}
	panic("no indicator " + id)
}

// fires returns the elements a rule reports on the document.
func fires(t *testing.T, d *impact.Model, rule string) []string {
	t.Helper()
	g, err := graph.Build(d)
	if err != nil {
		t.Fatalf("document does not build: %v", err)
	}
	res, err := Run(d, g, Options{Profile: ProfileFull})
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, f := range res.Findings {
		if f.Rule == rule {
			out = append(out, f.Element)
		}
	}
	return out
}

// Each case seeds one defect and checks the rule finds it, then repairs it
// and checks the rule falls silent about it. An empty element means a
// finding about the whole document.
func TestRulesFindSeededDefectsAndAcceptRepairs(t *testing.T) {
	horizon := func(d *impact.Model) {
		d.Context.TimeHorizon = &impact.TimeHorizon{Start: "2026-01-01", End: "2027-12-31"}
	}
	pathway := func(d *impact.Model, outcomes ...string) {
		d.Pathways = append(d.Pathways, impact.Pathway{ID: "pw-1", Title: "Pathway", OutcomeIDs: outcomes})
	}
	tests := []struct {
		rule, element string
		defect, fix   func(d *impact.Model)
	}{
		{"IMP-MEA-001", "oc-s1", func(d *impact.Model) { d.Ceiling = &impact.Ceiling{Level: impact.LevelLongTerm} },
			func(d *impact.Model) { addIndicator(d, "ind-1", "oc-s1") }},
		{"IMP-MEA-001", "oc-s2", func(d *impact.Model) {
			d.Ceiling = &impact.Ceiling{Level: impact.LevelLongTerm}
			addIndicator(d, "ind-1", "oc-s1")
		},
			func(d *impact.Model) {
				addIndicator(d, "ind-2", "oc-b", func(i *impact.Indicator) { i.ProxyFor = []string{"oc-s2"} })
			}},
		{"IMP-MEA-002", "ind-x", func(d *impact.Model) { addIndicator(d, "ind-x", "oc-s1") },
			func(d *impact.Model) {
				i := indicator(d, "ind-x")
				i.Baseline = &impact.Baseline{Level: "low"}
				i.Targets = []impact.Target{{Date: "2027-01-01", Level: "high"}}
				i.DataSource, i.Frequency = "Sign-in sheets", "P3M"
			}},
		{"IMP-MEA-003", "ind-x", func(d *impact.Model) { addIndicator(d, "ind-x", "oc-s1") },
			func(d *impact.Model) {
				indicator(d, "ind-x").Unit, indicator(d, "ind-x").Direction = "people", impact.DirectionIncrease
			}},
		{"IMP-MEA-004", "ind-x", func(d *impact.Model) {
			horizon(d)
			addIndicator(d, "ind-x", "oc-s1", func(i *impact.Indicator) { i.Targets = []impact.Target{{Date: "2030-01-01", Level: "x"}} })
		}, func(d *impact.Model) { indicator(d, "ind-x").Targets[0].Date = "2027-06-30" }},
		{"IMP-MEA-005", "oc-a", func(d *impact.Model) {
			addIndicator(d, "ind-s", "oc-s1", func(i *impact.Indicator) { i.Targets = []impact.Target{{Date: "2027-12-31", Level: "x"}} })
			addIndicator(d, "ind-a", "oc-a", func(i *impact.Indicator) { i.Targets = []impact.Target{{Date: "2026-12-31", Level: "x"}} })
		}, func(d *impact.Model) { indicator(d, "ind-a").Targets[0].Date = "2028-12-31" }},
		{"IMP-ACC-001", "", func(d *impact.Model) {},
			func(d *impact.Model) { d.Ceiling = &impact.Ceiling{Level: impact.LevelLongTerm} }},
		{"IMP-ACC-002", "oc-impact", func(d *impact.Model) {
			d.Ceiling = &impact.Ceiling{Level: impact.LevelLongTerm}
			outcome(d, "oc-impact").Responsibility = impact.ResponsibilityOwn
		}, func(d *impact.Model) { outcome(d, "oc-impact").Responsibility = impact.ResponsibilityShared }},
		{"IMP-CLM-001", "ind-x", func(d *impact.Model) {
			d.Ceiling = &impact.Ceiling{Level: impact.LevelLongTerm}
			addIndicator(d, "ind-x", "oc-impact", func(i *impact.Indicator) { i.Claim = impact.ClaimAttribution })
		}, func(d *impact.Model) { indicator(d, "ind-x").Claim = impact.ClaimContribution }},
		{"IMP-BUD-001", "", func(d *impact.Model) {
			d.Budget = &impact.MeasurementBudget{StaffMinutesPerWeek: 10}
			addIndicator(d, "ind-x", "oc-s1", selected, func(i *impact.Indicator) {
				i.Collection = &impact.Collection{Mode: impact.ModeManual, MinutesPerPeriod: 100, Period: "P1W"}
			})
		}, func(d *impact.Model) { d.Budget.StaffMinutesPerWeek = 200 }},
		{"IMP-BUD-002", "", func(d *impact.Model) {
			d.Budget = &impact.MeasurementBudget{MaxQuestionsPerRespondentPerYear: 5}
			addIndicator(d, "ind-x", "oc-s1", selected, func(i *impact.Indicator) {
				i.Collection = &impact.Collection{Mode: impact.ModeSurvey, Period: "P1M", QuestionsPerRespondent: 4}
			})
		}, func(d *impact.Model) { d.Budget.MaxQuestionsPerRespondentPerYear = 100 }},
		{"IMP-BUD-003", "", func(d *impact.Model) {
			d.Budget = &impact.MeasurementBudget{TargetSelectedMetrics: &impact.Range{Min: 3, Max: 5}}
			addIndicator(d, "ind-x", "oc-s1", selected)
		}, func(d *impact.Model) { d.Budget.TargetSelectedMetrics = &impact.Range{Min: 1, Max: 2} }},
		{"IMP-BUD-004", "ind-x", func(d *impact.Model) { addIndicator(d, "ind-x", "oc-s1", selected) },
			func(d *impact.Model) {
				indicator(d, "ind-x").Collection = &impact.Collection{Mode: impact.ModePassive}
			}},
		{"IMP-ASM-001", "asm-c", func(d *impact.Model) {
			d.Assumptions = append(d.Assumptions, impact.Assumption{ID: "asm-c", Statement: "S", Criticality: impact.CriticalityCritical})
		}, func(d *impact.Model) { d.Assumptions[1].Test = "Ask partners twice a year." }},
		{"IMP-ASM-002", "asm-c", func(d *impact.Model) {
			d.Assumptions = append(d.Assumptions, impact.Assumption{ID: "asm-c", Statement: "S", Criticality: impact.CriticalityHigh})
		}, func(d *impact.Model) {
			d.Evidence = append(d.Evidence, impact.Evidence{ID: "ev-1", Title: "Study"})
			d.Assumptions[1].EvidenceIDs = []string{"ev-1"}
		}},
		{"IMP-ASM-003", "", func(d *impact.Model) {},
			func(d *impact.Model) { d.Assumptions[0].Type = impact.AssumptionContextual }},
		{"IMP-ASM-004", "asm-c", func(d *impact.Model) {
			d.Assumptions = append(d.Assumptions, impact.Assumption{ID: "asm-c", Statement: "S", NeedsTesting: true})
		}, func(d *impact.Model) {
			d.Assumptions[1].Test = "Pilot it."
			d.Decisions = append(d.Decisions, impact.Decision{ID: "dec-1", Question: "Q", RelatedIDs: []string{"asm-c"}})
		}},
		{"IMP-ASM-005", "", func(d *impact.Model) {
			outcome(d, "oc-s1").Responsibility = impact.ResponsibilityPartner
		}, func(d *impact.Model) {
			d.Assumptions = append(d.Assumptions, impact.Assumption{ID: "asm-p", Statement: "Partners show up.", Type: impact.AssumptionPartnership})
		}},
		{"IMP-RSK-001", "asm-c", func(d *impact.Model) {
			d.Assumptions = append(d.Assumptions, impact.Assumption{ID: "asm-c", Statement: "S", Criticality: impact.CriticalityCritical})
		}, func(d *impact.Model) {
			d.Risks = append(d.Risks, impact.Risk{Risk: core.Risk{ID: "rsk-1", Description: "D", Mitigation: "Plan B"}, RelatedAssumptionIDs: []string{"asm-c"}})
		}},
		{"IMP-INT-002", "int-1", func(d *impact.Model) {},
			func(d *impact.Model) {
				d.Actors = append(d.Actors, impact.Actor{ID: "act-1", Name: "Network"})
				d.Interventions[0].OwnerID = "act-1"
			}},
		{"IMP-ACT-001", "oc-s1", func(d *impact.Model) { outcome(d, "oc-s1").Sphere = impact.SphereInfluence },
			func(d *impact.Model) {
				d.Actors = append(d.Actors, impact.Actor{ID: "act-1", Name: "Partner"})
				outcome(d, "oc-s1").ActorIDs = []string{"act-1"}
			}},
		{"IMP-EVD-001", "ln-a-i", func(d *impact.Model) {},
			func(d *impact.Model) {
				d.Evidence = append(d.Evidence, impact.Evidence{ID: "ev-1", Title: "Study"})
				d.Links[0].EvidenceIDs = []string{"ev-1"}
			}},
		{"IMP-PTH-001", "pw-1", func(d *impact.Model) {
			d.Outcomes = append(d.Outcomes, impact.Outcome{ID: "oc-iso", Title: "Isolated", Level: impact.LevelShortTerm})
			pathway(d, "oc-iso")
		}, func(d *impact.Model) { d.Pathways[0].OutcomeIDs = []string{"oc-s1"} }},
		{"IMP-PTH-002", "pw-1", func(d *impact.Model) { pathway(d, "oc-s1") },
			func(d *impact.Model) { addIndicator(d, "ind-x", "oc-s1", selected) }},
		{"IMP-PTH-003", "ind-x", func(d *impact.Model) {
			pathway(d, "oc-s1")
			d.Pathways = append(d.Pathways, impact.Pathway{ID: "pw-2", Title: "Other"})
			addIndicator(d, "ind-x", "oc-s1", func(i *impact.Indicator) { i.PathwayIDs = []string{"pw-2"} })
		}, func(d *impact.Model) { indicator(d, "ind-x").PathwayIDs = []string{"pw-1"} }},
		{"IMP-MET-001", "met-1", func(d *impact.Model) {
			pathway(d, "oc-s1")
			d.Metrics = append(d.Metrics, impact.Metric{ID: "met-1", Name: "M", Definition: "D", Formula: "F", Layer: impact.LayerCommon})
		}, func(d *impact.Model) {
			addIndicator(d, "ind-x", "oc-s1", func(i *impact.Indicator) { i.MetricID, i.PathwayIDs = "met-1", []string{"pw-1"} })
		}},
		{"IMP-MET-003", "met-w", func(d *impact.Model) {
			pathway(d, "oc-s1")
			d.Metrics = append(d.Metrics,
				impact.Metric{ID: "met-wt", Name: "Weight", Definition: "D", Formula: "F"},
				impact.Metric{ID: "met-w", Name: "W", Definition: "D", Formula: "F", Aggregation: impact.AggregationWeightedMean, WeightMetricID: "met-wt"})
			addIndicator(d, "ind-w", "oc-s1", func(i *impact.Indicator) { i.MetricID, i.PathwayIDs = "met-w", []string{"pw-1"} })
		}, func(d *impact.Model) {
			addIndicator(d, "ind-wt", "oc-s2", func(i *impact.Indicator) { i.MetricID, i.PathwayIDs = "met-wt", []string{"pw-1"} })
		}},
		{"IMP-AUD-001", "aud-1", func(d *impact.Model) {
			d.Audiences = append(d.Audiences, impact.Audience{ID: "aud-1", Name: "Board"})
		},
			func(d *impact.Model) { outcome(d, "oc-impact").AudienceIDs = []string{"aud-1"} }},
		{"IMP-CAU-001", "cause-1", func(d *impact.Model) {
			d.Causes = append(d.Causes, impact.Cause{ID: "cause-1", Statement: "S", Level: impact.CauseStructural})
		}, func(d *impact.Model) { outcome(d, "oc-a").AddressesCauses = []string{"cause-1"} }},
		{"IMP-CAU-002", "", func(d *impact.Model) {
			d.Causes = append(d.Causes, impact.Cause{ID: "cause-i", Statement: "S", Level: impact.CauseImmediate})
			outcome(d, "oc-a").AddressesCauses = []string{"cause-i"}
		}, func(d *impact.Model) {
			d.Causes = append(d.Causes, impact.Cause{ID: "cause-u", Statement: "S", Level: impact.CauseUnderlying})
			outcome(d, "oc-b").AddressesCauses = []string{"cause-u"}
		}},
		{"IMP-RSP-001", "oc-s1", func(d *impact.Model) { outcome(d, "oc-s1").Responsibility = impact.ResponsibilityPartner },
			func(d *impact.Model) {
				d.Actors = append(d.Actors, impact.Actor{ID: "act-1", Name: "Partner"})
				outcome(d, "oc-s1").DeliveredBy = []string{"act-1"}
			}},
		{"IMP-RSP-002", "oc-s1", func(d *impact.Model) { outcome(d, "oc-s1").Responsibility = impact.ResponsibilityPartner },
			func(d *impact.Model) { addIndicator(d, "ind-x", "oc-s1") }},
		{"IMP-EQU-001", "oc-s1", func(d *impact.Model) {
			d.TargetGroups = append(d.TargetGroups, impact.TargetGroup{ID: "tg-1", Name: "Small members"})
			outcome(d, "oc-s1").TargetGroupIDs = []string{"tg-1"}
			addIndicator(d, "ind-x", "oc-s1")
		}, func(d *impact.Model) { indicator(d, "ind-x").TargetGroupIDs = []string{"tg-1"} }},
		{"IMP-VAL-001", "", func(d *impact.Model) {},
			func(d *impact.Model) {
				d.Consultations = append(d.Consultations, impact.Consultation{ID: "con-1", Date: "2026-01-01", Step: impact.StepWhole})
			}},
		{"IMP-PRV-001", "oc-impact", func(d *impact.Model) {},
			func(d *impact.Model) {
				outcome(d, "oc-impact").Provenance = &impact.Provenance{Basis: impact.BasisInferred}
			}},
		{"IMP-PRV-002", "oc-s1", func(d *impact.Model) {
			outcome(d, "oc-s1").Provenance = &impact.Provenance{Basis: impact.BasisStated}
		}, func(d *impact.Model) {
			d.Evidence = append(d.Evidence, impact.Evidence{ID: "ev-1", Title: "Report", URL: "https://example.org/report"})
			outcome(d, "oc-s1").Provenance.EvidenceIDs = []string{"ev-1"}
		}},
		{"IMP-PRV-003", "oc-s1", func(d *impact.Model) {
			d.Metadata.Status = "approved"
			outcome(d, "oc-s1").Provenance = &impact.Provenance{Basis: impact.BasisHypothesis}
		}, func(d *impact.Model) { outcome(d, "oc-s1").Provenance.Basis = impact.BasisConfirmed }},
		{"IMP-PRV-004", "ind-x", func(d *impact.Model) { addIndicator(d, "ind-x", "oc-s1", selected) },
			func(d *impact.Model) { indicator(d, "ind-x").DataAvailability = impact.DataPublic }},
	}
	for _, tt := range tests {
		t.Run(tt.rule+"/"+tt.element, func(t *testing.T) {
			d := clean()
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

func TestEveryDocumentedRuleIsRegisteredOnce(t *testing.T) {
	// Rules the TRD lists that validation already enforces as errors, so
	// the document cannot reach lint in a state that breaks them.
	byValidation := []string{"IMP-INT-001", "IMP-MET-002"}
	want := []string{
		"IMP-STR-001", "IMP-STR-002", "IMP-STR-003", "IMP-STR-004", "IMP-STR-005", "IMP-STR-006", "IMP-STR-007",
		"IMP-LOG-001", "IMP-LOG-002", "IMP-LOG-003", "IMP-LOG-004",
		"IMP-ASM-001", "IMP-ASM-002", "IMP-ASM-003", "IMP-ASM-004", "IMP-ASM-005", "IMP-RSK-001",
		"IMP-MEA-001", "IMP-MEA-002", "IMP-MEA-003", "IMP-MEA-004", "IMP-MEA-005",
		"IMP-ACC-001", "IMP-ACC-002", "IMP-CLM-001", "IMP-INT-002", "IMP-ACT-001", "IMP-EVD-001",
		"IMP-PTH-001", "IMP-PTH-002", "IMP-PTH-003", "IMP-MET-001", "IMP-MET-003",
		"IMP-BUD-001", "IMP-BUD-002", "IMP-BUD-003", "IMP-BUD-004", "IMP-AUD-001",
		"IMP-CAU-001", "IMP-CAU-002", "IMP-RSP-001", "IMP-RSP-002", "IMP-EQU-001", "IMP-VAL-001",
		"IMP-PRV-001", "IMP-PRV-002", "IMP-PRV-003", "IMP-PRV-004",
		"IMP-INT-003", "IMP-INT-004", "IMP-INT-005", "IMP-INT-006", "IMP-INT-007",
		"IMP-AUD-002", "IMP-HLN-001", "IMP-HLN-002", "IMP-HLN-003", "IMP-HLN-004",
		"IMP-NSM-001", "IMP-NSM-002", "IMP-NSM-003", "IMP-NSM-004", "IMP-NSM-005", "IMP-NSM-006", "IMP-NSM-007",
		"IMP-CTX-001",
		"IMP-DMA-001", "IMP-DMA-002", "IMP-DMA-003",
		"IMP-EFF-001", "IMP-EFF-002", "IMP-CAU-003",
	}
	slices.Sort(want)
	var got []string
	for _, r := range Default.Rules() {
		got = append(got, r.ID())
		if r.Description() == "" {
			t.Errorf("%s has no description", r.ID())
		}
	}
	if !slices.Equal(got, want) {
		t.Errorf("registered rules differ from the documented set\n got: %v\nwant: %v", got, want)
	}
	if len(byValidation) != 2 {
		t.Fatal("update the list of rules enforced by validation")
	}
}

func TestPRV001RunsOnlyInTheFullProfile(t *testing.T) {
	r, ok := Default.Rule("IMP-PRV-001")
	if !ok {
		t.Fatal("IMP-PRV-001 is not registered")
	}
	if Runs(r, ProfileStandard) || !Runs(r, ProfileFull) {
		t.Error("IMP-PRV-001 should run under full only")
	}
}
