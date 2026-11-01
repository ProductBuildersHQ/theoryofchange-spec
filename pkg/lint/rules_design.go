package lint

import (
	"fmt"
	"slices"

	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
	"github.com/ProductBuildersHQ/impact-model/pkg/validate"
)

// provEl is an element that can carry a provenance.
type provEl struct {
	array string
	index int
	id    string
	prov  *impact.Provenance
	// accountable is false for an outcome the organization does not answer
	// for; every other kind counts as accountable.
	accountable bool
}

// provenanceElements lists the elements that can carry a provenance, in
// document order, without those that were rejected.
func provenanceElements(c *Context) []provEl {
	var out []provEl
	add := func(array string, i int, id string, p *impact.Provenance, accountable bool) {
		if !p.IsRejected() {
			out = append(out, provEl{array, i, id, p, accountable})
		}
	}
	d := c.Doc
	for i, x := range d.Outcomes {
		add("outcomes", i, x.ID, x.Provenance, x.Accountable() && c.Graph.BelowCeiling(x.ID))
	}
	for i, x := range d.Groups {
		add("preconditionGroups", i, x.ID, x.Provenance, true)
	}
	for i, x := range d.Links {
		add("links", i, x.ID, x.Provenance, true)
	}
	for i, x := range d.Interventions {
		add("interventions", i, x.ID, x.Provenance, true)
	}
	for i, x := range d.Assumptions {
		add("assumptions", i, x.ID, x.Provenance, true)
	}
	for i, x := range d.Indicators {
		add("indicators", i, x.ID, x.Provenance, true)
	}
	for i, x := range d.Actors {
		add("actors", i, x.ID, x.Provenance, true)
	}
	for i, x := range d.ExternalFactors {
		add("externalFactors", i, x.ID, x.Provenance, true)
	}
	for i, x := range d.Risks {
		add("risks", i, x.ID, x.Provenance, true)
	}
	for i, x := range d.Pathways {
		add("pathways", i, x.ID, x.Provenance, true)
	}
	for i, x := range d.Metrics {
		add("metrics", i, x.ID, x.Provenance, true)
	}
	return out
}

func activePathways(d *impact.Model) []indexed[impact.Pathway] {
	var out []indexed[impact.Pathway]
	for i, p := range d.Pathways {
		if !p.Rejected() {
			out = append(out, indexed[impact.Pathway]{i, p})
		}
	}
	return out
}

