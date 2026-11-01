package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/plexusone/structured-evaluation/rubric"
	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/impact-model/pkg/judge"
	"github.com/ProductBuildersHQ/impact-model/spec"
)

func newJudgeCmd(e *env) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "judge",
		Short: "Prepare an impact model for an LLM-as-a-Judge and combine the verdicts",
		Long: `Support judging an impact model against a rubric. This command makes no model calls: a
judge, human or agent, reads the bundle and returns a verdict.

  judge show     print the bundle: rubric, document, narrative, lint report,
                 instructions, and a hash of all of them
  judge record   check a verdict against the bundle and store it beside the impact model
  judge summary  combine the stored verdicts: per-category scores, agreement,
                 and whether the impact model passes

A verdict names the bundle hash it judged. If the impact model has changed since, the
hash differs and the verdict is rejected.`,
	}
	cmd.PersistentFlags().String("profile", spec.ProfileCore, "rubric profile: "+strings.Join(spec.Profiles(), " or "))
	cmd.AddCommand(newJudgeShowCmd(e), newJudgeRecordCmd(e), newJudgeSummaryCmd(e))
	return cmd
}

func judgeBundle(cmd *cobra.Command, path string) (*judge.Bundle, error) {
	doc, err := loadForRender(path)
	if err != nil {
		return nil, err
	}
	b, err := judge.Build(doc, flagString(cmd, "profile"))
	if err != nil {
		if errors.Is(err, spec.ErrUnknownProfile) {
			return nil, usageErr("%v", err)
		}
		return nil, renderError(path, err)
	}
	return b, nil
}

func newJudgeShowCmd(e *env) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show <model-file>",
		Short:   "Print the judge bundle",
		Example: "  impact-model judge show impact-model.yaml --profile development --out bundle.json",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			b, err := judgeBundle(cmd, args[0])
			if err != nil {
				return err
			}
			if out := flagString(cmd, "out"); out != "" {
				return writeJSONFile(out, b)
			}
			if err := writeJSON(e.out, b); err != nil {
				return &ExitError{Code: ExitUsageErr, Err: err}
			}
			return nil
		},
	}
	cmd.Flags().StringP("out", "o", "", "write the bundle to this file instead of standard output")
	return cmd
}

func newJudgeRecordCmd(e *env) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "record <model-file> <verdict-file>",
		Short: "Check a verdict and store it beside the impact model",
		Long: `Check a verdict against the current bundle and store it under
.toc/judgements/<bundle-hash>/<judge>.json beside the impact model file.

The verdict is rejected, with exit code 1, if it judged a different version of
the impact model, is not a valid evaluation report, scores a category the rubric does
not have, or leaves a required category unscored.`,
		Example: "  impact-model judge record impact-model.yaml verdict.json --judge model-a",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			b, err := judgeBundle(cmd, args[0])
			if err != nil {
				return err
			}
			data, err := os.ReadFile(args[1])
			if err != nil {
				return &ExitError{Code: ExitUsageErr, Err: fmt.Errorf("read verdict: %w", err)}
			}
			v, err := judge.ParseVerdict(data)
			if err == nil {
				var path string
				if path, err = judge.Record(args[0], b, v, flagString(cmd, "judge")); err == nil {
					fmt.Fprintf(e.out, "recorded %s\n", path)
					return nil
				}
			}
			if errors.Is(err, judge.ErrJudgeName) {
				return usageErr("%v", err)
			}
			fmt.Fprintf(e.errOut, "%s: %v\n", args[1], err)
			return invalid()
		},
	}
	cmd.Flags().String("judge", "", "name to store the verdict under (default: the report's judge ID)")
	return cmd
}

func newJudgeSummaryCmd(e *env) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "summary <model-file>",
		Short: "Combine the stored verdicts",
		Long: `Combine every verdict recorded for the current version of the impact model: per
category, how many judges gave each score and the combined score; how closely
the judges agree; and whether the impact model passes the rubric.

Aggregation is one of majority (a tie goes to partial), conservative (the
lowest score), optimistic (the highest), or unanimous (anything else is
partial). The impact model passes when every required category passes, no blocking
category fails, and the findings stay within the rubric's limits.

Exit code 1 means the impact model does not pass.`,
		Example: "  impact-model judge summary impact-model.yaml --aggregation majority",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			format, err := outputFormat(cmd)
			if err != nil {
				return err
			}
			b, err := judgeBundle(cmd, args[0])
			if err != nil {
				return err
			}
			verdicts, names, err := judge.Load(args[0], b)
			if err != nil {
				return &ExitError{Code: ExitUsageErr, Err: err}
			}
			s, err := judge.Summarize(b, names, verdicts, rubric.AggregationMethod(flagString(cmd, "aggregation")))
			if err != nil {
				return usageErr("%v", err)
			}
			if format == "json" {
				if err := writeJSON(e.out, s); err != nil {
					return &ExitError{Code: ExitUsageErr, Err: err}
				}
			} else {
				printJudgeSummary(e, s)
			}
			if !s.Pass {
				return invalid()
			}
			return nil
		},
	}
	cmd.Flags().String("aggregation", string(rubric.AggregationMajority), "majority, conservative, optimistic, or unanimous")
	return cmd
}

func printJudgeSummary(e *env, s *judge.Summary) {
	fmt.Fprintf(e.out, "%d judge(s): %s · %s rubric · %s\n", len(s.Judges), strings.Join(s.Judges, ", "), s.Profile, s.Aggregation)
	for _, c := range s.Categories {
		mark := ""
		if !c.Agreed {
			mark = "  (judges differ)"
		}
		fmt.Fprintf(e.out, "  %-22s %-8s pass %d · partial %d · fail %d%s\n", c.Category, c.Result,
			c.Scores["pass"], c.Scores["partial"], c.Scores["fail"], mark)
	}
	fmt.Fprintf(e.out, "agreement %.0f%%\n", s.Agreement*100)
	if s.Pass {
		fmt.Fprintln(e.out, "PASS")
		return
	}
	fmt.Fprintln(e.out, "DOES NOT PASS")
	for _, r := range s.Reasons {
		fmt.Fprintf(e.out, "  %s\n", r)
	}
}
