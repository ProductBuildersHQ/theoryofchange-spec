package measure

import (
	"slices"
	"strconv"
	"time"

	"github.com/ProductBuildersHQ/impact-model/pkg/graph"
	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
)

// Evaluate computes the status of an impact model at asOf from observation sets.
// g is the impact model's graph. The result is a pure function of its inputs: the
// same impact model, observations, date, and options always give a byte-identical
// snapshot, and the clock is never read.
//
// The steps are: derive values for theme-count and roll-up indicators from
// the observations; grade every indicator; roll indicators up to outcomes;
// work out each assumption's effective status and each link's state and
// confidence; combine preconditions into pathway states; cap the pathway
// states of outcomes downstream of broken assumptions; and score confidence.
func Evaluate(doc *impact.Model, g *graph.Graph, sets []*impact.ObservationSet, asOf time.Time, opts impact.StatusOptions) (*impact.Status, error) {
	observations := derivedObservations(doc, sets, asOf)

	var indicators []impact.IndicatorStatus
	byIndicator := map[string]impact.IndicatorStatus{}
	for _, ind := range doc.Indicators {
		if ind.Provenance.IsRejected() {
			continue
		}
		s := EvaluateIndicator(ind, observations, Options{AsOf: asOf, Tolerance: opts.Tolerance})
		indicators = append(indicators, s)
		byIndicator[ind.ID] = s
	}

	own := OutcomeStates(doc, byIndicator, opts.Rollup, opts.Tolerance)
	assumptions := EffectiveAssumptions(doc, sets, asOf)
	links := LinkStatuses(doc, assumptions, opts.Weights)
	pathway, err := PathwayStates(doc, g, own)
	if err != nil {
		return nil, err
	}
	prop := Propagate(doc, g, assumptions)
	prop.Apply(pathway)
	outcomeConf, impactConf := Confidence(doc, g, links)

	st := &impact.Status{
		Kind: impact.KindStatus, SpecVersion: impact.SpecVersion, ModelID: doc.ID, AsOf: asOf.Format(time.DateOnly), Options: opts,
		Indicators: indicators, Links: links, Assumptions: assumptions,
	}
	if st.Indicators == nil {
		st.Indicators = []impact.IndicatorStatus{}
	}
	for _, o := range doc.Outcomes {
		if o.Rejected() {
			continue
		}
		os := impact.OutcomeStatus{
			OutcomeID: o.ID, State: own[o.ID], PathwayState: pathway[o.ID], Health: Health(pathway[o.ID]),
			IndicatorIDs: indicatorIDs(doc, byIndicator, o.ID), AssumptionBroken: prop.Broken[o.ID],
		}
		if c, ok := outcomeConf[o.ID]; ok {
			os.Confidence = &c
		}
		if os.PathwayState == "" {
			os.PathwayState = impact.StateUnmeasured
		}
		st.Outcomes = append(st.Outcomes, os)
		if o.Level == impact.LevelImpact {
			is := impact.ImpactStatus{OutcomeID: o.ID, PathwayState: os.PathwayState, Health: os.Health}
			if c, ok := impactConf[o.ID]; ok {
				is.Confidence = &c
			}
			st.Overall = append(st.Overall, is)
		}
	}
	return st, nil
}

// indicatorIDs lists, in document order, the counted indicators that measure
// or stand in for an outcome.
func indicatorIDs(doc *impact.Model, statuses map[string]impact.IndicatorStatus, outcome string) []string {
	var out []string
	for _, ind := range doc.Indicators {
		if _, ok := statuses[ind.ID]; !ok || !Counted(ind) {
			continue
		}
		if ind.Measures == outcome || slices.Contains(ind.ProxyFor, outcome) {
			out = append(out, ind.ID)
		}
	}
	return out
}

