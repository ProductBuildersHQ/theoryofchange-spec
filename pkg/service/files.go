package service

import (
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"slices"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/edit"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/tocio"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/validate"
)

// ---- init -----------------------------------------------------------------

// InitOptions configures InitDocument.
type InitOptions struct {
	// Template names the starter; see Templates.
	Template string
	// ID is the document slug.
	ID string
	// Name is the human-readable title.
	Name string
	// Problem is the problem statement; a placeholder is used when empty.
	Problem string
}

// Templates lists the available starter templates.
func Templates() []string { return []string{"minimal"} }

// ErrUnknownTemplate is returned for a template that does not exist.
var ErrUnknownTemplate = errors.New("unknown template")

// InitDocument returns a starter document that already passes validation,
// with placeholders to replace.
func InitDocument(o InitOptions) (*toc.Document, error) {
	if !slices.Contains(Templates(), o.Template) {
		return nil, fmt.Errorf("%w %q (available: %v)", ErrUnknownTemplate, o.Template, Templates())
	}
	if o.ID == "" {
		o.ID = "my-theory-of-change"
	}
	if o.Name == "" {
		o.Name = "My Theory of Change"
	}
	if o.Problem == "" {
		o.Problem = "TODO: describe the problem this theory addresses, for whom, and where."
	}
	doc := toc.NewDocument(o.ID, o.Problem)
	doc.Metadata.Name = o.Name
	doc.Metadata.Status = "draft"
	doc.Outcomes = []toc.Outcome{
		{ID: "oc-impact", Title: "TODO: the long-term change you want to see", Level: toc.LevelImpact},
		{ID: "oc-first", Title: "TODO: the first condition that must exist", Level: toc.LevelShortTerm},
	}
	doc.Links = []toc.CausalLink{
		{ID: "ln-first-impact", From: "oc-first", To: "oc-impact", Rationale: "TODO: why this condition leads to the change"},
	}
	doc.Indicators = []toc.Indicator{
		{ID: "ind-first", Measures: "oc-first", Name: "TODO: how you will know this condition exists"},
	}
	if issues := validate.Document(doc); validate.HasErrors(issues) {
		return nil, fmt.Errorf("starter document is invalid (programming error): %v", issues)
	}
	return doc, nil
}

// ---- fmt ------------------------------------------------------------------

// SortElements orders every identified-element list by ID. Element order is
// otherwise preserved by formatting.
func SortElements(doc *toc.Document) {
	slices.SortStableFunc(doc.Outcomes, func(a, b toc.Outcome) int { return cmp.Compare(a.ID, b.ID) })
	slices.SortStableFunc(doc.Groups, func(a, b toc.PreconditionGroup) int { return cmp.Compare(a.ID, b.ID) })
	slices.SortStableFunc(doc.Links, func(a, b toc.CausalLink) int { return cmp.Compare(a.ID, b.ID) })
	slices.SortStableFunc(doc.Interventions, func(a, b toc.Intervention) int { return cmp.Compare(a.ID, b.ID) })
	slices.SortStableFunc(doc.Assumptions, func(a, b toc.Assumption) int { return cmp.Compare(a.ID, b.ID) })
	slices.SortStableFunc(doc.Indicators, func(a, b toc.Indicator) int { return cmp.Compare(a.ID, b.ID) })
	slices.SortStableFunc(doc.Actors, func(a, b toc.Actor) int { return cmp.Compare(a.ID, b.ID) })
	slices.SortStableFunc(doc.Evidence, func(a, b toc.Evidence) int { return cmp.Compare(a.ID, b.ID) })
	slices.SortStableFunc(doc.ExternalFactors, func(a, b toc.ExternalFactor) int { return cmp.Compare(a.ID, b.ID) })
	slices.SortStableFunc(doc.Risks, func(a, b toc.Risk) int { return cmp.Compare(a.ID, b.ID) })
	slices.SortStableFunc(doc.Decisions, func(a, b toc.Decision) int { return cmp.Compare(a.ID, b.ID) })
}

// FormatResult reports what formatting did to one file.
type FormatResult struct {
	Path string `json:"path"`
	// Changed reports whether the file differs from its canonical form.
	Changed bool `json:"changed"`
	// Written reports whether the file was rewritten.
	Written bool `json:"written"`
}

// FormatFile rewrites a file in canonical form. With check it only reports
// whether the file would change. With sortElements, a Theory of Change has
// its element lists ordered by ID first.
func FormatFile(path string, check, sortElements bool) (FormatResult, error) {
	data, f, err := readInput(path)
	if err != nil {
		return FormatResult{}, err
	}
	out, err := formatBytes(data, f, sortElements)
	if err != nil {
		return FormatResult{}, fmt.Errorf("%s: %w", path, err)
	}
	res := FormatResult{Path: path, Changed: !bytes.Equal(data, out)}
	if res.Changed && !check {
		if err := tocio.WriteFile(path, out); err != nil {
			return res, err
		}
		res.Written = true
	}
	return res, nil
}

func formatBytes(data []byte, f tocio.Format, sortElements bool) ([]byte, error) {
	if sortElements {
		if kind, err := tocio.DetectKind(data, f); err == nil && kind == toc.KindTheoryOfChange {
			doc, err := tocio.ParseDocument(data, f)
			if err != nil {
				return nil, err
			}
			SortElements(doc)
			return tocio.Canonical(doc, f)
		}
	}
	return tocio.Reformat(data, f)
}

// ---- patch ----------------------------------------------------------------

// PatchOptions configures PatchFile.
type PatchOptions struct {
	// DryRun computes the result without writing the file.
	DryRun bool
	// AllowErrors accepts a patch that adds structural errors.
	AllowErrors bool
}

// PatchResult is the outcome of applying a patch to a file.
type PatchResult struct {
	Path             string           `json:"path"`
	Applied          bool             `json:"applied"`
	Changed          bool             `json:"changed"`
	Written          bool             `json:"written"`
	DryRun           bool             `json:"dryRun,omitempty"`
	Reason           string           `json:"reason,omitempty"`
	IssuesIntroduced []validate.Issue `json:"issuesIntroduced"`
	IssuesResolved   []validate.Issue `json:"issuesResolved"`
	// Document is the patched document in canonical form, set only on a
	// dry run that would apply.
	Document []byte `json:"-"`
}

// PatchFile applies a patch to a Theory of Change file and, unless DryRun,
// writes the canonical result in the file's own format. Nothing is written
// when the patch is rejected or changes nothing.
func PatchFile(path string, p edit.Patch, o PatchOptions) (*PatchResult, error) {
	doc, err := tocio.LoadDocument(path)
	if err != nil {
		return nil, err
	}
	f, err := tocio.FormatFromPath(path)
	if err != nil {
		return nil, err
	}
	res, err := edit.Apply(doc, p, edit.Options{AllowErrors: o.AllowErrors})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	out := &PatchResult{
		Path: path, Applied: res.Applied, Changed: res.Changed, DryRun: o.DryRun, Reason: res.Reason,
		IssuesIntroduced: nonNilIssues(res.IssuesIntroduced), IssuesResolved: nonNilIssues(res.IssuesResolved),
	}
	if !res.Applied {
		return out, nil
	}
	data, err := tocio.Canonical(res.Document, f)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	switch {
	case o.DryRun:
		out.Document = data
	case res.Changed:
		if err := tocio.WriteFile(path, data); err != nil {
			return nil, err
		}
		out.Written = true
	}
	return out, nil
}

func nonNilIssues(i []validate.Issue) []validate.Issue {
	if i == nil {
		return []validate.Issue{}
	}
	return i
}
