package cli

import (
	"archive/zip"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/report"
)

func TestReportKinds(t *testing.T) {
	path := fixtureCopy(t, "full.json", "theory.json")
	for _, kind := range report.DefinitionKinds {
		r := run(t, "", "report", kind, path)
		if r.code != ExitOK || r.stdout == "" {
			t.Errorf("report %s: code %d, stderr %s", kind, r.code, r.err)
		}
	}
}

func TestReportDefinitionsHasEveryPart(t *testing.T) {
	path := fixtureCopy(t, "full.json", "theory.json")
	r := run(t, "", "report", "definitions", path)
	for _, want := range []string{"## Causal pathways", "## Logic model:", "## Metric definitions", "## Glossary"} {
		if !strings.Contains(r.stdout, want) {
			t.Errorf("missing %q", want)
		}
	}
}

func TestReportPathwayRestrictsLogicModel(t *testing.T) {
	path := fixtureCopy(t, "full.json", "theory.json")
	r := run(t, "", "report", "logic-model", path, "--pathway", "pw-supply")
	if r.code != ExitOK || strings.Count(r.stdout, "## Logic model:") != 1 {
		t.Errorf("code %d, output:\n%s%s", r.code, r.stdout, r.err)
	}
}

func TestReportUnknownKindAndPathway(t *testing.T) {
	path := fixtureCopy(t, "full.json", "theory.json")
	if r := run(t, "", "report", "nope", path); r.code != ExitUsageErr || !strings.Contains(r.err, "glossary") {
		t.Errorf("unknown kind: code %d, stderr %s", r.code, r.err)
	}
	if r := run(t, "", "report", "glossary", path, "--pathway", "pw-missing"); r.code != ExitUsageErr {
		t.Errorf("unknown pathway: code %d, stderr %s", r.code, r.err)
	}
}

