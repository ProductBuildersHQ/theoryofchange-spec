package cli

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/burden"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/graph"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/report"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/service"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/share"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/tocio"
)

// asOfFlags adds the flags that control the export stamp.
func asOfFlags(cmd *cobra.Command) {
	cmd.Flags().String("as-of", "", "date the artifact describes (YYYY-MM-DD), recorded in the export stamp")
	cmd.Flags().Bool("no-stamp", false, "leave the export stamp off")
}

// stampFor builds the export stamp for doc from the --as-of flag.
func stampFor(cmd *cobra.Command, doc *toc.Document) (share.Stamp, error) {
	s, err := share.NewStamp(doc, flagString(cmd, "as-of"))
	if err != nil {
		return share.Stamp{}, &ExitError{Code: ExitUsageErr, Err: err}
	}
	return s, nil
}

// loadForRender loads a document and maps a failure to the exit code.
func loadForRender(path string) (*toc.Document, error) {
	doc, err := tocio.LoadDocument(path)
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

// render loads a document, produces text with fn, stamps it (Markdown gets a
// metadata block, CSV a leading comment row) unless --no-stamp is set, and
// writes it to --out or standard output.
func render(e *env, cmd *cobra.Command, path string, csv bool, fn func(*toc.Document) (string, error)) error {
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
		if err := tocio.WriteFile(out, []byte(text)); err != nil {
			return &ExitError{Code: ExitUsageErr, Err: err}
		}
		return nil
	}
	fmt.Fprint(e.out, text)
	return nil
}

// statusReportKinds are the report kinds made from a status snapshot.
var statusReportKinds = []string{"exec", "results", "assumptions", "interventions", "evidence-gaps", "full", "story", "learning-review", "portfolio", "narrative", "actors", "custom"}

// optionalStatus are the kinds that work with or without observations.
var optionalStatus = []string{"narrative", "actors", "custom"}

func newReportCmd(e *env) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report <kind> <toc-file>",
		Short: "Render a report from a Theory of Change",
		Long: `Render a report as portable Markdown.

Definition reports need no observations, so a workshop can review them before
any data exists. Kinds: ` + strings.Join(report.DefinitionKinds, ", ") + `.
"definitions" renders all of them as one document.

Status reports describe how far the theory has come, from observation files
(-o) as of a date (--as-of, or the latest date in the observations). Kinds:
exec, results, assumptions, interventions, evidence-gaps, full, story
(--audience), learning-review (--period), portfolio, narrative, actors, and
custom (--template, a Go text/template over the report view). narrative,
actors, and custom also work with no observations.`,
		Example: `  theoryofchange report definitions theory.yaml --out definitions.md
  theoryofchange report logic-model theory.yaml --pathway pw-supply
  theoryofchange report exec theory.yaml -o obs-q3.yaml --prev status-q2.json
  theoryofchange report story theory.yaml -o obs-q3.yaml --audience aud-funders
  theoryofchange report custom theory.yaml --template board.tmpl`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if slices.Contains(statusReportKinds, args[0]) {
				return runStatusReport(e, cmd, args[0], args[1])
			}
			kind := args[0]
			if flagBool(cmd, "diagram") {
				if kind != "logic-model" {
					return usageErr("--diagram applies to the logic-model report")
				}
				kind = "logic-model-diagram"
			}
			return render(e, cmd, args[1], false, func(doc *toc.Document) (string, error) {
				return service.DefinitionReport(doc, kind, flagString(cmd, "pathway"))
			})
		},
	}
	f := cmd.Flags()
	f.String("pathway", "", "restrict a definition report to one pathway ID")
	f.Bool("diagram", false, "draw the logic-model report as a left-to-right diagram")
	f.String("out", "", "write to this file instead of standard output")
	f.String("audience", "", "the audience ID, for the story report")
	f.String("period", "", "the period label, for the learning-review report")
	f.StringSlice("prev-observations", nil, "the earlier period's observation files, for the learning-review report")
	f.String("template", "", "the Go text/template file, for the custom report")
	statusFlags(cmd)
	asOfFlags(cmd)
	return cmd
}

