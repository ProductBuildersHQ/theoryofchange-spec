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
	cmd.Flags().String("out", "", "the "+format+" file to write (required)")
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
	if ref := flagString(cmd, "reference-doc"); ref != "" {
		if _, err := os.Stat(ref); err != nil {
			return usageErr("--reference-doc: %v", err)
		}
	}
	doc, err := loadForRender(path)
	if err != nil {
		return err
	}
	st, err := stampFor(cmd, doc)
	if err != nil {
		return err
	}
	// Check the tools before doing any work.
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
	profile := share.Profile{Diagram: diagramRenderer(ctx, dir)}
	var md string
	if format == "pptx" {
		deck, err := service.WorkshopDeck(doc, "", "")
		if err != nil {
			return renderError(path, err)
		}
		md, err = profile.Deck(st, doc.Metadata.Name, deck)
		if err != nil {
			return err
		}
	} else {
		pathway := ""
		if format == "docx" {
			pathway = flagString(cmd, "pathway")
		}
		text, err := service.DefinitionReport(doc, kind, pathway)
		if err != nil {
			return renderError(path, err)
		}
		md, err = profile.Report(st, doc.Metadata.Name, text)
		if err != nil {
			return err
		}
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
