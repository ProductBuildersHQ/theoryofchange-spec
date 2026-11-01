package cli

import (
	"bytes"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
	"github.com/ProductBuildersHQ/impact-model/pkg/impactio"
	"github.com/ProductBuildersHQ/impact-model/pkg/prismx"
)

func newPrismExportCmd(e *env, kind string) *cobra.Command {
	short, long := "Export OKRs for the PRISM tools", `Export the outcomes the organization answers for as an OKR document: each
outcome with measured indicators is an objective, and each of its indicators a
key result with baseline, target, current value, score, status, and
confidence. With observations (-O), the key results carry the latest values.`
	if kind == "journey" {
		short, long = "Export outcome journeys for the PRISM tools", `Export each numeric indicator as an outcome journey: where the measure stands
now, from the latest observation or the baseline, and the dated targets it is
meant to reach. With observations (-O), target confidence comes from the
status.`
	}
	cmd := &cobra.Command{
		Use:     kind + " <model-file>",
		Short:   short,
		Long:    long + "\n\nThe output is JSON.",
		Example: fmt.Sprintf("  impact-model export %s impact-model.yaml -O obs-q3.yaml -o %s.json", kind, kind),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPrismExport(e, cmd, kind, args[0])
		},
	}
	statusFlags(cmd)
	cmd.Flags().String("as-of", "", "the status date, YYYY-MM-DD; default: the latest date in the observations")
	cmd.Flags().StringP("out", "o", "", "write to this file instead of standard output")
	return cmd
}

func runPrismExport(e *env, cmd *cobra.Command, kind, path string) error {
	doc, err := loadForRender(path)
	if err != nil {
		return err
	}
	var st *impact.Status
	if len(observationPaths(cmd)) > 0 || flagString(cmd, "as-of") != "" {
		in, _, err := loadStatus(e, cmd, path)
		if err != nil {
			return err
		}
		st = in.Status
	}
	var v any
	if kind == "okr" {
		v, err = prismx.ToOKR(doc, st)
	} else {
		v, err = prismx.ToOutcomeJourneys(doc, st)
	}
	if err != nil {
		return renderError(path, err)
	}
	var buf bytes.Buffer
	if err := writeJSON(&buf, v); err != nil {
		return &ExitError{Code: ExitUsageErr, Err: err}
	}
	if out := flagString(cmd, "out"); out != "" {
		if err := impactio.WriteFile(out, buf.Bytes()); err != nil {
			return &ExitError{Code: ExitUsageErr, Err: err}
		}
		return nil
	}
	fmt.Fprint(e.out, buf.String())
	return nil
}
