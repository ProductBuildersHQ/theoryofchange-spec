package cli

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/lint"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/service"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/share"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/validate"
)

func newLintCmd(e *env) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lint <toc-file>",
		Short: "Check the quality of a theory with deterministic rules",
		Long: `Check a theory against the lint rules. Validation says a document is well
formed; lint says whether the theory is sound.

Profiles choose how much is reported:

  draft     errors and structural rules; a small valid theory is GO
  standard  also the warnings (the default)
  full      also the advice

The result is GO unless a rule reports an error (exit code 1). With --format
json the output is a structured-evaluation summary report. The theory must
validate first; a document that does not is exit code 1.

Use --list to print the rules instead of running them.`,
		Example: `  theoryofchange lint theory.yaml
  theoryofchange lint theory.yaml --profile draft
  theoryofchange lint theory.yaml --disable TOC-PRV-001,TOC-AUD-001 --format json
  theoryofchange lint --list`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error { return runLint(e, cmd, args) },
	}
	f := cmd.Flags()
	f.String("profile", string(lint.ProfileStandard), "draft, standard, or full")
	f.StringSlice("disable", nil, "rule IDs to skip (comma separated)")
	f.String("as-of", "", "date for time-based rules and the report (YYYY-MM-DD)")
	f.Bool("list", false, "list the rules and exit")
	return cmd
}

func runLint(e *env, cmd *cobra.Command, args []string) error {
	if flagBool(cmd, "list") {
		listRules(e)
		return nil
	}
	if len(args) != 1 {
		return usageErr("give a theory file, or --list")
	}
	format, err := outputFormat(cmd)
	if err != nil {
		return err
	}
	asOf := flagString(cmd, "as-of")
	var at time.Time
	if asOf != "" {
		if at, err = time.Parse(time.DateOnly, asOf); err != nil {
			return usageErr("--as-of must be a date like 2026-03-01: %v", err)
		}
	}
	disable, err := cmd.Flags().GetStringSlice("disable")
	if err != nil {
		panic(fmt.Sprintf("cli: --disable is not a string slice flag: %v", err))
	}
	doc, err := loadForRender(args[0])
	if err != nil {
		return err
	}
	g, err := service.BuildGraph(doc)
	if err != nil {
		return renderError(args[0], err)
	}
	res, err := lint.Run(doc, g, lint.Options{Profile: lint.Profile(flagString(cmd, "profile")), Disable: disable, AsOf: asOf})
	if errors.Is(err, lint.ErrUnknownProfile) || errors.Is(err, lint.ErrUnknownRule) {
		return usageErr("%v", err)
	}
	if err != nil {
		return &ExitError{Code: ExitUsageErr, Err: err}
	}
	st, err := share.NewStamp(doc, asOf)
	if err != nil {
		return &ExitError{Code: ExitUsageErr, Err: err}
	}
	if format == "json" {
		if err := writeJSON(e.out, res.SummaryReport(doc.ID, st.Version, at)); err != nil {
			return &ExitError{Code: ExitUsageErr, Err: err}
		}
	} else {
		printLint(e, args[0], res)
	}
	if !res.Go() {
		return invalid()
	}
	return nil
}

func printLint(e *env, path string, res *lint.Result) {
	for _, f := range res.Findings {
		where := f.Element
		if where == "" {
			where = f.Path
		}
		fmt.Fprintf(e.out, "%s: %-7s %s %s: %s\n", path, f.Severity, f.Rule, where, f.Message)
	}
	verdict := "GO"
	if !res.Go() {
		verdict = "NO-GO"
	}
	fmt.Fprintf(e.out, "%s (%s profile): %d error(s), %d warning(s), %d note(s) from %d rule(s)\n", verdict, res.Profile,
		res.Count(validate.SeverityError), res.Count(validate.SeverityWarning), res.Count(validate.SeverityInfo), len(res.Ran))
}

func listRules(e *env) {
	for _, r := range lint.Default.Rules() {
		from := lint.ProfileFull
		for _, p := range lint.Profiles() {
			if lint.Runs(r, p) {
				from = p
				break
			}
		}
		fmt.Fprintf(e.out, "%s  %-7s  from %-8s  %s\n", r.ID(), r.Severity(), from, strings.TrimSpace(r.Description()))
	}
}
