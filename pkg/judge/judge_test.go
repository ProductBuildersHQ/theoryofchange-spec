package judge

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/plexusone/structured-evaluation/rubric"

	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
	"github.com/ProductBuildersHQ/impact-model/pkg/impactio"
)

func load(t *testing.T) *impact.Model {
	t.Helper()
	doc, err := impactio.LoadDocument("../../examples/intermediary/impact-model.yaml")
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

func bundle(t *testing.T) *Bundle {
	t.Helper()
	b, err := Build(load(t), "core")
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// verdict passes every category, except the overrides.
func verdict(b *Bundle, judgeID string, override map[string]rubric.ScoreValue) *Verdict {
	v := &Verdict{BundleHash: b.BundleHash}
	v.Metadata = rubric.ReportMetadata{Document: "impact-model.yaml", GeneratedAt: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), ReviewerID: judgeID}
	v.ReviewType = "toc"
	v.Judge = &rubric.JudgeMetadata{JudgeID: judgeID}
	v.Findings = []rubric.Finding{}
	for _, c := range b.Rubric.Categories {
		s := rubric.ScorePass
		if o, ok := override[c.ID]; ok {
			s = o
		}
		v.Categories = append(v.Categories, rubric.CategoryResult{Category: c.ID, Score: s, Reasoning: "cited oc-impact"})
	}
	v.Pass = len(override) == 0
	v.Decision = rubric.Decision{Status: rubric.DecisionPass, Rationale: "ok"}
	v.OverallDecision = "pass"
	return v
}

func TestBundleIsDeterministicAndHashed(t *testing.T) {
	a, b := bundle(t), bundle(t)
	ja, _ := json.Marshal(a)
	jb, _ := json.Marshal(b)
	if string(ja) != string(jb) || a.BundleHash == "" {
		t.Error("two bundles of the same impact model differ")
	}
	if a.Kind != BundleKind || a.Profile != "core" || a.Rubric == nil || a.Lint == nil || a.Instructions == "" || !strings.Contains(a.Narrative, "assessment") {
		t.Errorf("bundle is incomplete: %+v", a)
	}
	changed := load(t)
	changed.Outcomes[0].Title += " (edited)"
	c, err := Build(changed, "core")
	if err != nil {
		t.Fatal(err)
	}
	if c.BundleHash == a.BundleHash {
		t.Error("editing the impact model did not change the hash")
	}
	dev, err := Build(load(t), "development")
	if err != nil {
		t.Fatal(err)
	}
	if dev.BundleHash == a.BundleHash {
		t.Error("the profile is not part of the hash")
	}
	if _, err := Build(load(t), "nonsense"); err == nil {
		t.Error("an unknown profile was accepted")
	}
}

func TestRestrictedEvidenceNeverReachesTheJudge(t *testing.T) {
	doc := load(t)
	doc.Evidence = append(doc.Evidence,
		impact.Evidence{ID: "ev-secret", Title: "Board minutes", Excerpt: "SECRET-EXCERPT", Summary: "SECRET-SUMMARY", URL: "https://example.org/SECRET-URL", Sensitivity: impact.SensitivityRestricted},
		impact.Evidence{ID: "ev-unset", Title: "Interview", Excerpt: "UNSET-EXCERPT"},
		impact.Evidence{ID: "ev-open", Title: "Public report", Excerpt: "PUBLIC-EXCERPT", Sensitivity: impact.SensitivityPublic},
	)
	b, err := Build(doc, "core")
	if err != nil {
		t.Fatal(err)
	}
	all := string(b.Document) + b.Narrative
	for _, leak := range []string{"SECRET-EXCERPT", "SECRET-SUMMARY", "SECRET-URL", "UNSET-EXCERPT"} {
		if strings.Contains(all, leak) {
			t.Errorf("%s reached the bundle", leak)
		}
	}
	if !strings.Contains(string(b.Document), "PUBLIC-EXCERPT") || !strings.Contains(string(b.Document), "ev-secret") {
		t.Error("public evidence was removed, or the restricted record vanished (its title and ID should remain)")
	}
	if doc.Evidence[len(doc.Evidence)-3].Excerpt == "" {
		t.Error("redaction modified the caller's document")
	}
}

func TestRecordChecksTheVerdict(t *testing.T) {
	b := bundle(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "impact-model.yaml")

	v := verdict(b, "judge-a", nil)
	got, err := Record(path, b, v, "")
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(dir, ".toc", "judgements", b.BundleHash, "judge-a.json"); got != want {
		t.Errorf("stored at %s, want %s", got, want)
	}

	wrongHash := verdict(b, "judge-b", nil)
	wrongHash.BundleHash = "deadbeef"
	if _, err := Record(path, b, wrongHash, ""); !errors.Is(err, ErrWrongBundle) {
		t.Errorf("wrong hash error = %v", err)
	}

	unknown := verdict(b, "judge-c", nil)
	unknown.Categories[0].Category = "vibes"
	if _, err := Record(path, b, unknown, ""); !errors.Is(err, ErrInvalidReport) {
		t.Errorf("unknown category error = %v", err)
	}

	missing := verdict(b, "judge-d", nil)
	missing.Categories = missing.Categories[1:]
	if _, err := Record(path, b, missing, ""); !errors.Is(err, ErrInvalidReport) {
		t.Errorf("missing required category error = %v", err)
	}

	twice := verdict(b, "judge-e", nil)
	twice.Categories = append(twice.Categories, twice.Categories[0])
	if _, err := Record(path, b, twice, ""); !errors.Is(err, ErrInvalidReport) {
		t.Errorf("duplicate category error = %v", err)
	}
}

func TestJudgeNamesCannotEscapeTheStore(t *testing.T) {
	b := bundle(t)
	path := filepath.Join(t.TempDir(), "impact-model.yaml")
	for _, name := range []string{"../evil", "a/b", ".hidden", "UPPER", "a b", strings.Repeat("a", 65), "..", "a..b"} {
		if _, err := Record(path, b, verdict(b, "x", nil), name); !errors.Is(err, ErrJudgeName) {
			t.Errorf("name %q: error = %v", name, err)
		}
	}
	if _, err := Record(path, b, verdict(b, "Model.A_1", nil), ""); err != nil {
		t.Errorf("a judge ID is lowercased into a valid name: %v", err)
	}
}

func TestVerdictsAreFlatReports(t *testing.T) {
	b := bundle(t)
	data, err := json.Marshal(verdict(b, "judge-a", nil))
	if err != nil {
		t.Fatal(err)
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(data, &top); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"bundleHash", "metadata", "categories", "decision"} {
		if _, ok := top[key]; !ok {
			t.Errorf("verdict has no top-level %q", key)
		}
	}
	back, err := ParseVerdict(data)
	if err != nil || back.BundleHash != b.BundleHash || len(back.Categories) != len(b.Rubric.Categories) {
		t.Fatalf("round trip: %v %+v", err, back)
	}
	if _, err := ParseVerdict([]byte(`{"bundleHash":"x","bogusField":1}`)); !errors.Is(err, ErrInvalidReport) {
		t.Errorf("unknown field error = %v", err)
	}
}

func recordAll(t *testing.T, b *Bundle, vs map[string]*Verdict) (string, []string, map[string]*Verdict) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "impact-model.yaml")
	for name, v := range vs {
		if _, err := Record(path, b, v, name); err != nil {
			t.Fatal(err)
		}
	}
	loaded, names, err := Load(path, b)
	if err != nil {
		t.Fatal(err)
	}
	return path, names, loaded
}

