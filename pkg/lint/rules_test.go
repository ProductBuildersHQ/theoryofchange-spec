package lint

import (
	"slices"
	"testing"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/graph"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

// clean returns a small theory that breaks none of the structure or logic
// rules: one impact reached by two intermediate outcomes, each reached from a
// short-term outcome that an intervention brings about.
func clean() *toc.Document {
	d := toc.NewDocument("lint-base", "People cannot reach services.")
	all := &toc.Logic{Type: toc.LogicAll}
	d.Outcomes = []toc.Outcome{
		{ID: "oc-impact", Title: "Services are reachable", Level: toc.LevelImpact, Logic: all},
		{ID: "oc-a", Title: "Providers share rides", Level: toc.LevelIntermediate},
		{ID: "oc-b", Title: "Providers share schedules", Level: toc.LevelIntermediate},
		{ID: "oc-s1", Title: "Providers know each other", Level: toc.LevelShortTerm},
		{ID: "oc-s2", Title: "Providers meet monthly", Level: toc.LevelShortTerm},
	}
	link := func(id, from, to string) toc.CausalLink {
		return toc.CausalLink{ID: id, From: from, To: to, Rationale: "Because it helps.", AssumptionIDs: []string{"asm-1"}}
	}
	d.Links = []toc.CausalLink{
		link("ln-a-i", "oc-a", "oc-impact"), link("ln-b-i", "oc-b", "oc-impact"),
		link("ln-s1-a", "oc-s1", "oc-a"), link("ln-s2-b", "oc-s2", "oc-b"),
	}
	d.Assumptions = []toc.Assumption{{ID: "asm-1", Statement: "Providers want to cooperate.", AppliesTo: []string{"ln-a-i"}}}
	d.Interventions = []toc.Intervention{
		{ID: "int-1", Title: "Host mixers", Targets: []string{"oc-s1"}},
		{ID: "int-2", Title: "Host meetings", Targets: []string{"oc-s2"}},
	}
	return d
}

func runRule(t *testing.T, doc *toc.Document, id string) []Finding {
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
		seed    func(d *toc.Document)
	}{
		{"TOC-STR-001", "", func(d *toc.Document) { d.Outcomes[0].Level = toc.LevelLongTerm }},
		{"TOC-STR-002", "oc-impact", func(d *toc.Document) {
			d.Links = append(d.Links, toc.CausalLink{ID: "ln-i-a", From: "oc-impact", To: "oc-a", Rationale: "r", AssumptionIDs: []string{"asm-1"}})
		}},
		{"TOC-STR-003", "oc-lost", func(d *toc.Document) {
			d.Outcomes = append(d.Outcomes, toc.Outcome{ID: "oc-lost", Title: "Strays", Level: toc.LevelShortTerm})
			d.Interventions[0].Targets = append(d.Interventions[0].Targets, "oc-lost")
		}},
		{"TOC-STR-004", "oc-s2", func(d *toc.Document) { d.Interventions[1].Targets = []string{"oc-s1"} }},
		{"TOC-STR-005", "oc-impact", func(d *toc.Document) { d.Outcomes[0].Logic = &toc.Logic{Type: toc.LogicAtLeast, K: 3} }},
		{"TOC-STR-006", "ln-down", func(d *toc.Document) {
			d.Links = append(d.Links, toc.CausalLink{ID: "ln-down", From: "oc-a", To: "oc-s1", Rationale: "r"})
		}},
		{"TOC-STR-007", "grp-y", func(d *toc.Document) {
			d.Groups = []toc.PreconditionGroup{{ID: "grp-y", Target: "oc-impact", Members: []string{"oc-a", "oc-s1"}, Logic: toc.Logic{Type: toc.LogicAll}}}
		}},
		{"TOC-LOG-001", "ln-a-i", func(d *toc.Document) { d.Links[0].Rationale = "" }},
		{"TOC-LOG-002", "ln-s1-a", func(d *toc.Document) { d.Links[2].AssumptionIDs = nil }},
		{"TOC-LOG-003", "oc-impact", func(d *toc.Document) { d.Links = slices.Delete(d.Links, 1, 2) }},
		{"TOC-LOG-004", "oc-impact", func(d *toc.Document) { d.Outcomes[0].Logic = nil }},
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
	rejected := &toc.Provenance{Basis: toc.BasisRejected}
	doc.Links = append(doc.Links, toc.CausalLink{ID: "ln-dropped", From: "oc-s1", To: "oc-impact", Provenance: rejected})
	doc.Outcomes = append(doc.Outcomes, toc.Outcome{ID: "oc-dropped", Title: "Dropped", Level: toc.LevelShortTerm, Provenance: rejected})
	for _, rule := range []string{"TOC-LOG-001", "TOC-STR-003", "TOC-STR-004"} {
		if f := runRule(t, doc, rule); len(f) != 0 {
			t.Errorf("%s reported on a rejected element: %+v", rule, f)
		}
	}
}

func TestFeedbackLinksMayRunDownwardAndLoop(t *testing.T) {
	doc := clean()
	doc.Links = append(doc.Links, toc.CausalLink{ID: "ln-loop", From: "oc-impact", To: "oc-a", Feedback: true})
	for _, rule := range []string{"TOC-STR-002", "TOC-STR-006"} {
		if f := runRule(t, doc, rule); len(f) != 0 {
			t.Errorf("%s reported on a feedback link: %+v", rule, f)
		}
	}
}

func TestInterventionsDoNotCountAsPreconditions(t *testing.T) {
	doc := clean()
	// oc-s1 now has two targeting interventions but still no causal links
	// into it; TOC-LOG-004 is about links only.
	doc.Interventions = append(doc.Interventions, toc.Intervention{ID: "int-3", Title: "More", Targets: []string{"oc-s1"}})
	if f := runRule(t, doc, "TOC-LOG-004"); len(f) != 0 {
		t.Errorf("interventions were counted as incoming links: %+v", f)
	}
}

func TestNoImpactDoesNotAlsoReportEveryDeadEnd(t *testing.T) {
	doc := clean()
	doc.Outcomes[0].Level = toc.LevelLongTerm
	if f := runRule(t, doc, "TOC-STR-003"); len(f) != 0 {
		t.Errorf("TOC-STR-003 repeated the missing-impact problem: %+v", f)
	}
}

func TestReferenceExamplesReportNoErrors(t *testing.T) {
	for _, path := range []string{
		"../../examples/intermediary/theory.yaml", "../../examples/market-adoption/theory.yaml",
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
