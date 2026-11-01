package cli

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/impact-model/pkg/graph"
	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
	"github.com/ProductBuildersHQ/impact-model/pkg/impactio"
	"github.com/ProductBuildersHQ/impact-model/pkg/report"
	"github.com/ProductBuildersHQ/impact-model/pkg/service"
	"github.com/ProductBuildersHQ/impact-model/pkg/share"
)

// asOfFlags adds the flags that control the export stamp.
func asOfFlags(cmd *cobra.Command) {
	cmd.Flags().String("as-of", "", "date the artifact describes (YYYY-MM-DD), recorded in the export stamp")
	cmd.Flags().Bool("no-stamp", false, "leave the export stamp off")
}

// stampFor builds the export stamp for doc from the --as-of flag.
func stampFor(cmd *cobra.Command, doc *impact.Model) (share.Stamp, error) {
	s, err := share.NewStamp(doc, flagString(cmd, "as-of"))
	if err != nil {
		return share.Stamp{}, &ExitError{Code: ExitUsageErr, Err: err}
	}
	return s, nil
}

// loadForRender loads a document and maps a failure to the exit code.
func loadForRender(path string) (*impact.Model, error) {
	doc, err := impactio.LoadDocument(path)
	if err != nil {
		return nil, &ExitError{Code: ExitUsageErr, Err: err}
	}
	return doc, nil
}

// renderError maps an error from building a view to an exit code.
func renderError(path string, err error) error {
	if errors.Is(err, graph.ErrInvalidDocument) {
		return &ExitError{Code: ExitInvalid, Err: fmt.Errorf("%s: %w", path, err)}
	}
	return &ExitError{Code: ExitUsageErr, Err: err}
}

// renderJSON loads a document, produces a JSON document with fn, and writes
// it to --out or standard output, without a stamp.
func renderJSON(e *env, cmd *cobra.Command, path string, fn func(*impact.Model) (string, error)) error {
	doc, err := loadForRender(path)
	if err != nil {
		return err
	}
	text, err := fn(doc)
	if err != nil {
		return renderError(path, err)
	}
	if out := flagString(cmd, "out"); out != "" {
		if err := impactio.WriteFile(out, []byte(text)); err != nil {
			return &ExitError{Code: ExitUsageErr, Err: err}
		}
		return nil
	}
	fmt.Fprint(e.out, text)
	return nil
}

// render loads a document, produces text with fn, stamps it (Markdown gets a
// metadata block, CSV a leading comment row) unless --no-stamp is set, and
// writes it to --out or standard output.
func render(e *env, cmd *cobra.Command, path string, csv bool, fn func(*impact.Model) (string, error)) error {
	doc, err := loadForRender(path)
	if err != nil {
		return err
	}
	text, err := fn(doc)
	if err != nil {
		return renderError(path, err)
	}
	if !flagBool(cmd, "no-stamp") {
		st, err := stampFor(cmd, doc)
		if err != nil {
			return err
		}
		if csv {
			text = share.StampCSV(st, text)
		} else {
			text = share.StampMarkdown(st, doc.Metadata.Name, text)
		}
	}
	if out := flagString(cmd, "out"); out != "" {
		if err := impactio.WriteFile(out, []byte(text)); err != nil {
			return &ExitError{Code: ExitUsageErr, Err: err}
		}
		return nil
	}
	fmt.Fprint(e.out, text)
	return nil
}

func newReportCmd(e *env) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report <kind> <model-file>",
		Short: "Render a report from an impact model",
		Long: `Render a report as portable Markdown.

Definition reports need no observations, so a workshop can review them before
any data exists. Kinds: ` + strings.Join(report.DefinitionKinds, ", ") + `.
"definitions" renders all of them as one document.

Status reports describe how far the impact model has come, from observation files
(-o) as of a date (--as-of, or the latest date in the observations). Kinds:
exec, results, assumptions, interventions, evidence-gaps, full, story
(--audience), headline (--audience narrows it), metric-tree (--pathway narrows it), learning-review (--period), portfolio, narrative, actors, and
custom (--template, a Go text/template over the report view). narrative,
actors, and custom also work with no observations.`,
		Example: `  impact-model report definitions impact-model.yaml --out definitions.md
  impact-model report logic-model impact-model.yaml --pathway pw-supply
  impact-model report exec impact-model.yaml -O obs-q3.yaml --prev status-q2.json
  impact-model report story impact-model.yaml -O obs-q3.yaml --audience aud-funders
  impact-model report headline impact-model.yaml -O obs-q3.yaml
  impact-model report metric-tree impact-model.yaml -O obs-q3.yaml --pathway pw-supply
  impact-model report custom impact-model.yaml --template board.tmpl`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if service.IsStatusReport(args[0]) {
				return runStatusReport(e, cmd, args[0], args[1])
			}
			kind := args[0]
			if flagBool(cmd, "diagram") {
				if kind != "logic-model" {
					return usageErr("--diagram applies to the logic-model report")
				}
				kind = "logic-model-diagram"
			}
			fn := func(doc *impact.Model) (string, error) {
				return service.DefinitionReport(doc, kind, flagString(cmd, "pathway"))
			}
			if strings.HasSuffix(kind, "-json") {
				// A JSON document cannot carry the Markdown stamp.
				return renderJSON(e, cmd, args[1], fn)
			}
			return render(e, cmd, args[1], false, fn)
		},
	}
	f := cmd.Flags()
	f.String("pathway", "", "restrict a definition report, or the metric-tree report, to one pathway ID")
	f.Bool("diagram", false, "draw the logic-model report as a left-to-right diagram")
	f.StringP("out", "o", "", "write to this file instead of standard output")
	f.String("audience", "", "the audience ID, for the story report; narrows the headline report")
	f.String("period", "", "the period label, for the learning-review report")
	f.StringSlice("prev-observations", nil, "the earlier period's observation files, for the learning-review report")
	f.String("template", "", "the Go text/template file, for the custom report")
	statusFlags(cmd)
	asOfFlags(cmd)
	return cmd
}

