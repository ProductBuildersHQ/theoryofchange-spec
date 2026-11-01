// Command generate-schema writes the JSON Schemas generated from the Go
// types in pkg/impact. It is the only producer of schema/*.schema.json.
//
// Usage:
//
//	go generate ./schema/...
//	go run ./cmd/generate-schema -out schema
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ProductBuildersHQ/impact-model/pkg/schemagen"
)

func main() {
	out := flag.String("out", "schema", "output directory")
	flag.Parse()
	if err := run(*out); err != nil {
		fmt.Fprintln(os.Stderr, "generate-schema:", err)
		os.Exit(1)
	}
}

func run(dir string) error {
	files, err := schemagen.Generate()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create %s: %w", dir, err)
	}
	for _, f := range files {
		path := filepath.Join(dir, f.Name)
		if err := os.WriteFile(path, f.Data, 0o644); err != nil { //nolint:gosec // schema files are public, world-readable by design
			return fmt.Errorf("write %s: %w", path, err)
		}
		fmt.Println("wrote", path)
	}
	return nil
}
