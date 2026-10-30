package prismx

import (
	"os/exec"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/grokify/structured-goals/okr"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/graph"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/measure"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/tocio"
)

func example(t *testing.T) (*toc.Document, *toc.Status) {
	t.Helper()
	doc, err := tocio.LoadDocument("../../examples/intermediary/theory.yaml")
	if err != nil {
		t.Fatal(err)
	}
	g, err := graph.Build(doc)
	if err != nil {
		t.Fatal(err)
	}
	sets, err := tocio.LoadObservations("../../examples/intermediary/observations-2026-q3.yaml")
	if err != nil {
		t.Fatal(err)
	}
	asOf, _ := time.Parse(time.DateOnly, "2026-09-30")
	st, err := measure.Evaluate(doc, g, sets, asOf, toc.DefaultStatusOptions())
	if err != nil {
		t.Fatal(err)
	}
	return doc, st
}

func TestOKRExportValidates(t *testing.T) {
	doc, st := example(t)
	for name, status := range map[string]*toc.Status{"with status": st, "without": nil} {
		out, err := ToOKR(doc, status)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if len(out.Objectives) == 0 {
			t.Fatalf("%s: no objectives", name)
		}
		if errs := okr.Errors(out.Validate(okr.DefaultValidationOptions())); len(errs) != 0 {
			t.Errorf("%s: the exported OKR document does not validate: %+v", name, errs)
		}
	}
}

func TestOKRMapping(t *testing.T) {
	doc, st := example(t)
	out, err := ToOKR(doc, st)
	if err != nil {
		t.Fatal(err)
	}
	if out.Metadata.ID != doc.ID || out.Metadata.Period != "2026-09-30" || out.Metadata.Version != doc.Metadata.Version {
		t.Errorf("metadata = %+v", out.Metadata)
	}
	var obj *okr.Objective
	for i := range out.Objectives {
		if out.Objectives[i].ID == "oc-lt-collab" {
			obj = &out.Objectives[i]
		}
	}
	if obj == nil {
		t.Fatalf("no objective for an owned outcome with indicators: %v", ids(out.Objectives))
	}
	// Only accountable outcomes become objectives: the partner-delivered one does not.
	for _, notOurs := range []string{"oc-im-officials-engage", "oc-resident-access"} {
		if slices.Contains(ids(out.Objectives), notOurs) {
			t.Errorf("%s is partner-delivered or above the ceiling, but became an objective", notOurs)
		}
	}
	for _, kr := range obj.KeyResults {
		if kr.Title == "" || kr.Score < 0 || kr.Score > 1 {
			t.Errorf("key result = %+v", kr)
		}
	}
	// A key result with a numeric indicator carries baseline, target, current, unit, and a status.
	var found bool
	for _, o := range out.Objectives {
		for _, kr := range o.KeyResults {
			if kr.ID == "ind-grant-applications" {
				found = true
				if kr.Baseline != "120 applications" || kr.Target != "200 applications" || !strings.HasSuffix(kr.Current, " applications") || kr.DueDate != "2028-12-31" {
					t.Errorf("ind-grant-applications key result = %+v", kr)
				}
			}
		}
	}
	if !found {
		t.Error("ind-grant-applications is not a key result")
	}
	if len(out.Risks) == 0 || out.Risks[0].Impact == "" {
		t.Errorf("risks = %+v", out.Risks)
	}
}

func ids(objs []okr.Objective) []string {
	var out []string
	for _, o := range objs {
		out = append(out, o.ID)
	}
	return out
}

func TestOutcomeJourneys(t *testing.T) {
	doc, st := example(t)
	js, err := ToOutcomeJourneys(doc, st)
	if err != nil {
		t.Fatal(err)
	}
	if len(js) == 0 {
		t.Fatal("no journeys")
	}
	var hours *int
	for i, j := range js {
		if j.ID == "ind-hours-learning" {
			hours = &i
		}
		if j.CurrentState == nil || len(j.TargetStates) == 0 {
			t.Errorf("journey %s is incomplete", j.ID)
		}
		if strings.Contains(j.ID, "ind-practice-maturity") {
			t.Error("a qualitative indicator became a journey")
		}
	}
	if hours == nil {
		t.Fatal("ind-hours-learning is not a journey")
	}
	j := js[*hours]
	if j.CurrentState.Value != 4880 && j.CurrentState.MeasuredAt != "2026-09-30" {
		t.Errorf("current = %+v", j.CurrentState)
	}
	if len(j.TargetStates) != 3 || j.TargetStates[0].PeriodID != "2026-12-31" || j.Metadata["tocId"] != doc.ID {
		t.Errorf("journey = %+v", j)
	}
	// Without a status the journey starts from the baseline.
	bare, _ := ToOutcomeJourneys(doc, nil)
	for _, b := range bare {
		if b.ID == "ind-hours-learning" && b.CurrentState.Value != 3200 {
			t.Errorf("baseline journey = %+v", b.CurrentState)
		}
	}
	if _, err := ToOutcomeJourneys(nil, nil); err == nil {
		t.Error("a nil document was accepted")
	}
}

func TestExportsAreDeterministic(t *testing.T) {
	doc, st := example(t)
	a, _ := ToOKR(doc, st)
	b, _ := ToOKR(doc, st)
	ja, _ := a.JSON()
	jb, _ := b.JSON()
	if string(ja) != string(jb) {
		t.Error("two OKR exports differ")
	}
}

// The PRISM roadmap and goals modules are used here and nowhere else in the
// library, so the model and its tools stay usable without them. The command
// line may use this package to export; nothing else may import the modules.
func TestOnlyThisPackageImportsThePRISMFamily(t *testing.T) {
	if testing.Short() {
		t.Skip("runs go list")
	}
	out, err := exec.Command("go", "list", "-f", "{{.ImportPath}}|{{join .Imports \" \"}}", "../../...").CombinedOutput()
	if err != nil {
		t.Fatalf("go list: %v\n%s", err, out)
	}
	forbidden := []string{"github.com/grokify/prism-roadmap", "github.com/grokify/structured-goals"}
	allowed := map[string]bool{
		"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/prismx": true,
	}
	checked := 0
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		pkg, imports, _ := strings.Cut(line, "|")
		checked++
		if allowed[pkg] {
			continue
		}
		for _, imp := range strings.Fields(imports) {
			for _, f := range forbidden {
				if strings.HasPrefix(imp, f) {
					t.Errorf("%s imports %s; only pkg/prismx may", pkg, imp)
				}
			}
		}
	}
	if checked < 20 {
		t.Errorf("only %d packages were checked", checked)
	}
}
