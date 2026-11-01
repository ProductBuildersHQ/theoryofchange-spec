package lint

import (
	"slices"
	"testing"

	"github.com/ProductBuildersHQ/impact-model/pkg/graph"
	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
)

// clean returns a small impact model that breaks none of the structure or logic
// rules: one impact reached by two medium-term outcomes, each reached from a
// short-term outcome that an intervention brings about.
func clean() *impact.Model {
	d := impact.NewModel("lint-base", "People cannot reach services.")
	all := &impact.Logic{Type: impact.LogicAll}
	d.Outcomes = []impact.Outcome{
		{ID: "oc-impact", Title: "Services are reachable", Level: impact.LevelImpact, Logic: all},
		{ID: "oc-a", Title: "Providers share rides", Level: impact.LevelMediumTerm},
		{ID: "oc-b", Title: "Providers share schedules", Level: impact.LevelMediumTerm},
		{ID: "oc-s1", Title: "Providers know each other", Level: impact.LevelShortTerm},
		{ID: "oc-s2", Title: "Providers meet monthly", Level: impact.LevelShortTerm},
	}
	link := func(id, from, to string) impact.CausalLink {
		return impact.CausalLink{ID: id, From: from, To: to, Rationale: "Because it helps.", AssumptionIDs: []string{"asm-1"}}
	}
	d.Links = []impact.CausalLink{
		link("ln-a-i", "oc-a", "oc-impact"), link("ln-b-i", "oc-b", "oc-impact"),
		link("ln-s1-a", "oc-s1", "oc-a"), link("ln-s2-b", "oc-s2", "oc-b"),
	}
	d.Assumptions = []impact.Assumption{{ID: "asm-1", Statement: "Providers want to cooperate.", AppliesTo: []string{"ln-a-i"}}}
	d.Interventions = []impact.Intervention{
		{ID: "int-1", Title: "Host mixers", Targets: []string{"oc-s1"}},
		{ID: "int-2", Title: "Host meetings", Targets: []string{"oc-s2"}},
	}
	return d
}

func runRule(t *testing.T, doc *impact.Model, id string) []Finding {
	t.Helper()
	g, err := graph.Build(doc)
	if err != nil {
		t.Fatalf("seeded document does not build: %v", err)
	}
	res, err := Run(doc, g, Options{Profile: ProfileFull})
	if err != nil {
		t.Fatal(err)
	}
	var out []Finding
	for _, f := range res.Findings {
		if f.Rule == id {
			out = append(out, f)
		}
	}
	return out
}

func TestCleanBaseBreaksNoStructureOrLogicRule(t *testing.T) {
	doc := clean()
	g, err := graph.Build(doc)
	if err != nil {
		t.Fatal(err)
	}
	res, err := Run(doc, g, Options{Profile: ProfileFull})
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range res.Findings {
		if c := Category(f.Rule); c == "STR" || c == "LOG" {
			t.Errorf("the clean base trips %s: %s", f.Rule, f.Message)
		}
	}
}

