// Package review turns what reviewers said about an impact model into proposals.
//
// Importing a ReviewResponse never changes the impact model. It produces a patch a
// human can apply, notes for the rows that need a human decision, and a
// consultation record. Nothing here reads files or the clock.
package review

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/ProductBuildersHQ/impact-model/pkg/edit"
	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
)

// Candidate is an element a quoted text may refer to: its ID and the
// statement a reviewer saw for it.
type Candidate struct {
	ID        string
	Statement string
}

// Options configures an import.
type Options struct {
	// ContentHash is the hash of the current document. A response whose
	// hash differs was written against an earlier version.
	ContentHash string
	// Candidates resolve rows that carry quoted text but no element ID.
	Candidates []Candidate
	// Step is the consultation step recorded; the default is "whole".
	Step impact.ConsultationStep
}

// Flag is a row that needs a human decision, or a caution about a row that
// was turned into an operation.
type Flag struct {
	// Row is the 1-based position of the row in the response.
	Row       int    `json:"row"`
	ElementID string `json:"elementId,omitempty"`
	Decision  string `json:"decision,omitempty"`
	Origin    string `json:"origin"`
	// Reason says why the row needs a human.
	Reason  string `json:"reason"`
	Comment string `json:"comment,omitempty"`
	// SuggestedChange is repeated so the note stands alone.
	SuggestedChange string `json:"suggestedChange,omitempty"`
}

// Result is the outcome of an import.
type Result struct {
	// Patch holds the proposed operations. Apply it with the patch engine
	// after a human has looked at it.
	Patch edit.Patch `json:"patch"`
	// Flags are the rows that need a human decision.
	Flags []Flag `json:"flags"`
	// Stale is true when the response was written against a different
	// version of the impact model. Every row is then flagged, and operations are
	// still proposed against current IDs.
	Stale bool `json:"stale"`
	// Mapped is the number of rows that became operations.
	Mapped int `json:"mapped"`
	// Consultation records the review in the impact model's own terms.
	Consultation impact.Consultation `json:"consultation"`
}

// NeedsHuman reports whether anything in the result requires a person to
// look before the patch is applied.
func (r *Result) NeedsHuman() bool { return r.Stale || len(r.Flags) > 0 }

// Import maps the rows of resp onto doc. doc is never modified.
func Import(doc *impact.Model, resp *impact.ReviewResponse, o Options) (*Result, error) {
	if doc == nil || resp == nil {
		return nil, fmt.Errorf("review: nil document or response")
	}
	if resp.ModelID != "" && doc.ID != "" && resp.ModelID != doc.ID {
		return nil, fmt.Errorf("review: response is for impact model %q, not %q", resp.ModelID, doc.ID)
	}
	res := &Result{
		Patch: edit.Patch{},
		Flags: []Flag{},
		Stale: resp.ContentHash != o.ContentHash,
	}
	// work follows the document as accepted rows are applied, so later rows
	// see earlier ones and an inapplicable row is flagged rather than
	// breaking the whole patch.
	work := doc
	for i, row := range resp.Rows {
		flag := func(reason string) {
			res.Flags = append(res.Flags, Flag{
				Row: i + 1, ElementID: row.ElementID, Decision: string(row.Decision), Origin: string(row.Origin),
				Reason: reason, Comment: row.Comment, SuggestedChange: row.SuggestedChange,
			})
		}
		id, reason := resolve(doc, row, o.Candidates)
		if reason != "" {
			flag(reason)
			continue
		}
		row.ElementID = id
		patch, reason := propose(work, row, resp)
		if reason != "" {
			flag(reason)
			continue
		}
		applied, err := edit.Apply(work, patch, edit.Options{})
		switch {
		case err != nil:
			flag("the proposed change does not apply: " + err.Error())
			continue
		case !applied.Applied:
			flag("the proposed change is rejected: " + applied.Reason)
			continue
		}
		work = applied.Document
		res.Patch = append(res.Patch, patch...)
		res.Mapped++
		if res.Stale {
			flag("the impact model has changed since this was reviewed")
		}
	}
	res.Consultation = consultation(doc, resp, o.Step)
	return res, nil
}

// resolve returns the element a row is about. A row with an ID must name an
// element that exists; a row with only quoted text must match exactly one
// candidate statement.
func resolve(doc *impact.Model, row impact.ReviewItem, candidates []Candidate) (id, reason string) {
	if row.ElementID != "" {
		if _, ok := doc.Element(row.ElementID); !ok {
			return "", fmt.Sprintf("no element %q in the impact model", row.ElementID)
		}
		return row.ElementID, ""
	}
	if strings.TrimSpace(row.QuotedText) == "" {
		return "", "no element id and no quoted text; a comment about the whole impact model"
	}
	want := normalize(row.QuotedText)
	var found []string
	for _, c := range candidates {
		if normalize(c.Statement) == want && !slices.Contains(found, c.ID) {
			found = append(found, c.ID)
		}
	}
	switch len(found) {
	case 0:
		return "", "the quoted text matches no element"
	case 1:
		return found[0], ""
	default:
		return "", "the quoted text matches several elements: " + strings.Join(found, ", ")
	}
}

