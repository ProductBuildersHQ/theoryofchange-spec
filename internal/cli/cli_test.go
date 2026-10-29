package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/tocio"
)

const fixtures = "../../pkg/toc/testdata/"

type result struct {
	code        int
	stdout, err string
}

func run(t *testing.T, stdin string, args ...string) result {
	t.Helper()
	var out, errOut bytes.Buffer
	code := Execute(args, strings.NewReader(stdin), &out, &errOut)
	return result{code: code, stdout: out.String(), err: errOut.String()}
}

func fixtureCopy(t *testing.T, name, as string) string {
	t.Helper()
	data, err := os.ReadFile(fixtures + name)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	dir := t.TempDir()
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := root.Close(); err != nil {
			t.Errorf("close root: %v", err)
		}
	})
	if err := root.WriteFile(as, data, 0o600); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, as)
	return path
}

func mustRead(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func wantCode(t *testing.T, r result, code int) {
	t.Helper()
	if r.code != code {
		t.Fatalf("exit code = %d, want %d\nstdout: %s\nstderr: %s", r.code, code, r.stdout, r.err)
	}
}

func TestVersionAndHelp(t *testing.T) {
	r := run(t, "", "--version")
	wantCode(t, r, 0)
	if !strings.Contains(r.stdout, "theoryofchange version") {
		t.Fatalf("version output = %q", r.stdout)
	}
	r = run(t, "")
	wantCode(t, r, 0)
	if !strings.Contains(r.stdout, "validate") || !strings.Contains(r.stdout, "patch") {
		t.Fatalf("help should list commands: %q", r.stdout)
	}
}

func TestUsageErrorsExitTwo(t *testing.T) {
	for name, args := range map[string][]string{
		"unknown command": {"frobnicate"},
		"unknown flag":    {"validate", "--bogus", "x.json"},
		"missing args":    {"validate"},
		"bad format":      {"schema", "--format", "xml"},
		"patch one arg":   {"patch", "only.json"},
	} {
		r := run(t, "", args...)
		if r.code != ExitUsageErr || r.err == "" {
			t.Errorf("%s: code=%d stderr=%q", name, r.code, r.err)
		}
	}
}

func TestValidate(t *testing.T) {
	good := fixtureCopy(t, "full.json", "good.json")
	r := run(t, "", "validate", good)
	wantCode(t, r, 0)
	if !strings.Contains(r.stdout, "good.json: ok (TheoryOfChange)") {
		t.Fatalf("stdout = %q", r.stdout)
	}
	if r := run(t, "", "validate", "-q", good); r.stdout != "" {
		t.Fatalf("quiet mode should print nothing for valid files: %q", r.stdout)
	}

	dangling := filepath.Join(t.TempDir(), "bad.json")
	data := strings.Replace(mustRead(t, fixtures+"minimal.json"), `"to": "oc-funding"`, `"to": "oc-nowhere"`, 1)
	if err := os.WriteFile(dangling, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	r = run(t, "", "validate", good, dangling)
	wantCode(t, r, 1)
	if !strings.Contains(r.stdout, "bad.json: error VAL-REF-001 /links/0/to") || !strings.Contains(r.stdout, "good.json: ok") {
		t.Fatalf("stdout = %q", r.stdout)
	}

	r = run(t, "", "validate", "--format", "json", dangling)
	wantCode(t, r, 1)
	var reports []struct {
		Source string `json:"source"`
		Kind   string `json:"kind"`
		Issues []struct {
			Code string `json:"code"`
			Path string `json:"path"`
		} `json:"issues"`
	}
	if err := json.Unmarshal([]byte(r.stdout), &reports); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, r.stdout)
	}
	if len(reports) != 1 || reports[0].Kind != "TheoryOfChange" || reports[0].Issues[0].Code != "VAL-REF-001" {
		t.Fatalf("reports = %+v", reports)
	}

	wantCode(t, run(t, "", "validate", filepath.Join(t.TempDir(), "missing.json")), 2)
	wantCode(t, run(t, "", "validate", "notes.txt"), 2)
}

