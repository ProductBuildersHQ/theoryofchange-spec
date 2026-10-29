package toc

import "fmt"

// ElementKind names the kind of an identified element.
type ElementKind string

// Element kinds, in document order.
const (
	ElementOutcome        ElementKind = "outcome"
	ElementGroup          ElementKind = "preconditionGroup"
	ElementLink           ElementKind = "link"
	ElementIntervention   ElementKind = "intervention"
	ElementAssumption     ElementKind = "assumption"
	ElementIndicator      ElementKind = "indicator"
	ElementActor          ElementKind = "actor"
	ElementEvidence       ElementKind = "evidence"
	ElementExternalFactor ElementKind = "externalFactor"
	ElementRisk           ElementKind = "risk"
	ElementDecision       ElementKind = "decision"
	ElementPathway        ElementKind = "pathway"
)

// ElementRef locates an identified element in a document.
type ElementRef struct {
	ID   string
	Kind ElementKind
	// Path is the JSON pointer to the element, e.g. "/outcomes/3".
	Path string
}

// Elements lists every identified element in document order: outcomes,
// precondition groups, links, interventions, assumptions, indicators,
// actors, evidence, external factors, risks, then decisions. IDs are
// reported as written, including duplicates and empty IDs, so validation
// can report them.
func (d *Document) Elements() []ElementRef {
	if d == nil {
		return nil
	}
	var out []ElementRef
	add := func(kind ElementKind, field string, ids []string) {
		for i, id := range ids {
			out = append(out, ElementRef{ID: id, Kind: kind, Path: fmt.Sprintf("/%s/%d", field, i)})
		}
	}
	add(ElementOutcome, "outcomes", idsOf(d.Outcomes, func(v Outcome) string { return v.ID }))
	add(ElementGroup, "preconditionGroups", idsOf(d.Groups, func(v PreconditionGroup) string { return v.ID }))
	add(ElementLink, "links", idsOf(d.Links, func(v CausalLink) string { return v.ID }))
	add(ElementIntervention, "interventions", idsOf(d.Interventions, func(v Intervention) string { return v.ID }))
	add(ElementAssumption, "assumptions", idsOf(d.Assumptions, func(v Assumption) string { return v.ID }))
	add(ElementIndicator, "indicators", idsOf(d.Indicators, func(v Indicator) string { return v.ID }))
	add(ElementActor, "actors", idsOf(d.Actors, func(v Actor) string { return v.ID }))
	add(ElementEvidence, "evidence", idsOf(d.Evidence, func(v Evidence) string { return v.ID }))
	add(ElementExternalFactor, "externalFactors", idsOf(d.ExternalFactors, func(v ExternalFactor) string { return v.ID }))
	add(ElementRisk, "risks", idsOf(d.Risks, func(v Risk) string { return v.ID }))
	add(ElementDecision, "decisions", idsOf(d.Decisions, func(v Decision) string { return v.ID }))
	add(ElementPathway, "pathways", idsOf(d.Pathways, func(v Pathway) string { return v.ID }))
	return out
}

// Element returns the location of the first element with the given ID.
func (d *Document) Element(id string) (ElementRef, bool) {
	for _, e := range d.Elements() {
		if e.ID == id {
			return e, true
		}
	}
	return ElementRef{}, false
}

func idsOf[T any](items []T, id func(T) string) []string {
	out := make([]string, len(items))
	for i, it := range items {
		out[i] = id(it)
	}
	return out
}

func find[T any](items []T, id string, idOf func(T) string) (*T, bool) {
	for i := range items {
		if idOf(items[i]) == id {
			return &items[i], true
		}
	}
	return nil, false
}

// Outcome returns the outcome with the given ID.
func (d *Document) Outcome(id string) (*Outcome, bool) {
	return find(d.Outcomes, id, func(v Outcome) string { return v.ID })
}

// Group returns the precondition group with the given ID.
func (d *Document) Group(id string) (*PreconditionGroup, bool) {
	return find(d.Groups, id, func(v PreconditionGroup) string { return v.ID })
}

// Link returns the causal link with the given ID.
func (d *Document) Link(id string) (*CausalLink, bool) {
	return find(d.Links, id, func(v CausalLink) string { return v.ID })
}

// Intervention returns the intervention with the given ID.
func (d *Document) Intervention(id string) (*Intervention, bool) {
	return find(d.Interventions, id, func(v Intervention) string { return v.ID })
}

// Assumption returns the assumption with the given ID.
func (d *Document) Assumption(id string) (*Assumption, bool) {
	return find(d.Assumptions, id, func(v Assumption) string { return v.ID })
}

// Indicator returns the indicator with the given ID.
func (d *Document) Indicator(id string) (*Indicator, bool) {
	return find(d.Indicators, id, func(v Indicator) string { return v.ID })
}

// Actor returns the actor with the given ID.
func (d *Document) Actor(id string) (*Actor, bool) {
	return find(d.Actors, id, func(v Actor) string { return v.ID })
}

// EvidenceRecord returns the evidence record with the given ID.
func (d *Document) EvidenceRecord(id string) (*Evidence, bool) {
	return find(d.Evidence, id, func(v Evidence) string { return v.ID })
}

// IndicatorsMeasuring returns indicators whose Measures is id, in document
// order.
func (d *Document) IndicatorsMeasuring(id string) []Indicator {
	var out []Indicator
	for _, ind := range d.Indicators {
		if ind.Measures == id {
			out = append(out, ind)
		}
	}
	return out
}

// LinksTo returns links whose To is id (the outcome's preconditions), in
// document order.
func (d *Document) LinksTo(id string) []CausalLink {
	var out []CausalLink
	for _, l := range d.Links {
		if l.To == id {
			out = append(out, l)
		}
	}
	return out
}

// LinksFrom returns links whose From is id, in document order.
func (d *Document) LinksFrom(id string) []CausalLink {
	var out []CausalLink
	for _, l := range d.Links {
		if l.From == id {
			out = append(out, l)
		}
	}
	return out
}

// ImpactOutcomes returns outcomes at LevelImpact, in document order.
func (d *Document) ImpactOutcomes() []Outcome {
	var out []Outcome
	for _, o := range d.Outcomes {
		if o.Level == LevelImpact {
			out = append(out, o)
		}
	}
	return out
}

// EffectiveLogic returns the outcome's logic, defaulting to all-of.
func (o Outcome) EffectiveLogic() Logic {
	if o.Logic == nil {
		return Logic{Type: LogicAll}
	}
	return *o.Logic
}

// AllElementKinds returns element kinds in document order.
func AllElementKinds() []ElementKind {
	return []ElementKind{
		ElementOutcome, ElementGroup, ElementLink, ElementIntervention, ElementAssumption,
		ElementIndicator, ElementActor, ElementEvidence, ElementExternalFactor, ElementRisk,
		ElementDecision, ElementPathway,
	}
}

// Pathway returns the pathway with the given ID.
func (d *Document) Pathway(id string) (*Pathway, bool) {
	return find(d.Pathways, id, func(v Pathway) string { return v.ID })
}
