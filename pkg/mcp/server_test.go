package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	exTheory = "../../examples/intermediary/theory.yaml"
	exObs    = "../../examples/intermediary/observations-2026-q3.yaml"
)

// connect starts the server over an in-memory transport and returns a client
// session, rooted at dir.
func connect(t *testing.T, dir string) *sdk.ClientSession {
	t.Helper()
	srv, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	serverT, clientT := sdk.NewInMemoryTransports()
	ss, err := srv.MCP().Connect(ctx, serverT, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := ss.Close(); err != nil {
			t.Logf("close server session: %v", err)
		}
	})
	client := sdk.NewClient(&sdk.Implementation{Name: "test", Version: "0"}, nil)
	cs, err := client.Connect(ctx, clientT, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := cs.Close(); err != nil {
			t.Logf("close client session: %v", err)
		}
	})
	return cs
}

// call invokes a tool and returns its structured result, or the error text.
func call(t *testing.T, cs *sdk.ClientSession, name string, args map[string]any) (structured map[string]any, errText string) {
	t.Helper()
	res, err := cs.CallTool(context.Background(), &sdk.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		return nil, err.Error()
	}
	if res.IsError {
		var parts []string
		for _, c := range res.Content {
			if tc, ok := c.(*sdk.TextContent); ok {
				parts = append(parts, tc.Text)
			}
		}
		return nil, strings.Join(parts, " ")
	}
	raw, err := json.Marshal(res.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &structured); err != nil {
		t.Fatalf("%s: structured content is not an object: %s", name, raw)
	}
	return structured, ""
}