// runStatusReport renders a report made from a status snapshot.
func runStatusReport(e *env, cmd *cobra.Command, kind, path string) error {
	withStatus := !slices.Contains(optionalStatus, kind) || flagString(cmd, "as-of") != "" || len(observationPaths(cmd)) > 0
	var (
		in   report.StatusInput
		sets []*toc.ObservationSet
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
		g, err := service.BuildGraph(doc)
		if err != nil {
			return renderError(path, err)
		}
		v, err := report.BuildView(doc, g, nil, report.Options{})
		if err != nil {
			return renderError(path, err)
		}
		in = report.StatusInput{View: v, Doc: doc}
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
		if err := tocio.WriteFile(out, []byte(text)); err != nil {
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

func statusReportText(cmd *cobra.Command, kind string, in report.StatusInput, sets []*toc.ObservationSet, withStatus bool) (string, error) {
	var status *toc.Status
	if withStatus {
		status = in.Status
	}
	switch kind {
	case "exec":
		return report.Exec(in), nil
	case "results":
		return report.Results(in), nil
	case "assumptions":
		return report.Assumptions(in), nil
	case "interventions":
		return report.Interventions(in), nil
	case "evidence-gaps":
		return report.EvidenceGaps(in), nil
	case "full":
		return report.Full(in, report.Mermaid(in.View, report.MermaidOptions{})), nil
	case "story":
		aud := flagString(cmd, "audience")
		if aud == "" {
			return "", usageErr("the story report needs --audience")
		}
		out, err := report.Story(in, sets, aud)
		if err != nil {
			return "", usageErr("%v", err)
		}
		return out, nil
	case "learning-review":
		prev, err := cmd.Flags().GetStringSlice("prev-observations")
		if err != nil {
			panic(fmt.Sprintf("cli: --prev-observations is not a string slice flag: %v", err))
		}
		var prevSets []*toc.ObservationSet
		if len(prev) > 0 {
			if prevSets, err = tocio.LoadObservations(prev...); err != nil {
				return "", &ExitError{Code: ExitUsageErr, Err: err}
			}
		}
		return report.LearningReview(report.LearningInput{StatusInput: in, Sets: sets, PrevSets: prevSets, Period: flagString(cmd, "period")}), nil
	case "portfolio":
		cost, err := burden.Compute(in.Doc)
		if err != nil {
			return "", &ExitError{Code: ExitInvalid, Err: err}
		}
		return report.Portfolio(in, cost), nil
	case "narrative":
		return report.Narrative(in.View, in.Doc, status), nil
	case "actors":
		return report.Actors(in.View, in.Doc, status), nil
	case "custom":
		file := flagString(cmd, "template")
		if file == "" {
			return "", usageErr("the custom report needs --template")
		}
		data, err := os.ReadFile(file)
		if err != nil {
			return "", &ExitError{Code: ExitUsageErr, Err: fmt.Errorf("read template: %w", err)}
		}
		out, err := report.RenderTemplate(string(data), in.View)
		if err != nil {
			return "", &ExitError{Code: ExitUsageErr, Err: err}
		}
		return out, nil
	}
	return "", usageErr("unknown report kind %q", kind)
}

func newExportCmd(e *env) *cobra.Command {
	export := &cobra.Command{
		Use:   "export",
		Short: "Export tables from a Theory of Change",
	}
	csv := &cobra.Command{
		Use:   "csv <table> <toc-file>",
		Short: "Export a table as CSV (RFC 4180)",
		Long: `Export a table for spreadsheets. Tables: ` + strings.Join(report.CSVTables, ", ") + `.
Multi-valued cells are joined with "; ". The first row is a comment carrying the
export stamp ("# toc:<id> version:<v> hash:<h>"); --no-stamp leaves it off.`,
		Example: `  theoryofchange export csv indicators theory.yaml --out indicators.csv`,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return render(e, cmd, args[1], true, func(doc *toc.Document) (string, error) {
				return service.ExportCSV(doc, args[0])
			})
		},
	}
	csv.Flags().String("out", "", "write to this file instead of standard output")
	asOfFlags(csv)
	export.AddCommand(csv, newExportDocCmd("docx"), newExportDocCmd("pptx"), newReviewPacketCmd(e), newReviewSheetCmd(e))
	return export
}
