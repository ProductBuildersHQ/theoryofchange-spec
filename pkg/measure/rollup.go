package measure

import (
	"slices"
	"sort"

	"github.com/ProductBuildersHQ/impact-model/pkg/graph"
	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
)

// rank orders states from worst to best for combining. Baseline only ranks
// with not started, as the specification says.
func rank(s impact.State) int {
	switch s {
	case impact.StateOffTrack:
		return 0
	case impact.StateAtRisk:
		return 1
	case impact.StateStale:
		return 2
	case impact.StateNoData:
		return 3
	case impact.StateNotStarted, impact.StateBaselineOnly:
		return 4
	case impact.StateOnTrack:
		return 5
	case impact.StateAchieved:
		return 6
	case impact.StateUnmeasured:
		return -1
	}
	return -1
}

// Worse returns the worse of two states. Unmeasured has no rank and loses to
// any measured state, so it never hides a judgment.
func Worse(a, b impact.State) impact.State {
	switch {
	case a == impact.StateUnmeasured:
		return b
	case b == impact.StateUnmeasured:
		return a
	case rank(a) <= rank(b):
		return a
	default:
		return b
	}
}

// Counted reports whether an indicator is part of the outcome's measurement:
// not rejected, and not a candidate or retired one that nobody collects.
func Counted(ind impact.Indicator) bool {
	if ind.Provenance.IsRejected() {
		return false
	}
	if ind.Selection == nil {
		return true
	}
	return ind.Selection.Status == impact.SelectionSelected
}

// OutcomeStates rolls the statuses of each outcome's own indicators up into
// one state per outcome. An indicator belongs to the outcome it measures and
// to every outcome it stands in for through proxyFor. With the worst method
// the outcome takes the worst indicator state; with the mean method it takes
// the state that the mean progress and mean expected progress of the
// indicators that have them would earn, falling back to the worst state when
// none has. An outcome with no counted indicator is unmeasured.
func OutcomeStates(doc *impact.Model, statuses map[string]impact.IndicatorStatus, method impact.RollupMethod, tolerance float64) map[string]impact.State {
	byOutcome := map[string][]impact.IndicatorStatus{}
	for _, ind := range doc.Indicators {
		st, ok := statuses[ind.ID]
		if !ok || !Counted(ind) {
			continue
		}
		byOutcome[ind.Measures] = append(byOutcome[ind.Measures], st)
		for _, id := range ind.ProxyFor {
			if id != ind.Measures {
				byOutcome[id] = append(byOutcome[id], st)
			}
		}
	}
	out := map[string]impact.State{}
	for _, o := range doc.Outcomes {
		out[o.ID] = rollIndicators(byOutcome[o.ID], method, tolerance)
	}
	return out
}

func rollIndicators(sts []impact.IndicatorStatus, method impact.RollupMethod, tolerance float64) impact.State {
	if len(sts) == 0 {
		return impact.StateUnmeasured
	}
	worst := impact.StateUnmeasured
	var sumP, sumE float64
	n := 0
	for _, s := range sts {
		worst = Worse(worst, s.State)
		if s.Progress != nil && s.Expected != nil {
			sumP += *s.Progress
			sumE += *s.Expected
			n++
		}
	}
	if method != impact.RollupMean || n == 0 {
		return worst
	}
	p, e := sumP/float64(n), sumE/float64(n)
	if tolerance <= 0 {
		tolerance = DefaultTolerance
	}
	switch {
	case p >= 1:
		return impact.StateAchieved
	case p >= e-tolerance:
		return impact.StateOnTrack
	case p >= e-2*tolerance:
		return impact.StateAtRisk
	default:
		return impact.StateOffTrack
	}
}

// PathwayStates computes each outcome's pathway state from its own state and
// from the states of what leads to it.
//
// Outcomes are taken from the preconditions up. For an outcome, the
// precondition signal combines its incoming causal links: the sources that
// belong to a precondition group are first reduced through the group (nested
// groups innermost first), then the group results and the remaining sources
// are combined with the outcome's logic. All takes the worst, any the best,
// and at-least-k the k-th best. Sources with no state of their own are left
// out, and when nothing is left there is no signal. A link with negative
// polarity is an inhibitor, not a precondition: it does not enter the
// logic, but when its source is on track or achieved, the outcome's pathway
// state is no better than at risk. The pathway state is the worse of the
// outcome's own state and the signal; an unmeasured own state defers to the
// signal.
func PathwayStates(doc *impact.Model, g *graph.Graph, own map[string]impact.State) (map[string]impact.State, error) {
	order, err := g.TopologicalOrder()
	if err != nil {
		return nil, err
	}
	p := &pathwayCalc{doc: doc, own: own, done: map[string]impact.State{}}
	p.index()
	for _, id := range order {
		if _, ok := doc.Outcome(id); !ok {
			continue
		}
		p.done[id] = p.outcome(id)
	}
	return p.done, nil
}