// runStatusReport renders a report made from a status snapshot.
func runStatusReport(e *env, cmd *cobra.Command, kind, path string) error {
	withStatus := !slices.Contains(service.OptionalStatusKinds, kind) || flagString(cmd, "as-of") != "" || len(observationPaths(cmd)) > 0
	var (
		in   report.StatusInput
		sets []*impact.ObservationSet
		err  error
	)
	if withStatus {
		if in, sets, err = loadStatus(e, cmd, path); err != nil {
			return err
		}
	} else {
		doc, err := loadForRender(path)
		if err != nil {
			return err
		}
		if in, err = service.ViewWithoutStatus(doc); err != nil {
			return renderError(path, err)
		}
	}
	text, err := statusReportText(cmd, kind, in, sets, withStatus)
	if err != nil {
		return err
	}
	if !flagBool(cmd, "no-stamp") {
		asOf := ""
		if withStatus {
			asOf = in.Status.AsOf
		}
		st, err := share.NewStamp(in.Doc, asOf)
		if err != nil {
			return &ExitError{Code: ExitUsageErr, Err: err}
		}
		text = share.StampMarkdown(st, in.Doc.Metadata.Name, text)
	}
	if out := flagString(cmd, "out"); out != "" {
		if err := impactio.WriteFile(out, []byte(text)); err != nil {
			return &ExitError{Code: ExitUsageErr, Err: err}
		}
		return nil
	}
	fmt.Fprint(e.out, text)
	return nil
}

func observationPaths(cmd *cobra.Command) []string {
	paths, err := cmd.Flags().GetStringSlice("observations")
	if err != nil {
		panic(fmt.Sprintf("cli: --observations is not a string slice flag: %v", err))
	}
	return paths
}

func statusReportText(cmd *cobra.Command, kind string, in report.StatusInput, sets []*impact.ObservationSet, withStatus bool) (string, error) {
	p := service.ReportParams{Audience: flagString(cmd, "audience"), Period: flagString(cmd, "period"), Pathway: flagString(cmd, "pathway")}
	if kind == "story" && p.Audience == "" {
		return "", usageErr("the story report needs --audience")
	}
	if kind == "learning-review" {
		prev, err := cmd.Flags().GetStringSlice("prev-observations")
		if err != nil {
			panic(fmt.Sprintf("cli: --prev-observations is not a string slice flag: %v", err))
		}
		if len(prev) > 0 {
			if p.PrevSets, err = impactio.LoadObservations(prev...); err != nil {
				return "", &ExitError{Code: ExitUsageErr, Err: err}
			}
		}
	}
	if kind == "custom" {
		file := flagString(cmd, "template")
		if file == "" {
			return "", usageErr("the custom report needs --template")
		}
		data, err := os.ReadFile(file)
		if err != nil {
			return "", &ExitError{Code: ExitUsageErr, Err: fmt.Errorf("read template: %w", err)}
		}
		p.Template = string(data)
	}
	out, err := service.StatusReport(kind, in, sets, p, withStatus)
	if err != nil {
		return "", &ExitError{Code: ExitUsageErr, Err: err}
	}
	return out, nil
}

func newExportCmd(e *env) *cobra.Command {
	export := &cobra.Command{
		Use:   "export",
		Short: "Export tables from an impact model",
	}
	csv := &cobra.Command{
		Use:   "csv <table> <model-file>",
		Short: "Export a table as CSV (RFC 4180)",
		Long: `Export a table for spreadsheets. Tables: ` + strings.Join(report.CSVTables, ", ") + `.
Multi-valued cells are joined with "; ". The first row is a comment carrying the
export stamp ("# model:<id> version:<v> hash:<h>"); --no-stamp leaves it off.`,
		Example: `  impact-model export csv indicators impact-model.yaml --out indicators.csv`,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return render(e, cmd, args[1], true, func(doc *impact.Model) (string, error) {
				return service.ExportCSV(doc, args[0])
			})
		},
	}
	csv.Flags().StringP("out", "o", "", "write to this file instead of standard output")
	asOfFlags(csv)
	export.AddCommand(csv, newExportDocCmd("docx"), newExportDocCmd("pptx"), newReviewPacketCmd(e), newReviewSheetCmd(e),
		newPrismExportCmd(e, "okr"), newPrismExportCmd(e, "journey"), newPrismExportCmd(e, "v2mom"), newPrismExportCmd(e, "dmaic"))
	return export
}