func normalize(s string) string { return strings.Join(strings.Fields(s), " ") }

// kindOf maps a document element kind onto the patch engine's.
var kindOf = map[impact.ElementKind]string{
	impact.ElementOutcome: "outcome", impact.ElementGroup: "group", impact.ElementLink: "link",
	impact.ElementIntervention: "intervention", impact.ElementAssumption: "assumption",
	impact.ElementIndicator: "indicator", impact.ElementActor: "actor", impact.ElementEvidence: "evidence",
	impact.ElementExternalFactor: "factor", impact.ElementRisk: "risk", impact.ElementDecision: "decision",
	impact.ElementPathway: "pathway", impact.ElementMetric: "metric", impact.ElementTerm: "term",
	impact.ElementConsultation: "consultation",
}

// propose turns one resolved row into a patch, or says why it cannot.
func propose(doc *impact.Model, row impact.ReviewItem, resp *impact.ReviewResponse) (edit.Patch, string) {
	ref, _ := doc.Element(row.ElementID)
	kind, ok := edit.KindByName(kindOf[ref.Kind])
	if !ok {
		return nil, fmt.Sprintf("elements of kind %q cannot be edited", ref.Kind)
	}
	decided := map[string]any{
		"provenance.decidedAt": resp.Date,
		"provenance.decidedBy": strings.Join(resp.Reviewers, "; "),
	}
	switch row.Decision {
	case impact.DecisionAgree:
		decided["provenance.basis"] = string(impact.BasisConfirmed)
		return setFields(doc, kind, row.ElementID, decided)
	case impact.DecisionRemove:
		decided["provenance.basis"] = string(impact.BasisRejected)
		return setFields(doc, kind, row.ElementID, decided)
	case impact.DecisionChange:
		fields, ok := parseChange(row.SuggestedChange)
		if !ok {
			return nil, "the suggested change is free text; a person must decide how to apply it"
		}
		return setFields(doc, kind, row.ElementID, fields)
	case impact.DecisionUnsure:
		return nil, "the reviewer is unsure"
	case "":
		return nil, "a comment with no decision"
	default:
		return nil, fmt.Sprintf("unknown decision %q", row.Decision)
	}
}

func setFields(doc *impact.Model, k edit.ElementKind, id string, fields map[string]any) (edit.Patch, string) {
	p, err := edit.SetFields(doc, k, id, fields, nil)
	if err != nil {
		return nil, "the proposed change cannot be built: " + err.Error()
	}
	return p, ""
}

var fieldLine = regexp.MustCompile(`^([A-Za-z][A-Za-z0-9]*(?:\.[A-Za-z][A-Za-z0-9]*)*)\s*:\s*(.+)$`)

// parseChange reads a structured suggestion: one "field: value" per line.
// Anything else is free text and reports false. The id cannot be changed.
func parseChange(s string) (map[string]any, bool) {
	fields := map[string]any{}
	for line := range strings.SplitSeq(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		m := fieldLine.FindStringSubmatch(line)
		if m == nil || m[1] == "id" {
			return nil, false
		}
		fields[m[1]] = edit.ParseValue(m[2])
	}
	return fields, len(fields) > 0
}

// consultation builds the record of the review. Its ID avoids the IDs the
// Impact model already uses.
func consultation(doc *impact.Model, resp *impact.ReviewResponse, step impact.ConsultationStep) impact.Consultation {
	if step == "" {
		step = impact.StepWhole
	}
	method := impact.MethodDocumentReview
	if resp.Method == impact.ReviewWorkshop {
		method = impact.MethodWorkshop
	}
	counts := map[impact.ReviewDecision]int{}
	other := 0
	for _, r := range resp.Rows {
		if r.Decision == "" {
			other++
			continue
		}
		counts[r.Decision]++
	}
	var findings []string
	for _, d := range impact.AllReviewDecisions() {
		if counts[d] > 0 {
			findings = append(findings, fmt.Sprintf("%s: %d", d, counts[d]))
		}
	}
	if other > 0 {
		findings = append(findings, fmt.Sprintf("comment only: %d", other))
	}
	return impact.Consultation{
		ID:                consultationID(doc, resp.Date),
		Date:              resp.Date,
		Step:              step,
		Method:            method,
		StakeholderGroups: slices.Clone(resp.Reviewers),
		Findings:          findings,
	}
}

func consultationID(doc *impact.Model, date string) string {
	used := map[string]bool{}
	for _, c := range doc.Consultations {
		used[c.ID] = true
	}
	base := "consult-review-" + date
	id := base
	for n := 2; used[id]; n++ {
		id = fmt.Sprintf("%s-%d", base, n)
	}
	return id
}
