package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/report"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/service"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/share"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/tocio"
)

// workshopKind is the export kind that selects the workshop deck.
const workshopKind = "workshop"

func newExportDocCmd(format string) *cobra.Command {
	kinds := strings.Join(report.DefinitionKinds, ", ")
	short := "Export a report as a Word document through Pandoc"
	if format == "pptx" {
		kinds = workshopKind
		short = "Export the workshop deck as a PowerPoint file through Pandoc"
	}
	cmd := &cobra.Command{
		Use:   format + " <kind> <toc-file> --out <file>." + format,
		Short: short,
		Long: fmt.Sprintf(`%s.

The Markdown is converted to the Pandoc profile (diagrams as PNG, basis
badges as text, page breaks, and the export stamp) and then to %s by the
external tools mmdc (diagrams) and pandoc. If a needed tool is not
installed, nothing is written and the exit code is 3.

Kinds: %s.`, short, strings.ToUpper(format), kinds),
		Example: fmt.Sprintf("  theoryofchange export %s %s theory.yaml --out theory.%s --reference-doc template.%s",
			format, strings.Split(kinds, ",")[0], format, format),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return exportDoc(cmd, format, args[0], args[1])
		},
	}
	cmd.Flags().StringP("out", "o", "", "the "+format+" file to write (required)")
	if err := cmd.MarkFlagRequired("out"); err != nil {
		panic(fmt.Sprintf("cli: mark --out required: %v", err))
	}
	if format == "docx" {
		cmd.Flags().String("pathway", "", "restrict the report to one pathway ID")
	}
	cmd.Flags().String("reference-doc", "", "a Pandoc reference document that styles the output")
	asOfFlags(cmd)
	return cmd
}

func exportDoc(cmd *cobra.Command, format, kind, path string) error {
	out := flagString(cmd, "out")
	if !strings.EqualFold(filepath.Ext(out), "."+format) {
		return usageErr("--out must be a .%s file, not %q", format, out)
	}
	if (format == "pptx") != (kind == workshopKind) {
		if format == "pptx" {
			return usageErr("export pptx supports only the %q kind, not %q", workshopKind, kind)
		}
		return usageErr("%q is a deck: use export pptx %s", kind, kind)
	}
	doc, err := loadForRender(path)
	if err != nil {
		return err
	}
	st, err := stampFor(cmd, doc)
	if err != nil {
		return err
	}
	return withPandoc(cmd, format, func(profile share.Profile) (string, error) {
		if format == "pptx" {
			deck, err := service.WorkshopDeck(doc, "", "")
			if err != nil {
				return "", renderError(path, err)
			}
			return profile.Deck(st, doc.Metadata.Name, deck)
		}
		pathway := ""
		if format == "docx" {
			pathway = flagString(cmd, "pathway")
		}
		text, err := service.DefinitionReport(doc, kind, pathway)
		if err != nil {
			return "", renderError(path, err)
		}
		return profile.Report(st, doc.Metadata.Name, text)
	})
}

// withPandoc checks that pandoc is installed, lets build produce the
// profile Markdown (rendering diagrams into a scratch directory as needed),
// and converts it to format at --out. Nothing is written if a tool is
// missing or the build fails.
func withPandoc(cmd *cobra.Command, format string, build func(share.Profile) (string, error)) error {
	out := flagString(cmd, "out")
	if ref := flagString(cmd, "reference-doc"); ref != "" {
		if _, err := os.Stat(ref); err != nil {
			return usageErr("--reference-doc: %v", err)
		}
	}
	pandoc, err := exec.LookPath("pandoc")
	if err != nil {
		return toolMissing("pandoc")
	}
	dir, err := os.MkdirTemp("", "theoryofchange-export-")
	if err != nil {
		return &ExitError{Code: ExitUsageErr, Err: err}
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "theoryofchange: remove %s: %v\n", dir, err)
		}
	}()

	ctx := cmd.Context()
	md, err := build(share.Profile{Diagram: diagramRenderer(ctx, dir)})
	if err != nil {
		return err
	}
	src := filepath.Join(dir, "export.md")
	if err := tocio.WriteFile(src, []byte(md)); err != nil {
		return &ExitError{Code: ExitUsageErr, Err: err}
	}
	abs, err := filepath.Abs(out)
	if err != nil {
		return &ExitError{Code: ExitUsageErr, Err: err}
	}
	args := []string{"-f", "markdown", "-t", format, "-o", abs}
	if ref := flagString(cmd, "reference-doc"); ref != "" {
		args = append(args, "--reference-doc", ref)
	}
	args = append(args, src)
	if err := runTool(ctx, pandoc, args...); err != nil {
		return &ExitError{Code: ExitUsageErr, Err: err}
	}
	return nil
}

