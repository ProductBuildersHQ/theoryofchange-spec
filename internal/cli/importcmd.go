package cli

import (
	"bytes"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/review"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/service"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/tocio"
)

func newImportCmd(e *env) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Bring outside input into a review",
	}
	cmd.AddCommand(newImportReviewCmd(e), newImportObservationsCmd(e))
	return cmd
}

func newImportReviewCmd(e *env) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "review <toc-file> [response-file]",
		Short: "Turn reviewers' answers into proposed changes",
		Long: `Read what reviewers said about a theory and propose what to do about it. The
theory is never changed. The command writes:

  - a patch of the changes it could work out (--out), to be looked over and
    then applied with 'theoryofchange patch';
  - a list of the rows that need a person to decide;
  - a consultation record of the review (--consultation-out).

The answers come from a ReviewResponse file, or from a returned review sheet
with --from-csv (then --reviewer and --date are required).

How a row is handled:

  agree     the element's basis becomes confirmed
  remove    the element's basis becomes rejected
  change    when the suggestion is "field: value" lines, those fields are set;
            free text is flagged for a person
  unsure, or a comment with no decision: flagged

A row with no element ID is matched to an element by its quoted text, and
flagged if no single element matches. If the response was written against a
different version of the theory (its hash differs), every row is flagged,
though changes are still proposed.

Exit code 0 means every row became a change. Exit code 1 means a person has to
look at something: a flagged row or an out-of-date response.`,
		Example: `  theoryofchange import review theory.yaml response.json --out proposed.patch.json
  theoryofchange import review theory.yaml --from-csv sheet.csv --reviewer "program staff" --date 2026-03-01`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error { return runImportReview(e, cmd, args) },
	}
	f := cmd.Flags()
	f.String("from-csv", "", "read the answers from a returned review sheet")
	f.StringSlice("reviewer", nil, "reviewer group or role, for --from-csv (repeatable)")
	f.String("date", "", "date of the review (YYYY-MM-DD), for --from-csv")
	f.String("step", "", "consultation step to record (default whole)")
	f.String("out", "", "write the proposed patch to this file")
	f.String("consultation-out", "", "write the consultation record to this file")
	return cmd
}

func runImportReview(e *env, cmd *cobra.Command, args []string) error {
	format, err := outputFormat(cmd)
	if err != nil {
		return err
	}
	doc, err := loadForRender(args[0])
	if err != nil {
		return err
	}
	resp, err := readResponse(cmd, doc, args[1:])
	if err != nil {
		return err
	}
	step := toc.ConsultationStep(flagString(cmd, "step"))
	if step != "" && !step.Valid() {
		return usageErr("--step %q is not a consultation step", step)
	}
	res, err := service.ImportReview(doc, resp, step)
	if err != nil {
		return renderError(args[0], err)
	}
	if out := flagString(cmd, "out"); out != "" {
		if err := writeJSONFile(out, res.Patch); err != nil {
			return err
		}
	}
	if out := flagString(cmd, "consultation-out"); out != "" {
		if err := writeJSONFile(out, res.Consultation); err != nil {
			return err
		}
	}
	if format == "json" {
		if err := writeJSON(e.out, res); err != nil {
			return &ExitError{Code: ExitUsageErr, Err: err}
		}
	} else {
		printImport(e, args[0], len(resp.Rows), res)
	}
	if res.NeedsHuman() {
		return invalid()
	}
	return nil
}

// readResponse reads the response from the JSON/YAML file argument or from a
// returned sheet. Exactly one source is allowed.
func readResponse(cmd *cobra.Command, doc *toc.Document, rest []string) (*toc.ReviewResponse, error) {
	sheet := flagString(cmd, "from-csv")
	switch {
	case sheet == "" && len(rest) == 0:
		return nil, usageErr("give a response file or --from-csv")
	case sheet != "" && len(rest) > 0:
		return nil, usageErr("give a response file or --from-csv, not both")
	case sheet == "":
		return loadResponse(rest[0])
	}
	reviewers, err := cmd.Flags().GetStringSlice("reviewer")
	if err != nil {
		panic(fmt.Sprintf("cli: --reviewer is not a string slice flag: %v", err))
	}
	date := flagString(cmd, "date")
	if len(reviewers) == 0 || date == "" {
		return nil, usageErr("--from-csv needs --reviewer and --date")
	}
	data, err := os.ReadFile(sheet)
	if err != nil {
		return nil, &ExitError{Code: ExitUsageErr, Err: fmt.Errorf("read sheet: %w", err)}
	}
	resp, err := review.FromSheet(data, review.SheetMeta{Reviewers: reviewers, Date: date})
	if err != nil {
		return nil, &ExitError{Code: ExitUsageErr, Err: err}
	}
	if resp.TocID == "" {
		resp.TocID = doc.ID
	}
	return resp, nil
}

func loadResponse(path string) (*toc.ReviewResponse, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &ExitError{Code: ExitUsageErr, Err: fmt.Errorf("read response: %w", err)}
	}
	f, err := tocio.FormatFromPath(path)
	if err != nil {
		return nil, &ExitError{Code: ExitUsageErr, Err: err}
	}
	resp, err := tocio.ParseReviewResponse(data, f)
	if err != nil {
		return nil, &ExitError{Code: ExitUsageErr, Err: fmt.Errorf("%s: %w", path, err)}
	}
	return resp, nil
}

func writeJSONFile(path string, v any) error {
	var buf bytes.Buffer
	if err := writeJSON(&buf, v); err != nil {
		return &ExitError{Code: ExitUsageErr, Err: err}
	}
	if err := tocio.WriteFile(path, buf.Bytes()); err != nil {
		return &ExitError{Code: ExitUsageErr, Err: err}
	}
	return nil
}

func printImport(e *env, path string, rows int, res *review.Result) {
	fmt.Fprintf(e.out, "%s: %d row(s): %d became changes, %d need a person\n", path, rows, res.Mapped, len(res.Flags))
	if res.Stale {
		fmt.Fprintln(e.out, "warning: the theory has changed since it was reviewed; check every row")
	}
	for _, f := range res.Flags {
		target := f.ElementID
		if target == "" {
			target = "(no element)"
		}
		fmt.Fprintf(e.out, "  row %d %s: %s\n", f.Row, target, f.Reason)
	}
}
