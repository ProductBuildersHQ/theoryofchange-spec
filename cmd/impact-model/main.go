// Command impact-model captures, checks, edits, and reports on impact models
// of Change.
package main

import (
	"os"

	"github.com/ProductBuildersHQ/impact-model/internal/cli"
)

func main() {
	os.Exit(cli.Execute(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
