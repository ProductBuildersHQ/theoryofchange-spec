package prismx

import (
	"fmt"

	"github.com/grokify/prism-roadmap/journey"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/measure"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

// ToOutcomeJourneys exports each numeric indicator as an outcome journey: where
// the measure stands now and where its dated targets want it to be. An
// indicator becomes a journey when it is counted, is measured as a number,
// and has a baseline or latest value and at least one numeric target. The
// journey's name is the indicator's, its description the outcome it
// measures, and its persona the first actor who changes. A target's
// confidence is the confidence of the outcome's route to impact, when a
// status is given. st may be nil.
func ToOutcomeJourneys(doc *toc.Document, st *toc.Status) ([]journey.OutcomeJourney, error) {
	if doc == nil {
		return nil, fmt.Errorf("prismx: nil document")
	}
	latest := map[string]*toc.Observation{}
	confidence := map[string]float64{}
	if st != nil {
		for _, s := range st.Indicators {
			latest[s.IndicatorID] = s.Latest
		}
		for _, o := range st.Outcomes {
			if o.Confidence != nil {
				confidence[o.OutcomeID] = *o.Confidence
			}
		}
	}
	out := []journey.OutcomeJourney{}
	for _, ind := range doc.Indicators {
		if !measure.Counted(ind) || ind.MeasurementType == toc.MeasurementQualitative {
			continue
		}
		var targets []journey.OutcomeTargetState
		for _, t := range ind.Targets {
			if t.Value == nil {
				continue
			}
			targets = append(targets, journey.OutcomeTargetState{
				PeriodID: t.Date, Value: *t.Value, Unit: ind.Unit, Summary: t.Label, Confidence: confidence[ind.Measures],
			})
		}
		current := currentState(ind, latest[ind.ID])
		if len(targets) == 0 || current == nil {
			continue
		}
		oj := journey.OutcomeJourney{
			ID: ind.ID, Name: ind.Name, Description: outcomeTitle(doc, ind.Measures),
			CurrentState: current, TargetStates: targets,
			Metadata: map[string]any{"tocId": doc.ID, "indicatorId": ind.ID, "measures": ind.Measures},
		}
		if o, ok := doc.Outcome(ind.Measures); ok {
			oj.Tags = []string{string(o.Level)}
			if len(o.ActorIDs) > 0 {
				if a, ok := doc.Actor(o.ActorIDs[0]); ok {
					oj.Persona = a.Name
				}
			}
		}
		out = append(out, oj)
	}
	return out, nil
}

// currentState is the latest observation, else the baseline.
func currentState(ind toc.Indicator, latest *toc.Observation) *journey.OutcomeState {
	switch {
	case latest != nil && latest.Value != nil:
		return &journey.OutcomeState{Value: *latest.Value, Unit: ind.Unit, MeasuredAt: latest.Date, Summary: latest.Note}
	case ind.Baseline != nil && ind.Baseline.Value != nil:
		return &journey.OutcomeState{Value: *ind.Baseline.Value, Unit: ind.Unit, MeasuredAt: ind.Baseline.Date, Summary: ind.Baseline.Source}
	}
	return nil
}

func outcomeTitle(doc *toc.Document, id string) string {
	if o, ok := doc.Outcome(id); ok {
		return o.Title
	}
	return ""
}
