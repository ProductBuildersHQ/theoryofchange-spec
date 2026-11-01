package impact

import "slices"

// ReviewMethod is how a review was carried out.
type ReviewMethod string

const (
	ReviewDocument ReviewMethod = "document_review"
	ReviewSheet    ReviewMethod = "sheet_review"
	ReviewWorkshop ReviewMethod = "workshop"
)

// AllReviewMethods returns all review methods.
func AllReviewMethods() []ReviewMethod {
	return []ReviewMethod{ReviewDocument, ReviewSheet, ReviewWorkshop}
}

// Valid reports whether v is a known ReviewMethod.
func (v ReviewMethod) Valid() bool { return slices.Contains(AllReviewMethods(), v) }

// ReviewDecision is a reviewer's answer about one element.
type ReviewDecision string

const (
	DecisionAgree  ReviewDecision = "agree"
	DecisionChange ReviewDecision = "change"
	DecisionRemove ReviewDecision = "remove"
	DecisionUnsure ReviewDecision = "unsure"
)

// AllReviewDecisions returns all review decisions.
func AllReviewDecisions() []ReviewDecision {
	return []ReviewDecision{DecisionAgree, DecisionChange, DecisionRemove, DecisionUnsure}
}

// Valid reports whether v is a known ReviewDecision.
func (v ReviewDecision) Valid() bool { return slices.Contains(AllReviewDecisions(), v) }

// ReviewOrigin is where a response row came from.
type ReviewOrigin string

const (
	OriginSheet      ReviewOrigin = "sheet"
	OriginTable      ReviewOrigin = "table"
	OriginComment    ReviewOrigin = "comment"
	OriginSuggestion ReviewOrigin = "suggestion"
)

// AllReviewOrigins returns all review origins.
func AllReviewOrigins() []ReviewOrigin {
	return []ReviewOrigin{OriginSheet, OriginTable, OriginComment, OriginSuggestion}
}

// Valid reports whether v is a known ReviewOrigin.
func (v ReviewOrigin) Valid() bool { return slices.Contains(AllReviewOrigins(), v) }

// ReviewResponse carries what reviewers said about an impact model, in a form that
// does not depend on the channel it was collected through. Importing it
// proposes changes; it never changes the impact model.
type ReviewResponse struct {
	Schema      string `json:"$schema,omitempty"`
	Kind        Kind   `json:"kind" jsonschema:"required,enum=ReviewResponse"`
	SpecVersion string `json:"specVersion" jsonschema:"required"`
	ModelID     string `json:"modelId" jsonschema:"required"`
	Version     string `json:"version,omitempty"`
	// ContentHash is the hash of the impact model the reviewers saw, from the
	// export stamp. A different hash means the impact model has changed since.
	ContentHash string `json:"contentHash" jsonschema:"required"`
	// Reviewers name groups or roles, never individuals.
	Reviewers []string     `json:"reviewers" jsonschema:"required"`
	Method    ReviewMethod `json:"method" jsonschema:"required"`
	Date      string       `json:"date" jsonschema:"required,format=date"`
	Rows      []ReviewItem `json:"rows" jsonschema:"required"`
}

// ReviewItem is one answer from a reviewer.
type ReviewItem struct {
	// ElementID is empty when only the quoted text is known.
	ElementID       string         `json:"elementId,omitempty"`
	Decision        ReviewDecision `json:"decision,omitempty"`
	SuggestedChange string         `json:"suggestedChange,omitempty"`
	Comment         string         `json:"comment,omitempty"`
	Origin          ReviewOrigin   `json:"origin" jsonschema:"required"`
	QuotedText      string         `json:"quotedText,omitempty"`
}
