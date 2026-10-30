// Package cli builds the theoryofchange command line. It is a thin adapter:
// every operation lives in pkg/service.
//
// Exit codes: 0 success; 1 the input is invalid or a change was rejected;
// 2 usage or I/O error; 3 an external tool needed for export is missing. Results go to stdout, diagnostics to stderr.
package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

// Version is the CLI version, set at build time with -ldflags.
var Version = "dev"

// Exit codes.
const (
	ExitOK          = 0
	ExitInvalid     = 1
	ExitUsageErr    = 2
	ExitToolMissing = 3
)

// ExitError ends the command with a specific exit code. A nil Err means the
// command already reported the problem.
type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("exit status %d", e.Code)
	}
	return e.Err.Error()
}

func (e *ExitError) Unwrap() error { return e.Err }

func invalid() error { return &ExitError{Code: ExitInvalid} }

func usageErr(format string, args ...any) error {
	return &ExitError{Code: ExitUsageErr, Err: fmt.Errorf(format, args...)}
}

// env carries the command's standard streams.
type env struct {
	in     io.Reader
	out    io.Writer
	errOut io.Writer
}

// NewRootCmd builds the command tree.
func NewRootCmd(in io.Reader, out, errOut io.Writer) *cobra.Command {
	e := &env{in: in, out: out, errOut: errOut}
	root := &cobra.Command{
		Use:   "theoryofchange",
		Short: "Capture, check, edit, and report on Theories of Change",
		Long: `theoryofchange works with Theory of Change documents (JSON or YAML) and the
observation sets that measure them.

Exit codes: 0 success; 1 the input is invalid or a change was rejected;
2 usage or I/O error.`,
		Version:       Version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetIn(in)
	root.SetOut(out)
	root.SetErr(errOut)
	root.PersistentFlags().String("format", "text", "output format: text or json")
	root.PersistentFlags().BoolP("quiet", "q", false, "print only problems")
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return &ExitError{Code: ExitUsageErr, Err: err}
	})

	root.AddCommand(
		newValidateCmd(e),
		newGraphCmd(e),
		newSchemaCmd(e),
		newInitCmd(e),
		newFmtCmd(e),
		newPatchCmd(e),
		newDeckCmd(e),
		newReportCmd(e),
		newExportCmd(e),
		newImportCmd(e),
		newApproveCmd(e),
		newJudgeCmd(e),
		newLintCmd(e),
		newStatusCmd(e),
		newDiffCmd(e),
	)
	root.AddCommand(newElementCmds(e)...)
	return root
}

// Execute runs the CLI with the given arguments (without the program name)
// and returns the process exit code.
func Execute(args []string, in io.Reader, out, errOut io.Writer) int {
	root := NewRootCmd(in, out, errOut)
	root.SetArgs(args)
	err := root.Execute()
	if err == nil {
		return ExitOK
	}
	var ee *ExitError
	if errors.As(err, &ee) {
		if ee.Err != nil {
			fmt.Fprintf(errOut, "theoryofchange: %v\n", ee.Err)
		}
		return ee.Code
	}
	fmt.Fprintf(errOut, "theoryofchange: %v\n", err)
	return ExitUsageErr
}

// ---- helpers --------------------------------------------------------------

// flagString reads a string flag the command defines. A missing flag is a
// programming error.
func flagString(cmd *cobra.Command, name string) string {
	v, err := cmd.Flags().GetString(name)
	if err != nil {
		panic(fmt.Sprintf("cli: flag --%s is not a string flag: %v", name, err))
	}
	return v
}

func flagBool(cmd *cobra.Command, name string) bool {
	v, err := cmd.Flags().GetBool(name)
	if err != nil {
		panic(fmt.Sprintf("cli: flag --%s is not a bool flag: %v", name, err))
	}
	return v
}

func flagInt(cmd *cobra.Command, name string) int {
	v, err := cmd.Flags().GetInt(name)
	if err != nil {
		panic(fmt.Sprintf("cli: flag --%s is not an int flag: %v", name, err))
	}
	return v
}

// outputFormat returns "text" or "json".
func outputFormat(cmd *cobra.Command) (string, error) {
	f := flagString(cmd, "format")
	if f != "text" && f != "json" {
		return "", usageErr("--format must be text or json, not %q", f)
	}
	return f, nil
}

func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("write JSON: %w", err)
	}
	return nil
}
