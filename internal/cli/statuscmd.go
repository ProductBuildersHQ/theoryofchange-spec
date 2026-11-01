package cli

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/impact-model/pkg/graph"
	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
	"github.com/ProductBuildersHQ/impact-model/pkg/impactio"
	"github.com/ProductBuildersHQ/impact-model/pkg/measure"
	"github.com/ProductBuildersHQ/impact-model/pkg/report"
	"github.com/ProductBuildersHQ/impact-model/pkg/service"
	"github.com/ProductBuildersHQ/impact-model/pkg/tabular"
	"github.com/ProductBuildersHQ/impact-model/pkg/validate"
)

func newStatusCmd(e *env) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status <model-file> -O <observations>...",
		Short: "Compute how far an impact model has come, as of a date",
		Long: `Compute a status snapshot: for each indicator its progress against what the
targets call for by the date; for each outcome its own state and its pathway
state (what leads to it); for each link and assumption its state; and a
confidence score for the routes to impact.

The date is --as-of, or by default the latest date in the observations. The
snapshot is a pure function of the impact model, the observations, and the date.

By default a summary is printed. With --format json the snapshot is printed as
JSON. --out writes the snapshot as JSON to a file whatever the format. With
--prev, the changes since an earlier snapshot are listed too.

Exit code 1 means the impact model or the observations are invalid.`,
		Example: `  impact-model status impact-model.yaml -O obs-q3.yaml --as-of 2026-09-30 --out status-q3.json
  impact-model status impact-model.yaml -O obs-q3.yaml -O obs-q4.yaml --prev status-q3.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error { return runStatus(e, cmd, args[0]) },
	}
	statusFlags(cmd)
	cmd.Flags().String("as-of", "", "the date to describe (YYYY-MM-DD); default: the latest date in the observations")
	cmd.Flags().StringP("out", "o", "", "write the snapshot as JSON to this file")
	return cmd
}

// statusFlags adds the flags that choose what a status is computed from.
func statusFlags(cmd *cobra.Command) {
	f := cmd.Flags()
	f.StringSliceP("observations", "O", nil, "observation set files (repeatable)")
	f.String("prev", "", "an earlier status snapshot (JSON) to compare with")
	f.Float64("tolerance", measure.DefaultTolerance, "how far behind expected progress still counts as on track")
	f.String("rollup", string(impact.RollupWorst), "how indicators combine into an outcome: worst or mean")
}

// loadStatus evaluates the impact model at path against the observation files and
// date given by the flags. It reports observation problems on e.errOut and
// returns exit code 1 for them. The date is --as-of, else the latest date in
// the observations.
func loadStatus(e *env, cmd *cobra.Command, path string) (report.StatusInput, []*impact.ObservationSet, error) {
	opts := impact.DefaultStatusOptions()
	opts.Tolerance, _ = cmd.Flags().GetFloat64("tolerance")
	opts.Rollup = impact.RollupMethod(flagString(cmd, "rollup"))
	if !opts.Rollup.Valid() || opts.Tolerance <= 0 || opts.Tolerance >= 1 {
		return report.StatusInput{}, nil, usageErr("--rollup must be worst or mean, and --tolerance between 0 and 1")
	}
	doc, err := loadForRender(path)
	if err != nil {
		return report.StatusInput{}, nil, err
	}
	obsPaths, err := cmd.Flags().GetStringSlice("observations")
	if err != nil {
		panic(fmt.Sprintf("cli: --observations is not a string slice flag: %v", err))
	}
	var sets []*impact.ObservationSet
	if len(obsPaths) > 0 {
		if sets, err = impactio.LoadObservations(obsPaths...); err != nil {
			return report.StatusInput{}, nil, &ExitError{Code: ExitUsageErr, Err: err}
		}
		if issues := validate.Observations(doc, sets...); validate.HasErrors(issues) {
			for _, i := range issues {
				if i.Severity == validate.SeverityError {
					fmt.Fprintf(e.errOut, "observations: %s %s: %s\n", i.Code, i.Path, i.Message)
				}
			}
			return report.StatusInput{}, nil, invalid()
		}
	}
	asOf, err := statusDate(cmd, sets)
	if err != nil {
		return report.StatusInput{}, nil, err
	}
	var prev *impact.Status
	if prevPath := flagString(cmd, "prev"); prevPath != "" {
		if prev, err = loadSnapshot(prevPath); err != nil {
			return report.StatusInput{}, nil, err
		}
	}
	in, err := service.StatusInputFor(doc, sets, asOf, opts, prev)
	if err != nil {
		if errors.Is(err, graph.ErrInvalidDocument) {
			return report.StatusInput{}, nil, renderError(path, err)
		}
		return report.StatusInput{}, nil, &ExitError{Code: ExitUsageErr, Err: err}
	}
	return in, sets, nil
}

func runStatus(e *env, cmd *cobra.Command, path string) error {
	format, err := outputFormat(cmd)
	if err != nil {
		return err
	}
	in, _, err := loadStatus(e, cmd, path)
	if err != nil {
		return err
	}
	st, diff := in.Status, in.Diff
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

func statusDate(cmd *cobra.Command, sets []*impact.ObservationSet) (time.Time, error) {
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

func loadSnapshot(prevPath string) (*impact.Status, error) {
	data, err := os.ReadFile(prevPath)
	if err != nil {
		return nil, &ExitError{Code: ExitUsageErr, Err: fmt.Errorf("read previous snapshot: %w", err)}
	}
	prev, err := impactio.ParseStatus(data, impactio.FormatJSON)
	if err != nil {
		return nil, &ExitError{Code: ExitUsageErr, Err: fmt.Errorf("%s: %w", prevPath, err)}
	}
	return prev, nil
}

func printStatus(e *env, path string, st *impact.Status, diff *impact.StatusDiff) {
	fmt.Fprintf(e.out, "%s as of %s\n", path, st.AsOf)
	counts := map[impact.State]int{}
	for _, i := range st.Indicators {
		counts[i.State]++
	}
	fmt.Fprintf(e.out, "indicators (%d): %s\n", len(st.Indicators), stateCounts(counts))
	pathway := map[impact.State]int{}
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
		if l.State != impact.LinkStateOK {
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

func stateCounts(c map[impact.State]int) string {
	var parts []string
	for _, s := range impact.AllStates() {
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
		Use:   "observations --csv <file> --map <mapping> --model <model-file>",
		Short: "Turn a spreadsheet into an observation set",
		Long: `Read a CSV sheet through a column mapping and write an observation set. Every
