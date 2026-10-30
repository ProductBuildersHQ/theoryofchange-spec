package cli

import (
	"os"
	"os/exec"
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

func TestStatusDecks(t *testing.T) {
	dir := t.TempDir()
	prev := dir + "/prev.json"
	wantCode(t, run(t, "", "status", exTheory, "-o", exObs, "--as-of", "2026-06-30", "--out", prev), ExitOK)

	for name, args := range map[string]struct {
		args []string
		want []string
	}{
		"exec": {[]string{"deck", "exec", exTheory, "-o", exObs, "--prev", prev}, []string{"marp: true", "# Where we stand", "# What changed since 2026-06-30", "version 0.2"}},
		"learning-review": {[]string{"deck", "learning-review", exTheory, "-o", exObs, "--prev", prev, "--period", "2026-Q3"},
			[]string{"# Behind plan", "# What we decided", "Trial two daytime sessions"}},
		"story": {[]string{"deck", "story", exTheory, "-o", exObs, "--audience", "aud-funders"}, []string{"# What you told us you need", "# The numbers", "# What we ask of you"}},
	} {
		r := run(t, "", args.args...)
		wantCode(t, r, ExitOK)
		for _, want := range args.want {
			if !strings.Contains(r.stdout, want) {
				t.Errorf("%s: missing %q", name, want)
			}
		}
		if again := run(t, "", args.args...); again.stdout != r.stdout {
			t.Errorf("%s: two runs differ", name)
		}
	}

	// --no-stamp drops the stamp line.
	if r := run(t, "", "deck", "exec", exTheory, "-o", exObs, "--no-stamp"); strings.Contains(r.stdout, "version 0.2") {
		t.Error("--no-stamp left the stamp on")
	}
}

func TestStatusDecksExportNeedsATool(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	for _, kind := range []string{"exec", "learning-review"} {
		r := run(t, "", "deck", kind, exTheory, "-o", exObs, "--export", t.TempDir()+"/d.pdf")
		if r.code != ExitToolMissing || !strings.Contains(r.err, "marp") {
			t.Errorf("%s: code %d, stderr %s", kind, r.code, r.err)
		}
	}
}

func TestStatusDeckErrors(t *testing.T) {
	for name, args := range map[string][]string{
		"story needs an audience": {"deck", "story", exTheory, "-o", exObs},
		"unknown audience":        {"deck", "story", exTheory, "-o", exObs, "--audience", "aud-ghost"},
		"no date":                 {"deck", "exec", exTheory},
		"bad export extension":    {"deck", "exec", exTheory, "-o", exObs, "--export", "x.docx"},
	} {
		if r := run(t, "", args...); r.code != ExitUsageErr {
			t.Errorf("%s: exit %d, want %d (%s)", name, r.code, ExitUsageErr, r.err)
		}
	}
}

// A deck with no diagram needs only marp, so the story deck exports whenever
// marp is installed.
func TestStoryDeckExportsToHTMLWhenMarpIsInstalled(t *testing.T) {
	if _, err := exec.LookPath("marp"); err != nil {
		t.Skip("marp is not installed")
	}
	out := t.TempDir() + "/story.html"
	r := run(t, "", "deck", "story", exTheory, "-o", exObs, "--audience", "aud-funders", "--export", out)
	wantCode(t, r, ExitOK)
	data, err := os.ReadFile(out)
	if err != nil || !strings.Contains(string(data), "What you told us you need") {
		t.Errorf("export: %v %.200s", err, data)
	}
}