func TestSummarize(t *testing.T) {
	b := bundle(t)
	pa, f := rubric.ScorePartial, rubric.ScoreFail
	_, names, vs := recordAll(t, b, map[string]*Verdict{
		"a": verdict(b, "a", nil),
		"b": verdict(b, "b", map[string]rubric.ScoreValue{"evidence": pa}),
		"c": verdict(b, "c", map[string]rubric.ScoreValue{"evidence": f}),
	})
	if strings.Join(names, ",") != "a,b,c" {
		t.Fatalf("names = %v", names)
	}

	s, err := Summarize(b, names, vs, rubric.AggregationMajority)
	if err != nil {
		t.Fatal(err)
	}
	var ev CategorySummary
	for _, c := range s.Categories {
		if c.Category == "evidence" {
			ev = c
		}
	}
	// pass, partial, fail has no majority, so the tie goes to partial.
	if ev.Result != pa || ev.Agreed || ev.Scores["pass"] != 1 || ev.Scores["fail"] != 1 {
		t.Errorf("evidence = %+v", ev)
	}
	if s.Pass || len(s.Disagreements) != 1 || s.Disagreements[0] != "evidence" {
		t.Errorf("pass=%v disagreements=%v reasons=%v", s.Pass, s.Disagreements, s.Reasons)
	}
	want := float64(len(b.Rubric.Categories)-1) / float64(len(b.Rubric.Categories))
	if s.Agreement != want {
		t.Errorf("agreement = %v, want %v", s.Agreement, want)
	}
	if got := ids(s.Categories); got != ids2(b.Rubric.Categories) {
		t.Errorf("categories are not in rubric order:\n%s\n%s", got, ids2(b.Rubric.Categories))
	}

	cons, _ := Summarize(b, names, vs, rubric.AggregationConservative)
	for _, c := range cons.Categories {
		if c.Category == "evidence" && c.Result != f {
			t.Errorf("conservative evidence = %s", c.Result)
		}
	}
	opt, _ := Summarize(b, names, vs, rubric.AggregationOptimistic)
	if !opt.Pass {
		t.Errorf("optimistic aggregation should pass: %v", opt.Reasons)
	}
}

