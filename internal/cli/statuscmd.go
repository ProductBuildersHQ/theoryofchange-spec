package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/measure"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/service"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/tabular"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/tocio"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/validate"
)

func newStatusCmd(e *env) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status <toc-file> -o <observations>...",
		Short: "Compute how far a theory has come, as of a date",
		Long: `Compute a status snapshot: for each indicator its progress against what the
targets call for by the date; for each outcome its own state and its pathway
state (what leads to it); for each link and assumption its state; and a
confidence score for the routes to impact.

The date is --as-of, or by default the latest date in the observations. The
snapshot is a pure function of the theory, the observations, and the date.

By default a summary is printed. With --format json the snapshot is printed as
JSON. --out writes the snapshot as JSON to a file whatever the format. With
--prev, the changes since an earlier snapshot are listed too.

Exit code 1 means the theory or the observations are invalid.`,
		Example: `  theoryofchange status theory.yaml -o obs-q3.yaml --as-of 2026-09-30 --out status-q3.json
  theoryofchange status theory.yaml -o obs-q3.yaml -o obs-q4.yaml --prev status-q3.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error { return runStatus(e, cmd, args[0]) },
	}
	f := cmd.Flags()
	f.StringSliceP("observations", "o", nil, "observation set files (repeatable)")
	f.String("as-of", "", "the date to describe (YYYY-MM-DD); default: the latest date in the observations")
	f.String("prev", "", "an earlier status snapshot (JSON) to compare with")
	f.String("out", "", "write the snapshot as JSON to this file")
	f.Float64("tolerance", measure.DefaultTolerance, "how far behind expected progress still counts as on track")
	f.String("rollup", string(toc.RollupWorst), "how indicators combine into an outcome: worst or mean")
	return cmd
}

func runStatus(e *env, cmd *cobra.Command, path string) error {
	format, err := outputFormat(cmd)
	if err != nil {
		return err
	}
	opts := toc.DefaultStatusOptions()
	opts.Tolerance, _ = cmd.Flags().GetFloat64("tolerance")
	opts.Rollup = toc.RollupMethod(flagString(cmd, "rollup"))
	if !opts.Rollup.Valid() || opts.Tolerance <= 0 || opts.Tolerance >= 1 {
		return usageErr("--rollup must be worst or mean, and --tolerance between 0 and 1")
	}
	doc, err := loadForRender(path)
	if err != nil {
		return err
	}
	g, err := service.BuildGraph(doc)
	if err != nil {
		return renderError(path, err)
	}
	obsPaths, err := cmd.Flags().GetStringSlice("observations")
	if err != nil {
		panic(fmt.Sprintf("cli: --observations is not a string slice flag: %v", err))
	}
	var sets []*toc.ObservationSet
	if len(obsPaths) > 0 {
		if sets, err = tocio.LoadObservations(obsPaths...); err != nil {
			return &ExitError{Code: ExitUsageErr, Err: err}
		}
		if issues := validate.Observations(doc, sets...); validate.HasErrors(issues) {
			for _, i := range issues {
				if i.Severity == validate.SeverityError {
					fmt.Fprintf(e.errOut, "observations: %s %s: %s\n", i.Code, i.Path, i.Message)
				}
			}
			return invalid()
		}
	}
	asOf, err := statusDate(cmd, sets)
	if err != nil {
		return err
	}
	st, err := measure.Evaluate(doc, g, sets, asOf, opts)
	if err != nil {
		return &ExitError{Code: ExitUsageErr, Err: err}
	}

	var diff *toc.StatusDiff
	if prevPath := flagString(cmd, "prev"); prevPath != "" {
		if diff, err = diffAgainst(prevPath, st); err != nil {
			return err
		}
	}
	if out := flagString(cmd, "out"); out != "" {
		if err := writeJSONFile(out, st); err != nil {
			return err
		}
	}
	switch {
	case format == "json" && diff != nil:
		err = writeJSON(e.out, map[string]any{"status": st, "diff": diff})
	case format == "json":
		err = writeJSON(e.out, st)
	default:
		printStatus(e, path, st, diff)
	}
	if err != nil {
		return &ExitError{Code: ExitUsageErr, Err: err}
	}
	return nil
}

func statusDate(cmd *cobra.Command, sets []*toc.ObservationSet) (time.Time, error) {
	if s := flagString(cmd, "as-of"); s != "" {
		at, err := time.Parse(time.DateOnly, s)
		if err != nil {
			return time.Time{}, usageErr("--as-of must be a date like 2026-09-30: %v", err)
		}
		return at, nil
	}
	if at, ok := measure.LatestDate(sets); ok {
		return at, nil
	}
	return time.Time{}, usageErr("give --as-of: there are no observations to take a date from")
}

func diffAgainst(prevPath string, cur *toc.Status) (*toc.StatusDiff, error) {
	data, err := os.ReadFile(prevPath)
	if err != nil {
		return nil, &ExitError{Code: ExitUsageErr, Err: fmt.Errorf("read previous snapshot: %w", err)}
	}
	prev, err := tocio.ParseStatus(data, tocio.FormatJSON)
	if err != nil {
		return nil, &ExitError{Code: ExitUsageErr, Err: fmt.Errorf("%s: %w", prevPath, err)}
	}
	if prev.TocID != cur.TocID {
		return nil, usageErr("%s is a snapshot of %q, not %q", prevPath, prev.TocID, cur.TocID)
	}
	return measure.Diff(prev, cur), nil
}

func printStatus(e *env, path string, st *toc.Status, diff *toc.StatusDiff) {
	fmt.Fprintf(e.out, "%s as of %s\n", path, st.AsOf)
	counts := map[toc.State]int{}
	for _, i := range st.Indicators {
		counts[i.State]++
	}
	fmt.Fprintf(e.out, "indicators (%d): %s\n", len(st.Indicators), stateCounts(counts))
	pathway := map[toc.State]int{}
	for _, o := range st.Outcomes {
		pathway[o.PathwayState]++
	}
	fmt.Fprintf(e.out, "outcomes (%d), by pathway state: %s\n", len(st.Outcomes), stateCounts(pathway))
	for _, i := range st.Overall {
		conf := "n/a"
		if i.Confidence != nil {
			conf = fmt.Sprintf("%.2f", *i.Confidence)
		}
		fmt.Fprintf(e.out, "impact %s: %s (confidence %s)\n", i.OutcomeID, i.PathwayState, conf)
	}
	for _, l := range st.Links {
		if l.State != toc.LinkStateOK {
			fmt.Fprintf(e.out, "link %s is %s\n", l.LinkID, l.State)
		}
	}
	if diff != nil {
		fmt.Fprintf(e.out, "changes since %s: %d\n", diff.FromAsOf, len(diff.Changes))
		for _, c := range diff.Changes {
			fmt.Fprintf(e.out, "  %s %s %s: %s -> %s\n", c.ElementKind, c.ElementID, c.Field, absent(c.From), absent(c.To))
		}
	}
}

func absent(s string) string {
	if s == "" {
		return "(none)"
	}
	return s
}

func stateCounts(c map[toc.State]int) string {
	var parts []string
	for _, s := range toc.AllStates() {
		if c[s] > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", c[s], s))
		}
	}
	if len(parts) == 0 {
		return "none"
	}
	out := parts[0]
	for _, p := range parts[1:] {
		out += ", " + p
	}
	return out
}

// ---- import observations ----------------------------------------------------

func newImportObservationsCmd(e *env) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "observations --csv <file> --map <mapping> --toc <toc-file>",
		Short: "Turn a spreadsheet into an observation set",
		Long: `Read a CSV sheet through a column mapping and write an observation set. Every
