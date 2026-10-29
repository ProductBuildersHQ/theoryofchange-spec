package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/edit"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/tocio"
)

// patchJSON renders a patch as the JSON the patch command reads.
func patchJSON(t *testing.T, ops ...edit.Op) string {
	t.Helper()
	b, err := json.Marshal(edit.Patch(ops))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err) // test fixtures only
	}
	return v
}

// TestCommandsEqualTheirPatches runs each element command on one copy of the
// document and the hand-written equivalent patch on another, and requires
// identical files (TRD-030).
func TestCommandsEqualTheirPatches(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		patch func(t *testing.T) []edit.Op
	}{
		{
			"outcome add",
			[]string{"outcome", "add", "FILE", "--id", "oc-new", "--title", "A new outcome", "--level", "short_term", "--actors", "act-idp,act-enterprises"},
			func(t *testing.T) []edit.Op {
				return []edit.Op{must(edit.Add("/outcomes/-", map[string]any{
					"id": "oc-new", "title": "A new outcome", "level": "short_term", "actorIds": []string{"act-idp", "act-enterprises"},
				}))}
			},
		},
		{
			"group add with nested logic",
			[]string{"group", "add", "FILE", "--id", "grp-new", "--target", "oc-safer-agents", "--members", "oc-idp-support,oc-platform-support", "--logic", "at_least", "--k", "1"},
			func(t *testing.T) []edit.Op {
				return []edit.Op{must(edit.Add("/preconditionGroups/-", map[string]any{
					"id": "grp-new", "target": "oc-safer-agents", "members": []string{"oc-idp-support", "oc-platform-support"},
					"logic": map[string]any{"type": "at_least", "k": 1},
				}))}
			},
		},
		{
			"outcome set",
			[]string{"outcome", "set", "FILE", "oc-idp-support", "--title", "Providers ship it", "--set", "description=Shipping means GA", "--unset", "tags"},
			func(t *testing.T) []edit.Op {
				return []edit.Op{
					must(edit.Add("/outcomes/@oc-idp-support/description", "Shipping means GA")),
					must(edit.Add("/outcomes/@oc-idp-support/title", "Providers ship it")),
					edit.Remove("/outcomes/@oc-idp-support/tags"),
				}
			},
		},
		{
			"link set with JSON value",
			[]string{"link", "set", "FILE", "ln-spec-idp", "--set", "assumptionIds=[]", "--feedback=false"},
			func(t *testing.T) []edit.Op {
				return []edit.Op{
					must(edit.Add("/links/@ln-spec-idp/assumptionIds", []any{})),
					must(edit.Add("/links/@ln-spec-idp/feedback", false)),
				}
			},
		},
		{
			"risk rm",
			[]string{"risk", "rm", "FILE", "rsk-competing"},
			func(t *testing.T) []edit.Op { return []edit.Op{edit.Remove("/risks/@rsk-competing")} },
		},
		{
			"decision add",
			[]string{"decision", "add", "FILE", "--id", "dec-two", "--question", "Renew the grant?", "--options", "Yes,No", "--due", "2027-01-31"},
			func(t *testing.T) []edit.Op {
				return []edit.Op{must(edit.Add("/decisions/-", map[string]any{
					"id": "dec-two", "question": "Renew the grant?", "options": []string{"Yes", "No"}, "dueDate": "2027-01-31",
				}))}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viaCommand := fixtureCopy(t, "full.json", "a.json")
			viaPatch := fixtureCopy(t, "full.json", "b.json")

			args := append([]string(nil), tt.args...)
			for i, a := range args {
				if a == "FILE" {
					args[i] = viaCommand
				}
			}
			r := run(t, "", args...)
			wantCode(t, r, 0)
			r = run(t, patchJSON(t, tt.patch(t)...), "patch", viaPatch, "-")
			wantCode(t, r, 0)

			if mustRead(t, viaCommand) != mustRead(t, viaPatch) {
				t.Fatalf("command and patch produced different files:\n--- command\n%s\n--- patch\n%s", mustRead(t, viaCommand), mustRead(t, viaPatch))
			}
			if _, err := tocio.LoadDocument(viaCommand); err != nil {
				t.Fatalf("result does not load: %v", err)
			}
		})
	}
}

