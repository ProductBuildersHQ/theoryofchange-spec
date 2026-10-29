package cli

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/edit"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

// answeredSheet exports the review sheet of a copy of the full fixture and
// answers rows: for each element ID, the text that replaces the empty
// decision, suggested change, and comment columns.
func answeredSheet(t *testing.T, theory string, answers map[string]string) string {
	t.Helper()
	r := run(t, "", "export", "review-sheet", theory)
	wantCode(t, r, ExitOK)
	lines := strings.Split(r.stdout, "\r\n")
	for i, line := range lines {
		for id, answer := range answers {
			if strings.Contains(line, ","+id+",") && strings.HasSuffix(line, ",,,") {
				lines[i] = strings.TrimSuffix(line, ",,,") + answer
			}
		}
	}
	path := theory + ".sheet.csv"
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\r\n")), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestImportReviewFromSheet(t *testing.T) {
	theory := fixtureCopy(t, "full.json", "theory.json")
	before, err := os.ReadFile(theory)
	if err != nil {
		t.Fatal(err)
	}
	sheet := answeredSheet(t, theory, map[string]string{"oc-spec-stable": ",agree,,"})
	patchOut, consultOut := theory+".patch.json", theory+".consultation.json"

	r := run(t, "", "import", "review", theory, "--from-csv", sheet, "--reviewer", "program staff", "--date", "2026-03-01",
		"--out", patchOut, "--consultation-out", consultOut)
	wantCode(t, r, ExitOK)
	if !strings.Contains(r.stdout, "1 row(s): 1 became changes, 0 need a person") {
		t.Errorf("summary = %q", r.stdout)
	}

	after, err := os.ReadFile(theory)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Error("the theory was modified")
	}
	data, err := os.ReadFile(patchOut)
	if err != nil {
		t.Fatal(err)
	}
	var patch edit.Patch
	if err := json.Unmarshal(data, &patch); err != nil || len(patch) == 0 {
		t.Fatalf("patch %q: %v", data, err)
	}
	var c toc.Consultation
	data, err = os.ReadFile(consultOut)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &c); err != nil || c.Findings[0] != "agree: 1" {
		t.Fatalf("consultation %q: %v", data, err)
	}

	// The proposed patch applies with the ordinary patch command.
	wantCode(t, run(t, "", "patch", theory, patchOut), ExitOK)
}

func TestImportReviewFlagsRowsAndStaleness(t *testing.T) {
	theory := fixtureCopy(t, "full.json", "theory.json")
	sheet := answeredSheet(t, theory, map[string]string{"oc-spec-stable": ",,,Why this?"})
	r := run(t, "", "import", "review", theory, "--from-csv", sheet, "--reviewer", "program staff", "--date", "2026-03-01")
	wantCode(t, r, ExitInvalid)
	if !strings.Contains(r.stdout, "row 1 oc-spec-stable: a comment with no decision") {
		t.Errorf("stdout = %q", r.stdout)
	}

	// Changing the theory after export makes the sheet's hash stale.
	wantCode(t, run(t, "", "outcome", "set", theory, "oc-spec-stable", "--set", "title=Changed after export"), ExitOK)
	sheet = answeredSheet(t, theory, nil)
	stale := strings.Replace(mustRead(t, sheet), "hash:", "hash:0", 1)
	if err := os.WriteFile(sheet, []byte(stale), 0o600); err != nil {
		t.Fatal(err)
	}
	r = run(t, "", "import", "review", theory, "--from-csv", sheet, "--reviewer", "program staff", "--date", "2026-03-01", "--format", "json")
	wantCode(t, r, ExitInvalid)
	var out struct{ Stale bool }
	if err := json.Unmarshal([]byte(r.stdout), &out); err != nil || !out.Stale {
		t.Errorf("stdout = %q (%v)", r.stdout, err)
	}
}

func TestImportReviewFromJSON(t *testing.T) {
	theory := fixtureCopy(t, "full.json", "theory.json")
	resp := run(t, "", "export", "review-sheet", theory)
	wantCode(t, resp, ExitOK)
	hash := strings.TrimSpace(strings.SplitN(strings.SplitN(resp.stdout, "hash:", 2)[1], "\r\n", 2)[0])
	body := `{"kind":"ReviewResponse","specVersion":"0.1","tocId":"open-agent-auth","contentHash":"` + hash + `",
	"reviewers":["program staff"],"method":"workshop","date":"2026-03-01",
	"rows":[{"elementId":"oc-spec-stable","decision":"remove","origin":"table"}]}`
	path := theory + ".response.json"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	wantCode(t, run(t, "", "import", "review", theory, path), ExitOK)
}

func TestImportReviewUsageErrors(t *testing.T) {
	theory := fixtureCopy(t, "full.json", "theory.json")
	for name, args := range map[string][]string{
		"no source":           {"import", "review", theory},
		"csv without who":     {"import", "review", theory, "--from-csv", theory},
		"both sources":        {"import", "review", theory, theory, "--from-csv", theory, "--reviewer", "x", "--date", "2026-03-01"},
		"missing response":    {"import", "review", theory, theory + ".none.json"},
		"not a response file": {"import", "review", theory, theory},
		"bad step":            {"import", "review", theory, theory, "--step", "nonsense"},
	} {
		if r := run(t, "", args...); r.code != ExitUsageErr {
			t.Errorf("%s: exit %d, want %d (%s)", name, r.code, ExitUsageErr, r.err)
		}
	}
}
