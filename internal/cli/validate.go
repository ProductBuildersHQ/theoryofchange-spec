package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/service"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/tocio"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/validate"
)

func newValidateCmd(e *env) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate <file>...",
		Short: "Check documents against the schema and for structural problems",
		Long: `Validate Theory of Change documents, observation sets, status snapshots, and
actor registries. The kind is read from each file.

A Theory of Change is checked against its schema and for structural problems
(duplicate IDs, unresolved references, invalid enumerations, bad dates).
Pass --toc to also check observation sets against the theory they measure.
Pass --parent to also check that the given theories are children of a parent
theory and that their contributesTo references resolve there.
Pass --registry to resolve actors that live in a shared actor registry.
Pass --futures to check that the given theories, which describe competing
futures of one market, share one registry, name their futures and
perspectives, do not both speak for the same actor in the same future, and
refer to each other's elements correctly.

Exit code 1 means at least one file has errors.`,
		Example: `  theoryofchange validate theory.yaml
  theoryofchange validate observations-2026-q4.yaml --toc theory.yaml
  theoryofchange validate program-a.yaml program-b.yaml --parent organization.yaml
  theoryofchange validate future-a.yaml future-b.yaml future-c.yaml --registry actors.yaml --futures
  theoryofchange validate --format json theory.json`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			format, err := outputFormat(cmd)
			if err != nil {
				return err
			}
			var against *toc.Document
			if p := flagString(cmd, "toc"); p != "" {
				against, err = tocio.LoadDocument(p)
				if err != nil {
					return usageErr("--toc: %v", err)
				}
			}
			var opts []validate.Option
			if p := flagString(cmd, "registry"); p != "" {
				reg, err := service.LoadActorRegistry(p)
				if err != nil {
					return usageErr("--registry: %v", err)
				}
				opts = append(opts, validate.WithActorRegistry(reg))
			}
			reports := make([]service.Report, 0, len(args))
			for _, path := range args {
				r, err := service.ValidateFile(path, against, opts...)
				if err != nil {
					return &ExitError{Code: ExitUsageErr, Err: err}
				}
				reports = append(reports, r)
			}
			if flagBool(cmd, "futures") {
				fr, err := service.ValidateFutures(args...)
				if err != nil {
					return &ExitError{Code: ExitUsageErr, Err: fmt.Errorf("--futures: %w", err)}
				}
				reports = append(reports, fr)
			}
			if p := flagString(cmd, "parent"); p != "" {
				fr, err := service.ValidateFamily(p, args...)
				if err != nil {
					return &ExitError{Code: ExitUsageErr, Err: fmt.Errorf("--parent: %w", err)}
				}
				fr.Source = "parent " + p
				reports = append(reports, fr)
			}
			return e.printReports(cmd, format, reports)
		},
	}
	cmd.Flags().String("toc", "", "Theory of Change file to check observation sets against")
	cmd.Flags().String("registry", "", "shared actor registry that actor references resolve against")
	cmd.Flags().Bool("futures", false, "also check the given theories against each other as competing futures of one market")
	cmd.Flags().String("parent", "", "parent Theory of Change; also checks that the given files are its child theories")
	return cmd
}

func (e *env) printReports(cmd *cobra.Command, format string, reports []service.Report) error {
	allValid := true
	for _, r := range reports {
		allValid = allValid && r.Valid()
	}
	if format == "json" {
		if err := writeJSON(e.out, reports); err != nil {
			return &ExitError{Code: ExitUsageErr, Err: err}
		}
	} else {
		quiet := flagBool(cmd, "quiet")
		for _, r := range reports {
			for _, i := range r.Issues {
				fmt.Fprintf(e.out, "%s: %s\n", r.Source, i)
			}
			if r.Valid() && !quiet {
				fmt.Fprintf(e.out, "%s: ok (%s)\n", r.Source, kindLabel(r))
			}
		}
	}
	if !allValid {
		return invalid()
	}
	return nil
}

func kindLabel(r service.Report) string {
	if r.Kind == "" {
		return "unknown kind"
	}
	return string(r.Kind)
}