func ids(cs []CategorySummary) string {
	var out []string
	for _, c := range cs {
		out = append(out, c.Category)
	}
	return strings.Join(out, ",")
}

func ids2(cs []rubric.Category) string {
	var out []string
	for _, c := range cs {
		out = append(out, c.ID)
	}
	return strings.Join(out, ",")
}

func TestSummarizeSingleJudgeAndFailures(t *testing.T) {
	b := bundle(t)
	_, names, vs := recordAll(t, b, map[string]*Verdict{"solo": verdict(b, "solo", nil)})
	s, err := Summarize(b, names, vs, rubric.AggregationMajority)
	if err != nil || !s.Pass || s.Agreement != 1 || len(s.Disagreements) != 0 {
		t.Fatalf("single passing judge: %v %+v", err, s)
	}

	blocking := verdict(b, "blocker", map[string]rubric.ScoreValue{"structural_integrity": rubric.ScoreFail})
	_, names, vs = recordAll(t, b, map[string]*Verdict{"blocker": blocking})
	s, _ = Summarize(b, names, vs, rubric.AggregationMajority)
	if s.Pass || !strings.Contains(strings.Join(s.Reasons, ";"), "structural_integrity") {
		t.Errorf("a failed blocking category must fail the impact model: %+v", s.Reasons)
	}

	lots := verdict(b, "finder", nil)
	for i := 0; i < 5; i++ {
		lots.Findings = append(lots.Findings, rubric.Finding{ID: string(rune('a' + i)), Category: "evidence", Severity: rubric.SeverityHigh, Title: "t", Description: "d", Recommendation: "r"})
	}
	_, names, vs = recordAll(t, b, map[string]*Verdict{"finder": lots})
	s, _ = Summarize(b, names, vs, rubric.AggregationMajority)
	if s.Pass {
		t.Error("high findings above the rubric's limit must fail the impact model")
	}

	if _, err := Summarize(b, nil, nil, rubric.AggregationMajority); err == nil {
		t.Error("summarizing no verdicts was accepted")
	}
	if _, err := Summarize(b, names, vs, "average"); err == nil {
		t.Error("an unknown aggregation was accepted")
	}
}

func TestLoadRejectsVerdictsThatNoLongerFit(t *testing.T) {
	b := bundle(t)
	path, _, _ := recordAll(t, b, map[string]*Verdict{"a": verdict(b, "a", nil)})
	file := filepath.Join(Dir(path, b.BundleHash), "a.json")
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	tampered := strings.Replace(string(data), b.BundleHash, "0000", 1)
	if err := os.WriteFile(file, []byte(tampered), 0o600); err != nil { //nolint:gosec // file is under t.TempDir()
		t.Fatal(err)
	}
	if _, _, err := Load(path, b); !errors.Is(err, ErrWrongBundle) {
		t.Errorf("tampered verdict error = %v", err)
	}
	if _, names, err := Load(filepath.Join(t.TempDir(), "other.yaml"), b); err != nil || len(names) != 0 {
		t.Errorf("an empty store should load empty: %v %v", err, names)
	}
}
