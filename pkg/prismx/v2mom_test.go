package prismx

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/grokify/prism-roadmap/goals/v2mom"

	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
)

func TestV2MOMExportValidates(t *testing.T) {
	doc, st := example(t)
	for name, status := range map[string]*impact.Status{"with status": st, "without": nil} {
		out, err := ToV2MOM(doc, status)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if errs := v2mom.Errors(out.Validate(v2mom.DefaultValidationOptions())); len(errs) != 0 {
			t.Errorf("%s: the exported V2MOM does not validate: %+v", name, errs)
		}
	}
}

func TestV2MOMMapping(t *testing.T) {
	doc, st := example(t)
	out, err := ToV2MOM(doc, st)
	if err != nil {
		t.Fatal(err)
	}

	if out.Vision != doc.Context.Vision {
		t.Errorf("vision = %q, want context.vision", out.Vision)
	}
	if len(out.Values) != len(doc.Context.Values) {
		t.Fatalf("%d values, want %d", len(out.Values), len(doc.Context.Values))
	}
	for i, v := range out.Values {
		if v.Name != doc.Context.Values[i].Name || v.Priority != i+1 {
			t.Errorf("value %d = %q priority %d, want %q priority %d", i, v.Name, v.Priority, doc.Context.Values[i].Name, i+1)
		}
	}
	if out.Metadata.Structure != v2mom.StructureNested {
		t.Errorf("structure = %q, want nested", out.Metadata.Structure)
	}

	// Each pathway is a method, in file order; a rejected pathway is not.
	var pathways []string
	rejected := 0
	for _, p := range doc.Pathways {
		if p.Rejected() {
			rejected++
			continue
		}
		pathways = append(pathways, p.ID)
	}
	if rejected == 0 {
		t.Fatal("the example has no rejected pathway to leave out")
	}
	var methods []string
	for _, m := range out.Methods {
		methods = append(methods, m.ID)
	}
	if !slices.Equal(methods, pathways) {
		t.Fatalf("methods = %v, want the pathways %v", methods, pathways)
	}

	learning := out.Methods[0]
	p, _ := doc.Pathway(learning.ID)
	// The North Star leads the method's measures, and guardrails follow.
	if len(learning.Measures) == 0 || learning.Measures[0].ID != p.Focus.NorthStarID {
		t.Fatalf("first measure of %s is not its North Star %q: %+v", learning.ID, p.Focus.NorthStarID, learning.Measures)
	}
	if !strings.HasPrefix(learning.Measures[0].Description, "North Star") {
		t.Errorf("North Star measure description = %q", learning.Measures[0].Description)
	}
	for _, g := range p.Focus.GuardrailIDs {
		if !slices.ContainsFunc(learning.Measures, func(m v2mom.Measure) bool { return m.ID == g }) {
			t.Errorf("guardrail %s is not a measure of %s", g, learning.ID)
		}
	}
	if learning.Measures[0].Status == "" || learning.Measures[0].Current == "" {
		t.Errorf("with a status, the North Star has a current value and status: %+v", learning.Measures[0])
	}

	// The method's projects are its interventions, listed once at the top with
	// the method they belong to.
	if !slices.Equal(learning.Projects, p.InterventionIDs) {
		t.Errorf("projects of %s = %v, want its interventions %v", learning.ID, learning.Projects, p.InterventionIDs)
	}
	for _, id := range learning.Projects {
		i := slices.IndexFunc(out.Projects, func(pr v2mom.Project) bool { return pr.ID == id })
		if i < 0 || out.Projects[i].MethodID != learning.ID {
			t.Errorf("project %s is not listed under method %s", id, learning.ID)
		}
	}

	// An obstacle that applies within a pathway is under its method; one that
	// applies to none is listed for the whole V2MOM.
	under := map[string]bool{}
	for _, m := range out.Methods {
		for _, o := range m.Obstacles {
			under[o.ID] = true
		}
	}
	for _, o := range out.Obstacles {
		if under[o.ID] {
			t.Errorf("obstacle %s is listed both under a method and for the whole V2MOM", o.ID)
		}
	}
	if len(under) == 0 || len(out.Obstacles) == 0 {
		t.Errorf("the example has obstacles both within pathways and across them: under=%v global=%d", under, len(out.Obstacles))
	}

	// Only high or critical assumptions not yet shown to hold are obstacles.
	ids := map[string]bool{}
	for k := range under {
		ids[k] = true
	}
	for _, o := range out.Obstacles {
		ids[o.ID] = true
	}
	assessed := map[string]impact.AssumptionStatus{}
	for _, a := range st.Assumptions {
		assessed[a.AssumptionID] = a.Status
	}
	for _, a := range doc.Assumptions {
		want := !a.Rejected() && openAssumption(a.Criticality, assumptionStatus(a, assessed))
		if ids[a.ID] != want {
			t.Errorf("assumption %s (criticality %s, status %s): listed = %v, want %v", a.ID, a.Criticality, a.Status, ids[a.ID], want)
		}
	}

	// Without a status there are no current values, progress, or statuses.
	bare, err := ToV2MOM(doc, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range bare.Methods {
		for _, ms := range m.Measures {
			if ms.Current != "" || ms.Progress != 0 || ms.Status != "" {
				t.Errorf("measure %s has status fields without a status: %+v", ms.ID, ms)
			}
		}
	}
}

func TestV2MOMRefusesWhatItCannotDraw(t *testing.T) {
	tests := []struct {
		name  string
		edit  func(*impact.Model)
		names []string
	}{
		{"no vision", func(d *impact.Model) { d.Context.Vision = "" }, []string{"context.vision"}},
		{"no values", func(d *impact.Model) { d.Context.Values = nil }, []string{"context.values"}},
		{"no pathways", func(d *impact.Model) { d.Pathways = nil }, []string{"a pathway"}},
		{"nothing", func(d *impact.Model) {
			d.Context.Vision, d.Context.Values, d.Pathways = "", nil, nil
		}, []string{"context.vision", "context.values", "a pathway"}},
	}
	for _, tt := range tests {
		doc, _ := example(t)
		tt.edit(doc)
		_, err := ToV2MOM(doc, nil)
		if !errors.Is(err, ErrNotV2MOMReady) {
			t.Fatalf("%s: error = %v, want ErrNotV2MOMReady", tt.name, err)
		}
		for _, n := range tt.names {
			if !strings.Contains(err.Error(), n) {
				t.Errorf("%s: %q does not name %s", tt.name, err, n)
			}
		}
	}
	if _, err := ToV2MOM(nil, nil); err == nil {
		t.Error("a nil document was accepted")
	}
}

// With observations, an assumption's assessed status decides whether it is an
// obstacle; without them, the status written in the model does.
func TestV2MOMObstaclesFollowAssessedAssumptions(t *testing.T) {
	doc, st := example(t)
	listed := func(v *v2mom.V2MOM) map[string]bool {
		ids := map[string]bool{}
		for _, o := range v.Obstacles {
			ids[o.ID] = true
		}
		for _, m := range v.Methods {
			for _, o := range m.Obstacles {
				ids[o.ID] = true
			}
		}
		return ids
	}
	withStatus, err := ToV2MOM(doc, st)
	if err != nil {
		t.Fatal(err)
	}
	without, err := ToV2MOM(doc, nil)
	if err != nil {
		t.Fatal(err)
	}
	a, b := listed(withStatus), listed(without)
	changed := 0
	for _, s := range st.Assumptions {
		if a[s.AssumptionID] != b[s.AssumptionID] {
			changed++
		}
	}
	if changed == 0 {
		t.Fatal("no assumption's assessed status changes whether it is an obstacle; the example no longer tests this")
	}
}
