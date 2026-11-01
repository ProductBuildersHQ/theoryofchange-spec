package measure

import (
	"math"
	"slices"
	"sort"
	"time"

	"github.com/ProductBuildersHQ/impact-model/pkg/graph"
	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
)

// EffectiveAssumptions returns each assumption's status at asOf: the latest
// assessment at or before that date in any observation set, else the status
// written in the document (untested when it has none). The result follows
// document order.
func EffectiveAssumptions(doc *impact.Model, sets []*impact.ObservationSet, asOf time.Time) []impact.AssumptionState {
	type assessed struct {
		status impact.AssumptionStatus
		date   string
	}
	latest := map[string]assessed{}
	for _, set := range sets {
		for _, a := range set.Assessments {
			at, ok := parseDate(a.Date)
			if !ok || at.After(asOf) || a.Status == "" {
				continue
			}
			if cur, seen := latest[a.AssumptionID]; !seen || a.Date >= cur.date {
				latest[a.AssumptionID] = assessed{a.Status, a.Date}
			}
		}
	}
	out := make([]impact.AssumptionState, 0, len(doc.Assumptions))
	for _, a := range doc.Assumptions {
		if a.Provenance.IsRejected() {
			continue
		}
		s := impact.AssumptionState{AssumptionID: a.ID, Status: a.Status, Source: impact.SourceDocument}
		if s.Status == "" {
			s.Status = impact.AssumptionUntested
		}
		if as, ok := latest[a.ID]; ok {
			s.Status, s.Source, s.AssessedAt = as.status, impact.SourceAssessment, as.date
		}
		out = append(out, s)
	}
	return out
}

// Weights for the parts of link confidence, as in the specification.
var (
	strengthScore = map[impact.Strength]float64{
		impact.StrengthStrong: 1, impact.StrengthModerate: 0.7, impact.StrengthWeak: 0.4, impact.StrengthAnecdotal: 0.2,
	}
	assumptionScore = map[impact.AssumptionStatus]float64{
		impact.AssumptionValidated: 1, impact.AssumptionHolding: 0.8, impact.AssumptionUntested: 0.5,
		impact.AssumptionAtRisk: 0.25, impact.AssumptionInvalidated: 0,
	}
)

// noAssumptionScore is what a link that names no assumption scores for its
// assumptions: the same as an untested one, since nothing has been shown.
const noAssumptionScore = 0.5

// underpinning returns the assumptions that bear on a link: those it names,
// and those that name it in appliesTo.
func underpinning(doc *impact.Model, link impact.CausalLink) []string {
	ids := slices.Clone(link.AssumptionIDs)
	for _, a := range doc.Assumptions {
		if !a.Provenance.IsRejected() && slices.Contains(a.AppliesTo, link.ID) && !slices.Contains(ids, a.ID) {
			ids = append(ids, a.ID)
		}
	}
	return ids
}

// LinkStatuses computes the state and confidence of every link.
//
// A link is broken when an assumption it rests on is invalidated, and at
// risk when one is at risk. Confidence is, from 0 to 1,
// weights.Evidence·evidence + weights.Assumptions·assumptions +
// weights.Rationale·rationale, where evidence is the strength of the best
// supporting evidence less half that of the best contradicting evidence
// (strong 1, moderate 0.7, weak 0.4, anecdotal 0.2), assumptions is the
// lowest score among the link's assumptions (validated 1, holding 0.8,
// untested 0.5, at risk 0.25, invalidated 0; 0.5 if it names none), and
// rationale is 1 if the link gives one and 0 if not.
func LinkStatuses(doc *impact.Model, assumptions []impact.AssumptionState, w impact.ConfidenceWeights) []impact.LinkStatus {
	status := map[string]impact.AssumptionStatus{}
	for _, a := range assumptions {
		status[a.AssumptionID] = a.Status
	}
	out := make([]impact.LinkStatus, 0, len(doc.Links))
	for _, l := range doc.Links {
		if l.Rejected() {
			continue
		}
		state := impact.LinkStateOK
		aScore := noAssumptionScore
		first := true
		for _, id := range underpinning(doc, l) {
			s, ok := status[id]
			if !ok {
				continue // rejected or unknown
			}
			switch s {
			case impact.AssumptionInvalidated:
				state = impact.LinkStateBroken
			case impact.AssumptionAtRisk:
				if state != impact.LinkStateBroken {
					state = impact.LinkStateAtRisk
				}
			case impact.AssumptionUntested, impact.AssumptionHolding, impact.AssumptionValidated:
			}
			if sc := assumptionScore[s]; first || sc < aScore {
				aScore, first = sc, false
			}
		}
		rationale := 0.0
		if l.Rationale != "" {
			rationale = 1
		}
		conf := w.Evidence*evidenceScore(doc, l.EvidenceIDs) + w.Assumptions*aScore + w.Rationale*rationale
		out = append(out, impact.LinkStatus{LinkID: l.ID, State: state, Confidence: round6(math.Min(math.Max(conf, 0), 1))})
	}
	return out
}

func evidenceScore(doc *impact.Model, ids []string) float64 {
	var support, against float64
	for _, id := range ids {
		e, ok := doc.EvidenceRecord(id)
		if !ok {
			continue
		}
		s := strengthScore[e.Strength]
		if e.Direction == impact.EvidenceContradicts {
			against = math.Max(against, s)
		} else {
			support = math.Max(support, s)
		}
	}
	return math.Max(support-against/2, 0)
}

func round6(v float64) float64 { return math.Round(v*1e6) / 1e6 }

// Propagation is what broken assumptions mean downstream.
type Propagation struct {
	// Broken lists, per outcome, the invalidated or at-risk assumptions
	// that bear on it or on anything upstream of it, sorted.
	Broken map[string][]string
	// Cap is the best pathway state an outcome may have: at risk for an
	// outcome with a flagged assumption, off track when the assumption is
	// critical and invalidated. Outcomes with no flag have no entry.
	Cap map[string]impact.State
}

