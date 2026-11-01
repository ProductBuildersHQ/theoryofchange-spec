package skills_test

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/ProductBuildersHQ/impact-model/internal/cli"
)

const (
	theory  = "../examples/intermediary/impact-model.yaml"
	obs     = "../examples/intermediary/observations-2026-q3.yaml"
	mapping = "../examples/intermediary/csv/mapping.yaml"
	sheet   = "../examples/intermediary/csv/monthly-q4.csv"
)

func skillDirs(t *testing.T) []string {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	var dirs []string
	for _, e := range entries {
		if _, err := os.Stat(filepath.Join(e.Name(), "SKILL.md")); err == nil {
			dirs = append(dirs, e.Name())
		}
	}
	return dirs
}

func TestEverySkillIsWellFormed(t *testing.T) {
	want := []string{"impact-audience-story", "impact-author", "impact-draft-from-sources", "impact-learning-review", "impact-status-refresh"}
	got := skillDirs(t)
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("skills = %v, want %v", got, want)
	}
	front := regexp.MustCompile(`(?s)\A---\nname: ([^\n]+)\ndescription: ([^\n]+)\n---\n`)
	for _, dir := range got {
		data, err := os.ReadFile(filepath.Join(dir, "SKILL.md"))
		if err != nil {
			t.Fatal(err)
		}
		m := front.FindSubmatch(data)
		if m == nil {
			t.Errorf("%s: no front matter with a name and a description", dir)
			continue
		}
		if string(m[1]) != dir {
			t.Errorf("%s: the name is %q", dir, m[1])
		}
		if !strings.Contains(string(m[2]), "Use when") {
			t.Errorf("%s: the description should say when to use the skill", dir)
		}
		for _, heading := range []string{"## Purpose", "## Rules", "## Command sequence", "## Checks to report", "## Outputs"} {
			if !strings.Contains(string(data), heading) {
				t.Errorf("%s: missing %q", dir, heading)
			}
		}
	}
}

// commandLine joins a fenced command that continues over lines with a
// trailing backslash.
var fence = regexp.MustCompile("(?s)```bash\n(.*?)```")

// commands extracts every `impact-model ...` command from a skill.
func commands(t *testing.T, skill string) [][]string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(skill, "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	var out [][]string
	for _, block := range fence.FindAllStringSubmatch(string(data), -1) {
		joined := strings.ReplaceAll(block[1], "\\\n", " ")
		for _, line := range strings.Split(joined, "\n") {
			fields := strings.Fields(line)
			if len(fields) > 1 && fields[0] == "impact-model" {
				out = append(out, fields[1:])
			}
		}
	}
	return out
}

// Every command a skill tells an agent to run must exist, and so must every
// flag it passes, so a renamed flag cannot leave a skill giving bad advice.
func TestEveryCommandAndFlagInASkillExists(t *testing.T) {
	root := cli.NewRootCmd(strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
	flagName := regexp.MustCompile(`^--?([A-Za-z][A-Za-z0-9-]*)`)
	for _, skill := range skillDirs(t) {
		cmds := commands(t, skill)
		if len(cmds) == 0 {
			t.Errorf("%s: no commands", skill)
		}
		for _, args := range cmds {
			cmd, rest, err := root.Find(args)
			if err != nil || cmd == root {
				t.Errorf("%s: %q is not a command", skill, strings.Join(args, " "))
				continue
			}
			for _, a := range rest {
				m := flagName.FindStringSubmatch(a)
				if m == nil || strings.HasPrefix(a, "<") {
					continue
				}
				if !hasFlag(cmd, m[1], len(a) > 1 && a[1] != '-') {
					t.Errorf("%s: `%s` has no flag %s", skill, strings.Join(args, " "), a)
				}
			}
		}
	}
}

func hasFlag(cmd *cobra.Command, name string, short bool) bool {
	found := false
	visit := func(f *pflag.Flag) {
		if (short && f.Shorthand == name) || (!short && f.Name == name) {
			found = true
		}
	}
	cmd.Flags().VisitAll(visit)
	cmd.InheritedFlags().VisitAll(visit)
	return found
}

// session runs commands against files in a scratch directory and records
// everything they print.
type session struct {
	t   *testing.T
	dir string
	log strings.Builder
}

func newSession(t *testing.T, files map[string]string) *session {
	t.Helper()
	s := &session{t: t, dir: t.TempDir()}
	for name, src := range files {
		data, err := os.ReadFile(src)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(s.dir, name), data, 0o600); err != nil { //nolint:gosec // a scratch directory
			t.Fatal(err)
		}
	}
	return s
}

func (s *session) path(name string) string { return filepath.Join(s.dir, name) }

