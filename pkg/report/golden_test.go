package report

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "rewrite golden files")

// golden compares got with testdata/<name>, rewriting it with -update.
func golden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name)
	if *update {
		if err := os.WriteFile(path, []byte(got), 0o600); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s: %v (run with -update to create it)", path, err)
	}
	if string(want) != got {
		t.Errorf("%s differs from the golden file (run with -update to accept):\n--- got\n%s\n--- want\n%s", name, got, want)
	}
}
