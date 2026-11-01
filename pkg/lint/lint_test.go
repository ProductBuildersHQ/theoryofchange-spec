package lint

import (
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/plexusone/structured-evaluation/summary"

	"github.com/ProductBuildersHQ/impact-model/pkg/graph"
	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
	"github.com/ProductBuildersHQ/impact-model/pkg/impactio"
	"github.com/ProductBuildersHQ/impact-model/pkg/validate"
)

// fake is a rule that reports a fixed set of messages.
type fake struct {
	id       string
	severity validate.Severity
	messages []string
	calls    *int
}

func (f fake) ID() string                  { return f.id }
func (f fake) Severity() validate.Severity { return f.severity }
func (f fake) Description() string         { return "checks " + f.id }
func (f fake) Check(*Context) []Finding {
	if f.calls != nil {
		*f.calls++
	}
	var out []Finding
	for _, m := range f.messages {
		out = append(out, Finding{Message: m, Element: "el-" + m})
	}
	return out
}

func registry(t *testing.T, rules ...Rule) *Registry {
	t.Helper()
	r := NewRegistry()
	for _, rule := range rules {
		if err := r.Add(rule); err != nil {
			t.Fatal(err)
		}
	}
	return r
}

func load(t *testing.T, path string) (*impact.Model, *graph.Graph) {
	t.Helper()
	doc, err := impactio.LoadDocument(path)
	if err != nil {
		t.Fatal(err)
	}
	g, err := graph.Build(doc)
	if err != nil {
		t.Fatal(err)
	}
	return doc, g
}

func TestCategory(t *testing.T) {
	for id, want := range map[string]string{"IMP-STR-001": "STR", "IMP-MEA-004": "MEA", "bad": "", "IMP-X": ""} {
		if got := Category(id); got != want {
			t.Errorf("Category(%q) = %q, want %q", id, got, want)
		}
	}
}

func TestRegistryRejectsBadAndDuplicateIDs(t *testing.T) {
	r := NewRegistry()
	if err := r.Add(fake{id: "STR-001"}); err == nil {
		t.Error("an ID without the TOC prefix was accepted")
	}
	if err := r.Add(fake{id: "IMP-STR-001"}); err != nil {
		t.Fatal(err)
	}
	if err := r.Add(fake{id: "IMP-STR-001"}); !errors.Is(err, ErrDuplicateRule) {
		t.Errorf("duplicate error = %v", err)
	}
	if err := r.Add(fake{id: "IMP-ASM-001"}); err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, rule := range r.Rules() {
		ids = append(ids, rule.ID())
	}
	if !slices.Equal(ids, []string{"IMP-ASM-001", "IMP-STR-001"}) {
		t.Errorf("rules not sorted: %v", ids)
	}
}

func TestProfilesSelectRules(t *testing.T) {
	reg := registry(t,
		fake{id: "IMP-INT-001", severity: validate.SeverityError, messages: []string{"e"}},
		fake{id: "IMP-STR-004", severity: validate.SeverityWarning, messages: []string{"structural warning"}},
		fake{id: "IMP-ASM-001", severity: validate.SeverityWarning, messages: []string{"w"}},
		fake{id: "IMP-ASM-003", severity: validate.SeverityInfo, messages: []string{"i"}},
	)
	want := map[Profile][]string{
		ProfileDraft:    {"IMP-INT-001", "IMP-STR-004"},
		ProfileStandard: {"IMP-ASM-001", "IMP-INT-001", "IMP-STR-004"},
		ProfileFull:     {"IMP-ASM-001", "IMP-ASM-003", "IMP-INT-001", "IMP-STR-004"},
	}
	doc := &impact.Model{}
	for p, ids := range want {
		res, err := Run(doc, nil, Options{Profile: p, Registry: reg})
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(res.Ran, ids) {
			t.Errorf("%s ran %v, want %v", p, res.Ran, ids)
		}
	}
	// The default profile is standard.
	res, _ := Run(doc, nil, Options{Registry: reg})
	if res.Profile != ProfileStandard {
		t.Errorf("default profile = %q", res.Profile)
	}
}

func TestRunStampsRuleAndSortsFindings(t *testing.T) {
	reg := registry(t,
		fake{id: "IMP-MEA-001", severity: validate.SeverityWarning, messages: []string{"b", "a"}},
		fake{id: "IMP-ASM-001", severity: validate.SeverityWarning, messages: []string{"z"}},
	)
	res, err := Run(&impact.Model{}, nil, Options{Registry: reg})
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, f := range res.Findings {
		if f.Severity != validate.SeverityWarning {
			t.Errorf("finding severity = %q", f.Severity)
		}
		got = append(got, f.Rule+":"+f.Message)
	}
	if !slices.Equal(got, []string{"IMP-ASM-001:z", "IMP-MEA-001:a", "IMP-MEA-001:b"}) {
		t.Errorf("findings = %v", got)
	}
}

