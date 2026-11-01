package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/plexusone/structured-evaluation/rubric"

	"github.com/ProductBuildersHQ/impact-model/pkg/judge"
)

// bundleOf runs `judge show` and returns the parsed bundle.
func bundleOf(t *testing.T, theory string, extra ...string) *judge.Bundle {
	t.Helper()
	r := run(t, "", append([]string{"judge", "show", theory}, extra...)...)
	wantCode(t, r, ExitOK)
	var b judge.Bundle
	if err := json.Unmarshal([]byte(r.stdout), &b); err != nil {
		t.Fatalf("bundle: %v", err)
	}
	return &b
}

func writeVerdict(t *testing.T, dir string, b *judge.Bundle, judgeID string, def rubric.ScoreValue, hash string) string {
	t.Helper()
	v := &judge.Verdict{BundleHash: hash}
	v.Metadata = rubric.ReportMetadata{Document: "impact-model.json", GeneratedAt: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), ReviewerID: judgeID}
	v.ReviewType = "toc"
	v.Judge = &rubric.JudgeMetadata{JudgeID: judgeID}
	v.Findings = []rubric.Finding{}
	for _, c := range b.Rubric.Categories {
		v.Categories = append(v.Categories, rubric.CategoryResult{Category: c.ID, Score: def, Reasoning: "cited"})
	}
	v.Decision = rubric.Decision{Status: rubric.DecisionPass, Rationale: "ok"}
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, judgeID+".verdict.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestJudgeShowRecordSummary(t *testing.T) {
	theory := fixtureCopy(t, "full.json", "impact-model.json")
	dir := filepath.Dir(theory)
	b := bundleOf(t, theory)
	if b.Profile != "core" || b.BundleHash == "" {
		t.Fatalf("bundle = %+v", b)
	}
	if dev := bundleOf(t, theory, "--profile", "development"); dev.BundleHash == b.BundleHash {
		t.Error("the profile did not change the bundle")
	}

	// A verdict for another version is rejected and stores nothing.
	stale := writeVerdict(t, dir, b, "stale", rubric.ScorePass, "0000")
	r := run(t, "", "judge", "record", theory, stale)
	wantCode(t, r, ExitInvalid)
	if !strings.Contains(r.err, "different version") {
		t.Errorf("stderr = %q", r.err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".toc")); err == nil {
		t.Error("a rejected verdict created the store")
	}

	good := writeVerdict(t, dir, b, "model-a", rubric.ScorePass, b.BundleHash)
	wantCode(t, run(t, "", "judge", "record", theory, good), ExitOK)
	wantCode(t, run(t, "", "judge", "record", theory, writeVerdict(t, dir, b, "model-b", rubric.ScorePass, b.BundleHash)), ExitOK)

	r = run(t, "", "judge", "summary", theory)
	wantCode(t, r, ExitOK)
	for _, want := range []string{"2 judge(s): model-a, model-b", "agreement 100%", "PASS"} {
		if !strings.Contains(r.stdout, want) {
			t.Errorf("summary is missing %q:\n%s", want, r.stdout)
		}
	}

	// A third judge who fails everything tips the majority only if it outvotes.
	wantCode(t, run(t, "", "judge", "record", theory, writeVerdict(t, dir, b, "model-c", rubric.ScoreFail, b.BundleHash)), ExitOK)
	j := run(t, "", "judge", "summary", theory, "--aggregation", "conservative", "--format", "json")
	wantCode(t, j, ExitInvalid)
	var s judge.Summary
	if err := json.Unmarshal([]byte(j.stdout), &s); err != nil || s.Pass || len(s.Disagreements) == 0 {
		t.Errorf("conservative summary = %+v (%v)", s, err)
	}

	// Editing the impact model moves to a new bundle with no verdicts yet.
	wantCode(t, run(t, "", "outcome", "set", theory, "oc-spec-stable", "--title", "Edited after judging"), ExitOK)
	wantCode(t, run(t, "", "judge", "summary", theory), ExitUsageErr)
}

func TestJudgeUsageErrors(t *testing.T) {
	theory := fixtureCopy(t, "full.json", "impact-model.json")
	b := bundleOf(t, theory)
	good := writeVerdict(t, filepath.Dir(theory), b, "m", rubric.ScorePass, b.BundleHash)
	for name, args := range map[string][]string{
		"unknown profile":     {"judge", "show", theory, "--profile", "military"},
		"unsafe judge name":   {"judge", "record", theory, good, "--judge", "../evil"},
		"missing verdict":     {"judge", "record", theory, theory + ".none"},
		"unknown aggregation": {"judge", "summary", theory, "--aggregation", "average"},
	} {
		if r := run(t, "", args...); r.code != ExitUsageErr {
			t.Errorf("%s: exit %d, want %d (%s%s)", name, r.code, ExitUsageErr, r.stdout, r.err)
		}
	}
}