type pathwayCalc struct {
	doc  *impact.Model
	own  map[string]impact.State
	done map[string]impact.State

	groups   map[string]impact.PreconditionGroup
	memberOf map[string]bool
	byTarget map[string][]impact.PreconditionGroup
	incoming map[string][]impact.CausalLink
}

func (p *pathwayCalc) index() {
	p.groups = map[string]impact.PreconditionGroup{}
	p.memberOf = map[string]bool{}
	p.byTarget = map[string][]impact.PreconditionGroup{}
	p.incoming = map[string][]impact.CausalLink{}
	for _, g := range p.doc.Groups {
		if g.Rejected() {
			continue
		}
		p.groups[g.ID] = g
		p.byTarget[g.Target] = append(p.byTarget[g.Target], g)
		for _, m := range g.Members {
			p.memberOf[m] = true
		}
	}
	for _, l := range p.doc.Links {
		if !l.Rejected() && !l.Feedback {
			p.incoming[l.To] = append(p.incoming[l.To], l)
		}
	}
}

func (p *pathwayCalc) outcome(id string) impact.State {
	own := p.own[id]
	if own == "" {
		own = impact.StateUnmeasured
	}
	o, _ := p.doc.Outcome(id)

	grouped := map[string]bool{}
	for _, g := range p.byTarget[id] {
		for _, m := range g.Members {
			grouped[m] = true
		}
	}
	var items []impact.State
	var inhibited bool
	for _, l := range p.incoming[id] {
		if l.Polarity == impact.PolarityNegative {
			if s := p.done[l.From]; s == impact.StateOnTrack || s == impact.StateAchieved {
				inhibited = true
			}
			continue
		}
		if grouped[l.From] {
			continue
		}
		if s, ok := p.source(l.From); ok {
			items = append(items, s)
		}
	}
	for _, g := range p.byTarget[id] {
		if p.memberOf[g.ID] {
			continue // reached through the group that contains it
		}
		if s, ok := p.group(g); ok {
			items = append(items, s)
		}
	}

	state := own
	if signal, ok := combine(items, o.EffectiveLogic()); ok {
		state = Worse(own, signal)
	}
	if inhibited && (state == impact.StateOnTrack || state == impact.StateAchieved) {
		state = impact.StateAtRisk
	}
	return state
}

// source returns the pathway state of a link's source, if it is an outcome
// with a state of its own.
func (p *pathwayCalc) source(id string) (impact.State, bool) {
	s, ok := p.done[id]
	return s, ok && s != impact.StateUnmeasured
}

func (p *pathwayCalc) group(g impact.PreconditionGroup) (impact.State, bool) {
	var items []impact.State
	for _, m := range g.Members {
		if inner, ok := p.groups[m]; ok {
			if s, ok := p.group(inner); ok {
				items = append(items, s)
			}
			continue
		}
		if s, ok := p.source(m); ok {
			items = append(items, s)
		}
	}
	return combine(items, g.Logic)
}

// combine applies precondition logic to states. It reports false when there
// is nothing to combine.
func combine(items []impact.State, logic impact.Logic) (impact.State, bool) {
	if len(items) == 0 {
		return "", false
	}
	sorted := slices.Clone(items)
	sort.SliceStable(sorted, func(i, j int) bool { return rank(sorted[i]) > rank(sorted[j]) }) // best first
	switch logic.Type {
	case impact.LogicAny:
		return sorted[0], true
	case impact.LogicAtLeast:
		k := max(logic.K, 1)
		for len(sorted) < k {
			// Fewer preconditions have a state than the rule needs. The rest
			// are unknown, not failed, so they count as having no data.
			sorted = append(sorted, impact.StateNoData)
		}
		sort.SliceStable(sorted, func(i, j int) bool { return rank(sorted[i]) > rank(sorted[j]) })
		return sorted[k-1], true
	case impact.LogicAll:
		return sorted[len(sorted)-1], true
	}
	return sorted[len(sorted)-1], true
}
