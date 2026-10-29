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
