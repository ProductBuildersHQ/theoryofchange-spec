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

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/report"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/service"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/share"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/slides"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/tocio"
)

// mermaidFence matches a fenced Mermaid block in the deck.
var mermaidFence = regexp.MustCompile("(?s)```mermaid\n(.*?)```\n")

func newDeckCmd(e *env) *cobra.Command {
	deck := &cobra.Command{
		Use:   "deck",
		Short: "Generate slide decks",
	}
	deck.AddCommand(newWorkshopDeckCmd(e), newExecDeckCmd(e), newLearningReviewDeckCmd(e), newStoryDeckCmd(e))
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

By default the Marp Markdown is written to standard output, or to --out.
With --export <file>, the deck is also rendered by the external tools marp
(PDF, PPTX, or HTML by file extension) and mmdc (diagrams). If a needed tool
is not installed, nothing is rendered and the exit code is 3; the Markdown
can still be written without --export.`,
		Example: `  theoryofchange deck workshop theory.yaml > workshop.md
  theoryofchange deck workshop theory.yaml -o workshop.md
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
			return deliverDeck(e, cmd, md)
		},
	}
	asOfFlags(cmd)
	cmd.Flags().String("theme", "default", "Marp theme name")
	cmd.Flags().StringP("out", "o", "", "write the Markdown to this file instead of standard output")
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

// deckOutputFlags adds the flags every deck shares.
func deckOutputFlags(cmd *cobra.Command) {
	asOfFlags(cmd)
	cmd.Flags().String("theme", "default", "Marp theme name")
	cmd.Flags().StringP("out", "o", "", "write the Markdown to this file instead of standard output")
	cmd.Flags().String("export", "", "render the deck to this .pdf, .pptx, or .html file with marp and mmdc")
}

// deliverDeck writes the Markdown to --out or standard output and, with
// --export, renders it with the external tools.
func deliverDeck(e *env, cmd *cobra.Command, md string) error {
	if out := flagString(cmd, "out"); out != "" {
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
}

// statusDeckOptions builds the slide options, including the stamp line.
func statusDeckOptions(cmd *cobra.Command, in report.StatusInput) (slides.Options, error) {
	o := slides.Options{Theme: flagString(cmd, "theme")}
	if !flagBool(cmd, "no-stamp") {
		st, err := share.NewStamp(in.Doc, in.Status.AsOf)
		if err != nil {
			return o, &ExitError{Code: ExitUsageErr, Err: err}
		}
		o.Stamp = st.Line()
	}
	return o, nil
}

func newExecDeckCmd(e *env) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exec <toc-file> -O <observations>...",
		Short: "Generate the executive deck",
		Long: `Generate the executive deck from a status snapshot: where the theory stands,
status by level, the pathway coloured by status, where attention is needed,
assumptions to watch, what changed, decisions needed, and how to read the
numbers. The date is --as-of, or the latest date in the observations.

Markdown goes to standard output or --out. With --export <file>, marp and
mmdc render a PDF, PPTX, or HTML file; if a tool is missing the exit code is 3.`,
		Example: `  theoryofchange deck exec theory.yaml -O obs-q3.yaml --prev status-q2.json --export exec.pdf`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			in, _, err := loadStatus(e, cmd, args[0])
			if err != nil {
				return err
			}
			o, err := statusDeckOptions(cmd, in)
			if err != nil {
				return err
			}
			return deliverDeck(e, cmd, slides.Exec(in, o))
		},
	}
	statusFlags(cmd)
	deckOutputFlags(cmd)
	return cmd
}

func newLearningReviewDeckCmd(e *env) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "learning-review <toc-file> -O <observations>...",
		Short: "Generate the learning-review deck",
		Long: `Generate the deck for a learning review: what changed, the surprises, assumptions
that moved, what the stories are about, how far to trust the data, questions
for discussion, and any decisions already recorded for the period.

--period names the period whose recorded review is merged (for example
2026-Q3). --prev-observations names the earlier period's observation files, so
new themes can be marked. Only stories with granted consent are shown.`,
		Example: `  theoryofchange deck learning-review theory.yaml -O obs-q3.yaml --prev status-q2.json --period 2026-Q3`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			in, sets, err := loadStatus(e, cmd, args[0])
			if err != nil {
				return err
			}
			var prevSets []*toc.ObservationSet
			if paths, err := cmd.Flags().GetStringSlice("prev-observations"); err != nil {
				panic(fmt.Sprintf("cli: --prev-observations is not a string slice flag: %v", err))
			} else if len(paths) > 0 {
				if prevSets, err = tocio.LoadObservations(paths...); err != nil {
					return &ExitError{Code: ExitUsageErr, Err: err}
				}
			}
			o, err := statusDeckOptions(cmd, in)
			if err != nil {
				return err
			}
			li := report.LearningInput{StatusInput: in, Sets: sets, PrevSets: prevSets, Period: flagString(cmd, "period")}
			return deliverDeck(e, cmd, slides.LearningReview(li, o))
		},
	}
	statusFlags(cmd)
	deckOutputFlags(cmd)
	cmd.Flags().String("period", "", "the period label, to merge its recorded review")
	cmd.Flags().StringSlice("prev-observations", nil, "the earlier period's observation files (repeatable)")
	return cmd
}

func newStoryDeckCmd(e *env) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "story <toc-file> -O <observations>... --audience <id>",
		Short: "Generate a deck for one audience",
		Long: `Generate a short deck for one audience: what they need, three to five headline
numbers with their caveats, one or two stories they may be told, and a closing
ask to fill in. A story appears only with granted consent, and its quote only
where the story may be quoted.`,
		Example: `  theoryofchange deck story theory.yaml -O obs-q3.yaml --audience aud-funders --export funders.pdf`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			in, sets, err := loadStatus(e, cmd, args[0])
			if err != nil {
				return err
			}
			o, err := statusDeckOptions(cmd, in)
			if err != nil {
				return err
			}
			md, err := slides.Story(in, sets, flagString(cmd, "audience"), o)
			if err != nil {
				return usageErr("%v", err)
			}
			return deliverDeck(e, cmd, md)
		},
	}
	statusFlags(cmd)
	deckOutputFlags(cmd)
	cmd.Flags().String("audience", "", "the audience ID (required)")
	if err := cmd.MarkFlagRequired("audience"); err != nil {
		panic(fmt.Sprintf("cli: mark --audience required: %v", err))
	}
	return cmd
}
