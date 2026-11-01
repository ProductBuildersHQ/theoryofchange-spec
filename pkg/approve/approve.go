// Package approve builds the patch that records an impact model as approved, and
// refuses when the impact model is not ready. It only builds the patch; applying it
// is the patch engine's job, so approval is an ordinary, reviewable edit.
package approve

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	core "github.com/grokify/prism-core"

	"github.com/ProductBuildersHQ/impact-model/pkg/edit"
	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
	"github.com/ProductBuildersHQ/impact-model/pkg/validate"
)

// StatusApproved is the metadata status of an approved impact model.
const StatusApproved = "approved"

// Request describes an approval.
type Request struct {
	// Version is the label the approved impact model carries. It must differ from
	// the version already recorded.
	Version string
	// Date is the day of approval (YYYY-MM-DD).
	Date string
	// Summary says what was approved; it defaults to a stock sentence.
	Summary string
	// Approvers are the people who approved.
	Approvers []core.Approver
	// Force approves despite open hypotheses; Reason is then required and
	// is recorded in the revision.
	Force  bool
	Reason string
}

// Refusal is one reason approval was refused.
type Refusal struct {
	// Code is "errors", "hypotheses", or "request".
	Code    string `json:"code"`
	Message string `json:"message"`
	// IDs lists the elements concerned.
	IDs []string `json:"ids,omitempty"`
}

// ErrRefused is returned (wrapped) when approval is refused.
var ErrRefused = errors.New("approval refused")

// RefusedError carries the reasons.
type RefusedError struct{ Refusals []Refusal }

func (e *RefusedError) Error() string {
	msgs := make([]string, len(e.Refusals))
	for i, r := range e.Refusals {
		msgs[i] = r.Message
	}
	return ErrRefused.Error() + ": " + strings.Join(msgs, "; ")
}

func (e *RefusedError) Unwrap() error { return ErrRefused }

// Build returns the patch that approves doc. It refuses when the request is
// incomplete, when the impact model has validation errors, or when it still holds
// hypotheses (unless forced with a reason).
func Build(doc *impact.Model, req Request) (edit.Patch, error) {
	var refusals []Refusal
	add := func(code, msg string, ids ...string) { refusals = append(refusals, Refusal{code, msg, ids}) }

	if req.Version == "" {
		add("request", "a version label is required")
	} else if cur := currentVersion(doc); req.Version == cur {
		add("request", fmt.Sprintf("version %q is already recorded; choose a new label", req.Version))
	}
	if _, err := time.Parse(time.DateOnly, req.Date); err != nil {
		add("request", "a date in the form YYYY-MM-DD is required")
	}
	if len(req.Approvers) == 0 {
		add("request", "at least one approver is required")
	}
	if req.Force && strings.TrimSpace(req.Reason) == "" {
		add("request", "forcing approval requires a reason")
	}
	if errs := errorIssues(doc); len(errs) > 0 {
		add("errors", fmt.Sprintf("the impact model has %d validation error(s)", len(errs)), issueIDs(errs)...)
	}
	hyp, err := Hypotheses(doc)
	if err != nil {
		return nil, err
	}
	if len(hyp) > 0 && !req.Force {
		add("hypotheses", fmt.Sprintf("%d element(s) are still hypotheses; confirm or reject them, or force with a reason", len(hyp)), hyp...)
	}
	if len(refusals) > 0 {
		return nil, &RefusedError{Refusals: refusals}
	}
	return patch(doc, req, hyp)
}

func patch(doc *impact.Model, req Request, hypotheses []string) (edit.Patch, error) {
	summary := strings.TrimSpace(req.Summary)
	if summary == "" {
		summary = "Approved after review."
	}
	if len(hypotheses) > 0 {
		summary += fmt.Sprintf(" Approved with %d open hypothesis element(s): %s", len(hypotheses), strings.TrimSpace(req.Reason))
	}
	at, err := time.Parse(time.DateOnly, req.Date)
	if err != nil {
		return nil, fmt.Errorf("approve: %w", err)
	}
	approvers := make([]core.Approver, len(req.Approvers))
	for i, a := range req.Approvers {
		a.Approved, a.ApprovedAt = true, at
		approvers[i] = a
	}
	rev := impact.Revision{
		Version: req.Version, Date: req.Date, Summary: summary,
		Trigger: impact.TriggerReview, Approvers: approvers,
	}
	var ops []edit.Op
	for _, step := range []struct {
		path  string
		value any
	}{
		{edit.Pointer("metadata", "status"), StatusApproved},
		{edit.Pointer("metadata", "version"), req.Version},
	} {
		op, err := edit.Add(step.path, step.value)
		if err != nil {
			return nil, err
		}
		ops = append(ops, op)
	}
	var revOp edit.Op
	if len(doc.Revisions) == 0 {
		revOp, err = edit.Add(edit.Pointer("revisions"), []impact.Revision{rev})
	} else {
		revOp, err = edit.Add(edit.Pointer("revisions", "-"), rev)
	}
	if err != nil {
		return nil, err
	}
	return append(ops, revOp), nil
}

func currentVersion(doc *impact.Model) string {
	if doc.Metadata.Version != "" {
		return doc.Metadata.Version
	}
	if n := len(doc.Revisions); n > 0 {
		return doc.Revisions[n-1].Version
	}
	return ""
}

func errorIssues(doc *impact.Model) []validate.Issue {
	var out []validate.Issue
	for _, i := range validate.Document(doc) {
		if i.Severity == validate.SeverityError {
			out = append(out, i)
		}
	}
	return out
}

func issueIDs(issues []validate.Issue) []string {
	var ids []string
	for _, i := range issues {
		if i.Element != "" {
			ids = append(ids, i.Element)
		}
	}
	return ids
}

// Hypotheses lists the IDs of active elements whose basis is hypothesis, in
// document order. Rejected elements are not counted.
func Hypotheses(doc *impact.Model) ([]string, error) {
	data, err := json.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("approve: encode document: %w", err)
	}
	var tree map[string]json.RawMessage
	if err := json.Unmarshal(data, &tree); err != nil {
		return nil, fmt.Errorf("approve: decode document: %w", err)
	}
	var ids []string
	for _, field := range elementFields {
		var items []struct {
			ID         string `json:"id"`
			Provenance *struct {
				Basis string `json:"basis"`
			} `json:"provenance"`
		}
		raw, ok := tree[field]
		if !ok {
			continue
		}
		if err := json.Unmarshal(raw, &items); err != nil {
			return nil, fmt.Errorf("approve: decode %s: %w", field, err)
		}
		for _, it := range items {
			if it.Provenance != nil && it.Provenance.Basis == string(impact.BasisHypothesis) {
				ids = append(ids, it.ID)
			}
		}
	}
	return ids, nil
}

var elementFields = func() []string {
	kinds := edit.ElementKinds()
	out := make([]string, len(kinds))
	for i, k := range kinds {
		out[i] = k.Field
	}
	return out
}()
