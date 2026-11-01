package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ProductBuildersHQ/impact-model/pkg/diff"
	"github.com/ProductBuildersHQ/impact-model/pkg/impactio"
)

// twoVersions copies the full fixture twice and edits the second with the
// patch engine, as an author would.
func twoVersions(t *testing.T) (oldPath, newPath string) {
	t.Helper()
	oldPath = fixtureCopy(t, "full.json", "v1.json")
	data, err := os.ReadFile(oldPath)
	if err != nil {
		t.Fatal(err)
	}
	newPath = filepath.Join(filepath.Dir(oldPath), "v2.json")
	if err := os.WriteFile(newPath, data, 0o600); err != nil { //nolint:gosec // a file in the test's temporary directory
		t.Fatal(err)
	}
	wantCode(t, run(t, "", "outcome", "set", newPath, "oc-spec-stable", "--title", "A stable specification exists"), ExitOK)
	wantCode(t, run(t, "", "outcome", "add", newPath, "--id", "oc-extra", "--title", "Extra", "--level", "output"), ExitOK)
	return oldPath, newPath
}

func TestDiffMarkdownAndJSON(t *testing.T) {
	oldPath, newPath := twoVersions(t)
	r := run(t, "", "diff", oldPath, newPath)
	wantCode(t, r, ExitOK)
	for _, want := range []string{"1 element(s) added, 1 changed, 0 removed.", "**changed** `oc-spec-stable`", "**added** `oc-extra`", "`title`"} {
		if !strings.Contains(r.stdout, want) {
			t.Errorf("missing %q:\n%s", want, r.stdout)
		}
	}
	j := run(t, "", "diff", oldPath, newPath, "--format", "json")
	wantCode(t, j, ExitOK)
	var res diff.Result
	if err := json.Unmarshal([]byte(j.stdout), &res); err != nil || len(res.Changes) != 2 {
		t.Errorf("json: %v %+v", err, res)
	}
	same := run(t, "", "diff", oldPath, oldPath)
	wantCode(t, same, ExitOK)
	if !strings.Contains(same.stdout, "Nothing changed.") {
		t.Errorf("same = %s", same.stdout)
	}
}

func TestDiffRevisionPatchAppliesWithThePatchCommand(t *testing.T) {
	oldPath, newPath := twoVersions(t)
	patchFile := filepath.Join(filepath.Dir(newPath), "revision.patch.json")
	r := run(t, "", "diff", oldPath, newPath, "--revision", "--version", "0.9", "--date", "2026-03-01", "--trigger", "review", "-o", patchFile)
	wantCode(t, r, ExitOK)
	wantCode(t, run(t, "", "patch", newPath, patchFile), ExitOK)
	doc, err := impactio.LoadDocument(newPath)
	if err != nil {
		t.Fatal(err)
	}
	last := doc.Revisions[len(doc.Revisions)-1]
	if last.Version != "0.9" || last.Trigger != "review" || len(last.ChangedIDs) != 2 {
		t.Errorf("revision = %+v", last)
	}
}

func TestDiffErrors(t *testing.T) {
	oldPath, newPath := twoVersions(t)
	for name, args := range map[string][]string{
		"revision needs a date":    {"diff", oldPath, newPath, "--revision", "--version", "1"},
		"revision needs a version": {"diff", oldPath, newPath, "--revision", "--date", "2026-03-01"},
		"bad trigger":              {"diff", oldPath, newPath, "--revision", "--version", "1", "--date", "2026-03-01", "--trigger", "whim"},
		"missing file":             {"diff", oldPath, newPath + ".none"},
	} {
		if r := run(t, "", args...); r.code != ExitUsageErr {
			t.Errorf("%s: exit %d, want %d (%s)", name, r.code, ExitUsageErr, r.err)
		}
	}
}