func TestValidateObservationsAgainstTheory(t *testing.T) {
	theory := fixtureCopy(t, "full.json", "theory.json")
	obs := fixtureCopy(t, "observations.json", "obs.json")
	wantCode(t, run(t, "", "validate", obs), 0)
	wantCode(t, run(t, "", "validate", obs, "--toc", theory), 0)

	bad := filepath.Join(t.TempDir(), "obs-bad.json")
	data := strings.Replace(mustRead(t, obs), `"indicatorId": "ind-idp-count"`, `"indicatorId": "ind-ghost"`, 1)
	if err := os.WriteFile(bad, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	r := run(t, "", "validate", bad, "--toc", theory)
	wantCode(t, r, 1)
	if !strings.Contains(r.stdout, "VAL-REF-001") {
		t.Fatalf("stdout = %q", r.stdout)
	}
	wantCode(t, run(t, "", "validate", bad), 0)
	wantCode(t, run(t, "", "validate", bad, "--toc", filepath.Join(t.TempDir(), "none.json")), 2)
}

func TestGraph(t *testing.T) {
	doc := fixtureCopy(t, "full.json", "theory.json")

	r := run(t, "", "graph", doc)
	wantCode(t, r, 0)
	for _, want := range []string{"nodes:           8", "(1 feedback)", "impact outcomes: oc-safer-agents", "cycles:          none"} {
		if !strings.Contains(r.stdout, want) {
			t.Errorf("summary missing %q:\n%s", want, r.stdout)
		}
	}

	r = run(t, "", "graph", doc, "--paths", "--from", "oc-spec-stable")
	wantCode(t, r, 0)
	if strings.TrimSpace(r.stdout) != "oc-spec-stable -> oc-idp-support -> oc-enterprise-require -> oc-safer-agents" {
		t.Fatalf("paths = %q", r.stdout)
	}

	r = run(t, "", "graph", doc, "--downstream", "--from", "oc-idp-support", "--format", "json")
	wantCode(t, r, 0)
	var ids []string
	if err := json.Unmarshal([]byte(r.stdout), &ids); err != nil || len(ids) != 2 {
		t.Fatalf("downstream JSON = %q (%v)", r.stdout, err)
	}

	r = run(t, "", "graph", doc, "--levels")
	wantCode(t, r, 0)
	if !strings.HasPrefix(r.stdout, "0\toc-safer-agents\n") {
		t.Fatalf("levels = %q", r.stdout)
	}

	r = run(t, "", "graph", doc, "--format", "dot")
	wantCode(t, r, 0)
	if !strings.HasPrefix(r.stdout, "digraph theoryofchange {") || !strings.Contains(r.stdout, "[style=dashed]") {
		t.Fatalf("dot = %q", r.stdout)
	}

	for name, args := range map[string][]string{
		"two queries":         {"graph", doc, "--cycles", "--orphans"},
		"upstream needs from": {"graph", doc, "--upstream"},
		"unknown from":        {"graph", doc, "--upstream", "--from", "oc-ghost"},
		"dot with query":      {"graph", doc, "--cycles", "--format", "dot"},
		"bad format":          {"graph", doc, "--format", "xml"},
		"missing file":        {"graph", filepath.Join(t.TempDir(), "none.json")},
	} {
		if r := run(t, "", args...); r.code != ExitUsageErr {
			t.Errorf("%s: code = %d, stderr = %q", name, r.code, r.err)
		}
	}
}

func TestGraphOfInvalidDocumentExitsOne(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	data := strings.Replace(mustRead(t, fixtures+"minimal.json"), `"to": "oc-funding"`, `"to": "oc-nowhere"`, 1)
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	r := run(t, "", "graph", path)
	wantCode(t, r, 1)
	if !strings.Contains(r.err, "oc-nowhere") {
		t.Fatalf("stderr = %q", r.err)
	}
}

func TestSchema(t *testing.T) {
	r := run(t, "", "schema")
	wantCode(t, r, 0)
	if !strings.Contains(r.stdout, "theory-of-change\n") || !strings.Contains(r.stdout, "observation-set\n") {
		t.Fatalf("names = %q", r.stdout)
	}
	r = run(t, "", "schema", "--format", "json")
	var names []string
	if err := json.Unmarshal([]byte(r.stdout), &names); err != nil || len(names) != 5 {
		t.Fatalf("names JSON = %q (%v)", r.stdout, err)
	}
	r = run(t, "", "schema", "theory-of-change")
	wantCode(t, r, 0)
	var s map[string]any
	if err := json.Unmarshal([]byte(r.stdout), &s); err != nil || s["$id"] == nil {
		t.Fatalf("schema output is not a JSON Schema: %v", err)
	}
	if r := run(t, "", "schema", "nope"); r.code != ExitUsageErr || !strings.Contains(r.err, "available") {
		t.Fatalf("unknown schema: code=%d stderr=%q", r.code, r.err)
	}
}

func TestInit(t *testing.T) {
	r := run(t, "", "init", "--id", "demo", "--name", "Demo")
	wantCode(t, r, 0)
	if !strings.HasPrefix(r.stdout, "kind: TheoryOfChange\n") {
		t.Fatalf("default output should be YAML: %q", r.stdout)
	}
	if _, err := tocio.ParseDocument([]byte(r.stdout), tocio.FormatYAML); err != nil {
		t.Fatalf("init output does not parse: %v", err)
	}

	r = run(t, "", "init", "--format", "json")
	wantCode(t, r, 0)
	if _, err := tocio.ParseDocument([]byte(r.stdout), tocio.FormatJSON); err != nil {
		t.Fatalf("init JSON does not parse: %v", err)
	}

	out := filepath.Join(t.TempDir(), "theory.json")
	r = run(t, "", "init", "--out", out, "--problem", "Funders see anecdotes.")
	wantCode(t, r, 0)
	if !strings.HasPrefix(mustRead(t, out), "{") || !strings.Contains(mustRead(t, out), "Funders see anecdotes.") {
		t.Fatalf("--out with .json should write JSON: %q", mustRead(t, out))
	}
	if r := run(t, "", "init", "--out", out); r.code != ExitUsageErr || !strings.Contains(r.err, "already exists") {
		t.Fatalf("overwrite should be refused: code=%d stderr=%q", r.code, r.err)
	}
	wantCode(t, run(t, "", "init", "--out", out, "--force", "-q"), 0)

	for name, args := range map[string][]string{
		"unknown template": {"init", "--template", "nope"},
		"bad format":       {"init", "--format", "toml"},
		"bad id":           {"init", "--id", "Not A Slug"},
	} {
		if r := run(t, "", args...); r.code != ExitUsageErr && r.code != ExitInvalid {
			t.Errorf("%s: code = %d", name, r.code)
		}
	}
}

func TestFmt(t *testing.T) {
	path := fixtureCopy(t, "minimal.json", "doc.json")
	compact, err := json.Marshal(json.RawMessage(mustRead(t, path)))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, compact, 0o600); err != nil {
		t.Fatal(err)
	}

	r := run(t, "", "fmt", "--check", path)
	wantCode(t, r, 1)
	if !strings.Contains(r.stdout, "would reformat") || mustRead(t, path) != string(compact) {
		t.Fatalf("check mode: stdout=%q", r.stdout)
	}
	r = run(t, "", "fmt", path)
	wantCode(t, r, 0)
	if !strings.Contains(r.stdout, "formatted") || mustRead(t, path) == string(compact) {
		t.Fatalf("write mode: stdout=%q", r.stdout)
	}
	wantCode(t, run(t, "", "fmt", "--check", path), 0)
	r = run(t, "", "fmt", "--format", "json", path)
	wantCode(t, r, 0)
	if !strings.Contains(r.stdout, `"changed": false`) {
		t.Fatalf("json output = %q", r.stdout)
	}
	wantCode(t, run(t, "", "fmt", filepath.Join(t.TempDir(), "none.json")), 2)
}

