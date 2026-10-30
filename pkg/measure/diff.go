package measure

import (
	"fmt"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

// Diff lists what changed between two status snapshots of the same theory:
// the state of indicators, links, and assumptions, and both states and the
// confidence of outcomes (their pathway state is the pathway's status).
// Changes come in a fixed order: indicators, outcomes, links, assumptions,
// each in the order of the later snapshot and then elements that only the
// earlier one had. An element that appears or disappears is a change in its
// presence.
func Diff(prev, cur *toc.Status) *toc.StatusDiff {
	d := &toc.StatusDiff{TocID: cur.TocID, FromAsOf: prev.AsOf, ToAsOf: cur.AsOf, Changes: []toc.StatusChange{}}
	add := func(id string, kind toc.ElementKind, field, from, to string) {
		if from != to {
			d.Changes = append(d.Changes, toc.StatusChange{ElementID: id, ElementKind: kind, Field: field, From: from, To: to})
		}
	}

	diffKind(d, toc.ElementIndicator, ids(cur.Indicators, func(s toc.IndicatorStatus) string { return s.IndicatorID }),
		ids(prev.Indicators, func(s toc.IndicatorStatus) string { return s.IndicatorID }),
		func(id string) {
			a, b := findBy(prev.Indicators, id, func(s toc.IndicatorStatus) string { return s.IndicatorID }), findBy(cur.Indicators, id, func(s toc.IndicatorStatus) string { return s.IndicatorID })
			add(id, toc.ElementIndicator, "state", string(a.State), string(b.State))
		})
	diffKind(d, toc.ElementOutcome, ids(cur.Outcomes, func(s toc.OutcomeStatus) string { return s.OutcomeID }),
		ids(prev.Outcomes, func(s toc.OutcomeStatus) string { return s.OutcomeID }),
		func(id string) {
			a, b := findBy(prev.Outcomes, id, func(s toc.OutcomeStatus) string { return s.OutcomeID }), findBy(cur.Outcomes, id, func(s toc.OutcomeStatus) string { return s.OutcomeID })
			add(id, toc.ElementOutcome, "state", string(a.State), string(b.State))
			add(id, toc.ElementOutcome, "pathwayState", string(a.PathwayState), string(b.PathwayState))
			add(id, toc.ElementOutcome, "confidence", confidence(a.Confidence), confidence(b.Confidence))
		})
	diffKind(d, toc.ElementLink, ids(cur.Links, func(s toc.LinkStatus) string { return s.LinkID }),
		ids(prev.Links, func(s toc.LinkStatus) string { return s.LinkID }),
		func(id string) {
			a, b := findBy(prev.Links, id, func(s toc.LinkStatus) string { return s.LinkID }), findBy(cur.Links, id, func(s toc.LinkStatus) string { return s.LinkID })
			add(id, toc.ElementLink, "state", string(a.State), string(b.State))
		})
	diffKind(d, toc.ElementAssumption, ids(cur.Assumptions, func(s toc.AssumptionState) string { return s.AssumptionID }),
		ids(prev.Assumptions, func(s toc.AssumptionState) string { return s.AssumptionID }),
		func(id string) {
			a, b := findBy(prev.Assumptions, id, func(s toc.AssumptionState) string { return s.AssumptionID }), findBy(cur.Assumptions, id, func(s toc.AssumptionState) string { return s.AssumptionID })
			add(id, toc.ElementAssumption, "status", string(a.Status), string(b.Status))
		})
	return d
}

// diffKind walks the later snapshot's elements, then those only the earlier
// one had, recording presence changes and calling compare for elements in
// both.
func diffKind(d *toc.StatusDiff, kind toc.ElementKind, cur, prev []string, compare func(id string)) {
	inPrev := map[string]bool{}
	for _, id := range prev {
		inPrev[id] = true
	}
	inCur := map[string]bool{}
	for _, id := range cur {
		inCur[id] = true
		if inPrev[id] {
			compare(id)
		} else {
			d.Changes = append(d.Changes, toc.StatusChange{ElementID: id, ElementKind: kind, Field: "presence", To: "added"})
		}
	}
	for _, id := range prev {
		if !inCur[id] {
			d.Changes = append(d.Changes, toc.StatusChange{ElementID: id, ElementKind: kind, Field: "presence", From: "present", To: "removed"})
		}
	}
}

func ids[T any](items []T, id func(T) string) []string {
	out := make([]string, len(items))
	for i, it := range items {
		out[i] = id(it)
	}
	return out
}

func findBy[T any](items []T, id string, idOf func(T) string) T {
	for _, it := range items {
		if idOf(it) == id {
			return it
		}
	}
	var zero T
	return zero
}

// confidence renders a confidence to two places, the precision at which a
// change is worth reporting.
func confidence(c *float64) string {
	if c == nil {
		return ""
	}
	return fmt.Sprintf("%.2f", *c)
}