func read(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func workDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, name := range []string{"theory.yaml", "obs.yaml"} {
		src := exTheory
		if name == "obs.yaml" {
			src = exObs
		}
		if err := os.WriteFile(filepath.Join(dir, name), []byte(read(t, src)), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestEveryToolIsListed(t *testing.T) {
	cs := connect(t, workDir(t))
	res, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, tool := range res.Tools {
		got[tool.Name] = true
		if tool.Description == "" || tool.InputSchema == nil {
			t.Errorf("tool %s lacks a description or a schema", tool.Name)
		}
	}
	for _, want := range []string{"toc_schema", "toc_validate", "toc_lint", "toc_graph", "toc_status", "toc_report", "toc_patch", "toc_deck", "toc_judge_bundle", "toc_diff"} {
		if !got[want] {
			t.Errorf("tool %s is missing", want)
		}
	}
}

func TestSchemaTool(t *testing.T) {
	cs := connect(t, workDir(t))
	out, errText := call(t, cs, "toc_schema", nil)
	if errText != "" || len(out["names"].([]any)) < 6 {
		t.Fatalf("names: %v %q", out, errText)
	}
	out, errText = call(t, cs, "toc_schema", map[string]any{"name": "theory-of-change"})
	if errText != "" || out["title"] != "Theory of Change" {
		t.Fatalf("schema: %v %q", out["title"], errText)
	}
	if _, errText := call(t, cs, "toc_schema", map[string]any{"name": "nope"}); errText == "" {
		t.Error("an unknown schema was returned")
	}
}

func TestValidateAndLintByPathAndInline(t *testing.T) {
	dir := workDir(t)
	cs := connect(t, dir)
	out, errText := call(t, cs, "toc_validate", map[string]any{"path": "theory.yaml"})
	if errText != "" || out["valid"] != true || out["kind"] != "TheoryOfChange" {
		t.Fatalf("validate by path: %v %q", out, errText)
	}
	var inline any
	if err := json.Unmarshal([]byte(`{"kind":"TheoryOfChange","specVersion":"0.1","id":"x","context":{"problemStatement":"p"},"outcomes":[{"id":"oc","title":"t","level":"impact"}],"links":[{"id":"l","from":"oc","to":"ghost"}]}`), &inline); err != nil {
		t.Fatal(err)
	}
	out, errText = call(t, cs, "toc_validate", map[string]any{"document": inline})
	if errText != "" || out["valid"] != false || len(out["issues"].([]any)) == 0 {
		t.Fatalf("an invalid inline document: %v %q", out, errText)
	}
	out, errText = call(t, cs, "toc_lint", map[string]any{"path": "theory.yaml", "profile": "draft"})
	if errText != "" || out["status"] == nil || out["project"] != "riverbend-network" {
		t.Fatalf("lint: %v %q", out, errText)
	}
	if _, errText := call(t, cs, "toc_lint", map[string]any{"path": "theory.yaml", "disable": []string{"TOC-NOPE-001"}}); errText == "" {
		t.Error("an unknown rule was accepted")
	}
	// The same theory as a string of YAML.
	out, errText = call(t, cs, "toc_validate", map[string]any{"document": read(t, exTheory)})
	if errText != "" || out["valid"] != true {
		t.Fatalf("validate YAML string: %v %q", out, errText)
	}
}

func TestGraphStatusReportDeckJudgeAndDiff(t *testing.T) {
	dir := workDir(t)
	cs := connect(t, dir)

	out, errText := call(t, cs, "toc_graph", map[string]any{"path": "theory.yaml", "query": "summary"})
	if errText != "" || out["result"] == nil {
		t.Fatalf("graph summary: %v %q", out, errText)
	}
	out, errText = call(t, cs, "toc_graph", map[string]any{"path": "theory.yaml", "query": "mermaid"})
	if errText != "" || !strings.HasPrefix(out["content"].(string), "flowchart BT") {
		t.Fatalf("graph mermaid: %v %q", out, errText)
	}
	if _, errText := call(t, cs, "toc_graph", map[string]any{"path": "theory.yaml", "query": "upstream"}); errText == "" {
		t.Error("upstream with no id was accepted")
	}

	out, errText = call(t, cs, "toc_status", map[string]any{"path": "theory.yaml", "observationPaths": []string{"obs.yaml"}})
	if errText != "" || out["asOf"] != "2026-09-30" || out["kind"] != "TheoryOfChangeStatus" {
		t.Fatalf("status: asOf %v %q", out["asOf"], errText)
	}
	if _, errText := call(t, cs, "toc_status", map[string]any{"path": "theory.yaml"}); errText == "" {
		t.Error("status with no observations and no date was accepted")
	}

	for kind, want := range map[string]string{"exec": "executive summary", "glossary": "Glossary", "outcome-map": "Outcome map", "narrative": "programme narrative"} {
		out, errText = call(t, cs, "toc_report", map[string]any{"path": "theory.yaml", "observationPaths": []string{"obs.yaml"}, "kind": kind})
		if errText != "" || out["format"] != "markdown" || !strings.Contains(out["content"].(string), want) {
			t.Errorf("report %s: %q %.100v", kind, errText, out["content"])
		}
	}
	out, errText = call(t, cs, "toc_report", map[string]any{"path": "theory.yaml", "kind": "story", "observationPaths": []string{"obs.yaml"}, "audience": "aud-funders"})
	if errText != "" || !strings.Contains(out["content"].(string), "story for Funders") {
		t.Errorf("story report: %q", errText)
	}
	if _, errText := call(t, cs, "toc_report", map[string]any{"path": "theory.yaml", "kind": "mystery"}); errText == "" {
		t.Error("an unknown report kind was accepted")
	}

	for kind, want := range map[string]string{"workshop": "marp: true", "exec": "# Where we stand"} {
		out, errText = call(t, cs, "toc_deck", map[string]any{"path": "theory.yaml", "observationPaths": []string{"obs.yaml"}, "kind": kind})
		if errText != "" || out["format"] != "marp" || !strings.Contains(out["content"].(string), want) {
			t.Errorf("deck %s: %q", kind, errText)
		}
	}
	if _, errText := call(t, cs, "toc_deck", map[string]any{"path": "theory.yaml", "kind": "story", "observationPaths": []string{"obs.yaml"}}); errText == "" {
		t.Error("a story deck with no audience was accepted")
	}

	out, errText = call(t, cs, "toc_judge_bundle", map[string]any{"path": "theory.yaml"})
	if errText != "" || out["kind"] != "JudgeBundle" || out["bundleHash"] == "" || out["profile"] != "core" {
		t.Fatalf("judge bundle: %q", errText)
	}

	out, errText = call(t, cs, "toc_diff", map[string]any{"old": map[string]any{"path": "theory.yaml"}, "new": map[string]any{"path": "theory.yaml"}})
	if errText != "" || len(out["changes"].([]any)) != 0 {
		t.Fatalf("diff of a theory with itself: %v %q", out, errText)
	}
}

func TestPatchToolEditsInlineAndByPath(t *testing.T) {
	dir := workDir(t)
	cs := connect(t, dir)
	title := []map[string]any{{"op": "replace", "path": "/outcomes/@oc-st-leader-skills/title", "value": "Leaders gain coaching skills"}}

	// By path, a dry run changes nothing on disk and returns the result.
	before := read(t, filepath.Join(dir, "theory.yaml"))
	out, errText := call(t, cs, "toc_patch", map[string]any{"path": "theory.yaml", "patch": title, "dryRun": true})
	if errText != "" || out["applied"] != true || out["document"] == nil {
		t.Fatalf("dry run: %v %q", out, errText)
	}
	if read(t, filepath.Join(dir, "theory.yaml")) != before {
		t.Fatal("a dry run wrote the file")
	}

	// By path, not a dry run, writes the canonical file.
	out, errText = call(t, cs, "toc_patch", map[string]any{"path": "theory.yaml", "patch": title})
	if errText != "" || out["applied"] != true || out["written"] != true {
		t.Fatalf("write: %v %q", out, errText)
	}
	if !strings.Contains(read(t, filepath.Join(dir, "theory.yaml")), "Leaders gain coaching skills") {
		t.Error("the file was not updated")
	}

	// A patch that leaves a dangling reference is rejected and writes nothing.
	bad := []map[string]any{{"op": "replace", "path": "/links/@ln-skills-practice/to", "value": "oc-ghost"}}
	after := read(t, filepath.Join(dir, "theory.yaml"))
	out, errText = call(t, cs, "toc_patch", map[string]any{"path": "theory.yaml", "patch": bad})
	if errText != "" || out["applied"] != false || !strings.Contains(out["reason"].(string), "structural errors") && len(out["issuesIntroduced"].([]any)) == 0 {
		t.Fatalf("an error-introducing patch must be reported as not applied, with the new issues: %v %q", out, errText)
	}
	if read(t, filepath.Join(dir, "theory.yaml")) != after {
		t.Error("a rejected patch changed the file")
	}

	// Inline documents are patched in memory.
	var inline any
	if err := json.Unmarshal([]byte(`{"kind":"TheoryOfChange","specVersion":"0.1","id":"x","context":{"problemStatement":"p"},"outcomes":[{"id":"oc","title":"t","level":"impact"}]}`), &inline); err != nil {
		t.Fatal(err)
	}
	out, errText = call(t, cs, "toc_patch", map[string]any{"document": inline, "patch": []map[string]any{{"op": "replace", "path": "/outcomes/@oc/title", "value": "renamed"}}})
	if errText != "" || out["applied"] != true || !strings.Contains(string(mustJSON(t, out["document"])), "renamed") {
		t.Fatalf("inline patch: %v %q", out, errText)
	}
	if _, errText := call(t, cs, "toc_patch", map[string]any{"document": inline, "patch": []map[string]any{{"op": "frobnicate", "path": "/x"}}}); errText == "" {
		t.Error("an unknown operation was accepted")
	}
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestPathsAreConfinedToTheRoot(t *testing.T) {
	dir := workDir(t)
	outside := t.TempDir()
	secret := filepath.Join(outside, "secret.yaml")
	if err := os.WriteFile(secret, []byte(read(t, exTheory)), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(dir, "link")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(secret, filepath.Join(dir, "file-link.yaml")); err != nil {
		t.Fatal(err)
	}
	cs := connect(t, dir)
	for name, path := range map[string]string{
		"parent":              "../" + filepath.Base(outside) + "/secret.yaml",
		"deeper parent":       "sub/../../x.yaml",
		"absolute outside":    secret,
		"symlinked directory": "link/secret.yaml",
		"symlinked file":      "file-link.yaml",
		"nul byte":            "theory.yaml\x00.txt",
	} {
		for _, tool := range []string{"toc_validate", "toc_lint", "toc_judge_bundle", "toc_patch"} {
			args := map[string]any{"path": path}
			if tool == "toc_patch" {
				args["patch"] = []map[string]any{{"op": "add", "path": "/x", "value": 1}}
			}
			if out, errText := call(t, cs, tool, args); errText == "" {
				t.Errorf("%s accepted the %s path %q: %v", tool, name, path, out)
			}
		}
	}
	if _, errText := call(t, cs, "toc_status", map[string]any{"path": "theory.yaml", "observationPaths": []string{"../escape.yaml"}}); errText == "" {
		t.Error("an observation path outside the root was accepted")
	}
	// The file outside was never touched.
	if read(t, secret) != read(t, exTheory) {
		t.Error("a file outside the root changed")
	}
}

func TestRootResolve(t *testing.T) {
	dir := t.TempDir()
	r, err := NewRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	// A new file in an existing directory resolves; a new file in a missing one does too, inside the root.
	for _, p := range []string{"new.yaml", "sub/new.yaml", filepath.Join(r.Dir(), "abs.yaml"), "./dot.yaml"} {
		got, err := r.Resolve(p)
		if err != nil || !strings.HasPrefix(got, r.Dir()) {
			t.Errorf("Resolve(%q) = %q, %v", p, got, err)
		}
	}
	// "/etc/passwd", "\\x" and "C:x" are rooted or name a drive on every platform.
	for _, p := range []string{"", "..", "a/../../b", "/etc/passwd", `\x`, "C:x.yaml", "../x"} {
		if _, err := r.Resolve(p); err == nil {
			t.Errorf("Resolve(%q) was accepted", p)
		}
	}
	if _, err := NewRoot(filepath.Join(dir, "missing")); err == nil {
		t.Error("a root that does not exist was accepted")
	}
	file := filepath.Join(dir, "file")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := NewRoot(file); err == nil {
		t.Error("a file was accepted as a root")
	}
}

func TestInlineDocumentSizeIsLimited(t *testing.T) {
	cs := connect(t, workDir(t))
	huge := strings.Repeat("x", 10<<20+1)
	if _, errText := call(t, cs, "toc_validate", map[string]any{"document": huge}); errText == "" {
		t.Error("an oversized inline document was accepted")
	}
}
