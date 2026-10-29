package cli

import (
	"os"
	"strings"
	"testing"
)

func TestDeckWorkshopPrintsMarp(t *testing.T) {
	path := fixtureCopy(t, "full.json", "theory.json")
	r := run(t, "", "deck", "workshop", path)
	if r.code != ExitOK {
		t.Fatalf("code = %d, stderr = %s", r.code, r.err)
	}
	if !strings.HasPrefix(r.stdout, "---\nmarp: true\n") || !strings.Contains(r.stdout, "# Logic model: ") {
		t.Errorf("unexpected deck:\n%s", r.stdout)
	}
}

func TestDeckWorkshopWritesOutputFile(t *testing.T) {
	path := fixtureCopy(t, "full.json", "theory.json")
	out := path + ".md"
	r := run(t, "", "deck", "workshop", path, "--output", out, "--theme", "gaia")
	if r.code != ExitOK || r.stdout != "" {
		t.Fatalf("code = %d, stdout = %q, stderr = %s", r.code, r.stdout, r.err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "theme: gaia") {
		t.Errorf("theme flag ignored:\n%s", data)
	}
}

func TestDeckExportNeedsATool(t *testing.T) {
	path := fixtureCopy(t, "full.json", "theory.json")
	t.Setenv("PATH", t.TempDir())
	r := run(t, "", "deck", "workshop", path, "--export", path+".pdf")
	if r.code != ExitToolMissing || !strings.Contains(r.err, "marp") {
		t.Errorf("code = %d, stderr = %s; want exit 3 naming marp", r.code, r.err)
	}
}

func TestDeckExportRejectsUnknownExtension(t *testing.T) {
	path := fixtureCopy(t, "full.json", "theory.json")
	r := run(t, "", "deck", "workshop", path, "--export", path+".docx")
	if r.code != ExitUsageErr {
		t.Errorf("code = %d, want %d", r.code, ExitUsageErr)
	}
}

func TestDeckWorkshopMissingFile(t *testing.T) {
	if r := run(t, "", "deck", "workshop", "/nonexistent/theory.yaml"); r.code != ExitUsageErr {
		t.Errorf("code = %d, want %d", r.code, ExitUsageErr)
	}
}