func init() {
	// ---- interventions, actors, evidence ----------------------------------
	register("IMP-INT-002", validate.SeverityWarning, "Every intervention has an owner.", func(c *Context) []Finding {
		var out []Finding
		for i, in := range c.Doc.Interventions {
			if !in.Provenance.IsRejected() && in.OwnerID == "" {
				out = append(out, Finding{Path: at("interventions", i), Element: in.ID, Message: "nobody is named as responsible for this intervention"})
			}
		}
		return out
	})

	register("IMP-ACT-001", validate.SeverityInfo, "Outcomes in the sphere of influence name the actors who must change.", func(c *Context) []Finding {
		var out []Finding
		for _, o := range activeOutcomes(c.Doc) {
			if o.v.Sphere == impact.SphereInfluence && len(o.v.ActorIDs) == 0 {
				out = append(out, Finding{Path: at("outcomes", o.i), Element: o.v.ID, Message: "this outcome depends on others changing, but no actor is named"})
			}
		}
		return out
	})

	register("IMP-EVD-001", validate.SeverityInfo, "Links into impact outcomes cite evidence.", func(c *Context) []Finding {
		var out []Finding
		for _, l := range activeLinks(c.Doc) {
			if lv, ok := outcomeLevel(c.Doc, l.v.To); ok && lv == impact.LevelImpact && len(l.v.EvidenceIDs) == 0 {
				out = append(out, Finding{Path: at("links", l.i), Element: l.v.ID, Message: "the claim that this leads to the impact cites no evidence"})
			}
		}
		return out
	})

	// ---- pathways and metrics ---------------------------------------------
	register("IMP-PTH-001", validate.SeverityWarning, "Every pathway contains an outcome that leads to impact.", func(c *Context) []Finding {
		var out []Finding
		for _, p := range activePathways(c.Doc) {
			leads := false
			for _, id := range p.v.OutcomeIDs {
				lv, _ := outcomeLevel(c.Doc, id)
				leads = leads || lv == impact.LevelImpact || c.Graph.HasPathToImpact(id)
			}
			if !leads {
				out = append(out, Finding{Path: at("pathways", p.i), Element: p.v.ID, Message: "no outcome in this pathway leads to an impact outcome"})
			}
		}
		return out
	})

	register("IMP-PTH-002", validate.SeverityInfo, "Every pathway has a selected indicator.", func(c *Context) []Finding {
		var out []Finding
		for _, p := range activePathways(c.Doc) {
			found := false
			for _, ind := range activeIndicators(c.Doc) {
				found = found || (isSelected(ind.v) && c.Doc.IndicatorServes(ind.v, p.v.ID))
			}
			if !found {
				out = append(out, Finding{Path: at("pathways", p.i), Element: p.v.ID, Message: "no selected indicator serves this pathway"})
			}
		}
		return out
	})

	register("IMP-PTH-003", validate.SeverityInfo, "A pathway's outcomes agree with the pathways its indicators name.", func(c *Context) []Finding {
		var out []Finding
		for _, ind := range activeIndicators(c.Doc) {
			if len(ind.v.PathwayIDs) == 0 {
				continue
			}
			for _, p := range activePathways(c.Doc) {
				inList := slices.Contains(ind.v.PathwayIDs, p.v.ID)
				inPathway := slices.Contains(p.v.OutcomeIDs, ind.v.Measures)
				switch {
				case inPathway && !inList:
					out = append(out, Finding{Path: at("indicators", ind.i) + "/pathwayIds", Element: ind.v.ID,
						Message: fmt.Sprintf("it measures %q, which pathway %q lists, but does not name that pathway", ind.v.Measures, p.v.ID)})
				case inList && !inPathway && isOutcomeID(c.Doc, ind.v.Measures):
					out = append(out, Finding{Path: at("indicators", ind.i) + "/pathwayIds", Element: ind.v.ID,
						Message: fmt.Sprintf("it names pathway %q, which does not list %q among its outcomes", p.v.ID, ind.v.Measures)})
				}
			}
		}
		return out
	})

	register("IMP-MET-001", validate.SeverityWarning, "A common metric is measured in every pathway, or the omission is explained.", func(c *Context) []Finding {
		var out []Finding
		for i, m := range c.Doc.Metrics {
			if m.Layer != impact.LayerCommon || m.Provenance.IsRejected() {
				continue
			}
			explained := false
			var uses []impact.Indicator
			for _, ind := range activeIndicators(c.Doc) {
				if ind.v.MetricID == m.ID {
					uses = append(uses, ind.v)
					explained = explained || (ind.v.Selection != nil && ind.v.Selection.Rationale != "")
				}
			}
			if explained {
				continue
			}
			for _, p := range activePathways(c.Doc) {
				if !slices.ContainsFunc(uses, func(x impact.Indicator) bool { return c.Doc.IndicatorServes(x, p.v.ID) }) {
					out = append(out, Finding{Path: at("metrics", i), Element: m.ID, Message: fmt.Sprintf("common metric is not measured in pathway %q; add an indicator, or explain the omission in an indicator's selection rationale", p.v.ID)})
				}
			}
		}
		return out
	})

	register("IMP-MET-003", validate.SeverityWarning, "A weighted mean can be computed: its weight is measured wherever it is used.", func(c *Context) []Finding {
		var out []Finding
		for i, m := range c.Doc.Metrics {
			if m.Aggregation != impact.AggregationWeightedMean || m.WeightMetricID == "" || m.Provenance.IsRejected() {
				continue
			}
			for _, p := range activePathways(c.Doc) {
				uses, weighted := false, false
				for _, ind := range activeIndicators(c.Doc) {
					in := c.Doc.IndicatorServes(ind.v, p.v.ID)
					uses = uses || (in && ind.v.MetricID == m.ID)
					weighted = weighted || (in && ind.v.MetricID == m.WeightMetricID)
				}
				if uses && !weighted {
					out = append(out, Finding{Path: at("metrics", i), Element: m.ID, Message: fmt.Sprintf("pathway %q measures this metric but not its weight %q, so the roll-up cannot be computed", p.v.ID, m.WeightMetricID)})
				}
			}
		}
		return out
	})

	// ---- audiences, causes, responsibility, equity --------------------------
	register("IMP-AUD-001", validate.SeverityInfo, "Every audience is the target of something reported.", func(c *Context) []Finding {
		targeted := map[string]bool{}
		for _, ind := range activeIndicators(c.Doc) {
			if isSelected(ind.v) {
				for _, a := range ind.v.AudienceIDs {
					targeted[a] = true
				}
			}
		}
		for _, o := range activeOutcomes(c.Doc) {
			for _, a := range o.v.AudienceIDs {
				targeted[a] = true
			}
		}
		var out []Finding
		for i, a := range c.Doc.Audiences {
			if !targeted[a.ID] {
				out = append(out, Finding{Path: at("audiences", i), Element: a.ID, Message: "no selected indicator or outcome is reported to this audience"})
			}
		}
		return out
	})

	register("IMP-CAU-001", validate.SeverityWarning, "Each deeper cause is addressed by an outcome or set aside with a reason.", func(c *Context) []Finding {
		addressed := addressedCauses(c.Doc)
		var out []Finding
		for i, cause := range c.Doc.Causes {
			if !cause.Rejected() && cause.Level != impact.CauseImmediate && !addressed[cause.ID] && cause.OutOfScope == nil {
				out = append(out, Finding{Path: at("causes", i), Element: cause.ID, Message: "no outcome addresses this cause; address it, or set it aside with a reason"})
			}
		}
		return out
	})

	register("IMP-CAU-002", validate.SeverityInfo, "Outcomes address root causes, not only immediate ones.", func(c *Context) []Finding {
		if len(c.Doc.Causes) == 0 {
			return nil
		}
		addressed := addressedCauses(c.Doc)
		for _, cause := range c.Doc.Causes {
			if !cause.Rejected() && cause.Level != impact.CauseImmediate && addressed[cause.ID] {
				return nil
			}
		}
		return []Finding{{Path: "/causes", Message: "causes are declared but no outcome addresses an underlying or structural one"}}
	})

	register("IMP-RSP-001", validate.SeverityWarning, "Shared and partner outcomes name who delivers them.", func(c *Context) []Finding {
		var out []Finding
		for _, o := range activeOutcomes(c.Doc) {
			need := 0
			switch o.v.Responsibility {
			case impact.ResponsibilityPartner:
				need = 1
			case impact.ResponsibilityShared:
				need = 2
			case impact.ResponsibilityOwn, impact.ResponsibilityUnaddressed, "":
			}
			if len(o.v.DeliveredBy) < need {
				out = append(out, Finding{Path: at("outcomes", o.i) + "/deliveredBy", Element: o.v.ID,
					Message: fmt.Sprintf("a %s outcome names at least %d delivering actor(s); this names %d", o.v.Responsibility, need, len(o.v.DeliveredBy))})
			}
		}
		return out
	})

	register("IMP-RSP-002", validate.SeverityInfo, "Partner and unaddressed outcomes on the way to impact are still measured.", func(c *Context) []Finding {
		measured := map[string]bool{}
		for _, ind := range activeIndicators(c.Doc) {
			measured[ind.v.Measures] = true
			for _, id := range ind.v.ProxyFor {
				measured[id] = true
			}
		}
		var out []Finding
		for _, o := range activeOutcomes(c.Doc) {
			r := o.v.Responsibility
			if (r == impact.ResponsibilityPartner || r == impact.ResponsibilityUnaddressed) && !measured[o.v.ID] && (o.v.Level == impact.LevelImpact || c.Graph.HasPathToImpact(o.v.ID)) {
				out = append(out, Finding{Path: at("outcomes", o.i), Element: o.v.ID, Message: "this outcome is not the organization's to deliver but the impact model depends on it; monitor it with an indicator"})
			}
		}
		return out
	})

	register("IMP-EQU-001", validate.SeverityWarning, "Outcomes for target groups are measured by those groups.", func(c *Context) []Finding {
		var out []Finding
		for _, o := range activeOutcomes(c.Doc) {
			var measuring []impact.Indicator
			for _, ind := range activeIndicators(c.Doc) {
				if ind.v.Measures == o.v.ID {
					measuring = append(measuring, ind.v)
				}
			}
			if len(measuring) == 0 {
				continue // IMP-MEA-001 reports an unmeasured outcome
			}
			for _, g := range o.v.TargetGroupIDs {
				if !slices.ContainsFunc(measuring, func(i impact.Indicator) bool { return slices.Contains(i.TargetGroupIDs, g) }) {
					out = append(out, Finding{Path: at("outcomes", o.i), Element: o.v.ID, Message: fmt.Sprintf("this outcome is for target group %q but no indicator breaks its result down by that group", g)})
				}
			}
		}
		return out
	})

	register("IMP-VAL-001", validate.SeverityInfo, "The main steps of the impact model have been checked with the people it concerns.", func(c *Context) []Finding {
		if len(c.Doc.Consultations) == 0 && len(c.Doc.Outcomes) < 3 {
			return nil // too small a draft to need it
		}
		have := map[impact.ConsultationStep]bool{}
		for _, k := range c.Doc.Consultations {
			have[k.Step] = true
		}
		var out []Finding
		for _, step := range []impact.ConsultationStep{impact.StepFocus, impact.StepChangeAnalysis, impact.StepAssumptionsRisks, impact.StepActors} {
			if !have[step] && !have[impact.StepWhole] {
				out = append(out, Finding{Path: "/consultations", Message: fmt.Sprintf("no consultation records that the %s step was checked with stakeholders", step)})
			}
		}
		return out
	})

	// ---- provenance ---------------------------------------------------------
	rule := funcRule{id: "IMP-PRV-001", severity: validate.SeverityWarning, from: ProfileFull,
		desc: "Every element says where it comes from.",
		check: func(c *Context) []Finding {
			var out []Finding
			for _, e := range provenanceElements(c) {
				if e.prov == nil {
					out = append(out, Finding{Path: at(e.array, e.index), Element: e.id, Message: "no provenance: say whether this is stated, inferred, a hypothesis, or confirmed"})
				}
			}
			return out
		}}
	Register(rule)

	register("IMP-PRV-002", validate.SeverityWarning, "Every stated element cites a source that can be followed.", func(c *Context) []Finding {
		var out []Finding
		for _, e := range provenanceElements(c) {
			if e.prov == nil || e.prov.Basis != impact.BasisStated {
				continue
			}
			followable := false
			for _, id := range e.prov.EvidenceIDs {
				if ev, ok := c.Doc.EvidenceRecord(id); ok {
					followable = followable || ev.URL != "" || ev.Source != ""
				}
			}
			if !followable {
				out = append(out, Finding{Path: at(e.array, e.index) + "/provenance", Element: e.id, Message: "marked stated, but it cites no evidence with a URL or a source reference"})
			}
		}
		return out
	})

	register("IMP-PRV-003", validate.SeverityWarning, "An approved impact model has no hypotheses among what the organization answers for.", func(c *Context) []Finding {
		if c.Doc.Metadata.Status != "approved" {
			return nil
		}
		var out []Finding
		for _, e := range provenanceElements(c) {
			if e.prov != nil && e.prov.Basis == impact.BasisHypothesis && e.accountable {
				out = append(out, Finding{Path: at(e.array, e.index) + "/provenance", Element: e.id, Message: "the impact model is approved but this element is still a hypothesis; confirm it or reject it"})
			}
		}
		return out
	})

	register("IMP-PRV-004", validate.SeverityInfo, "Selected indicators say whether their data is available.", func(c *Context) []Finding {
		var out []Finding
		for _, ind := range activeIndicators(c.Doc) {
			if isSelected(ind.v) && ind.v.DataAvailability == "" {
				out = append(out, Finding{Path: at("indicators", ind.i) + "/dataAvailability", Element: ind.v.ID, Message: "it is not stated whether the data for this selected indicator exists"})
			}
		}
		return out
	})
}

func isOutcomeID(d *impact.Model, id string) bool {
	_, ok := d.Outcome(id)
	return ok
}

// addressedCauses returns the causes some active outcome addresses.
func addressedCauses(d *impact.Model) map[string]bool {
	out := map[string]bool{}
	for _, o := range activeOutcomes(d) {
		for _, id := range o.v.AddressesCauses {
			out[id] = true
		}
	}
	return out
}