// TestEveryKindCanBeAdded adds one minimal element of each kind to a real
// document.
func TestEveryKindCanBeAdded(t *testing.T) {
	tests := map[string][]string{
		"outcome":      {"--id", "oc-x", "--title", "T", "--level", "output"},
		"group":        {"--id", "grp-x", "--target", "oc-safer-agents", "--members", "oc-idp-support,oc-platform-support", "--logic", "any"},
		"link":         {"--id", "ln-x", "--from", "oc-reference-impl", "--to", "oc-idp-support"},
		"intervention": {"--id", "int-x", "--title", "T", "--targets", "oc-spec-stable"},
		"assumption":   {"--id", "asm-x", "--statement", "S", "--applies-to", "ln-spec-idp", "--needs-testing"},
		"indicator":    {"--id", "ind-x", "--measures", "oc-spec-stable", "--name", "N", "--type", "quantitative"},
		"actor":        {"--id", "act-x", "--name", "N", "--roles", "adopter,funder"},
		"evidence":     {"--id", "ev-x", "--title", "T", "--url", "https://example.org/x", "--accessed", "2026-10-29"},
		"factor":       {"--id", "ef-x", "--description", "D", "--affects", "oc-safer-agents"},
		"risk":         {"--id", "rsk-x", "--description", "D", "--category", "market"},
		"decision":     {"--id", "dec-x", "--question", "Q"},
		"pathway":      {"--id", "pw-x", "--title", "T", "--outcomes", "oc-idp-support", "--kind", "program"},
		"consultation": {"--id", "con-x", "--date", "2026-10-29", "--step", "whole", "--method", "workshop"},
		"metric":       {"--id", "met-x", "--name", "N", "--definition", "D", "--formula", "count of things"},
		"term":         {"--id", "term-x", "--term", "Outcome", "--definition", "A change that happens to someone."},
	}
	for _, k := range edit.ElementKinds() {
		args, ok := tests[k.Name]
		if !ok {
			t.Errorf("no test case for kind %q", k.Name)
			continue
		}
		t.Run(k.Name, func(t *testing.T) {
			path := fixtureCopy(t, "full.json", "doc.json")
			r := run(t, "", append([]string{k.Name, "add", path}, args...)...)
			wantCode(t, r, 0)
			doc, err := tocio.LoadDocument(path)
			if err != nil {
				t.Fatalf("result does not load: %v", err)
			}
			if _, ok := doc.Element(args[1]); !ok {
				t.Fatalf("element %q was not added", args[1])
			}
			// Removing it again restores a valid document.
			wantCode(t, run(t, "", k.Name, "rm", path, args[1], "--cascade", "-q"), 0)
			if _, err := tocio.LoadDocument(path); err != nil {
				t.Fatalf("after rm: %v", err)
			}
		})
	}
}

func TestElementAddFailures(t *testing.T) {
	path := fixtureCopy(t, "full.json", "doc.json")
	before := mustRead(t, path)
	tests := map[string]struct {
		args []string
		code int
	}{
		"missing id":            {[]string{"outcome", "add", path, "--title", "T", "--level", "output"}, ExitUsageErr},
		"missing required flag": {[]string{"outcome", "add", path, "--id", "oc-x", "--title", "T"}, ExitUsageErr},
		"duplicate id":          {[]string{"outcome", "add", path, "--id", "oc-idp-support", "--title", "T", "--level", "output"}, ExitInvalid},
		"dangling reference":    {[]string{"link", "add", path, "--id", "ln-x", "--from", "oc-spec-stable", "--to", "oc-ghost"}, ExitInvalid},
		"invalid enum":          {[]string{"outcome", "add", path, "--id", "oc-x", "--title", "T", "--level", "huge"}, ExitInvalid},
		"bad id":                {[]string{"outcome", "add", path, "--id", "Not A Slug", "--title", "T", "--level", "output"}, ExitInvalid},
		"bad set":               {[]string{"outcome", "add", path, "--id", "oc-x", "--title", "T", "--level", "output", "--set", "novalue"}, ExitUsageErr},
		"field given twice":     {[]string{"outcome", "add", path, "--id", "oc-x", "--title", "T", "--level", "output", "--set", "title=U"}, ExitUsageErr},
		"unknown field":         {[]string{"outcome", "add", path, "--id", "oc-x", "--title", "T", "--level", "output", "--set", "bogus=1"}, ExitInvalid},
		"missing file":          {[]string{"outcome", "add", filepath.Join(t.TempDir(), "none.json"), "--id", "oc-x", "--title", "T", "--level", "output"}, ExitUsageErr},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			r := run(t, "", tt.args...)
			if r.code != tt.code {
				t.Fatalf("exit code = %d, want %d\nstderr: %s", r.code, tt.code, r.err)
			}
		})
	}
	if mustRead(t, path) != before {
		t.Fatal("a failing add changed the file")
	}
}

