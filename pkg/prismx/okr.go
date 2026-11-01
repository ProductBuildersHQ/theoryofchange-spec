// Package prismx exports an impact model and its status to PRISM
// documents: OKRs and outcome journeys. It is the only package that imports
// them. The model and the tools never do, so an impact model stays usable without
// the PRISM family; a test checks that boundary.
package prismx

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/grokify/prism-roadmap/goals/okr"

	"github.com/ProductBuildersHQ/impact-model/pkg/graph"
	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
	"github.com/ProductBuildersHQ/impact-model/pkg/measure"
)

// ToOKR exports the outcomes an organization answers for (owned or shared,
// and within the ceiling of accountability) as an OKR document.
// Each such outcome that has measured indicators becomes an objective, and
// each of its counted indicators a key result: the baseline, the final
// target, the latest value, a 0 to 1 score from progress, and a status and
// confidence from the indicator's state. Risks carry over. st may be nil;
// the key results then have no current value, score, or status.
//
// An OKR document holds a handful of objectives. This exports every
// qualifying outcome and leaves choosing among them to the caller; the
// result's own validation warns when there are more than the usual five.
func ToOKR(doc *impact.Model, st *impact.Status) (*okr.OKRDocument, error) {
	if doc == nil {
		return nil, fmt.Errorf("prismx: nil document")
	}
	g, err := graph.Build(doc)
	if err != nil {
		return nil, err
	}
	statuses := map[string]impact.IndicatorStatus{}
	if st != nil {
		for _, s := range st.Indicators {
			statuses[s.IndicatorID] = s
		}
	}

	out := okr.New(doc.ID, doc.Metadata.Name, doc.Metadata.Owner)
	out.Metadata.Version = doc.Metadata.Version
	out.Metadata.Status = okrStatus(doc.Metadata.Status)
	out.Metadata.CreatedAt, out.Metadata.UpdatedAt = time.Time{}, time.Time{}
	if st != nil {
		out.Metadata.Period = st.AsOf
	}
	out.Objectives = nil

	for _, o := range doc.Outcomes {
		if o.Rejected() || !o.Accountable() || !g.BelowCeiling(o.ID) {
			continue
		}
		obj := okr.Objective{ID: o.ID, Title: o.Title, Description: o.Description, Status: out.Metadata.Status}
		for _, ind := range doc.Indicators {
			if ind.Measures != o.ID || !measure.Counted(ind) {
				continue
			}
			kr := keyResult(ind, statuses[ind.ID], st != nil)
			obj.KeyResults = append(obj.KeyResults, kr)
			if kr.DueDate > obj.Timeframe {
				obj.Timeframe = kr.DueDate
			}
		}
		if len(obj.KeyResults) == 0 {
			continue
		}
		if obj.Timeframe != "" {
			obj.Timeframe = obj.Timeframe[:4]
		}
		if o.Responsibility != "" {
			obj.Owner = ownerName(doc, o)
		}
		obj.UpdateProgress()
		out.Objectives = append(out.Objectives, obj)
	}
	for _, r := range doc.Risks {
		if r.Provenance.IsRejected() {
			continue
		}
		out.Risks = append(out.Risks, okr.Risk{
			ID: r.ID, Title: r.Description, Impact: capitalize(r.Impact), Likelihood: capitalize(r.Probability),
			Mitigation: r.Mitigation, Status: r.Status,
		})
	}
	return out, nil
}

func okrStatus(s string) string {
	switch s {
	case "active", "approved":
		return okr.StatusActive
	case "archived":
		return okr.StatusCompleted
	default:
		return okr.StatusDraft
	}
}

func capitalize(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func ownerName(doc *impact.Model, o impact.Outcome) string {
	if len(o.DeliveredBy) > 0 {
		if a, ok := doc.Actor(o.DeliveredBy[0]); ok {
			return a.Name
		}
	}
	return ""
}

func keyResult(ind impact.Indicator, s impact.IndicatorStatus, withStatus bool) okr.KeyResult {
	kr := okr.KeyResult{ID: ind.ID, Title: ind.Name, Description: ind.Definition, Metric: ind.Name, Unit: ind.Unit}
	if ind.Baseline != nil {
		kr.Baseline = valueText(ind.Baseline.Value, ind.Baseline.Level, ind.Unit)
	}
	if n := len(ind.Targets); n > 0 {
		final := ind.Targets[0]
		for _, t := range ind.Targets[1:] {
			if t.Date > final.Date {
				final = t
			}
		}
		kr.Target = valueText(final.Value, final.Level, ind.Unit)
		kr.DueDate = final.Date
	}
	if !withStatus {
		return kr
	}
	if s.Latest != nil {
		kr.Current = valueText(s.Latest.Value, s.Latest.Level, ind.Unit)
	}
	if s.Progress != nil {
		kr.Score = math.Round(math.Min(*s.Progress, 1)*100) / 100
	}
	kr.Status = krStatus(s.State)
	kr.Confidence = krConfidence(s.Health)
	return kr
}

func valueText(v *float64, level, unit string) string {
	switch {
	case v != nil && unit != "":
		return strconv.FormatFloat(*v, 'f', -1, 64) + " " + unit
	case v != nil:
		return strconv.FormatFloat(*v, 'f', -1, 64)
	default:
		return level
	}
}

func krStatus(s impact.State) string {
	switch s {
	case impact.StateAchieved:
		return "Achieved"
	case impact.StateOnTrack:
		return "On Track"
	case impact.StateAtRisk, impact.StateStale:
		return "At Risk"
	case impact.StateOffTrack:
		return "Behind"
	case impact.StateNotStarted, impact.StateNoData, impact.StateBaselineOnly, impact.StateUnmeasured:
		return ""
	}
	return ""
}

func krConfidence(health string) string {
	switch health {
	case "green":
		return okr.ConfidenceHigh
	case "yellow":
		return okr.ConfidenceMedium
	case "red":
		return okr.ConfidenceLow
	}
	return ""
}