// Each case seeds one defect; the rule must find it, naming the element.
func TestStructureAndLogicRulesFindSeededDefects(t *testing.T) {
	tests := []struct {
		rule    string
		element string
		seed    func(d *impact.Model)
	}{
		{"IMP-STR-001", "", func(d *impact.Model) { d.Outcomes[0].Level = impact.LevelLongTerm }},
		{"IMP-STR-002", "oc-impact", func(d *impact.Model) {
			d.Links = append(d.Links, impact.CausalLink{ID: "ln-i-a", From: "oc-impact", To: "oc-a", Rationale: "r", AssumptionIDs: []string{"asm-1"}})
		}},
		{"IMP-STR-003", "oc-lost", func(d *impact.Model) {
			d.Outcomes = append(d.Outcomes, impact.Outcome{ID: "oc-lost", Title: "Strays", Level: impact.LevelShortTerm})
			d.Interventions[0].Targets = append(d.Interventions[0].Targets, "oc-lost")
		}},
		{"IMP-STR-004", "oc-s2", func(d *impact.Model) { d.Interventions[1].Targets = []string{"oc-s1"} }},
		{"IMP-STR-005", "oc-impact", func(d *impact.Model) { d.Outcomes[0].Logic = &impact.Logic{Type: impact.LogicAtLeast, K: 3} }},
		{"IMP-STR-006", "ln-down", func(d *impact.Model) {
			d.Links = append(d.Links, impact.CausalLink{ID: "ln-down", From: "oc-a", To: "oc-s1", Rationale: "r"})
		}},
		{"IMP-STR-007", "grp-y", func(d *impact.Model) {
			d.Groups = []impact.PreconditionGroup{{ID: "grp-y", Target: "oc-impact", Members: []string{"oc-a", "oc-s1"}, Logic: impact.Logic{Type: impact.LogicAll}}}
		}},
		{"IMP-LOG-001", "ln-a-i", func(d *impact.Model) { d.Links[0].Rationale = "" }},
		{"IMP-LOG-002", "ln-s1-a", func(d *impact.Model) { d.Links[2].AssumptionIDs = nil }},
		{"IMP-LOG-003", "oc-impact", func(d *impact.Model) { d.Links = slices.Delete(d.Links, 1, 2) }},
		{"IMP-LOG-004", "oc-impact", func(d *impact.Model) { d.Outcomes[0].Logic = nil }},
	}
	for _, tt := range tests {
		t.Run(tt.rule+"/"+tt.element, func(t *testing.T) {
			doc := clean()
			tt.seed(doc)
			var elements []string
			for _, f := range runRule(t, doc, tt.rule) {
				elements = append(elements, f.Element)
			}
			if !slices.Contains(elements, tt.element) {
				t.Errorf("%s did not report %q; reported %v", tt.rule, tt.element, elements)
			}
		})
	}
}

func TestRejectedElementsAreIgnored(t *testing.T) {
	doc := clean()
	rejected := &impact.Provenance{Basis: impact.BasisRejected}
	doc.Links = append(doc.Links, impact.CausalLink{ID: "ln-dropped", From: "oc-s1", To: "oc-impact", Provenance: rejected})
	doc.Outcomes = append(doc.Outcomes, impact.Outcome{ID: "oc-dropped", Title: "Dropped", Level: impact.LevelShortTerm, Provenance: rejected})
	for _, rule := range []string{"IMP-LOG-001", "IMP-STR-003", "IMP-STR-004"} {
		if f := runRule(t, doc, rule); len(f) != 0 {
			t.Errorf("%s reported on a rejected element: %+v", rule, f)
		}
	}
}

func TestFeedbackLinksMayRunDownwardAndLoop(t *testing.T) {
	doc := clean()
	doc.Links = append(doc.Links, impact.CausalLink{ID: "ln-loop", From: "oc-impact", To: "oc-a", Feedback: true})
	for _, rule := range []string{"IMP-STR-002", "IMP-STR-006"} {
		if f := runRule(t, doc, rule); len(f) != 0 {
			t.Errorf("%s reported on a feedback link: %+v", rule, f)
		}
	}
}

func TestInterventionsDoNotCountAsPreconditions(t *testing.T) {
	doc := clean()
	// oc-s1 now has two targeting interventions but still no causal links
	// into it; IMP-LOG-004 is about links only.
	doc.Interventions = append(doc.Interventions, impact.Intervention{ID: "int-3", Title: "More", Targets: []string{"oc-s1"}})
	if f := runRule(t, doc, "IMP-LOG-004"); len(f) != 0 {
		t.Errorf("interventions were counted as incoming links: %+v", f)
	}
}

func TestNoImpactDoesNotAlsoReportEveryDeadEnd(t *testing.T) {
	doc := clean()
	doc.Outcomes[0].Level = impact.LevelLongTerm
	if f := runRule(t, doc, "IMP-STR-003"); len(f) != 0 {
		t.Errorf("IMP-STR-003 repeated the missing-impact problem: %+v", f)
	}
}

func TestReferenceExamplesReportNoErrors(t *testing.T) {
	for _, path := range []string{
		"../../examples/intermediary/impact-model.yaml", "../../examples/market-adoption/impact-model.yaml",
	} {
		doc, g := load(t, path)
		res, err := Run(doc, g, Options{Profile: ProfileFull})
		if err != nil {
			t.Fatal(err)
		}
		for _, f := range res.Findings {
			if f.Severity == "error" {
				t.Errorf("%s: %s on %s: %s", path, f.Rule, f.Element, f.Message)
			}
		}
	}
}
