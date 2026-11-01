package service

import (
	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
	"github.com/ProductBuildersHQ/impact-model/pkg/impactio"
	"github.com/ProductBuildersHQ/impact-model/pkg/report"
	"github.com/ProductBuildersHQ/impact-model/pkg/review"
	"github.com/ProductBuildersHQ/impact-model/pkg/share"
)

// ImportReview maps a review response onto doc as proposals. doc is never
// modified. Quoted text is matched against the statements the review packet
// showed, and the response's hash against the current document's. step is
// the consultation step to record; empty means the whole impact model.
func ImportReview(doc *impact.Model, resp *impact.ReviewResponse, step impact.ConsultationStep) (*review.Result, error) {
	g, err := BuildGraph(doc)
	if err != nil {
		return nil, err
	}
	v, err := report.BuildView(doc, g, nil, report.Options{})
	if err != nil {
		return nil, err
	}
	rows := share.ReviewRows(v)
	candidates := make([]review.Candidate, len(rows))
	for i, r := range rows {
		candidates[i] = review.Candidate{ID: r.ID, Statement: r.Statement}
	}
	hash, err := impactio.ContentHash(doc)
	if err != nil {
		return nil, err
	}
	return review.Import(doc, resp, review.Options{ContentHash: hash, Candidates: candidates, Step: step})
}