row either becomes an observation or is reported as unmatched with its row
number and the reason; none is dropped silently. Unmatched rows are listed on
standard error, and --unmatched writes them to a sheet you can fix and import
again.

The observation set goes to --out (YAML or JSON by extension) or to standard
output as YAML. Exit code 0 means every row was imported; 1 means some rows
were unmatched (the rest are still written); 2 means an error.`,
		Example: `  theoryofchange import observations --csv q3.csv --map mapping.yaml --toc theory.yaml --out obs-q3.yaml --unmatched q3-unmatched.csv`,
		Args:    cobra.NoArgs,
		RunE:    func(cmd *cobra.Command, _ []string) error { return runImportObservations(e, cmd) },
	}
	f := cmd.Flags()
	f.String("csv", "", "the spreadsheet to read (required)")
	f.String("map", "", "the column mapping file (required)")
	f.String("toc", "", "the theory the sheet measures (required)")
	f.String("out", "", "write the observation set here (default: standard output, YAML)")
	f.String("unmatched", "", "write the unmatched rows to this CSV file")
	for _, name := range []string{"csv", "map", "toc"} {
		if err := cmd.MarkFlagRequired(name); err != nil {
			panic(fmt.Sprintf("cli: mark --%s required: %v", name, err))
		}
	}
	return cmd
}

func runImportObservations(e *env, cmd *cobra.Command) error {
	doc, err := loadForRender(flagString(cmd, "toc"))
	if err != nil {
		return err
	}
	mapData, err := os.ReadFile(flagString(cmd, "map"))
	if err != nil {
		return &ExitError{Code: ExitUsageErr, Err: fmt.Errorf("read mapping: %w", err)}
	}
	mf, err := tocio.FormatFromPath(flagString(cmd, "map"))
	if err != nil {
		return &ExitError{Code: ExitUsageErr, Err: err}
	}
	m, err := tocio.ParseCSVMapping(mapData, mf)
	if err != nil {
		return &ExitError{Code: ExitUsageErr, Err: fmt.Errorf("%s: %w", flagString(cmd, "map"), err)}
	}
	f, err := os.Open(flagString(cmd, "csv"))
	if err != nil {
		return &ExitError{Code: ExitUsageErr, Err: fmt.Errorf("open sheet: %w", err)}
	}
	res, err := tabular.Import(f, m, doc)
	if cerr := f.Close(); cerr != nil && err == nil {
		err = fmt.Errorf("close sheet: %w", cerr)
	}
	if err != nil {
		return &ExitError{Code: ExitUsageErr, Err: err}
	}

	format := tocio.FormatYAML
	if out := flagString(cmd, "out"); out != "" {
		if format, err = tocio.FormatFromPath(out); err != nil {
			return &ExitError{Code: ExitUsageErr, Err: err}
		}
	}
	data, err := tocio.Canonical(res.Set, format)
	if err != nil {
		return &ExitError{Code: ExitUsageErr, Err: err}
	}
	if out := flagString(cmd, "out"); out != "" {
		if err := tocio.WriteFile(out, data); err != nil {
			return &ExitError{Code: ExitUsageErr, Err: err}
		}
	} else if _, err := e.out.Write(data); err != nil {
		return &ExitError{Code: ExitUsageErr, Err: fmt.Errorf("write observations: %w", err)}
	}

	fmt.Fprintf(e.errOut, "%d row(s): %d imported, %d unmatched, %d blank\n", res.Rows(), res.Imported, len(res.Unmatched), res.Blank)
	for _, u := range res.Unmatched {
		fmt.Fprintf(e.errOut, "  row %d: %s\n", u.Row, u.Reason)
	}
	if path := flagString(cmd, "unmatched"); path != "" && len(res.Unmatched) > 0 {
		sheet, err := tabular.UnmatchedCSV(res)
		if err != nil {
			return &ExitError{Code: ExitUsageErr, Err: err}
		}
		if err := tocio.WriteFile(path, []byte(sheet)); err != nil {
			return &ExitError{Code: ExitUsageErr, Err: err}
		}
	}
	if len(res.Unmatched) > 0 {
		return invalid()
	}
	return nil
}