// run executes a command, expecting one of the exit codes.
func (s *session) run(wantCodes []int, args ...string) string {
	s.t.Helper()
	var out, errOut bytes.Buffer
	code := cli.Execute(args, strings.NewReader(""), &out, &errOut)
	ok := false
	for _, c := range wantCodes {
		ok = ok || c == code
	}
	if !ok {
		s.t.Fatalf("impact-model %s: exit %d, want one of %v\n%s%s", strings.Join(args, " "), code, wantCodes, out.String(), errOut.String())
	}
	s.log.WriteString("$ impact-model " + strings.ReplaceAll(strings.Join(args, " "), s.dir, "$DIR") + "\n")
	s.log.WriteString(strings.ReplaceAll(out.String()+errOut.String(), s.dir, "$DIR"))
	return out.String()
}

func (s *session) ok(args ...string) string { s.t.Helper(); return s.run([]int{cli.ExitOK}, args...) }

func (s *session) file(name string) string {
	s.t.Helper()
	data, err := os.ReadFile(s.path(name))
	if err != nil {
		s.t.Fatal(err)
	}
	return string(data)
}

// repeatable runs a skill's sequence twice in separate scratch directories
// and requires everything it printed and wrote to be identical.
func repeatable(t *testing.T, sequence func(t *testing.T) (transcript string, outputs map[string]string)) {
	t.Helper()
	logA, filesA := sequence(t)
	logB, filesB := sequence(t)
	if logA != logB {
		t.Errorf("the transcripts differ between runs:\n--- first\n%s\n--- second\n%s", logA, logB)
	}
	for name, a := range filesA {
		if filesB[name] != a {
			t.Errorf("%s differs between runs", name)
		}
	}
	if len(filesA) == 0 {
		t.Error("the sequence produced no files to compare")
	}
}

func TestAuthorSequenceIsRepeatable(t *testing.T) {
	repeatable(t, func(t *testing.T) (string, map[string]string) {
		s := newSession(t, map[string]string{"impact-model.yaml": theory, "previous.yaml": theory})
		f := s.path("impact-model.yaml")
		s.ok("validate", f)
		// A preview changes nothing; the edit then does.
		before := s.file("impact-model.yaml")
		s.ok("outcome", "set", f, "oc-st-leader-skills", "--title", "Leaders gain coaching skills", "--dry-run")
		if s.file("impact-model.yaml") != before {
			t.Fatal("--dry-run wrote the file")
		}
		s.ok("outcome", "set", f, "oc-st-leader-skills", "--title", "Leaders gain coaching skills")
		s.ok("assumption", "add", f, "--id", "asm-new", "--statement", "Leaders have time to attend.", "--applies-to", "ln-cohorts-skills")
		// An edit that would leave a dangling reference is refused and writes nothing.
		kept := s.file("impact-model.yaml")
		s.run([]int{cli.ExitInvalid}, "link", "set", f, "ln-cohorts-skills", "--to", "oc-ghost")
		if s.file("impact-model.yaml") != kept {
			t.Fatal("a rejected edit changed the file")
		}
		s.ok("validate", f)
		s.run([]int{cli.ExitOK, cli.ExitInvalid}, "lint", f)
		s.ok("diff", s.path("previous.yaml"), f)
		return s.log.String(), map[string]string{"impact-model.yaml": s.file("impact-model.yaml")}
	})
}

func TestStatusRefreshSequenceIsRepeatable(t *testing.T) {
	repeatable(t, func(t *testing.T) (string, map[string]string) {
		s := newSession(t, map[string]string{"impact-model.yaml": theory, "q3.yaml": obs, "period.csv": sheet, "mapping.yaml": mapping})
		f := s.path("impact-model.yaml")
		s.ok("status", f, "-O", s.path("q3.yaml"), "--as-of", "2026-09-30", "--out", s.path("status-prev.json"))
		s.ok("import", "observations", "--csv", s.path("period.csv"), "--map", s.path("mapping.yaml"), "--model", f,
			"--out", s.path("observations.yaml"), "--unmatched", s.path("unmatched.csv"))
		s.ok("validate", s.path("observations.yaml"), "--model", f)
		s.ok("status", f, "-O", s.path("q3.yaml"), "-O", s.path("observations.yaml"), "--prev", s.path("status-prev.json"), "--out", s.path("status.json"))
		s.ok("report", "exec", f, "-O", s.path("q3.yaml"), "-O", s.path("observations.yaml"), "--prev", s.path("status-prev.json"), "--out", s.path("exec.md"))
		return s.log.String(), map[string]string{
			"observations.yaml": s.file("observations.yaml"), "status.json": s.file("status.json"), "exec.md": s.file("exec.md"),
		}
	})
}

