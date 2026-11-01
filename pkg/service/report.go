package service

import (
	"fmt"
	"slices"

	"github.com/ProductBuildersHQ/impact-model/pkg/burden"
	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
	"github.com/ProductBuildersHQ/impact-model/pkg/report"
)

// StatusReportKinds are the report kinds made from a status snapshot.
var StatusReportKinds = []string{
	"exec", "results", "assumptions", "interventions", "evidence-gaps", "full",
	"story", "headline", "metric-tree", "learning-review", "portfolio", "narrative", "actors", "custom",
}

// OptionalStatusKinds are the status report kinds that also work with no
// observations at all.
var OptionalStatusKinds = []string{"narrative", "actors", "custom"}

// IsStatusReport reports whether kind is made from a status snapshot.
func IsStatusReport(kind string) bool { return slices.Contains(StatusReportKinds, kind) }

// ReportParams are the extra inputs some status reports take.
type ReportParams struct {
	// Audience is required by the story report, and narrows the headline report.
	Audience string
	// Period names the review merged by the learning-review report.
	Period string
	// PrevSets are the earlier period's observations, so new themes show.
	PrevSets []*impact.ObservationSet
	// Pathway narrows the metric-tree report to one pathway.
	Pathway string
	// Template is the text/template source of the custom report.
	Template string
}

// StatusReport renders a report made from a status snapshot. sets are the
// observation sets the snapshot was computed from. withStatus is false only
// for a narrative, actors, or custom report of an impact model with no
// observations, whose input then has a view without status.
func StatusReport(kind string, in report.StatusInput, sets []*impact.ObservationSet, p ReportParams, withStatus bool) (string, error) {
	var status *impact.Status
	if withStatus {
		status = in.Status
	}
	switch kind {
	case "exec":
		return report.Exec(in), nil
	case "results":
		return report.Results(in), nil
	case "assumptions":
		return report.Assumptions(in), nil
	case "interventions":
		return report.Interventions(in), nil
	case "evidence-gaps":
		return report.EvidenceGaps(in), nil
	case "full":
		return report.Full(in, report.Mermaid(in.View, report.MermaidOptions{})), nil
	case "story":
		if p.Audience == "" {
			return "", fmt.Errorf("the story report needs an audience")
		}
		return report.Story(in, sets, p.Audience)
	case "headline":
		return report.Headline(in, p.Audience)
	case "metric-tree":
		return report.MetricTree(in, p.Pathway)
	case "learning-review":
		return report.LearningReview(report.LearningInput{StatusInput: in, Sets: sets, PrevSets: p.PrevSets, Period: p.Period}), nil
	case "portfolio":
		cost, err := burden.Compute(in.Doc)
		if err != nil {
			return "", err
		}
		return report.Portfolio(in, cost), nil
	case "narrative":
		return report.Narrative(in.View, in.Doc, status), nil
	case "actors":
		return report.Actors(in.View, in.Doc, status), nil
	case "custom":
		if p.Template == "" {
			return "", fmt.Errorf("the custom report needs a template")
		}
		return report.RenderTemplate(p.Template, in.View)
	}
	return "", fmt.Errorf("unknown status report %q", kind)
}

// ViewWithoutStatus builds the report input for an impact model that has no
// observations.
func ViewWithoutStatus(doc *impact.Model) (report.StatusInput, error) {
	g, err := BuildGraph(doc)
	if err != nil {
		return report.StatusInput{}, err
	}
	v, err := report.BuildView(doc, g, nil, report.Options{})
	if err != nil {
		return report.StatusInput{}, err
	}
	return report.StatusInput{View: v, Doc: doc}, nil
}
