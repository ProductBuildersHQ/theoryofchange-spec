package cli

import (
	"strings"
	"testing"

	"github.com/ProductBuildersHQ/impact-model/pkg/approve"
	"github.com/ProductBuildersHQ/impact-model/pkg/impactio"
)

func approveArgs(theory string, extra ...string) []string {
	args := []string{"approve", theory, "--version", "9.0.0", "--date", "2026-04-01", "--approver", "Board chair:chair"}
	return append(args, extra...)
}

func TestApproveRefusesOpenHypotheses(t *testing.T) {
	theory := fixtureCopy(t, "full.json", "impact-model.json")
	before := mustRead(t, theory)
	r := run(t, "", approveArgs(theory)...)
	wantCode(t, r, ExitInvalid)
	if !strings.Contains(r.err, "approval refused") || !strings.Contains(r.err, "hypotheses") {
		t.Errorf("stderr = %q", r.err)
	}
	if mustRead(t, theory) != before {
		t.Error("a refused approval changed the file")
	}
	j := run(t, "", approveArgs(theory, "--format", "json")...)
	wantCode(t, j, ExitInvalid)
	if !strings.Contains(j.stdout, `"code": "hypotheses"`) {
		t.Errorf("json = %q", j.stdout)
	}
}

func TestApproveForced(t *testing.T) {
	theory := fixtureCopy(t, "full.json", "impact-model.json")
	wantCode(t, run(t, "", approveArgs(theory, "--force")...), ExitInvalid)
	wantCode(t, run(t, "", approveArgs(theory, "--force", "--reason", "sign-off needed now", "--dry-run")...), ExitOK)
	wantCode(t, run(t, "", approveArgs(theory, "--force", "--reason", "sign-off needed now")...), ExitOK)
	doc, err := impactio.LoadDocument(theory)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Metadata.Status != approve.StatusApproved || doc.Metadata.Version != "9.0.0" {
		t.Errorf("metadata = %+v", doc.Metadata)
	}
	rev := doc.Revisions[len(doc.Revisions)-1]
	if !strings.Contains(rev.Summary, "sign-off needed now") || rev.Approvers[0].Name != "Board chair" || rev.Approvers[0].Role != "chair" {
		t.Errorf("revision = %+v", rev)
	}
	// The version is now taken.
	wantCode(t, run(t, "", approveArgs(theory, "--force", "--reason", "again")...), ExitInvalid)
}
