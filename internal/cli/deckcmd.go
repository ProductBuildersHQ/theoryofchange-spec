package cli

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/service"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/tocio"
)

// mermaidFence matches a fenced Mermaid block in the deck.
var mermaidFence = regexp.MustCompile("(?s)```mermaid\n(.*?)```\n")

func newDeckCmd(e *env) *cobra.Command {
	deck := &cobra.Command{
		Use:   "deck",
		Short: "Generate slide decks",
	}
	deck.AddCommand(newWorkshopDeckCmd(e))
	return deck
}

func newWorkshopDeckCmd(e *env) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workshop <toc-file>",
		Short: "Generate a workshop deck as Marp Markdown",
		Long: `Generate the workshop deck: the problem, the goal and ceiling of
accountability, the pathway diagrams, a logic model per pathway, the key
assumptions, metric definitions, the glossary, and what is still our reading.
It needs no observations.

By default the Marp Markdown is written to standard output, or to --output.
With --export <file>, the deck is also rendered by the external tools marp
(PDF, PPTX, or HTML by file extension) and mmdc (diagrams). If a needed tool
is not installed, nothing is rendered and the exit code is 3; the Markdown
can still be written without --export.`,
		Example: `  theoryofchange deck workshop theory.yaml > workshop.md
  theoryofchange deck workshop theory.yaml --output workshop.md
  theoryofchange deck workshop theory.yaml --export workshop.pdf`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			doc, err := tocio.LoadDocument(args[0])
			if err != nil {
				return &ExitError{Code: ExitUsageErr, Err: err}
			}
			stamp := ""
			if !flagBool(cmd, "no-stamp") {
				st, err := stampFor(cmd, doc)
				if err != nil {
					return err
				}
				stamp = st.Line()
			}
			md, err := service.WorkshopDeck(doc, flagString(cmd, "theme"), stamp)
			if err != nil {
				return renderError(args[0], err)
			}
			if out := flagString(cmd, "output"); out != "" {
				if err := tocio.WriteFile(out, []byte(md)); err != nil {
					return &ExitError{Code: ExitUsageErr, Err: err}
				}
			} else if flagString(cmd, "export") == "" {
				fmt.Fprint(e.out, md)
			}
			if target := flagString(cmd, "export"); target != "" {
				return exportDeck(cmd.Context(), md, target)
			}
			return nil
		},
	}
	asOfFlags(cmd)
	cmd.Flags().String("theme", "default", "Marp theme name")
	cmd.Flags().StringP("output", "o", "", "write the Markdown to this file instead of standard output")
	cmd.Flags().String("export", "", "render the deck to this .pdf, .pptx, or .html file with marp and mmdc")
	return cmd
}

// toolMissing is the exit error for an absent external tool.
func toolMissing(tool string) error {
	return &ExitError{Code: ExitToolMissing, Err: fmt.Errorf("%s is not installed or not on PATH; install it, or write the Markdown without --export", tool)}
}

// exportDeck renders each Mermaid diagram to SVG with mmdc, then the deck to
// the target with marp. All rendering happens in a scratch directory.
func exportDeck(ctx context.Context, md, target string) error {
	switch strings.ToLower(filepath.Ext(target)) {
	case ".pdf", ".pptx", ".html":
	default:
		return usageErr("--export needs a .pdf, .pptx, or .html file, not %q", target)
	}
	marp, err := exec.LookPath("marp")
	if err != nil {
		return toolMissing("marp")
	}
	mmdc := ""
	if mermaidFence.MatchString(md) {
		if mmdc, err = exec.LookPath("mmdc"); err != nil {
			return toolMissing("mmdc")
		}
	}
	dir, err := os.MkdirTemp("", "theoryofchange-deck-")
	if err != nil {
		return &ExitError{Code: ExitUsageErr, Err: err}
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			fmt.Fprintf(os.Stderr, "theoryofchange: remove %s: %v\n", dir, err)
		}
	}()

	var renderErr error
	n := 0
	md = mermaidFence.ReplaceAllStringFunc(md, func(block string) string {
		if renderErr != nil {
			return block
		}
		n++
		src := filepath.Join(dir, fmt.Sprintf("d%d.mmd", n))
		svg := filepath.Join(dir, fmt.Sprintf("d%d.svg", n))
		body := mermaidFence.FindStringSubmatch(block)[1]
		if renderErr = tocio.WriteFile(src, []byte(body)); renderErr != nil {
			return block
		}
		if renderErr = runTool(ctx, mmdc, "-i", src, "-o", svg); renderErr != nil {
			return block
		}
		return fmt.Sprintf("![w:900](%s)\n", filepath.Base(svg))
	})
	if renderErr != nil {
		return &ExitError{Code: ExitUsageErr, Err: renderErr}
	}
	deckFile := filepath.Join(dir, "deck.md")
	if err := tocio.WriteFile(deckFile, []byte(md)); err != nil {
		return &ExitError{Code: ExitUsageErr, Err: err}
	}
	abs, err := filepath.Abs(target)
	if err != nil {
		return &ExitError{Code: ExitUsageErr, Err: err}
	}
	if err := runTool(ctx, marp, deckFile, "--allow-local-files", "-o", abs); err != nil {
		return &ExitError{Code: ExitUsageErr, Err: err}
	}
	return nil
}

func runTool(ctx context.Context, tool string, args ...string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	c := exec.CommandContext(ctx, tool, args...)
	var stderr bytes.Buffer
	c.Stderr = &stderr
	if err := c.Run(); err != nil {
		return fmt.Errorf("%s failed: %w: %s", filepath.Base(tool), err, strings.TrimSpace(stderr.String()))
	}
	return nil
}
