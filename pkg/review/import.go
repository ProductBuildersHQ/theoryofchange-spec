// Package review turns what reviewers said about a theory into proposals.
//
// Importing a ReviewResponse never changes the theory. It produces a patch a
// human can apply, notes for the rows that need a human decision, and a
// consultation record. Nothing here reads files or the clock.
package review

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/edit"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
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
	Step toc.ConsultationStep
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
	// version of the theory. Every row is then flagged, and operations are
	// still proposed against current IDs.
	Stale bool `json:"stale"`
	// Mapped is the number of rows that became operations.
	Mapped int `json:"mapped"`
	// Consultation records the review in the theory's own terms.
	Consultation toc.Consultation `json:"consultation"`
}

// NeedsHuman reports whether anything in the result requires a person to
// look before the patch is applied.
func (r *Result) NeedsHuman() bool { return r.Stale || len(r.Flags) > 0 }

// Import maps the rows of resp onto doc. doc is never modified.
func Import(doc *toc.Document, resp *toc.ReviewResponse, o Options) (*Result, error) {
	if doc == nil || resp == nil {
		return nil, fmt.Errorf("review: nil document or response")
	}
	if resp.TocID != "" && doc.ID != "" && resp.TocID != doc.ID {
		return nil, fmt.Errorf("review: response is for theory %q, not %q", resp.TocID, doc.ID)
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
			flag("the theory has changed since this was reviewed")
		}
	}
	res.Consultation = consultation(doc, resp, o.Step)
	return res, nil
}

// resolve returns the element a row is about. A row with an ID must name an
// element that exists; a row with only quoted text must match exactly one
// candidate statement.
func resolve(doc *toc.Document, row toc.ReviewItem, candidates []Candidate) (id, reason string) {
	if row.ElementID != "" {
		if _, ok := doc.Element(row.ElementID); !ok {
			return "", fmt.Sprintf("no element %q in the theory", row.ElementID)
		}
		return row.ElementID, ""
	}
	if strings.TrimSpace(row.QuotedText) == "" {
		return "", "no element id and no quoted text; a comment about the whole theory"
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
var kindOf = map[toc.ElementKind]string{
	toc.ElementOutcome: "outcome", toc.ElementGroup: "group", toc.ElementLink: "link",
	toc.ElementIntervention: "intervention", toc.ElementAssumption: "assumption",
	toc.ElementIndicator: "indicator", toc.ElementActor: "actor", toc.ElementEvidence: "evidence",
	toc.ElementExternalFactor: "factor", toc.ElementRisk: "risk", toc.ElementDecision: "decision",
	toc.ElementPathway: "pathway", toc.ElementMetric: "metric", toc.ElementTerm: "term",
	toc.ElementConsultation: "consultation",
}

// propose turns one resolved row into a patch, or says why it cannot.
func propose(doc *toc.Document, row toc.ReviewItem, resp *toc.ReviewResponse) (edit.Patch, string) {
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
	case toc.DecisionAgree:
		decided["provenance.basis"] = string(toc.BasisConfirmed)
		return setFields(doc, kind, row.ElementID, decided)
	case toc.DecisionRemove:
		decided["provenance.basis"] = string(toc.BasisRejected)
		return setFields(doc, kind, row.ElementID, decided)
	case toc.DecisionChange:
		fields, ok := parseChange(row.SuggestedChange)
		if !ok {
			return nil, "the suggested change is free text; a person must decide how to apply it"
		}
		return setFields(doc, kind, row.ElementID, fields)
	case toc.DecisionUnsure:
		return nil, "the reviewer is unsure"
	case "":
		return nil, "a comment with no decision"
	default:
		return nil, fmt.Sprintf("unknown decision %q", row.Decision)
	}
}

func setFields(doc *toc.Document, k edit.ElementKind, id string, fields map[string]any) (edit.Patch, string) {
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
// theory already uses.
func consultation(doc *toc.Document, resp *toc.ReviewResponse, step toc.ConsultationStep) toc.Consultation {
	if step == "" {
		step = toc.StepWhole
	}
	method := toc.MethodDocumentReview
	if resp.Method == toc.ReviewWorkshop {
		method = toc.MethodWorkshop
	}
	counts := map[toc.ReviewDecision]int{}
	other := 0
	for _, r := range resp.Rows {
		if r.Decision == "" {
			other++
			continue
		}
		counts[r.Decision]++
	}
	var findings []string
	for _, d := range toc.AllReviewDecisions() {
		if counts[d] > 0 {
			findings = append(findings, fmt.Sprintf("%s: %d", d, counts[d]))
		}
	}
	if other > 0 {
		findings = append(findings, fmt.Sprintf("comment only: %d", other))
	}
	return toc.Consultation{
		ID:                consultationID(doc, resp.Date),
		Date:              resp.Date,
		Step:              step,
		Method:            method,
		StakeholderGroups: slices.Clone(resp.Reviewers),
		Findings:          findings,
	}
}

func consultationID(doc *toc.Document, date string) string {
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
