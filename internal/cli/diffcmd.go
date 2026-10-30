package cli

import (
	"bytes"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/diff"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/tocio"
)

func newDiffCmd(e *env) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff <old> <new>",
		Short: "Show what changed between two versions of a theory",
		Long: `Compare two versions of a theory by element. A change is keyed by the element's
ID, so a retitled outcome is one changed element, and a rejected one shows as a
rejection. Field order and file format do not matter: the same theory in YAML
and in JSON has no differences. The revision log is not compared.

The result is Markdown, or JSON with --format json.

With --revision, nothing is shown. Instead the command prints a patch that
appends a draft revision to the newer theory, recording the version, date,
and trigger given and listing every element ID that changed. Review it, then
apply it with 'theoryofchange patch'.`,
		Example: `  theoryofchange diff theory-v1.yaml theory-v2.yaml
  theoryofchange diff old.yaml new.yaml --revision --version 0.2 --date 2026-03-01 > revision.patch.json
  theoryofchange patch new.yaml revision.patch.json`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error { return runDiff(e, cmd, args[0], args[1]) },
	}
	f := cmd.Flags()
	f.Bool("revision", false, "print a patch that adds a draft revision for the change")
	f.String("version", "", "version label for the draft revision (default: the newer theory's version)")
	f.String("date", "", "date of the draft revision, YYYY-MM-DD (required with --revision)")
	f.String("trigger", string(toc.TriggerLearning), "what prompted the change: "+triggerNames())
	f.StringP("out", "o", "", "write to this file instead of standard output")
	return cmd
}

func triggerNames() string {
	out := ""
	for i, t := range toc.AllRevisionTriggers() {
		if i > 0 {
			out += ", "
		}
		out += string(t)
	}
	return out
}

func runDiff(e *env, cmd *cobra.Command, oldPath, newPath string) error {
	format, err := outputFormat(cmd)
	if err != nil {
		return err
	}
	old, err := loadForRender(oldPath)
	if err != nil {
		return err
	}
	cur, err := loadForRender(newPath)
	if err != nil {
		return err
	}
	res, err := diff.Compare(old, cur)
	if err != nil {
		return &ExitError{Code: ExitUsageErr, Err: err}
	}

	var text string
	switch {
	case flagBool(cmd, "revision"):
		if text, err = revisionPatchText(cmd, res, cur); err != nil {
			return err
		}
	case format == "json":
		var buf bytes.Buffer
		if err := writeJSON(&buf, res); err != nil {
			return &ExitError{Code: ExitUsageErr, Err: err}
		}
		text = buf.String()
	default:
		text = res.Markdown()
	}
	if out := flagString(cmd, "out"); out != "" {
		if err := tocio.WriteFile(out, []byte(text)); err != nil {
			return &ExitError{Code: ExitUsageErr, Err: err}
		}
		return nil
	}
	fmt.Fprint(e.out, text)
	return nil
}

func revisionPatchText(cmd *cobra.Command, res *diff.Result, cur *toc.Document) (string, error) {
	date := flagString(cmd, "date")
	if _, err := time.Parse(time.DateOnly, date); err != nil {
		return "", usageErr("--revision needs --date as YYYY-MM-DD")
	}
	version := flagString(cmd, "version")
	if version == "" {
		version = cur.Metadata.Version
	}
	if version == "" {
		return "", usageErr("--revision needs --version: the newer theory declares none")
	}
	trigger := toc.RevisionTrigger(flagString(cmd, "trigger"))
	if !trigger.Valid() {
		return "", usageErr("--trigger %q is not one of %s", trigger, triggerNames())
	}
	patch, err := diff.RevisionPatch(res.DraftRevision(version, date, trigger), len(cur.Revisions) > 0)
	if err != nil {
		return "", &ExitError{Code: ExitUsageErr, Err: err}
	}
	var buf bytes.Buffer
	if err := writeJSON(&buf, patch); err != nil {
		return "", &ExitError{Code: ExitUsageErr, Err: err}
	}
	return buf.String(), nil
}
