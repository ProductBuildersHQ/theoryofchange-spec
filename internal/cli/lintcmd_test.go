package cli

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/plexusone/structured-evaluation/summary"
)

func TestLintMinimalIsGoUnderDraft(t *testing.T) {
	path := fixtureCopy(t, "minimal.json", "impact-model.json")
	r := run(t, "", "lint", path, "--profile", "draft")
	wantCode(t, r, ExitOK)
	if !strings.Contains(r.stdout, "GO (draft profile)") {
		t.Errorf("stdout = %q", r.stdout)
	}
}

func TestLintFindsErrorsAndExitsOne(t *testing.T) {
	path := fixtureCopy(t, "full.json", "impact-model.json")
	r := run(t, "", "lint", path)
	wantCode(t, r, ExitInvalid)
	if !strings.Contains(r.stdout, "NO-GO") || !strings.Contains(r.stdout, "IMP-MEA-001 oc-enterprise-require") {
		t.Errorf("stdout = %.400s", r.stdout)
	}
	// Disabling the failing rule makes the impact model GO.
	wantCode(t, run(t, "", "lint", path, "--disable", "IMP-MEA-001"), ExitOK)
}

func TestLintJSONIsASummaryReport(t *testing.T) {
	path := fixtureCopy(t, "full.json", "impact-model.json")
	r := run(t, "", "lint", path, "--format", "json", "--as-of", "2026-03-01")
	wantCode(t, r, ExitInvalid)
	var rep summary.SummaryReport
	if err := json.Unmarshal([]byte(r.stdout), &rep); err != nil {
		t.Fatalf("not a summary report: %v", err)
	}
	if rep.Project != "open-agent-auth" || rep.Status != summary.StatusNoGo || len(rep.Teams) == 0 || rep.GeneratedAt.Format("2006-01-02") != "2026-03-01" {
		t.Errorf("report = %+v", rep)
	}
	again := run(t, "", "lint", path, "--format", "json", "--as-of", "2026-03-01")
	if again.stdout != r.stdout {
		t.Error("two runs differ")
	}
}

func TestLintListAndUsageErrors(t *testing.T) {
	r := run(t, "", "lint", "--list")
	wantCode(t, r, ExitOK)
	for _, want := range []string{"IMP-STR-001  error", "IMP-PRV-001  warning  from full"} {
		if !strings.Contains(r.stdout, want) {
			t.Errorf("rule list is missing %q:\n%s", want, r.stdout)
		}
	}
	path := fixtureCopy(t, "minimal.json", "impact-model.json")
	for name, args := range map[string][]string{
		"no file":      {"lint"},
		"bad profile":  {"lint", path, "--profile", "strict"},
		"unknown rule": {"lint", path, "--disable", "IMP-NOPE-001"},
		"bad date":     {"lint", path, "--as-of", "March"},
		"missing file": {"lint", path + ".none"},
	} {
		if r := run(t, "", args...); r.code != ExitUsageErr {
			t.Errorf("%s: exit %d, want %d (%s)", name, r.code, ExitUsageErr, r.err)
		}
	}
}

func TestLintInvalidDocumentIsExitOne(t *testing.T) {
	path := fixtureCopy(t, "minimal.json", "impact-model.json")
	bad := strings.Replace(mustRead(t, path), `"to": "oc-funding"`, `"to": "oc-ghost"`, 1)
	if err := os.WriteFile(path, []byte(bad), 0o600); err != nil {
		t.Fatal(err)
	}
	wantCode(t, run(t, "", "lint", path), ExitInvalid)
}