// Propagate marks the outcomes affected by assumptions that are invalidated
// or at risk. An assumption bears on what it applies to: a link, an
// outcome, or an intervention (through the outcomes it targets). It flags
// those outcomes and everything downstream of them.
func Propagate(doc *impact.Model, g *graph.Graph, assumptions []impact.AssumptionState) Propagation {
	status := map[string]impact.AssumptionStatus{}
	for _, a := range assumptions {
		status[a.AssumptionID] = a.Status
	}
	// direct[outcome] holds the assumptions that bear on it directly.
	direct := map[string]map[string]bool{}
	flag := func(outcome, assumption string) {
		if direct[outcome] == nil {
			direct[outcome] = map[string]bool{}
		}
		direct[outcome][assumption] = true
	}
	for _, a := range doc.Assumptions {
		if s := status[a.ID]; s != impact.AssumptionInvalidated && s != impact.AssumptionAtRisk {
			continue
		}
		targets := map[string]bool{}
		for _, id := range a.AppliesTo {
			targets[id] = true
		}
		for _, l := range doc.Links {
			if !l.Rejected() && (slices.Contains(l.AssumptionIDs, a.ID) || targets[l.ID]) {
				flag(l.To, a.ID)
			}
		}
		for _, o := range doc.Outcomes {
			if targets[o.ID] && !o.Rejected() {
				flag(o.ID, a.ID)
			}
		}
		for _, in := range doc.Interventions {
			if targets[in.ID] && !in.Rejected() {
				for _, t := range in.Targets {
					flag(t, a.ID)
				}
			}
		}
	}

	critical := map[string]bool{}
	for _, a := range doc.Assumptions {
		critical[a.ID] = a.Criticality == impact.CriticalityCritical
	}
	p := Propagation{Broken: map[string][]string{}, Cap: map[string]impact.State{}}
	add := func(outcome, assumption string) {
		if !slices.Contains(p.Broken[outcome], assumption) {
			p.Broken[outcome] = append(p.Broken[outcome], assumption)
		}
		limit := impact.StateAtRisk
		if critical[assumption] && status[assumption] == impact.AssumptionInvalidated {
			limit = impact.StateOffTrack
		}
		if cur, ok := p.Cap[outcome]; ok {
			limit = Worse(cur, limit)
		}
		p.Cap[outcome] = limit
	}
	for outcome, ids := range direct {
		for id := range ids {
			add(outcome, id)
			for _, d := range g.Downstream(outcome) {
				if _, isOutcome := doc.Outcome(d); isOutcome {
					add(d, id)
				}
			}
		}
	}
	for _, ids := range p.Broken {
		sort.Strings(ids)
	}
	return p
}

// Apply lowers each pathway state, in place, to the best its flags allow.
func (p Propagation) Apply(pathway map[string]impact.State) {
	for id, limit := range p.Cap {
		cur, ok := pathway[id]
		if !ok || cur == "" {
			cur = impact.StateUnmeasured
		}
		pathway[id] = Worse(cur, limit)
	}
}

// Confidence computes how well supported each outcome's route to impact is.
//
// A path's confidence is the product of the confidences of its links, taking
// the strongest link when two outcomes are joined by several. An outcome's
// confidence is the best of its paths to an impact outcome; an impact
// outcome has confidence 1, and one with no path has none. The confidence of
// an impact outcome as a headline is the best complete chain into it: the
// highest outcome confidence among the outcomes that start a chain, that is,
// that no causal link leads into, counting only paths that end at that
// impact.
func Confidence(doc *impact.Model, g *graph.Graph, links []impact.LinkStatus) (outcomes map[string]float64, impacts map[string]float64) {
	byPair := map[[2]string]float64{}
	for _, l := range doc.Links {
		if l.Rejected() || l.Feedback {
			continue
		}
		for _, ls := range links {
			if ls.LinkID == l.ID {
				k := [2]string{l.From, l.To}
				byPair[k] = math.Max(byPair[k], ls.Confidence)
			}
		}
	}
	pathConf := func(path []string) (float64, bool) {
		c := 1.0
		for i := 0; i+1 < len(path); i++ {
			lc, ok := byPair[[2]string{path[i], path[i+1]}]
			if !ok {
				return 0, false // an edge that is an intervention, not a link
			}
			c *= lc
		}
		return c, true
	}

	outcomes, impacts = map[string]float64{}, map[string]float64{}
	hasIncoming := map[string]bool{}
	for _, l := range doc.Links {
		if !l.Rejected() && !l.Feedback {
			hasIncoming[l.To] = true
		}
	}
	for _, o := range doc.Outcomes {
		if o.Rejected() {
			continue
		}
		best, found := 0.0, false
		for _, path := range g.PathsToImpact(o.ID, 0) {
			if c, ok := pathConf(path); ok && (!found || c > best) {
				best, found = c, true
			}
		}
		if found {
			outcomes[o.ID] = round6(best)
		}
	}
	for _, o := range doc.Outcomes {
		if o.Rejected() || o.Level != impact.LevelImpact {
			continue
		}
		best, found := 0.0, false
		for _, start := range doc.Outcomes {
			if start.Rejected() || start.ID == o.ID || hasIncoming[start.ID] {
				continue
			}
			for _, path := range g.PathsToImpact(start.ID, 0) {
				if path[len(path)-1] != o.ID {
					continue
				}
				if c, ok := pathConf(path); ok && (!found || c > best) {
					best, found = c, true
				}
			}
		}
		if found {
			impacts[o.ID] = round6(best)
		}
	}
	return outcomes, impacts
}
