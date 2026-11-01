package cli

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// runAsCLIEnv makes the test binary behave as the impact-model command, for
// tests that need a real process with its own standard input and output.
const runAsCLIEnv = "IMPACT_MODEL_RUN_AS_CLI"

func TestMain(m *testing.M) {
	if os.Getenv(runAsCLIEnv) == "1" {
		os.Exit(Execute(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
	}
	os.Exit(m.Run())
}

// The server speaks MCP over the real process's standard input and output.
func TestMCPServesOverStdio(t *testing.T) {
	root := t.TempDir()
	// The test binary runs as the command (see TestMain), so nothing is built
	// and no platform-specific executable name is needed.
	cmd := exec.Command(os.Args[0], "mcp", "--root", root) //nolint:gosec // re-runs this test binary, not input
	cmd.Env = append(os.Environ(), runAsCLIEnv+"=1")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = stdin.Close()
		done := make(chan struct{})
		go func() { _ = cmd.Wait(); close(done) }()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			_ = cmd.Process.Kill()
		}
	})
	r := bufio.NewReader(stdout)
	send := func(msg string) {
		if _, err := io.WriteString(stdin, msg+"\n"); err != nil {
			t.Fatal(err)
		}
	}
	readResponse := func() map[string]any {
		line, err := r.ReadString('\n')
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("response %q: %v", line, err)
		}
		return m
	}
	send(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"t","version":"0"}}}`)
	init := readResponse()
	if res, _ := init["result"].(map[string]any); res == nil || res["serverInfo"].(map[string]any)["name"] != "impact-model" {
		t.Fatalf("initialize: %v", init)
	}
	send(`{"jsonrpc":"2.0","method":"notifications/initialized"}`)
	send(`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`)
	list := readResponse()
	if !strings.Contains(string(mustMarshal(t, list)), "impact_patch") {
		t.Errorf("tools/list: %v", list)
	}
}

func mustMarshal(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
