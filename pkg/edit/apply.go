package edit

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
	"github.com/ProductBuildersHQ/impact-model/pkg/impactio"
	"github.com/ProductBuildersHQ/impact-model/pkg/validate"
)

// Options configures Apply.
type Options struct {
	// AllowErrors accepts a patch that leaves the document with more
	// structural errors than before. The result must still conform to the
	// schema; AllowErrors never accepts a document that does not.
	AllowErrors bool
}

// Result is the outcome of Apply.
type Result struct {
	// Applied reports whether the patch was accepted. When false, nothing
	// changed.
	Applied bool
	// Reason says why a patch was not applied.
	Reason string
	// Document is the patched document, set only when Applied. It never
	// aliases the input.
	Document *impact.Model
	// Changed reports whether the patched document differs from the input.
	Changed bool
	// IssuesIntroduced are validation issues present after but not before.
	IssuesIntroduced []validate.Issue
	// IssuesResolved are validation issues present before but not after.
	IssuesResolved []validate.Issue
}

// Apply applies p to doc atomically and returns the outcome. doc is never
// modified.
//
// A malformed or inapplicable patch (unknown path or @id, failed test, bad
// operation) returns an error wrapping ErrPatch, with an *OpError naming the
// failing operation. A well-formed patch that is rejected — because the
// result violates the schema, or adds structural errors without AllowErrors
// — returns a Result with Applied false and a nil error.
func Apply(doc *impact.Model, p Patch, opts Options) (*Result, error) {
	if doc == nil {
		return nil, fmt.Errorf("%w: nil document", ErrPatch)
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	orig, err := json.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("encode document: %w", err)
	}
	before := issuesOf(orig, doc)

	tree, err := decodeTree(orig)
	if err != nil {
		return nil, fmt.Errorf("decode document: %w", err)
	}
	for i, op := range p {
		tree, err = applyOp(tree, op)
		if err != nil {
			return nil, &OpError{Index: i, Op: op, Err: err}
		}
	}
	patched, err := json.Marshal(tree)
	if err != nil {
		return nil, fmt.Errorf("encode patched document: %w", err)
	}

	if schemaIssues := validate.Schema(patched, impactio.FormatJSON); len(schemaIssues) > 0 {
		introduced, resolved := validate.Delta(before, schemaIssues)
		return &Result{
			Reason:           "the patched document does not conform to the schema",
			IssuesIntroduced: introduced,
			IssuesResolved:   resolved,
		}, nil
	}
	out, err := decodeDocument(patched)
	if err != nil {
		return nil, err
	}
	after := validate.Document(out)
	introduced, resolved := validate.Delta(before, after)
	res := &Result{IssuesIntroduced: introduced, IssuesResolved: resolved}

	if n, m := validate.CountSeverity(after, validate.SeverityError), validate.CountSeverity(before, validate.SeverityError); n > m && !opts.AllowErrors {
		res.Reason = fmt.Sprintf("the patch would increase structural errors from %d to %d", m, n)
		return res, nil
	}
	res.Applied = true
	res.Document = out
	canonical, err := json.Marshal(out)
	if err != nil {
		return nil, fmt.Errorf("encode patched document: %w", err)
	}
	res.Changed = !bytes.Equal(orig, canonical)
	return res, nil
}

// issuesOf returns the schema and structural issues of an existing
// document.
func issuesOf(encoded []byte, doc *impact.Model) []validate.Issue {
	issues := validate.Schema(encoded, impactio.FormatJSON)
	return append(issues, validate.Document(doc)...)
}

func decodeDocument(data []byte) (*impact.Model, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var d impact.Model
	if err := dec.Decode(&d); err != nil {
		return nil, fmt.Errorf("decode patched document: %w", err)
	}
	return &d, nil
}
