package cli

import (
	"errors"
	"fmt"
	"strings"

	core "github.com/grokify/prism-core"
	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/impact-model/pkg/approve"
	"github.com/ProductBuildersHQ/impact-model/pkg/impactio"
	"github.com/ProductBuildersHQ/impact-model/pkg/service"
)

func newApproveCmd(e *env) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approve <model-file> --version <label> --date <YYYY-MM-DD> --approver <name[:role]>...",
		Short: "Record an impact model as approved",
		Long: `Mark an impact model approved: set its status to approved and its version label, and
append a revision (trigger: review) naming the approvers. The change is made
through the patch engine, in the file's own format.

Approval is refused, with exit code 1 and nothing written, when:

  - the impact model has validation errors;
  - elements are still hypotheses. Confirm or reject them, or use --force with
    --reason, which is recorded in the revision;
  - the version label is missing or already in use, the date is missing, or
    there is no approver.

Approver names are people: use a name or role you are happy to publish.`,
		Example: `  impact-model approve impact-model.yaml --version 1.0.0 --date 2026-04-01 \
    --approver "Board chair:chair" --approver "Executive director:director"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error { return runApprove(e, cmd, args[0]) },
	}
	f := cmd.Flags()
	f.String("version", "", "version label for the approved impact model (required)")
	f.String("date", "", "date of approval, YYYY-MM-DD (required)")
	f.StringSlice("approver", nil, "approver as name or name:role (repeatable, required)")
	f.String("summary", "", "what was approved")
	f.Bool("force", false, "approve despite open hypotheses (needs --reason)")
	f.String("reason", "", "why approval goes ahead with open hypotheses")
	f.Bool("dry-run", false, "do not write; print the resulting document")
	return cmd
}

func runApprove(e *env, cmd *cobra.Command, path string) error {
	format, err := outputFormat(cmd)
	if err != nil {
		return err
	}
	doc, err := impactio.LoadDocument(path)
	if err != nil {
		return &ExitError{Code: ExitUsageErr, Err: err}
	}
	names, err := cmd.Flags().GetStringSlice("approver")
	if err != nil {
		panic(fmt.Sprintf("cli: --approver is not a string slice flag: %v", err))
	}
	req := approve.Request{
		Version: flagString(cmd, "version"), Date: flagString(cmd, "date"), Summary: flagString(cmd, "summary"),
		Force: flagBool(cmd, "force"), Reason: flagString(cmd, "reason"),
	}
	for _, n := range names {
		name, role, _ := strings.Cut(n, ":")
		req.Approvers = append(req.Approvers, core.Approver{Person: core.Person{Name: strings.TrimSpace(name), Role: strings.TrimSpace(role)}})
	}
	p, err := approve.Build(doc, req)
	var refused *approve.RefusedError
	switch {
	case errors.As(err, &refused):
		return reportRefusal(e, format, path, refused)
	case err != nil:
		return &ExitError{Code: ExitUsageErr, Err: err}
	}
	res, err := service.PatchFile(path, p, service.PatchOptions{DryRun: flagBool(cmd, "dry-run")})
	if err != nil {
		return &ExitError{Code: ExitUsageErr, Err: err}
	}
	return e.reportPatch(cmd, format, res)
}

func reportRefusal(e *env, format, path string, r *approve.RefusedError) error {
	if format == "json" {
		if err := writeJSON(e.out, map[string]any{"approved": false, "refusals": r.Refusals}); err != nil {
			return &ExitError{Code: ExitUsageErr, Err: err}
		}
		return invalid()
	}
	fmt.Fprintf(e.errOut, "%s: approval refused\n", path)
	for _, x := range r.Refusals {
		fmt.Fprintf(e.errOut, "  %s\n", x.Message)
		if len(x.IDs) > 0 {
			fmt.Fprintf(e.errOut, "    %s\n", strings.Join(x.IDs, ", "))
		}
	}
	return invalid()
}