// reviewView builds the whole-theory view the review exports render.
func reviewView(path string) (*toc.Document, *report.View, error) {
	doc, err := loadForRender(path)
	if err != nil {
		return nil, nil, err
	}
	g, err := service.BuildGraph(doc)
	if err != nil {
		return nil, nil, renderError(path, err)
	}
	v, err := report.BuildView(doc, g, nil, report.Options{})
	if err != nil {
		return nil, nil, renderError(path, err)
	}
	return doc, v, nil
}

func newReviewPacketCmd(e *env) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "review-packet <toc-file>",
		Short: "Export the review packet: one ID-tagged table per element kind",
		Long: `Export the review packet for reviewers who work in a document. Each element
kind has a table with the element's ID, statement, basis, and sources, and
empty Decision, Suggested change, and Comment columns.

With --format md (the default) the Markdown is written to --out or standard
output. With --format docx it is converted by pandoc and --out is required;
if pandoc is missing, nothing is written and the exit code is 3. Sources are
named by ID and title only: excerpts and URLs are never included.`,
		Example: `  theoryofchange export review-packet theory.yaml --out packet.md
  theoryofchange export review-packet theory.yaml --format docx --out packet.docx`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			format := flagString(cmd, "format")
			if format != "md" && format != "docx" {
				return usageErr("--format must be md or docx, not %q", format)
			}
			doc, v, err := reviewView(args[0])
			if err != nil {
				return err
			}
			st, err := stampFor(cmd, doc)
			if err != nil {
				return err
			}
			packet := share.ReviewPacket(v)
			if format == "docx" {
				if out := flagString(cmd, "out"); !strings.EqualFold(filepath.Ext(out), ".docx") {
					return usageErr("--format docx needs --out <file>.docx")
				}
				return withPandoc(cmd, "docx", func(profile share.Profile) (string, error) {
					return profile.Report(st, doc.Metadata.Name, packet)
				})
			}
			return writeText(e, cmd, packetStamp(cmd, st, doc.Metadata.Name, packet))
		},
	}
	cmd.Flags().String("format", "md", "md or docx")
	cmd.Flags().StringP("out", "o", "", "write to this file instead of standard output")
	cmd.Flags().String("reference-doc", "", "a Pandoc reference document that styles the docx")
	asOfFlags(cmd)
	return cmd
}

func packetStamp(cmd *cobra.Command, st share.Stamp, title, body string) string {
	if flagBool(cmd, "no-stamp") {
		return body
	}
	return share.StampMarkdown(st, title+": review packet", body)
}

func newReviewSheetCmd(e *env) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "review-sheet <toc-file>",
		Short: "Export the review sheet: the same rows as CSV",
		Long: `Export the review rows for reviewers who work in a spreadsheet: one row per
element with kind, id, statement, basis, and sources, and empty decision,
suggested_change, and comment columns. The first row is a comment carrying the
export stamp; --no-stamp leaves it off.`,
		Example: `  theoryofchange export review-sheet theory.yaml --out sheet.csv`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			doc, v, err := reviewView(args[0])
			if err != nil {
				return err
			}
			sheet, err := share.ReviewSheet(v)
			if err != nil {
				return &ExitError{Code: ExitUsageErr, Err: err}
			}
			if !flagBool(cmd, "no-stamp") {
				st, err := stampFor(cmd, doc)
				if err != nil {
					return err
				}
				sheet = share.StampCSV(st, sheet)
			}
			return writeText(e, cmd, sheet)
		},
	}
	cmd.Flags().StringP("out", "o", "", "write to this file instead of standard output")
	asOfFlags(cmd)
	return cmd
}

// writeText writes text to --out or standard output.
func writeText(e *env, cmd *cobra.Command, text string) error {
	if out := flagString(cmd, "out"); out != "" {
		if err := tocio.WriteFile(out, []byte(text)); err != nil {
			return &ExitError{Code: ExitUsageErr, Err: err}
		}
		return nil
	}
	fmt.Fprint(e.out, text)
	return nil
}

// diagramRenderer returns a function that renders Mermaid source to a PNG in
// dir with mmdc and returns the PNG's absolute path. mmdc is looked up on
// first use, so documents without diagrams do not need it.
func diagramRenderer(ctx context.Context, dir string) func(string) (string, error) {
	n := 0
	var mmdc string
	return func(source string) (string, error) {
		if mmdc == "" {
			p, err := exec.LookPath("mmdc")
			if err != nil {
				return "", toolMissing("mmdc")
			}
			mmdc = p
		}
		n++
		in := filepath.Join(dir, fmt.Sprintf("d%d.mmd", n))
		png := filepath.Join(dir, fmt.Sprintf("d%d.png", n))
		if err := tocio.WriteFile(in, []byte(source)); err != nil {
			return "", &ExitError{Code: ExitUsageErr, Err: err}
		}
		if err := runTool(ctx, mmdc, "-i", in, "-o", png, "-s", "2"); err != nil {
			return "", &ExitError{Code: ExitUsageErr, Err: err}
		}
		return png, nil
	}
}
