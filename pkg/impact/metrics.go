package impact

import "slices"

// Tier says where in the results chain the thing an indicator measures sits.
// It is derived from what the indicator measures and is never stored, so
// there is one source of truth for it.
type Tier string

// Tiers, from the start of the results chain to its end.
const (
	// TierActivity is a metric on an intervention: what was done, and the
	// resources it used. Resource use is measured here, as an indicator on the
	// intervention with a unit such as hours or dollars.
	TierActivity Tier = "activity"
	// TierOutput is a metric on an output-level outcome, such as reach.
	TierOutput Tier = "output"
	// TierOutcome is a metric on a short-term, medium-term, or long-term
	// outcome. These are the drivers of the change the program is for.
	TierOutcome Tier = "outcome"
	// TierImpact is a metric on an impact outcome, usually above the ceiling
	// of accountability and reported as a contribution.
	TierImpact Tier = "impact"
	// TierAssumption is a metric that monitors an assumption.
	TierAssumption Tier = "assumption"
	// TierPathway is a metric on a pathway as a whole.
	TierPathway Tier = "pathway"
)

// AllTiers returns every tier, from the start of the results chain to its end.
func AllTiers() []Tier {
	return []Tier{TierActivity, TierOutput, TierOutcome, TierImpact, TierAssumption, TierPathway}
}

// IndicatorTier returns the tier of the element the indicator measures, or the
// empty tier when it measures nothing that exists.
func (d *Model) IndicatorTier(ind Indicator) Tier {
	if o, ok := d.Outcome(ind.Measures); ok {
		switch o.Level {
		case LevelOutput:
			return TierOutput
		case LevelImpact:
			return TierImpact
		default:
			return TierOutcome
		}
	}
	if _, ok := d.Intervention(ind.Measures); ok {
		return TierActivity
	}
	if _, ok := d.Assumption(ind.Measures); ok {
		return TierAssumption
	}
	if _, ok := d.Pathway(ind.Measures); ok {
		return TierPathway
	}
	return ""
}

// IndicatorServes reports whether the indicator belongs to the pathway. It
// does when the indicator names the pathway in pathwayIds, or measures an
// outcome or an intervention the pathway lists, or measures the pathway
// itself. A pathway that was rejected is never served.
//
// Every rule and report asks this one question, so they agree on it.
func (d *Model) IndicatorServes(ind Indicator, pathwayID string) bool {
	p, ok := d.Pathway(pathwayID)
	if !ok || p.Rejected() {
		return false
	}
	return slices.Contains(ind.PathwayIDs, p.ID) ||
		slices.Contains(p.OutcomeIDs, ind.Measures) ||
		slices.Contains(p.InterventionIDs, ind.Measures) ||
		ind.Measures == p.ID
}

// IndicatorPathways returns the IDs of the pathways the indicator serves, in
// document order and without repeats.
func (d *Model) IndicatorPathways(ind Indicator) []string {
	var out []string
	for _, p := range d.Pathways {
		if d.IndicatorServes(ind, p.ID) {
			out = append(out, p.ID)
		}
	}
	return out
}