func TestPatch(t *testing.T) {
	path := fixtureCopy(t, "full.json", "theory.json")
	rename := `[{"op":"replace","path":"/outcomes/@oc-idp-support/title","value":"Providers ship it"}]`

	// Dry run prints the document, writes nothing.
	before := mustRead(t, path)
	r := run(t, rename, "patch", path, "-", "--dry-run")
	wantCode(t, r, 0)
	if !strings.Contains(r.stdout, "Providers ship it") || !strings.Contains(r.err, "dry run") || mustRead(t, path) != before {
		t.Fatalf("dry run: stdout has title=%v stderr=%q", strings.Contains(r.stdout, "Providers"), r.err)
	}

	// Real run via stdin.
	r = run(t, rename, "patch", path, "-")
	wantCode(t, r, 0)
	if !strings.Contains(mustRead(t, path), "Providers ship it") || !strings.Contains(r.err, "patch applied") {
		t.Fatalf("apply: stderr=%q", r.err)
	}

	// Patch from a file, no change.
	patchFile := filepath.Join(t.TempDir(), "p.json")
	if err := os.WriteFile(patchFile, []byte(rename), 0o600); err != nil {
		t.Fatal(err)
	}
	r = run(t, "", "patch", path, patchFile)
	wantCode(t, r, 0)
	if !strings.Contains(r.err, "no change") {
		t.Fatalf("no-op: stderr=%q", r.err)
	}

	// Rejection: exit 1, nothing written, issues listed.
	snapshot := mustRead(t, path)
	bad := `[{"op":"add","path":"/links/-","value":{"id":"ln-bad","from":"oc-spec-stable","to":"oc-ghost"}}]`
	r = run(t, bad, "patch", path, "-")
	wantCode(t, r, 1)
	if !strings.Contains(r.err, "rejected") || !strings.Contains(r.err, "introduced: error VAL-REF-001") || mustRead(t, path) != snapshot {
		t.Fatalf("rejection: stderr=%q", r.err)
	}
	wantCode(t, run(t, bad, "patch", path, "-", "--allow-errors"), 0)

	// JSON output.
	r = run(t, `[{"op":"test","path":"/id","value":"open-agent-auth"}]`, "patch", path, "-", "--format", "json")
	wantCode(t, r, 0)
	var res struct {
		Applied bool   `json:"applied"`
		Changed bool   `json:"changed"`
		Path    string `json:"path"`
	}
	if err := json.Unmarshal([]byte(r.stdout), &res); err != nil || !res.Applied || res.Changed {
		t.Fatalf("json result = %q (%v)", r.stdout, err)
	}

	// Errors exit 2 and write nothing.
	snapshot = mustRead(t, path)
	for name, in := range map[string]string{
		"unknown id":    `[{"op":"remove","path":"/outcomes/@ghost"}]`,
		"failed test":   `[{"op":"test","path":"/id","value":"other"}]`,
		"malformed":     `[{`,
		"unknown op":    `[{"op":"explode","path":"/id"}]`,
		"missing value": `[{"op":"replace","path":"/id"}]`,
		"not an array":  `{"op":"remove","path":"/id"}`,
	} {
		if r := run(t, in, "patch", path, "-"); r.code != ExitUsageErr {
			t.Errorf("%s: code = %d, stderr = %q", name, r.code, r.err)
		}
	}
	if mustRead(t, path) != snapshot {
		t.Fatal("a failing patch changed the file")
	}
	wantCode(t, run(t, "", "patch", path, filepath.Join(t.TempDir(), "none.json")), 2)
	wantCode(t, run(t, rename, "patch", filepath.Join(t.TempDir(), "none.json"), "-"), 2)
}

