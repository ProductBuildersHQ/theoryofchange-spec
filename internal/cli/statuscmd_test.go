package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/tocio"
)

const (
	exTheory = "../../examples/intermediary/theory.yaml"
	exObs    = "../../examples/intermediary/observations-2026-q3.yaml"
	exMap    = "../../examples/intermediary/csv/mapping.yaml"
	exSheet  = "../../examples/intermediary/csv/monthly-q4.csv"
)

func TestStatusSummaryAndJSON(t *testing.T) {
	r := run(t, "", "status", exTheory, "-o", exObs)
	wantCode(t, r, ExitOK)
	for _, want := range []string{"as of 2026-09-30", "indicators (", "impact oc-resident-access:", "confidence"} {
		if !strings.Contains(r.stdout, want) {
			t.Errorf("summary is missing %q:\n%s", want, r.stdout)
		}
	}
	j := run(t, "", "status", exTheory, "-o", exObs, "--format", "json", "--as-of", "2026-09-30")
	wantCode(t, j, ExitOK)
	if _, err := tocio.ParseStatus([]byte(j.stdout), tocio.FormatJSON); err != nil {
		t.Fatalf("the JSON is not a status snapshot: %v", err)
	}
	if again := run(t, "", "status", exTheory, "-o", exObs, "--format", "json", "--as-of", "2026-09-30"); again.stdout != j.stdout {
		t.Error("two runs differ")
	}
}

func TestStatusAgainstAPreviousSnapshot(t *testing.T) {
	dir := t.TempDir()
	prev := filepath.Join(dir, "prev.json")
	wantCode(t, run(t, "", "status", exTheory, "-o", exObs, "--as-of", "2026-06-30", "--out", prev), ExitOK)
	if _, err := os.Stat(prev); err != nil {
		t.Fatalf("--out wrote nothing: %v", err)
	}
	r := run(t, "", "status", exTheory, "-o", exObs, "--as-of", "2026-09-30", "--prev", prev)
	wantCode(t, r, ExitOK)
	if !strings.Contains(r.stdout, "changes since 2026-06-30:") {
		t.Errorf("stdout = %s", r.stdout)
	}
	j := run(t, "", "status", exTheory, "-o", exObs, "--as-of", "2026-09-30", "--prev", prev, "--format", "json")
	var both struct {
		Status toc.Status
		Diff   toc.StatusDiff
	}
	if err := json.Unmarshal([]byte(j.stdout), &both); err != nil || both.Diff.FromAsOf != "2026-06-30" || both.Status.AsOf != "2026-09-30" {
		t.Errorf("json = %v %+v", err, both.Diff)
	}
}

func TestStatusErrors(t *testing.T) {
	for name, args := range map[string][]string{
		"no date available":  {"status", exTheory},
		"bad date":           {"status", exTheory, "--as-of", "soon"},
		"bad rollup":         {"status", exTheory, "--as-of", "2026-09-30", "--rollup", "median"},
		"bad tolerance":      {"status", exTheory, "--as-of", "2026-09-30", "--tolerance", "2"},
		"missing obs file":   {"status", exTheory, "-o", exObs + ".none"},
		"missing prev":       {"status", exTheory, "-o", exObs, "--prev", exObs + ".none"},
		"prev is not status": {"status", exTheory, "-o", exObs, "--prev", exTheory},
	} {
		if r := run(t, "", args...); r.code != ExitUsageErr {
			t.Errorf("%s: exit %d, want %d (%s)", name, r.code, ExitUsageErr, r.err)
		}
	}
}

func TestStatusRejectsInvalidObservations(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "bad.yaml")
	body := "kind: ObservationSet\nspecVersion: \"0.1\"\ntocId: riverbend-network\nobservations:\n  - indicatorId: ind-ghost\n    date: \"2026-09-01\"\n    value: 1\n"
	if err := os.WriteFile(bad, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	r := run(t, "", "status", exTheory, "-o", bad)
	wantCode(t, r, ExitInvalid)
	if !strings.Contains(r.err, "ind-ghost") {
		t.Errorf("stderr = %q", r.err)
	}
}

func TestImportObservationsThenStatus(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "obs-q4.yaml")
	r := run(t, "", "import", "observations", "--csv", exSheet, "--map", exMap, "--toc", exTheory, "--out", out)
	wantCode(t, r, ExitOK)
	if !strings.Contains(r.err, "5 row(s): 5 imported, 0 unmatched") {
		t.Errorf("stderr = %q", r.err)
	}
	wantCode(t, run(t, "", "validate", out, "--toc", exTheory), ExitOK)
	wantCode(t, run(t, "", "status", exTheory, "-o", exObs, "-o", out), ExitOK)
}

func TestImportObservationsReportsUnmatchedRows(t *testing.T) {
	dir := t.TempDir()
	sheet := filepath.Join(dir, "sheet.csv")
	body := "Measure,Month end,Value,Program,Quality,Source\r\nAttendee hours,31/10/2026,430,Learning Exchange,verified,\r\nNo such measure,31/10/2026,1,,,\r\nAttendee hours,not a date,2,Civic Voice,,\r\n"
	if err := os.WriteFile(sheet, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	out, unmatched := filepath.Join(dir, "obs.json"), filepath.Join(dir, "unmatched.csv")
	r := run(t, "", "import", "observations", "--csv", sheet, "--map", exMap, "--toc", exTheory, "--out", out, "--unmatched", unmatched)
	wantCode(t, r, ExitInvalid)
	for _, want := range []string{"3 row(s): 1 imported, 2 unmatched", "row 3: unknown indicator", "row 4: the date"} {
		if !strings.Contains(r.err, want) {
			t.Errorf("stderr is missing %q:\n%s", want, r.err)
		}
	}
	if _, err := os.Stat(out); err != nil {
		t.Error("the matched rows were not written")
	}
	sheetOut := mustRead(t, unmatched)
	if !strings.HasPrefix(sheetOut, "Measure,Month end,Value,Program,Quality,Source,row,reason") || !strings.Contains(sheetOut, "No such measure") {
		t.Errorf("unmatched sheet = %q", sheetOut)
	}
}

func TestImportObservationsUsageErrors(t *testing.T) {
	for name, args := range map[string][]string{
		"no flags":        {"import", "observations"},
		"missing sheet":   {"import", "observations", "--csv", exSheet + ".none", "--map", exMap, "--toc", exTheory},
		"missing map":     {"import", "observations", "--csv", exSheet, "--map", exMap + ".none", "--toc", exTheory},
		"map is a theory": {"import", "observations", "--csv", exSheet, "--map", exTheory, "--toc", exTheory},
	} {
		if r := run(t, "", args...); r.code != ExitUsageErr {
			t.Errorf("%s: exit %d, want %d (%s)", name, r.code, ExitUsageErr, r.err)
		}
	}
}
