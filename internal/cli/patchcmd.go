package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/impact-model/pkg/edit"
	"github.com/ProductBuildersHQ/impact-model/pkg/impactio"
	"github.com/ProductBuildersHQ/impact-model/pkg/service"
)

func newPatchCmd(e *env) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "patch <model-file> <patch-file|->",
		Short: "Apply a JSON Patch to an impact model",
		Long: `Apply an RFC 6902 JSON Patch to an impact model, atomically.

Pointers may address array elements by ID: /outcomes/@oc-adoption/title
selects the outcome with id oc-adoption, and /outcomes/- appends. The patch
is applied to a copy. If any operation fails, the result no longer conforms
to the schema, or the patch would leave the document with more structural
errors than before, nothing is written. The file is rewritten in canonical
form in its own format (JSON or YAML).

Read the patch from standard input with "-". With --dry-run, nothing is
written; the resulting document is printed to stdout and the summary to
stderr.

Exit codes: 0 applied (or nothing to change); 1 rejected; 2 the patch could
not be applied (bad path, unknown ID, failed test, or unreadable input).`,
		Example: `  echo '[{"op":"replace","path":"/outcomes/@oc-skills/title","value":"Staff write strong applications"}]' \
    | impact-model patch impact-model.yaml -
  impact-model patch impact-model.yaml changes.json --dry-run`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			format, err := outputFormat(cmd)
			if err != nil {
				return err
			}
			raw, err := readPatchSource(e, args[1])
			if err != nil {
				return &ExitError{Code: ExitUsageErr, Err: err}
			}
			p, err := edit.ParsePatch(raw)
			if err != nil {
				return &ExitError{Code: ExitUsageErr, Err: fmt.Errorf("%s: %w", args[1], err)}
			}
			res, err := service.PatchFile(args[0], p, service.PatchOptions{
				DryRun:      flagBool(cmd, "dry-run"),
				AllowErrors: flagBool(cmd, "allow-errors"),
			})
			if err != nil {
				return &ExitError{Code: ExitUsageErr, Err: err}
			}
			return e.reportPatch(cmd, format, res)
		},
	}
	cmd.Flags().Bool("dry-run", false, "do not write; print the resulting document")
	cmd.Flags().Bool("allow-errors", false, "accept a patch that adds structural errors (never a schema violation)")
	return cmd
}

func readPatchSource(e *env, src string) ([]byte, error) {
	if src == "-" {
		data, err := io.ReadAll(io.LimitReader(e.in, impactio.MaxFileBytes+1))
		if err != nil {
			return nil, fmt.Errorf("read patch from stdin: %w", err)
		}
		if len(data) > impactio.MaxFileBytes {
			return nil, fmt.Errorf("%w: patch on stdin exceeds %d bytes", impactio.ErrTooLarge, impactio.MaxFileBytes)
		}
		return data, nil
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return nil, fmt.Errorf("read patch %s: %w", src, err)
	}
	return data, nil
}

func (e *env) reportPatch(cmd *cobra.Command, format string, res *service.PatchResult) error {
	if format == "json" {
		out := struct {
			*service.PatchResult
			Document string `json:"document,omitempty"`
		}{PatchResult: res, Document: string(res.Document)}
		if err := writeJSON(e.out, out); err != nil {
			return &ExitError{Code: ExitUsageErr, Err: err}
		}
	} else {
		if res.Applied && res.DryRun {
			if _, err := e.out.Write(res.Document); err != nil {
				return &ExitError{Code: ExitUsageErr, Err: fmt.Errorf("write document: %w", err)}
			}
		}
		e.patchSummary(cmd, res)
	}
	if !res.Applied {
		return invalid()
	}
	return nil
}

func (e *env) patchSummary(cmd *cobra.Command, res *service.PatchResult) {
	switch {
	case !res.Applied:
		fmt.Fprintf(e.errOut, "%s: patch rejected: %s\n", res.Path, res.Reason)
	case res.DryRun && res.Changed:
		fmt.Fprintf(e.errOut, "%s: would apply (dry run, nothing written)\n", res.Path)
	case res.DryRun:
		fmt.Fprintf(e.errOut, "%s: no change (dry run)\n", res.Path)
	case res.Written:
		if !flagBool(cmd, "quiet") {
			fmt.Fprintf(e.errOut, "%s: patch applied\n", res.Path)
		}
	default:
		if !flagBool(cmd, "quiet") {
			fmt.Fprintf(e.errOut, "%s: no change\n", res.Path)
		}
	}
	for _, i := range res.IssuesIntroduced {
		fmt.Fprintf(e.errOut, "  introduced: %s\n", i)
	}
	for _, i := range res.IssuesResolved {
		fmt.Fprintf(e.errOut, "  resolved:   %s\n", i)
	}
}