// derivedObservations returns the observations of all sets plus the values
// that follow from them: the count of tagged stories for each theme-count
// indicator, the combined value for each roll-up indicator that has none of its
// own, and the value worked out for each composed indicator that has none.
// Derived values are dated asOf, since they describe that date.
func derivedObservations(doc *impact.Model, sets []*impact.ObservationSet, asOf time.Time) []impact.Observation {
	var obs []impact.Observation
	for _, set := range sets {
		obs = append(obs, set.Observations...)
	}
	date := asOf.Format(time.DateOnly)

	for _, ind := range doc.Indicators {
		// With no observation sets there are no stories to count, which is
		// no data rather than zero.
		if len(sets) == 0 || ind.Provenance.IsRejected() || ind.MeasurementType != impact.MeasurementThemeCount {
			continue
		}
		n := float64(ThemeCount(doc, ind, sets, asOf))
		obs = append(obs, impact.Observation{IndicatorID: ind.ID, Date: date, Value: &n, Source: "story theme count", Quality: impact.QualityReported})
	}

	// Roll-ups can depend on other roll-ups, so repeat until nothing new is
	// derived; an impact model with more passes than indicators has a loop, which
	// validation does not forbid, and is cut off there.
	done := map[string]bool{}
	for pass := 0; pass < len(doc.Indicators); pass++ {
		added := false
		for _, ind := range doc.Indicators {
			if done[ind.ID] || ind.Provenance.IsRejected() {
				continue
			}
			if cres, composed := Compose(ind, obs, asOf); composed {
				if cres.Direct {
					done[ind.ID] = true
				} else if cres.Value != nil {
					obs = append(obs, impact.Observation{IndicatorID: ind.ID, Date: date, Value: cres.Value, Quality: cres.Quality, Source: "worked out from " + strconv.Itoa(len(ind.ComposedOf.IndicatorIDs)) + " indicators"})
					done[ind.ID] = true
					added = true
				}
				continue
			}
			res, ok := Rollup(doc, ind, obs, asOf)
			if !ok {
				continue
			}
			if res.Direct || res.Value == nil {
				if res.Direct {
					done[ind.ID] = true
				}
				continue
			}
			obs = append(obs, impact.Observation{IndicatorID: ind.ID, Date: date, Value: res.Value, Source: "rolled up from " + strconv.Itoa(len(ind.RollupFrom)) + " indicators"})
			done[ind.ID] = true
			added = true
		}
		if !added {
			break
		}
	}
	return obs
}

// ThemeCount is the number of distinct stories, dated at or before asOf and
// tagged with any of the indicator's themes, in the observation sets. If the
// indicator names pathways, only stories about an outcome of one of those
// pathways count.
func ThemeCount(doc *impact.Model, ind impact.Indicator, sets []*impact.ObservationSet, asOf time.Time) int {
	var allowed map[string]bool
	if len(ind.PathwayIDs) > 0 {
		allowed = map[string]bool{}
		for _, id := range ind.PathwayIDs {
			if p, ok := doc.Pathway(id); ok {
				for _, o := range p.OutcomeIDs {
					allowed[o] = true
				}
			}
		}
	}
	seen := map[string]bool{}
	for _, set := range sets {
		for _, s := range set.Stories {
			at, ok := parseDate(s.Date)
			if !ok || at.After(asOf) || seen[s.ID] {
				continue
			}
			if !slices.ContainsFunc(s.ThemeIDs, func(t string) bool { return slices.Contains(ind.ThemeIDs, t) }) {
				continue
			}
			if allowed != nil && !slices.ContainsFunc(s.OutcomeIDs, func(o string) bool { return allowed[o] }) {
				continue
			}
			seen[s.ID] = true
		}
	}
	return len(seen)
}

// LatestDate returns the latest date that appears in the observation sets:
// an observation, an assessment, a story, or a note. It is the natural as-of
// date when none is given, and reports false when the sets hold no date.
func LatestDate(sets []*impact.ObservationSet) (time.Time, bool) {
	var latest time.Time
	found := false
	consider := func(s string) {
		if at, ok := parseDate(s); ok && (!found || at.After(latest)) {
			latest, found = at, true
		}
	}
	for _, set := range sets {
		for _, o := range set.Observations {
			consider(o.Date)
		}
		for _, a := range set.Assessments {
			consider(a.Date)
		}
		for _, s := range set.Stories {
			consider(s.Date)
		}
		for _, n := range set.Notes {
			consider(n.Date)
		}
	}
	return latest, found
}