func TestPatchYAMLFile(t *testing.T) {
	out := filepath.Join(t.TempDir(), "theory.yaml")
	wantCode(t, run(t, "", "init", "--id", "demo", "--out", out, "-q"), 0)
	patch := `[{"op":"replace","path":"/outcomes/@oc-first/title","value":"Staff write strong applications"}]`
	wantCode(t, run(t, patch, "patch", out, "-", "-q"), 0)
	data := mustRead(t, out)
	if strings.HasPrefix(data, "{") || !strings.Contains(data, "Staff write strong applications") {
		t.Fatalf("file should stay YAML:\n%s", data)
	}
	wantCode(t, run(t, "", "validate", out), 0)
}

func TestExitErrorText(t *testing.T) {
	if (&ExitError{Code: 7}).Error() != "exit status 7" {
		t.Fatal("ExitError without Err")
	}
	e := usageErr("bad %s", "thing")
	if !strings.Contains(e.Error(), "bad thing") {
		t.Fatalf("usageErr = %v", e)
	}
}

func TestValidateAgainstParent(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) string {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	parent := write("parent.json", `{
  "kind": "TheoryOfChange", "specVersion": "0.1", "id": "agent-ecosystem-trust",
  "metadata": {}, "context": {"problemStatement": "p"},
  "outcomes": [{"id": "oc-trusted-agents", "title": "Agents are trusted", "level": "impact"}],
  "pathways": [{"id": "pw-authz", "title": "Authorization", "childToc": "toc:open-agent-auth"}]
}`)
	child := fixtureCopy(t, "full.json", "child.json")
	r := run(t, "", "validate", child, "--parent", parent)
	wantCode(t, r, 0)
	if !strings.Contains(r.stdout, "parent "+parent+": ok") {
		t.Fatalf("stdout = %q", r.stdout)
	}

	broken := write("broken.json", strings.Replace(mustRead(t, parent), "oc-trusted-agents", "oc-renamed", 1))
	r = run(t, "", "validate", child, "--parent", broken)
	wantCode(t, r, 1)
	if !strings.Contains(r.stdout, "VAL-FAM-002") {
		t.Fatalf("stdout = %q", r.stdout)
	}
	wantCode(t, run(t, "", "validate", child, "--parent", filepath.Join(dir, "none.json")), 2)
}