func TestDisableAndUnknowns(t *testing.T) {
	calls := 0
	reg := registry(t, fake{id: "IMP-ASM-001", severity: validate.SeverityWarning, messages: []string{"w"}, calls: &calls})
	res, err := Run(&impact.Model{}, nil, Options{Registry: reg, Disable: []string{"IMP-ASM-001"}})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 0 || len(res.Findings) != 0 || !slices.Equal(res.Disabled, []string{"IMP-ASM-001"}) || len(res.Ran) != 0 {
		t.Errorf("a disabled rule ran: calls=%d result=%+v", calls, res)
	}
	if _, err := Run(&impact.Model{}, nil, Options{Registry: reg, Disable: []string{"IMP-NOPE-001"}}); !errors.Is(err, ErrUnknownRule) {
		t.Errorf("unknown rule error = %v", err)
	}
	if _, err := Run(&impact.Model{}, nil, Options{Registry: reg, Profile: "strict"}); !errors.Is(err, ErrUnknownProfile) {
		t.Errorf("unknown profile error = %v", err)
	}
}

func TestSummaryReportShape(t *testing.T) {
	reg := registry(t,
		fake{id: "IMP-STR-001", severity: validate.SeverityError},
		fake{id: "IMP-ASM-001", severity: validate.SeverityWarning, messages: []string{"no test"}},
		fake{id: "IMP-ASM-003", severity: validate.SeverityInfo, messages: []string{"advice"}},
		fake{id: "IMP-MEA-001", severity: validate.SeverityError, messages: []string{"unmeasured"}},
		fake{id: "IMP-INT-002", severity: validate.SeverityWarning},
	)
	res, err := Run(&impact.Model{}, nil, Options{Profile: ProfileFull, Registry: reg, Disable: []string{"IMP-INT-002"}})
	if err != nil {
		t.Fatal(err)
	}
	at := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	rep := res.SummaryReport("riverbend", "1.0.0", at)

	if rep.Project != "riverbend" || rep.Version != "1.0.0" || !rep.GeneratedAt.Equal(at) || !strings.Contains(rep.Phase, "FULL") {
		t.Errorf("header = %+v", rep)
	}
	if rep.Status != summary.StatusNoGo || rep.IsGo() {
		t.Errorf("status = %q, want NO-GO", rep.Status)
	}
	teams := map[string]summary.TeamSection{}
	var order []string
	for _, tm := range rep.Teams {
		teams[tm.Name] = tm
		order = append(order, tm.ID)
	}
	if !slices.Equal(order, []string{"lint-asm", "lint-int", "lint-mea", "lint-str"}) {
		t.Errorf("teams = %v", order)
	}
	status := func(team, task string) summary.Status {
		for _, tk := range teams[team].Tasks {
			if tk.ID == task {
				return tk.Status
			}
		}
		t.Fatalf("no task %s in %s", task, team)
		return ""
	}
	for _, c := range []struct {
		team, task string
		want       summary.Status
	}{
		{"Structure", "IMP-STR-001", summary.StatusGo},
		{"Assumptions", "IMP-ASM-001", summary.StatusWarn},
		{"Assumptions", "IMP-ASM-003", summary.StatusGo},
		{"Measurement", "IMP-MEA-001", summary.StatusNoGo},
		{"Interventions", "IMP-INT-002", summary.StatusSkip},
	} {
		if got := status(c.team, c.task); got != c.want {
			t.Errorf("%s: %s = %s, want %s", c.team, c.task, got, c.want)
		}
	}
	if teams["Assumptions"].Status != summary.StatusWarn {
		t.Errorf("team status = %q", teams["Assumptions"].Status)
	}
	if _, err := json.Marshal(rep); err != nil {
		t.Errorf("report does not marshal: %v", err)
	}
	// Same inputs, same report.
	a, _ := json.Marshal(rep)
	b, _ := json.Marshal(res.SummaryReport("riverbend", "1.0.0", at))
	if string(a) != string(b) {
		t.Error("two reports differ")
	}
}

func TestMinimalDocumentIsGoUnderDraft(t *testing.T) {
	doc, g := load(t, "../impact/testdata/minimal.json")
	res, err := Run(doc, g, Options{Profile: ProfileDraft})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Go() {
		t.Errorf("the minimal document is not GO under draft: %+v", res.Findings)
	}
}

func TestGoMeansNoErrors(t *testing.T) {
	r := &Result{Findings: []Finding{{Severity: validate.SeverityWarning}, {Severity: validate.SeverityInfo}}}
	if !r.Go() {
		t.Error("warnings must not block")
	}
	r.Findings = append(r.Findings, Finding{Severity: validate.SeverityError})
	if r.Go() || r.Count(validate.SeverityError) != 1 {
		t.Error("an error must block")
	}
}