func TestExportCSVToFile(t *testing.T) {
	path := fixtureCopy(t, "full.json", "theory.json")
	out := path + ".csv"
	r := run(t, "", "export", "csv", "metrics", path, "--out", out, "--no-stamp")
	if r.code != ExitOK || r.stdout != "" {
		t.Fatalf("code %d, stdout %q, stderr %s", r.code, r.stdout, r.err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(data), "id,name,definition,formula") {
		t.Errorf("unexpected CSV:\n%s", data)
	}
	if r := run(t, "", "export", "csv", "nope", path); r.code != ExitUsageErr {
		t.Errorf("unknown table: code %d", r.code)
	}
}

func TestExportsCarryTheStamp(t *testing.T) {
	path := fixtureCopy(t, "full.json", "theory.json")
	md := run(t, "", "report", "glossary", path, "--as-of", "2026-09-30")
	for _, want := range []string{"---\ntitle: ", "version: ", `date: "2026-09-30"`, "content-hash: "} {
		if !strings.Contains(md.stdout, want) {
			t.Errorf("report is missing %q:\n%.300s", want, md.stdout)
		}
	}
	csv := run(t, "", "export", "csv", "glossary", path)
	if !strings.HasPrefix(csv.stdout, "# toc:open-agent-auth version:") || !strings.Contains(csv.stdout, "\r\nid,term,") {
		t.Errorf("CSV should start with the stamp row:\n%.200s", csv.stdout)
	}
	deck := run(t, "", "deck", "workshop", path)
	if !strings.Contains(deck.stdout, "_Theory of Change open-agent-auth · version ") {
		t.Errorf("deck title slide is missing the stamp:\n%.600s", deck.stdout)
	}
	bare := run(t, "", "report", "glossary", path, "--no-stamp")
	if strings.Contains(bare.stdout, "content-hash") {
		t.Error("--no-stamp should leave the stamp off")
	}
}

func TestStampHashMatchesAcrossFormats(t *testing.T) {
	path := fixtureCopy(t, "full.json", "theory.json")
	md := run(t, "", "report", "glossary", path).stdout
	csv := run(t, "", "export", "csv", "glossary", path).stdout
	hash := md[strings.Index(md, "content-hash: \"")+len("content-hash: \""):]
	hash = hash[:strings.Index(hash, "\"")]
	if len(hash) != 64 || !strings.Contains(csv, "hash:"+hash) {
		t.Errorf("hash %q differs between formats; CSV row: %.120s", hash, csv)
	}
}

func TestExportDocRejectsMismatchedKinds(t *testing.T) {
	path := fixtureCopy(t, "full.json", "theory.json")
	dir := filepath.Dir(path)
	for _, args := range [][]string{
		{"export", "docx", "workshop", path, "--out", filepath.Join(dir, "a.docx")},
		{"export", "pptx", "glossary", path, "--out", filepath.Join(dir, "a.pptx")},
		{"export", "docx", "glossary", path, "--out", filepath.Join(dir, "a.pdf")},
		{"export", "docx", "nope", path, "--out", filepath.Join(dir, "a.docx")},
	} {
		if r := run(t, "", args...); r.code != ExitUsageErr {
			t.Errorf("%v: code %d, want %d (stderr %s)", args, r.code, ExitUsageErr, r.err)
		}
	}
}

func TestExportDocNeedsPandoc(t *testing.T) {
	path := fixtureCopy(t, "full.json", "theory.json")
	t.Setenv("PATH", t.TempDir())
	r := run(t, "", "export", "docx", "glossary", path, "--out", filepath.Join(filepath.Dir(path), "a.docx"))
	if r.code != ExitToolMissing || !strings.Contains(r.err, "pandoc") {
		t.Errorf("code %d, stderr %s; want exit 3 naming pandoc", r.code, r.err)
	}
}

func TestExportDocNeedsMmdcForDiagrams(t *testing.T) {
	if _, err := exec.LookPath("pandoc"); err != nil {
		t.Skip("pandoc is not installed")
	}
	if _, err := exec.LookPath("mmdc"); err == nil {
		t.Skip("mmdc is installed; this test needs it absent")
	}
	path := fixtureCopy(t, "full.json", "theory.json")
	out := filepath.Join(filepath.Dir(path), "a.docx")
	r := run(t, "", "export", "docx", "pathway", path, "--out", out)
	if r.code != ExitToolMissing || !strings.Contains(r.err, "mmdc") {
		t.Errorf("code %d, stderr %s; want exit 3 naming mmdc", r.code, r.err)
	}
	if _, err := os.Stat(out); err == nil {
		t.Error("nothing should be written when a tool is missing")
	}
}

func TestExportDocxEndToEnd(t *testing.T) {
	if _, err := exec.LookPath("pandoc"); err != nil {
		t.Skip("pandoc is not installed")
	}
	path := fixtureCopy(t, "full.json", "theory.json")
	out := filepath.Join(filepath.Dir(path), "glossary.docx")
	r := run(t, "", "export", "docx", "glossary", path, "--out", out, "--as-of", "2026-09-30")
	if r.code != ExitOK {
		t.Fatalf("code %d, stderr %s", r.code, r.err)
	}
	zr, err := zip.OpenReader(out)
	if err != nil {
		t.Fatalf("the output is not a DOCX: %v", err)
	}
	defer func() {
		if err := zr.Close(); err != nil {
			t.Errorf("close: %v", err)
		}
	}()
	var body string
	for _, f := range zr.File {
		if f.Name != "word/document.xml" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(rc)
		if err != nil {
			t.Fatal(err)
		}
		if err := rc.Close(); err != nil {
			t.Fatal(err)
		}
		body = string(data)
	}
	for _, want := range []string{"General availability", "Theory of Change open-agent-auth", "as of 2026-09-30"} {
		if !strings.Contains(body, want) {
			t.Errorf("the DOCX is missing %q", want)
		}
	}
}

func TestReviewSheetIsStampedAndEmptyResponses(t *testing.T) {
	path := fixtureCopy(t, "full.json", "theory.json")
	r := run(t, "", "export", "review-sheet", path)
	if r.code != ExitOK {
		t.Fatalf("code %d, stderr %s", r.code, r.err)
	}
	if !strings.HasPrefix(r.stdout, "# toc:open-agent-auth version:") ||
		!strings.Contains(r.stdout, "\r\nkind,id,statement,basis,sources,decision,suggested_change,comment\r\n") {
		t.Errorf("unexpected sheet:\n%.300s", r.stdout)
	}
	bare := run(t, "", "export", "review-sheet", path, "--no-stamp")
	if !strings.HasPrefix(bare.stdout, "kind,id,") {
		t.Errorf("--no-stamp sheet should start with the header:\n%.100s", bare.stdout)
	}
}

func TestReviewPacketMarkdownToFile(t *testing.T) {
	path := fixtureCopy(t, "full.json", "theory.json")
	out := path + ".md"
	if r := run(t, "", "export", "review-packet", path, "--out", out); r.code != ExitOK {
		t.Fatalf("code %d, stderr %s", r.code, r.err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"content-hash: ", "## Outcomes", "| ID | Statement | Basis | Sources | Decision | Suggested change | Comment |", "| oc-spec-stable |"} {
		if !strings.Contains(string(data), want) {
			t.Errorf("packet is missing %q", want)
		}
	}
}

func TestReviewPacketFormatChecks(t *testing.T) {
	path := fixtureCopy(t, "full.json", "theory.json")
	if r := run(t, "", "export", "review-packet", path, "--format", "pdf"); r.code != ExitUsageErr {
		t.Errorf("bad format: code %d", r.code)
	}
	if r := run(t, "", "export", "review-packet", path, "--format", "docx"); r.code != ExitUsageErr {
		t.Errorf("docx without --out: code %d", r.code)
	}
	t.Setenv("PATH", t.TempDir())
	r := run(t, "", "export", "review-packet", path, "--format", "docx", "--out", filepath.Join(filepath.Dir(path), "p.docx"))
	if r.code != ExitToolMissing || !strings.Contains(r.err, "pandoc") {
		t.Errorf("code %d, stderr %s; want exit 3 naming pandoc", r.code, r.err)
	}
}

func TestReviewPacketDocx(t *testing.T) {
	if _, err := exec.LookPath("pandoc"); err != nil {
		t.Skip("pandoc is not installed")
	}
	path := fixtureCopy(t, "full.json", "theory.json")
	out := filepath.Join(filepath.Dir(path), "packet.docx")
	if r := run(t, "", "export", "review-packet", path, "--format", "docx", "--out", out); r.code != ExitOK {
		t.Fatalf("code %d, stderr %s", r.code, r.err)
	}
	zr, err := zip.OpenReader(out)
	if err != nil {
		t.Fatalf("not a DOCX: %v", err)
	}
	if err := zr.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestSelectionAndCollectionReportsAndTables(t *testing.T) {
	path := fixtureCopy(t, "full.json", "theory.json")
	for _, kind := range []string{"metric-selection", "collection-plan"} {
		r := run(t, "", "report", kind, path)
		wantCode(t, r, ExitOK)
		if !strings.Contains(r.stdout, "content-hash: ") || !strings.Contains(r.stdout, "## ") {
			t.Errorf("report %s = %.200s", kind, r.stdout)
		}
		c := run(t, "", "export", "csv", kind, path)
		wantCode(t, c, ExitOK)
		if !strings.HasPrefix(c.stdout, "# toc:open-agent-auth") || !strings.Contains(c.stdout, "\r\nid,name,") {
			t.Errorf("csv %s = %.200s", kind, c.stdout)
		}
	}
	if r := run(t, "", "report", "collection-plan", path); !strings.Contains(r.stdout, "Against the budget") {
		t.Errorf("budget comparison missing:\n%s", r.stdout)
	}
}

func TestProjectionReports(t *testing.T) {
	path := fixtureCopy(t, "full.json", "theory.json")
	for kind, want := range map[string]string{
		"outcome-map": "## Outcome map:", "logframe": "## Logframe:",
	} {
		r := run(t, "", "report", kind, path)
		wantCode(t, r, ExitOK)
		if !strings.Contains(r.stdout, want) {
			t.Errorf("%s: %.200s", kind, r.stdout)
		}
	}
	r := run(t, "", "report", "logic-model", path, "--diagram")
	wantCode(t, r, ExitOK)
	if !strings.Contains(r.stdout, "flowchart LR") {
		t.Errorf("diagram = %.200s", r.stdout)
	}
	wantCode(t, run(t, "", "report", "glossary", path, "--diagram"), ExitUsageErr)
	if r := run(t, "", "report", "definitions", path); !strings.Contains(r.stdout, "## Outcome map:") || !strings.Contains(r.stdout, "## Logframe:") {
		t.Error("the definitions pack should include the projections")
	}
}

func TestStatusReportKinds(t *testing.T) {
	want := map[string]string{
		"exec": "executive summary", "results": ": results", "assumptions": ": assumptions", "interventions": ": interventions",
		"evidence-gaps": "evidence gaps", "full": "## Pathway", "portfolio": ": portfolio", "narrative": "programme narrative", "actors": "who does what",
	}
	for kind, text := range want {
		r := run(t, "", "report", kind, exTheory, "-o", exObs)
		wantCode(t, r, ExitOK)
		if !strings.Contains(r.stdout, text) {
			t.Errorf("%s: missing %q in %.200s", kind, text, r.stdout)
		}
		if !strings.Contains(r.stdout, "content-hash: ") || !strings.Contains(r.stdout, `date: "2026-09-30"`) {
			t.Errorf("%s: the stamp should carry the status date:\n%.300s", kind, r.stdout)
		}
	}
	r := run(t, "", "report", "story", exTheory, "-o", exObs, "--audience", "aud-funders")
	wantCode(t, r, ExitOK)
	if !strings.Contains(r.stdout, "story for Funders") || strings.Contains(r.stdout, "QUOTE-") {
		t.Errorf("story = %.300s", r.stdout)
	}
	r = run(t, "", "report", "learning-review", exTheory, "-o", exObs, "--period", "2026-Q3")
	wantCode(t, r, ExitOK)
	if !strings.Contains(r.stdout, "learning review 2026-Q3") || !strings.Contains(r.stdout, "Trial two daytime sessions") {
		t.Errorf("learning review = %.300s", r.stdout)
	}
}

func TestNarrativeAndActorsNeedNoObservations(t *testing.T) {
	for _, kind := range []string{"narrative", "actors"} {
		r := run(t, "", "report", kind, exTheory)
		wantCode(t, r, ExitOK)
		if strings.Contains(r.stdout, "Status as of") {
			t.Errorf("%s claims a status date with no observations", kind)
		}
	}
}

func TestCustomReport(t *testing.T) {
	dir := t.TempDir()
	tmpl := dir + "/board.tmpl"
	if err := os.WriteFile(tmpl, []byte("# {{.Document.Title}}\n{{range .Outcomes}}{{if .Status}}- {{.Title}}: {{label .Status.PathwayState}}\n{{end}}{{end}}"), 0o600); err != nil {
		t.Fatal(err)
	}
	r := run(t, "", "report", "custom", exTheory, "--template", tmpl, "-o", exObs, "--no-stamp")
	wantCode(t, r, ExitOK)
	if !strings.HasPrefix(r.stdout, "# Riverbend") || !strings.Contains(r.stdout, "- Residents get timely") {
		t.Errorf("custom = %.200s", r.stdout)
	}
	bad := dir + "/bad.tmpl"
	if err := os.WriteFile(bad, []byte(`{{readFile "/etc/passwd"}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	wantCode(t, run(t, "", "report", "custom", exTheory, "--template", bad), ExitUsageErr)
}

func TestStatusReportErrors(t *testing.T) {
	for name, args := range map[string][]string{
		"no date":           {"report", "exec", exTheory},
		"story no audience": {"report", "story", exTheory, "-o", exObs},
		"bad audience":      {"report", "story", exTheory, "-o", exObs, "--audience", "aud-ghost"},
		"custom no file":    {"report", "custom", exTheory},
		"missing template":  {"report", "custom", exTheory, "--template", exTheory + ".none"},
		"unknown kind":      {"report", "mystery", exTheory},
	} {
		if r := run(t, "", args...); r.code != ExitUsageErr {
			t.Errorf("%s: exit %d, want %d (%s)", name, r.code, ExitUsageErr, r.err)
		}
	}
}