func TestElementSetFailures(t *testing.T) {
	path := fixtureCopy(t, "full.json", "doc.json")
	before := mustRead(t, path)
	for name, tt := range map[string]struct {
		args []string
		code int
	}{
		"change id":       {[]string{"outcome", "set", path, "oc-idp-support", "--set", "id=other"}, ExitUsageErr},
		"nothing to do":   {[]string{"outcome", "set", path, "oc-idp-support"}, ExitUsageErr},
		"unknown element": {[]string{"outcome", "set", path, "oc-ghost", "--title", "T"}, ExitUsageErr},
		"invalid value":   {[]string{"outcome", "set", path, "oc-idp-support", "--level", "huge"}, ExitInvalid},
		"wrong kind":      {[]string{"outcome", "set", path, "rsk-competing", "--title", "T"}, ExitUsageErr},
	} {
		if r := run(t, "", tt.args...); r.code != tt.code {
			t.Errorf("%s: exit code = %d, want %d\nstderr: %s", name, r.code, tt.code, r.err)
		}
	}
	if mustRead(t, path) != before {
		t.Fatal("a failing set changed the file")
	}
}

func TestRmRejectsDanglingReferencesAndCascadeFixesThem(t *testing.T) {
	path := fixtureCopy(t, "full.json", "doc.json")
	before := mustRead(t, path)

	r := run(t, "", "outcome", "rm", path, "oc-idp-support")
	wantCode(t, r, 1)
	if !strings.Contains(r.err, "rejected") || !strings.Contains(r.err, "oc-idp-support") {
		t.Fatalf("rejection should name the references: %q", r.err)
	}
	if mustRead(t, path) != before {
		t.Fatal("rejected rm changed the file")
	}

	r = run(t, "", "outcome", "rm", path, "oc-idp-support", "--cascade", "--dry-run")
	wantCode(t, r, 0)
	if mustRead(t, path) != before {
		t.Fatal("dry run changed the file")
	}
	if strings.Contains(r.stdout, `"id": "oc-idp-support"`) {
		t.Fatal("dry-run output still contains the removed outcome")
	}

	wantCode(t, run(t, "", "outcome", "rm", path, "oc-idp-support", "--cascade"), 0)
	doc, err := tocio.LoadDocument(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := doc.Outcome("oc-idp-support"); ok {
		t.Fatal("outcome still present")
	}
	if _, ok := doc.Link("ln-spec-idp"); ok {
		t.Fatal("link into the outcome survived the cascade")
	}
	wantCode(t, run(t, "", "validate", path), 0)

	for name, args := range map[string][]string{
		"unknown id":   {"outcome", "rm", path, "oc-ghost"},
		"wrong kind":   {"outcome", "rm", path, "rsk-competing"},
		"missing file": {"outcome", "rm", filepath.Join(t.TempDir(), "none.json"), "x"},
	} {
		if r := run(t, "", args...); r.code != ExitUsageErr {
			t.Errorf("%s: exit code = %d\nstderr: %s", name, r.code, r.err)
		}
	}
}

func TestElementCommandsKeepYAML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "theory.yaml")
	wantCode(t, run(t, "", "init", "--id", "demo", "--out", path, "-q"), 0)
	wantCode(t, run(t, "", "outcome", "add", path, "--id", "oc-second", "--title", "Second condition", "--level", "short_term", "-q"), 0)
	wantCode(t, run(t, "", "link", "add", path, "--id", "ln-second", "--from", "oc-second", "--to", "oc-first", "--rationale", "Needed first.", "-q"), 0)
	data := mustRead(t, path)
	if strings.HasPrefix(data, "{") || !strings.Contains(data, "oc-second") {
		t.Fatalf("file should stay YAML:\n%s", data)
	}
	wantCode(t, run(t, "", "validate", path), 0)
	r := run(t, "", "graph", path, "--upstream", "--from", "oc-impact")
	wantCode(t, r, 0)
	if !strings.Contains(r.stdout, "oc-second") {
		t.Fatalf("upstream = %q", r.stdout)
	}
}

