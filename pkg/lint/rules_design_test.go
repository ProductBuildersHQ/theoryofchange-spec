package lint

import (
	"slices"
	"testing"

	core "github.com/grokify/prism-core"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/graph"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

// addIndicator appends an indicator to the document and lets the caller
// adjust it.
func addIndicator(d *toc.Document, id, measures string, adjust ...func(*toc.Indicator)) {
	ind := toc.Indicator{ID: id, Measures: measures, Name: id}
	for _, f := range adjust {
		f(&ind)
	}
	d.Indicators = append(d.Indicators, ind)
}

func selected(i *toc.Indicator) { i.Selection = &toc.Selection{Status: toc.SelectionSelected} }

func outcome(d *toc.Document, id string) *toc.Outcome {
	for i := range d.Outcomes {
		if d.Outcomes[i].ID == id {
			return &d.Outcomes[i]
		}
	}
	panic("no outcome " + id)
}

func indicator(d *toc.Document, id string) *toc.Indicator {
	for i := range d.Indicators {
		if d.Indicators[i].ID == id {
			return &d.Indicators[i]
		}
	}
	panic("no indicator " + id)
}

// fires returns the elements a rule reports on the document.
func fires(t *testing.T, d *toc.Document, rule string) []string {
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
	horizon := func(d *toc.Document) {
		d.Context.TimeHorizon = &toc.TimeHorizon{Start: "2026-01-01", End: "2027-12-31"}
	}
	pathway := func(d *toc.Document, outcomes ...string) {
		d.Pathways = append(d.Pathways, toc.Pathway{ID: "pw-1", Title: "Pathway", OutcomeIDs: outcomes})
	}
	tests := []struct {
		rule, element string
		defect, fix   func(d *toc.Document)
	}{
		{"TOC-MEA-001", "oc-s1", func(d *toc.Document) { d.Ceiling = &toc.Ceiling{Level: toc.LevelLongTerm} },
			func(d *toc.Document) { addIndicator(d, "ind-1", "oc-s1") }},
		{"TOC-MEA-001", "oc-s2", func(d *toc.Document) {
			d.Ceiling = &toc.Ceiling{Level: toc.LevelLongTerm}
			addIndicator(d, "ind-1", "oc-s1")
		},
			func(d *toc.Document) {
				addIndicator(d, "ind-2", "oc-b", func(i *toc.Indicator) { i.ProxyFor = []string{"oc-s2"} })
			}},
		{"TOC-MEA-002", "ind-x", func(d *toc.Document) { addIndicator(d, "ind-x", "oc-s1") },
			func(d *toc.Document) {
				i := indicator(d, "ind-x")
				i.Baseline = &toc.Baseline{Level: "low"}
				i.Targets = []toc.Target{{Date: "2027-01-01", Level: "high"}}
				i.DataSource, i.Frequency = "Sign-in sheets", "P3M"
			}},
		{"TOC-MEA-003", "ind-x", func(d *toc.Document) { addIndicator(d, "ind-x", "oc-s1") },
			func(d *toc.Document) {
				indicator(d, "ind-x").Unit, indicator(d, "ind-x").Direction = "people", toc.DirectionIncrease
			}},
		{"TOC-MEA-004", "ind-x", func(d *toc.Document) {
			horizon(d)
			addIndicator(d, "ind-x", "oc-s1", func(i *toc.Indicator) { i.Targets = []toc.Target{{Date: "2030-01-01", Level: "x"}} })
		}, func(d *toc.Document) { indicator(d, "ind-x").Targets[0].Date = "2027-06-30" }},
		{"TOC-MEA-005", "oc-a", func(d *toc.Document) {
			addIndicator(d, "ind-s", "oc-s1", func(i *toc.Indicator) { i.Targets = []toc.Target{{Date: "2027-12-31", Level: "x"}} })
			addIndicator(d, "ind-a", "oc-a", func(i *toc.Indicator) { i.Targets = []toc.Target{{Date: "2026-12-31", Level: "x"}} })
		}, func(d *toc.Document) { indicator(d, "ind-a").Targets[0].Date = "2028-12-31" }},
		{"TOC-ACC-001", "", func(d *toc.Document) {},
			func(d *toc.Document) { d.Ceiling = &toc.Ceiling{Level: toc.LevelLongTerm} }},
		{"TOC-ACC-002", "oc-impact", func(d *toc.Document) {
			d.Ceiling = &toc.Ceiling{Level: toc.LevelLongTerm}
			outcome(d, "oc-impact").Responsibility = toc.ResponsibilityOwn
		}, func(d *toc.Document) { outcome(d, "oc-impact").Responsibility = toc.ResponsibilityShared }},
		{"TOC-CLM-001", "ind-x", func(d *toc.Document) {
			d.Ceiling = &toc.Ceiling{Level: toc.LevelLongTerm}
			addIndicator(d, "ind-x", "oc-impact", func(i *toc.Indicator) { i.Claim = toc.ClaimAttribution })
		}, func(d *toc.Document) { indicator(d, "ind-x").Claim = toc.ClaimContribution }},
		{"TOC-BUD-001", "", func(d *toc.Document) {
			d.Budget = &toc.MeasurementBudget{StaffMinutesPerWeek: 10}
			addIndicator(d, "ind-x", "oc-s1", selected, func(i *toc.Indicator) {
				i.Collection = &toc.Collection{Mode: toc.ModeManual, MinutesPerPeriod: 100, Period: "P1W"}
			})
		}, func(d *toc.Document) { d.Budget.StaffMinutesPerWeek = 200 }},
		{"TOC-BUD-002", "", func(d *toc.Document) {
			d.Budget = &toc.MeasurementBudget{MaxQuestionsPerRespondentPerYear: 5}
			addIndicator(d, "ind-x", "oc-s1", selected, func(i *toc.Indicator) {
				i.Collection = &toc.Collection{Mode: toc.ModeSurvey, Period: "P1M", QuestionsPerRespondent: 4}
			})
		}, func(d *toc.Document) { d.Budget.MaxQuestionsPerRespondentPerYear = 100 }},
		{"TOC-BUD-003", "", func(d *toc.Document) {
			d.Budget = &toc.MeasurementBudget{TargetSelectedMetrics: &toc.Range{Min: 3, Max: 5}}
			addIndicator(d, "ind-x", "oc-s1", selected)
		}, func(d *toc.Document) { d.Budget.TargetSelectedMetrics = &toc.Range{Min: 1, Max: 2} }},
		{"TOC-BUD-004", "ind-x", func(d *toc.Document) { addIndicator(d, "ind-x", "oc-s1", selected) },
			func(d *toc.Document) { indicator(d, "ind-x").Collection = &toc.Collection{Mode: toc.ModePassive} }},
		{"TOC-ASM-001", "asm-c", func(d *toc.Document) {
			d.Assumptions = append(d.Assumptions, toc.Assumption{ID: "asm-c", Statement: "S", Criticality: toc.CriticalityCritical})
		}, func(d *toc.Document) { d.Assumptions[1].Test = "Ask partners twice a year." }},
		{"TOC-ASM-002", "asm-c", func(d *toc.Document) {
			d.Assumptions = append(d.Assumptions, toc.Assumption{ID: "asm-c", Statement: "S", Criticality: toc.CriticalityHigh})
		}, func(d *toc.Document) {
			d.Evidence = append(d.Evidence, toc.Evidence{ID: "ev-1", Title: "Study"})
			d.Assumptions[1].EvidenceIDs = []string{"ev-1"}
		}},
		{"TOC-ASM-003", "", func(d *toc.Document) {},
			func(d *toc.Document) { d.Assumptions[0].Type = toc.AssumptionContextual }},
		{"TOC-ASM-004", "asm-c", func(d *toc.Document) {
			d.Assumptions = append(d.Assumptions, toc.Assumption{ID: "asm-c", Statement: "S", NeedsTesting: true})
		}, func(d *toc.Document) {
			d.Assumptions[1].Test = "Pilot it."
			d.Decisions = append(d.Decisions, toc.Decision{ID: "dec-1", Question: "Q", RelatedIDs: []string{"asm-c"}})
		}},
		{"TOC-ASM-005", "", func(d *toc.Document) {
			outcome(d, "oc-s1").Responsibility = toc.ResponsibilityPartner
		}, func(d *toc.Document) {
			d.Assumptions = append(d.Assumptions, toc.Assumption{ID: "asm-p", Statement: "Partners show up.", Type: toc.AssumptionPartnership})
		}},
		{"TOC-RSK-001", "asm-c", func(d *toc.Document) {
			d.Assumptions = append(d.Assumptions, toc.Assumption{ID: "asm-c", Statement: "S", Criticality: toc.CriticalityCritical})
		}, func(d *toc.Document) {
			d.Risks = append(d.Risks, toc.Risk{Risk: core.Risk{ID: "rsk-1", Description: "D", Mitigation: "Plan B"}, RelatedAssumptionIDs: []string{"asm-c"}})
		}},
		{"TOC-INT-002", "int-1", func(d *toc.Document) {},
			func(d *toc.Document) {
				d.Actors = append(d.Actors, toc.Actor{ID: "act-1", Name: "Network"})
				d.Interventions[0].OwnerID = "act-1"
			}},
		{"TOC-ACT-001", "oc-s1", func(d *toc.Document) { outcome(d, "oc-s1").Sphere = toc.SphereInfluence },
			func(d *toc.Document) {
				d.Actors = append(d.Actors, toc.Actor{ID: "act-1", Name: "Partner"})
				outcome(d, "oc-s1").ActorIDs = []string{"act-1"}
			}},
		{"TOC-EVD-001", "ln-a-i", func(d *toc.Document) {},
			func(d *toc.Document) {
				d.Evidence = append(d.Evidence, toc.Evidence{ID: "ev-1", Title: "Study"})
				d.Links[0].EvidenceIDs = []string{"ev-1"}
			}},
		{"TOC-PTH-001", "pw-1", func(d *toc.Document) {
			d.Outcomes = append(d.Outcomes, toc.Outcome{ID: "oc-iso", Title: "Isolated", Level: toc.LevelShortTerm})
			pathway(d, "oc-iso")
		}, func(d *toc.Document) { d.Pathways[0].OutcomeIDs = []string{"oc-s1"} }},
		{"TOC-PTH-002", "pw-1", func(d *toc.Document) { pathway(d, "oc-s1") },
			func(d *toc.Document) { addIndicator(d, "ind-x", "oc-s1", selected) }},
		{"TOC-PTH-003", "ind-x", func(d *toc.Document) {
			pathway(d, "oc-s1")
			d.Pathways = append(d.Pathways, toc.Pathway{ID: "pw-2", Title: "Other"})
			addIndicator(d, "ind-x", "oc-s1", func(i *toc.Indicator) { i.PathwayIDs = []string{"pw-2"} })
		}, func(d *toc.Document) { indicator(d, "ind-x").PathwayIDs = []string{"pw-1"} }},
		{"TOC-MET-001", "met-1", func(d *toc.Document) {
			pathway(d, "oc-s1")
			d.Metrics = append(d.Metrics, toc.Metric{ID: "met-1", Name: "M", Definition: "D", Formula: "F", Layer: toc.LayerCommon})
		}, func(d *toc.Document) {
			addIndicator(d, "ind-x", "oc-s1", func(i *toc.Indicator) { i.MetricID, i.PathwayIDs = "met-1", []string{"pw-1"} })
		}},
		{"TOC-MET-003", "met-w", func(d *toc.Document) {
			pathway(d, "oc-s1")
			d.Metrics = append(d.Metrics,
				toc.Metric{ID: "met-wt", Name: "Weight", Definition: "D", Formula: "F"},
				toc.Metric{ID: "met-w", Name: "W", Definition: "D", Formula: "F", Aggregation: toc.AggregationWeightedMean, WeightMetricID: "met-wt"})
			addIndicator(d, "ind-w", "oc-s1", func(i *toc.Indicator) { i.MetricID, i.PathwayIDs = "met-w", []string{"pw-1"} })
		}, func(d *toc.Document) {
			addIndicator(d, "ind-wt", "oc-s2", func(i *toc.Indicator) { i.MetricID, i.PathwayIDs = "met-wt", []string{"pw-1"} })
		}},
		{"TOC-AUD-001", "aud-1", func(d *toc.Document) { d.Audiences = append(d.Audiences, toc.Audience{ID: "aud-1", Name: "Board"}) },
			func(d *toc.Document) { outcome(d, "oc-impact").AudienceIDs = []string{"aud-1"} }},
		{"TOC-CAU-001", "cause-1", func(d *toc.Document) {
			d.Causes = append(d.Causes, toc.Cause{ID: "cause-1", Statement: "S", Level: toc.CauseStructural})
		}, func(d *toc.Document) { outcome(d, "oc-a").AddressesCauses = []string{"cause-1"} }},
		{"TOC-CAU-002", "", func(d *toc.Document) {
			d.Causes = append(d.Causes, toc.Cause{ID: "cause-i", Statement: "S", Level: toc.CauseImmediate})
			outcome(d, "oc-a").AddressesCauses = []string{"cause-i"}
		}, func(d *toc.Document) {
			d.Causes = append(d.Causes, toc.Cause{ID: "cause-u", Statement: "S", Level: toc.CauseUnderlying})
			outcome(d, "oc-b").AddressesCauses = []string{"cause-u"}
		}},
		{"TOC-RSP-001", "oc-s1", func(d *toc.Document) { outcome(d, "oc-s1").Responsibility = toc.ResponsibilityPartner },
			func(d *toc.Document) {
				d.Actors = append(d.Actors, toc.Actor{ID: "act-1", Name: "Partner"})
				outcome(d, "oc-s1").DeliveredBy = []string{"act-1"}
			}},
		{"TOC-RSP-002", "oc-s1", func(d *toc.Document) { outcome(d, "oc-s1").Responsibility = toc.ResponsibilityPartner },
			func(d *toc.Document) { addIndicator(d, "ind-x", "oc-s1") }},
		{"TOC-EQU-001", "oc-s1", func(d *toc.Document) {
			d.TargetGroups = append(d.TargetGroups, toc.TargetGroup{ID: "tg-1", Name: "Small members"})
			outcome(d, "oc-s1").TargetGroupIDs = []string{"tg-1"}
			addIndicator(d, "ind-x", "oc-s1")
		}, func(d *toc.Document) { indicator(d, "ind-x").TargetGroupIDs = []string{"tg-1"} }},
		{"TOC-VAL-001", "", func(d *toc.Document) {},
			func(d *toc.Document) {
				d.Consultations = append(d.Consultations, toc.Consultation{ID: "con-1", Date: "2026-01-01", Step: toc.StepWhole})
			}},
		{"TOC-PRV-001", "oc-impact", func(d *toc.Document) {},
			func(d *toc.Document) { outcome(d, "oc-impact").Provenance = &toc.Provenance{Basis: toc.BasisInferred} }},
		{"TOC-PRV-002", "oc-s1", func(d *toc.Document) {
			outcome(d, "oc-s1").Provenance = &toc.Provenance{Basis: toc.BasisStated}
		}, func(d *toc.Document) {
			d.Evidence = append(d.Evidence, toc.Evidence{ID: "ev-1", Title: "Report", URL: "https://example.org/report"})
			outcome(d, "oc-s1").Provenance.EvidenceIDs = []string{"ev-1"}
		}},
		{"TOC-PRV-003", "oc-s1", func(d *toc.Document) {
			d.Metadata.Status = "approved"
			outcome(d, "oc-s1").Provenance = &toc.Provenance{Basis: toc.BasisHypothesis}
		}, func(d *toc.Document) { outcome(d, "oc-s1").Provenance.Basis = toc.BasisConfirmed }},
		{"TOC-PRV-004", "ind-x", func(d *toc.Document) { addIndicator(d, "ind-x", "oc-s1", selected) },
			func(d *toc.Document) { indicator(d, "ind-x").DataAvailability = toc.DataPublic }},
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
	byValidation := []string{"TOC-INT-001", "TOC-MET-002"}
	want := []string{
		"TOC-STR-001", "TOC-STR-002", "TOC-STR-003", "TOC-STR-004", "TOC-STR-005", "TOC-STR-006", "TOC-STR-007",
		"TOC-LOG-001", "TOC-LOG-002", "TOC-LOG-003", "TOC-LOG-004",
		"TOC-ASM-001", "TOC-ASM-002", "TOC-ASM-003", "TOC-ASM-004", "TOC-ASM-005", "TOC-RSK-001",
		"TOC-MEA-001", "TOC-MEA-002", "TOC-MEA-003", "TOC-MEA-004", "TOC-MEA-005",
		"TOC-ACC-001", "TOC-ACC-002", "TOC-CLM-001", "TOC-INT-002", "TOC-ACT-001", "TOC-EVD-001",
		"TOC-PTH-001", "TOC-PTH-002", "TOC-PTH-003", "TOC-MET-001", "TOC-MET-003",
		"TOC-BUD-001", "TOC-BUD-002", "TOC-BUD-003", "TOC-BUD-004", "TOC-AUD-001",
		"TOC-CAU-001", "TOC-CAU-002", "TOC-RSP-001", "TOC-RSP-002", "TOC-EQU-001", "TOC-VAL-001",
		"TOC-PRV-001", "TOC-PRV-002", "TOC-PRV-003", "TOC-PRV-004",
		"TOC-INT-003", "TOC-INT-004", "TOC-INT-005", "TOC-INT-006", "TOC-INT-007",
		"TOC-AUD-002", "TOC-HLN-001", "TOC-HLN-002", "TOC-HLN-003", "TOC-HLN-004",
		"TOC-NSM-001", "TOC-NSM-002", "TOC-NSM-003", "TOC-NSM-004", "TOC-NSM-005", "TOC-NSM-006", "TOC-NSM-007",
		"TOC-CTX-001",
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
	r, ok := Default.Rule("TOC-PRV-001")
	if !ok {
		t.Fatal("TOC-PRV-001 is not registered")
	}
	if Runs(r, ProfileStandard) || !Runs(r, ProfileFull) {
		t.Error("TOC-PRV-001 should run under full only")
	}
}