func TestStatusRefreshReportsUnmatchedRows(t *testing.T) {
	s := newSession(t, map[string]string{"impact-model.yaml": theory, "mapping.yaml": mapping})
	bad := "Measure,Month end,Value,Program,Quality,Source\r\nNo such measure,31/10/2026,1,,,\r\n"
	if err := os.WriteFile(s.path("period.csv"), []byte(bad), 0o600); err != nil {
		t.Fatal(err)
	}
	s.run([]int{cli.ExitInvalid}, "import", "observations", "--csv", s.path("period.csv"), "--map", s.path("mapping.yaml"),
		"--model", s.path("impact-model.yaml"), "--out", s.path("observations.yaml"), "--unmatched", s.path("unmatched.csv"))
	if !strings.Contains(s.file("unmatched.csv"), "unknown indicator") {
		t.Errorf("the unmatched sheet should give the reason:\n%s", s.file("unmatched.csv"))
	}
}

func TestLearningReviewSequenceIsRepeatable(t *testing.T) {
	repeatable(t, func(t *testing.T) (string, map[string]string) {
		s := newSession(t, map[string]string{"impact-model.yaml": theory, "observations.yaml": obs})
		f, o := s.path("impact-model.yaml"), s.path("observations.yaml")
		s.ok("status", f, "-O", o, "--as-of", "2026-06-30", "--out", s.path("status-prev.json"))
		s.ok("status", f, "-O", o, "--prev", s.path("status-prev.json"))
		s.ok("report", "learning-review", f, "-O", o, "--prev", s.path("status-prev.json"), "--period", "2026-Q3", "--out", s.path("review.md"))
		s.ok("deck", "learning-review", f, "-O", o, "--prev", s.path("status-prev.json"), "--period", "2026-Q3", "--out", s.path("review-deck.md"))
		s.ok("validate", o, "--model", f)
		return s.log.String(), map[string]string{"review.md": s.file("review.md"), "review-deck.md": s.file("review-deck.md")}
	})
}

func TestAudienceStorySequenceIsRepeatable(t *testing.T) {
	repeatable(t, func(t *testing.T) (string, map[string]string) {
		s := newSession(t, map[string]string{"impact-model.yaml": theory, "observations.yaml": obs})
		f, o := s.path("impact-model.yaml"), s.path("observations.yaml")
		s.ok("report", "story", f, "-O", o, "--audience", "aud-funders", "--out", s.path("scaffold.md"))
		s.ok("deck", "story", f, "-O", o, "--audience", "aud-funders", "--out", s.path("story-deck.md"))
		return s.log.String(), map[string]string{"scaffold.md": s.file("scaffold.md"), "story-deck.md": s.file("story-deck.md")}
	})
}

var number = regexp.MustCompile(`\d[\d,]*(?:\.\d+)?%?`)

// unsupported returns the figures and quotations in a draft that the scaffold
// does not contain: the check the audience-story skill asks for.
func unsupported(scaffold, draft string) []string {
	var bad []string
	have := map[string]bool{}
	for _, n := range number.FindAllString(scaffold, -1) {
		have[n] = true
	}
	for _, n := range number.FindAllString(draft, -1) {
		if !have[n] {
			bad = append(bad, "figure "+n)
		}
	}
	for _, q := range regexp.MustCompile(`“([^”]+)”|"([^"]+)"`).FindAllStringSubmatch(draft, -1) {
		text := q[1] + q[2]
		if !strings.Contains(scaffold, text) {
			bad = append(bad, "quotation "+text)
		}
	}
	return bad
}

func TestAudienceStoryDraftCheck(t *testing.T) {
	s := newSession(t, map[string]string{"impact-model.yaml": theory, "observations.yaml": obs})
	scaffold := s.ok("report", "story", s.path("impact-model.yaml"), "-O", s.path("observations.yaml"), "--audience", "aud-funders", "--no-stamp")

	faithful := "The network reached 84 member organizations this year, and as one director put it, " +
		"“I would have spent a month on that alone.”"
	if bad := unsupported(scaffold, faithful); len(bad) != 0 {
		t.Errorf("a faithful draft was flagged: %v", bad)
	}
	invented := "Members spent 5,200 hours together, and one said “We saved a fortune.”"
	bad := unsupported(scaffold, invented)
	if len(bad) != 2 {
		t.Errorf("an invented figure and an invented quotation should both be flagged: %v", bad)
	}
	// A story that was withheld (consent not granted) is not in the scaffold, so it cannot be quoted.
	if strings.Contains(scaffold, "dreading month end") || strings.Contains(scaffold, "closes its monthly books") {
		t.Error("a story without consent is in the scaffold")
	}
	if _, err := os.Stat("impact-audience-story/templates/story-draft.md"); err != nil {
		t.Errorf("the draft template is missing: %v", err)
	}
}