func TestFlagTableCoversEveryKind(t *testing.T) {
	for _, k := range edit.ElementKinds() {
		flags, ok := elementFlags[k.Name]
		if !ok || len(flags) == 0 {
			t.Errorf("kind %q has no flag table", k.Name)
			continue
		}
		seen := map[string]bool{}
		required := 0
		for _, f := range flags {
			if seen[f.name] {
				t.Errorf("%s: duplicate flag --%s", k.Name, f.name)
			}
			seen[f.name] = true
			if f.required {
				required++
			}
			switch f.name {
			case "id", "set", "unset", "dry-run", "allow-errors", "format", "quiet", "help":
				t.Errorf("%s: flag --%s collides with a built-in flag", k.Name, f.name)
			}
		}
		if required == 0 {
			t.Errorf("%s: no required flag; an empty element would be added", k.Name)
		}
	}
	if article("outcome") != "an outcome" || article("link") != "a link" || article("evidence") != "an evidence" {
		t.Error("article() mismatch")
	}
	if shortPrefix("outcome") != "oc" || shortPrefix("decision") != "dec" {
		t.Error("shortPrefix() mismatch")
	}
}

func TestSetCreatesMissingParents(t *testing.T) {
	path := fixtureCopy(t, "full.json", "doc.json")
	r := run(t, "", "outcome", "set", path, "oc-spec-stable", "--basis", "stated", "--sources", "ev-survey", "--basis-note", "Stated on the project page.", "--set", "logic.type=any")
	wantCode(t, r, 0)
	doc, err := tocio.LoadDocument(path)
	if err != nil {
		t.Fatal(err)
	}
	o, _ := doc.Outcome("oc-spec-stable")
	if o.Provenance == nil || o.Provenance.Basis != "stated" || o.Provenance.Note != "Stated on the project page." || o.Logic == nil {
		t.Fatalf("outcome = %+v", o)
	}
}

func TestEvidenceAddDerivesTheIDFromTheTitle(t *testing.T) {
	path := fixtureCopy(t, "minimal.json", "theory.json")
	for i, want := range []string{"ev-annual-report-2026", "ev-annual-report-2026-2"} {
		r := run(t, "", "evidence", "add", path, "--title", "Annual Report, 2026!", "--url", "https://example.org/r")
		if r.code != ExitOK {
			t.Fatalf("add %d: code %d, stderr %s", i, r.code, r.err)
		}
		if got := run(t, "", "validate", path); got.code != ExitOK {
			t.Fatalf("validate: %s", got.stdout+got.err)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), `"id": "`+want+`"`) {
			t.Errorf("want evidence %s in:\n%s", want, data)
		}
	}
}

func TestEvidenceAddNeedsAnIDWhenTheTitleHasNoLetters(t *testing.T) {
	path := fixtureCopy(t, "minimal.json", "theory.json")
	if r := run(t, "", "evidence", "add", path, "--title", "???"); r.code != ExitUsageErr {
		t.Errorf("code %d, want %d", r.code, ExitUsageErr)
	}
}

func TestOtherElementsStillRequireAnID(t *testing.T) {
	path := fixtureCopy(t, "minimal.json", "theory.json")
	if r := run(t, "", "outcome", "add", path, "--title", "X", "--level", "impact"); r.code != ExitUsageErr {
		t.Errorf("code %d, want %d", r.code, ExitUsageErr)
	}
}
