package cli

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/grokify/structured-goals/okr"
)

func TestExportOKRAndJourneys(t *testing.T) {
	out := filepath.Join(t.TempDir(), "okr.json")
	r := run(t, "", "export", "okr", exTheory, "-O", exObs, "-o", out)
	wantCode(t, r, ExitOK)
	doc, err := okr.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if errs := okr.Errors(doc.Validate(okr.DefaultValidationOptions())); len(errs) != 0 {
		t.Errorf("exported OKR does not validate: %+v", errs)
	}
	if doc.Metadata.Period != "2026-09-30" {
		t.Errorf("period = %q", doc.Metadata.Period)
	}

	// Without observations the key results have no current value.
	bare := run(t, "", "export", "okr", exTheory)
	wantCode(t, bare, ExitOK)
	if strings.Contains(bare.stdout, `"current"`) {
		t.Error("an export with no observations has current values")
	}

	j := run(t, "", "export", "journey", exTheory, "-O", exObs)
	wantCode(t, j, ExitOK)
	var journeys []map[string]any
	if err := json.Unmarshal([]byte(j.stdout), &journeys); err != nil || len(journeys) == 0 {
		t.Errorf("journeys: %v %.200s", err, j.stdout)
	}
	if again := run(t, "", "export", "journey", exTheory, "-O", exObs); again.stdout != j.stdout {
		t.Error("two exports differ")
	}
	wantCode(t, run(t, "", "export", "okr", exTheory+".none"), ExitUsageErr)
}
