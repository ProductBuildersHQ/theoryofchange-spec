// Command theoryofchange captures, checks, edits, and reports on Theories
// of Change.
package main

import (
	"os"

	"github.com/ProductBuildersHQ/theoryofchange-spec/internal/cli"
)

func main() {
	os.Exit(cli.Execute(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