row either becomes an observation or is reported as unmatched with its row
number and the reason; none is dropped silently. Unmatched rows are listed on
standard error, and --unmatched writes them to a sheet you can fix and import
again.

The observation set goes to --out (YAML or JSON by extension) or to standard
output as YAML. Exit code 0 means every row was imported; 1 means some rows
were unmatched (the rest are still written); 2 means an error.`,
		Example: `  impact-model import observations --csv q3.csv --map mapping.yaml --model impact-model.yaml --out obs-q3.yaml --unmatched q3-unmatched.csv`,
		Args:    cobra.NoArgs,
		RunE:    func(cmd *cobra.Command, _ []string) error { return runImportObservations(e, cmd) },
	}
	f := cmd.Flags()
	f.String("csv", "", "the spreadsheet to read (required)")
	f.String("map", "", "the column mapping file (required)")
	f.String("model", "", "the impact model the sheet measures (required)")
	f.StringP("out", "o", "", "write the observation set here (default: standard output, YAML)")
	f.String("unmatched", "", "write the unmatched rows to this CSV file")
	for _, name := range []string{"csv", "map", "model"} {
		if err := cmd.MarkFlagRequired(name); err != nil {
			panic(fmt.Sprintf("cli: mark --%s required: %v", name, err))
		}
	}
	return cmd
}

func runImportObservations(e *env, cmd *cobra.Command) error {
	doc, err := loadForRender(flagString(cmd, "model"))
	if err != nil {
		return err
	}
	mapData, err := os.ReadFile(flagString(cmd, "map"))
	if err != nil {
		return &ExitError{Code: ExitUsageErr, Err: fmt.Errorf("read mapping: %w", err)}
	}
	mf, err := impactio.FormatFromPath(flagString(cmd, "map"))
	if err != nil {
		return &ExitError{Code: ExitUsageErr, Err: err}
	}
	m, err := impactio.ParseCSVMapping(mapData, mf)
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

	format := impactio.FormatYAML
	if out := flagString(cmd, "out"); out != "" {
		if format, err = impactio.FormatFromPath(out); err != nil {
			return &ExitError{Code: ExitUsageErr, Err: err}
		}
	}
	data, err := impactio.Canonical(res.Set, format)
	if err != nil {
		return &ExitError{Code: ExitUsageErr, Err: err}
	}
	if out := flagString(cmd, "out"); out != "" {
		if err := impactio.WriteFile(out, data); err != nil {
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
		if err := impactio.WriteFile(path, []byte(sheet)); err != nil {
			return &ExitError{Code: ExitUsageErr, Err: err}
		}
	}
	if len(res.Unmatched) > 0 {
		return invalid()
	}
	return nil
}
