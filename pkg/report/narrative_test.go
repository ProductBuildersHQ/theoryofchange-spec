package report

import (
	"strings"
	"testing"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/graph"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

func TestNarrativeAndActorsGoldens(t *testing.T) {
	in := exampleInput(t)
	noStatus, err := BuildView(in.Doc, nil, nil, Options{})
	if err != nil {
		t.Fatal(err)
	}
	golden(t, "narrative.md", Narrative(noStatus, in.Doc, nil))
	golden(t, "narrative-with-status.md", Narrative(in.View, in.Doc, in.Status))
	golden(t, "actors.md", Actors(in.View, in.Doc, in.Status))
	golden(t, "status-results.md", Results(in))
}

// mixed is a small theory with every kind of responsibility.
func mixed(t *testing.T) (*toc.Document, *View) {
	t.Helper()
	d := toc.NewDocument("mixed", "People cannot reach services.")
	d.Metadata.Name = "Mixed"
	d.Ceiling = &toc.Ceiling{Level: toc.LevelLongTerm}
	d.Actors = []toc.Actor{{ID: "act-us", Name: "Us", Roles: []toc.ActorRole{toc.RoleImplementer}}, {ID: "act-them", Name: "Partner Org"}}
	d.TargetGroups = []toc.TargetGroup{{ID: "tg-1", Name: "Rural residents"}}
	d.Causes = []toc.Cause{
		{ID: "c-struct", Statement: "No shared transport.", Level: toc.CauseStructural, AffectedGroupIDs: []string{"tg-1"}},
		{ID: "c-under", Statement: "Providers work alone.", Level: toc.CauseUnderlying},
		{ID: "c-imm", Statement: "Schedules clash.", Level: toc.CauseImmediate, OutOfScope: &toc.OutOfScope{Reason: "Providers set their own hours."}},
	}
	d.Outcomes = []toc.Outcome{
		{ID: "oc-impact", Title: "Services reachable", Level: toc.LevelImpact},
		{ID: "oc-own", Title: "We run rides", Level: toc.LevelIntermediate, AddressesCauses: []string{"c-under"}},
		{ID: "oc-shared", Title: "Providers share rides", Level: toc.LevelIntermediate, Responsibility: toc.ResponsibilityShared, DeliveredBy: []string{"act-us", "act-them"}},
		{ID: "oc-partner", Title: "The county funds buses", Level: toc.LevelShortTerm, Responsibility: toc.ResponsibilityPartner, DeliveredBy: []string{"act-them"}},
		{ID: "oc-nobody", Title: "Someone maps routes", Level: toc.LevelShortTerm, Responsibility: toc.ResponsibilityUnaddressed},
		{ID: "oc-orphan", Title: "We train drivers", Level: toc.LevelShortTerm},
	}
	d.Links = []toc.CausalLink{
		{ID: "l1", From: "oc-own", To: "oc-impact"}, {ID: "l2", From: "oc-shared", To: "oc-impact"},
		{ID: "l3", From: "oc-partner", To: "oc-own"}, {ID: "l4", From: "oc-nobody", To: "oc-shared"}, {ID: "l5", From: "oc-orphan", To: "oc-own"},
	}
	d.Interventions = []toc.Intervention{{ID: "int-1", Title: "Run rides", Targets: []string{"oc-own"}, OwnerID: "act-us"}}
	g, err := graph.Build(d)
	if err != nil {
		t.Fatal(err)
	}
	v, err := BuildView(d, g, nil, Options{})
	if err != nil {
		t.Fatal(err)
	}
	return d, v
}

func TestAccountableAndMonitoredOnlyAreSeparated(t *testing.T) {
	_, v := mixed(t)
	want := map[string]bool{"oc-own": true, "oc-shared": true, "oc-orphan": true, "oc-partner": false, "oc-nobody": false, "oc-impact": false}
	for _, o := range v.Outcomes {
		if o.Accountable != want[o.ID] {
			t.Errorf("%s: accountable = %v, want %v (responsibility %s, owned %v)", o.ID, o.Accountable, want[o.ID], o.Responsibility, o.Owned)
		}
	}
}

func TestResultsSeparateWhatWeAnswerForFromWhatWeMonitor(t *testing.T) {
	in := exampleInput(t)
	out := Results(in)
	a, m := strings.Index(out, "## Outcomes the organization answers for"), strings.Index(out, "## Monitored only")
	if a < 0 || m < a {
		t.Fatalf("sections missing or out of order:\n%.600s", out)
	}
	// The partner-delivered outcome sits in the monitored section, with its responsibility named.
	officials := strings.Index(out, "Local officials and funders engage members in planning")
	if officials < m || !strings.Contains(out[m:], "Responsibility: partner.") {
		t.Errorf("a partner-delivered outcome is not in the monitored-only section")
	}
}

func TestNarrativeChecksTheSolutionAgainstTheProblem(t *testing.T) {
	d, v := mixed(t)
	out := Narrative(v, d, nil)
	for _, want := range []string{
		"### Structural causes", "**Not addressed**", // c-struct has nothing addressing it
		"Set aside: Providers set their own hours.", // c-imm is out of scope
		"We run rides (`oc-own`)",                   // c-under is addressed by oc-own
		"### Delivered by partners", "The county funds buses", "### Needed, and delivered by no one", "Someone maps routes",
		"Rural residents",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("narrative is missing %q", want)
		}
	}
	if strings.Contains(out, "Status as of") {
		t.Error("a narrative without observations must not claim a status date")
	}
}

func TestNarrativeWithoutAProblemTreeFallsBackToRootCauses(t *testing.T) {
	d, _ := mixed(t)
	d.Causes = nil
	d.Outcomes[1].AddressesCauses = nil
	d.Context.RootCauses = []string{"Distance."}
	g, err := graph.Build(d)
	if err != nil {
		t.Fatal(err)
	}
	v, err := BuildView(d, g, nil, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if out := Narrative(v, d, nil); !strings.Contains(out, "No problem tree has been drawn") || !strings.Contains(out, "Distance.") {
		t.Errorf("fallback missing:\n%s", out)
	}
}

func TestActorsReport(t *testing.T) {
	d, v := mixed(t)
	out := Actors(v, d, nil)
	// The shared outcome needs collaboration; the partner outcome is listed as monitored.
	for _, want := range []string{"Providers share rides (`oc-shared`), shared: Us, Partner Org", "The county funds buses (`oc-partner`), partner", "## Outcomes partners deliver, which are monitored"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q:\n%s", want, out)
		}
	}
	// oc-orphan is ours but nobody is named and no owned intervention targets it.
	section := out[strings.Index(out, "## Outcomes with no one delivering them"):]
	if !strings.Contains(section, "We train drivers (`oc-orphan`)") || strings.Contains(section[:strings.Index(section, "## Outcomes partners")], "We run rides") {
		t.Errorf("undelivered outcomes wrong:\n%s", section)
	}
	// An actor's row lists what it runs.
	if !strings.Contains(out, "| Us (`act-us`) | Implementer |") || !strings.Contains(out, "Run rides") {
		t.Errorf("division of labour row missing:\n%s", out)
	}
}

func TestActorsWithNoPartnerOutcomes(t *testing.T) {
	d, v := mixed(t)
	for i := range v.Outcomes {
		v.Outcomes[i].Responsibility = string(toc.ResponsibilityOwn)
	}
	if out := Actors(v, d, nil); !strings.Contains(out, "None.") {
		t.Error("with no partner outcomes the section should say none")
	}
}
