package lint

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ProductBuildersHQ/impact-model/pkg/burden"
	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
	"github.com/ProductBuildersHQ/impact-model/pkg/validate"
)

type indexedIndicator = indexed[impact.Indicator]

// activeIndicators returns the indicators that were not rejected.
func activeIndicators(d *impact.Model) []indexedIndicator {
	var out []indexedIndicator
	for i, x := range d.Indicators {
		if !x.Provenance.IsRejected() {
			out = append(out, indexedIndicator{i, x})
		}
	}
	return out
}

func isSelected(i impact.Indicator) bool {
	return i.Selection != nil && i.Selection.Status == impact.SelectionSelected
}

func lastTarget(i impact.Indicator) string {
	last := ""
	for _, t := range i.Targets {
		if t.Date > last {
			last = t.Date
		}
	}
	return last
}

func init() {
	// ---- measurement -------------------------------------------------------
	register("IMP-MEA-001", validate.SeverityError, "Every outcome within the declared ceiling of accountability is measured by an indicator, or stood in for by one.", func(c *Context) []Finding {
		if c.Doc.Ceiling == nil {
			return nil // without a ceiling nothing is known to be owned; IMP-ACC-001 asks for one
		}
		measured := map[string]bool{}
		for _, ind := range activeIndicators(c.Doc) {
			measured[ind.v.Measures] = true
			for _, id := range ind.v.ProxyFor {
				measured[id] = true
			}
		}
		var out []Finding
		for _, o := range activeOutcomes(c.Doc) {
			if c.Graph.BelowCeiling(o.v.ID) && !measured[o.v.ID] {
				out = append(out, Finding{Path: at("outcomes", o.i), Element: o.v.ID, Message: "this outcome is within what the organization answers for but nothing measures it; add an indicator or name it in another indicator's proxyFor"})
			}
		}
		return out
	})

	register("IMP-MEA-002", validate.SeverityWarning, "Every indicator has a baseline, a target, a data source, and a frequency.", func(c *Context) []Finding {
		var out []Finding
		for _, ind := range activeIndicators(c.Doc) {
			var missing []string
			if ind.v.Baseline == nil {
				missing = append(missing, "baseline")
			}
			if len(ind.v.Targets) == 0 {
				missing = append(missing, "target")
			}
			if ind.v.DataSource == "" {
				missing = append(missing, "data source")
			}
			if ind.v.Frequency == "" {
				missing = append(missing, "frequency")
			}
			if len(missing) > 0 {
				out = append(out, Finding{Path: at("indicators", ind.i), Element: ind.v.ID, Message: "missing " + strings.Join(missing, ", ")})
			}
		}
		return out
	})

	register("IMP-MEA-003", validate.SeverityWarning, "Quantitative indicators have a unit and a direction.", func(c *Context) []Finding {
		var out []Finding
		for _, ind := range activeIndicators(c.Doc) {
			if t := ind.v.MeasurementType; t != impact.MeasurementQuantitative && t != "" {
				continue
			}
			var missing []string
			if ind.v.Unit == "" {
				missing = append(missing, "unit")
			}
			if ind.v.Direction == "" {
				missing = append(missing, "direction")
			}
			if len(missing) > 0 {
				out = append(out, Finding{Path: at("indicators", ind.i), Element: ind.v.ID, Message: "a quantitative indicator needs a " + strings.Join(missing, " and a ")})
			}
		}
		return out
	})

	register("IMP-MEA-004", validate.SeverityWarning, "Target dates fall within the time horizon of the impact model.", func(c *Context) []Finding {
		h := c.Doc.Context.TimeHorizon
		if h == nil {
			return nil
		}
		var out []Finding
		for _, ind := range activeIndicators(c.Doc) {
			for j, t := range ind.v.Targets {
				if t.Date != "" && (t.Date < h.Start || t.Date > h.End) {
					out = append(out, Finding{Path: fmt.Sprintf("%s/targets/%d/date", at("indicators", ind.i), j), Element: ind.v.ID,
						Message: fmt.Sprintf("target date %s is outside the time horizon %s to %s", t.Date, h.Start, h.End)})
				}
			}
		}
		return out
	})

	register("IMP-MEA-005", validate.SeverityWarning, "An outcome is not due before the outcomes that lead to it.", func(c *Context) []Finding {
		due := map[string]string{} // outcome ID to its latest target date
		for _, ind := range activeIndicators(c.Doc) {
			if d := lastTarget(ind.v); d > due[ind.v.Measures] {
				due[ind.v.Measures] = d
			}
		}
		var out []Finding
		for _, o := range activeOutcomes(c.Doc) {
			mine, ok := due[o.v.ID]
			if !ok {
				continue
			}
			for _, src := range incomingSources(c.Doc, o.v.ID) {
				if theirs, ok := due[src]; ok && mine < theirs {
					out = append(out, Finding{Path: at("outcomes", o.i), Element: o.v.ID,
						Message: fmt.Sprintf("this outcome is due %s, before %s, which leads to it, is due (%s)", mine, src, theirs)})
				}
			}
		}
		return out
	})

	// ---- accountability and claims ----------------------------------------
	register("IMP-ACC-001", validate.SeverityWarning, "A ceiling of accountability is declared.", func(c *Context) []Finding {
		if c.Doc.Ceiling == nil {
			return []Finding{{Path: "/ceilingOfAccountability", Message: "no ceiling of accountability; say which outcomes the organization answers for and which it only contributes to"}}
		}
		return nil
	})

	register("IMP-ACC-002", validate.SeverityWarning, "Outcomes above the ceiling of accountability are not owned.", func(c *Context) []Finding {
		if c.Doc.Ceiling == nil {
			return nil
		}
		var out []Finding
		for _, o := range activeOutcomes(c.Doc) {
			if !c.Graph.BelowCeiling(o.v.ID) && o.v.EffectiveResponsibility() == impact.ResponsibilityOwn && o.v.Responsibility != "" {
				out = append(out, Finding{Path: at("outcomes", o.i) + "/responsibility", Element: o.v.ID,
					Message: "this outcome is above the ceiling of accountability but is marked as owned; mark it shared or partner, or move the ceiling"})
			}
		}
		return out
	})

	register("IMP-CLM-001", validate.SeverityWarning, "Indicators above the ceiling of accountability claim contribution, not attribution.", func(c *Context) []Finding {
		var out []Finding
		for _, ind := range activeIndicators(c.Doc) {
			if _, isOutcome := c.Doc.Outcome(ind.v.Measures); isOutcome && !c.Graph.BelowCeiling(ind.v.Measures) && ind.v.EffectiveClaim() == impact.ClaimAttribution {
				out = append(out, Finding{Path: at("indicators", ind.i) + "/claim", Element: ind.v.ID,
					Message: "this indicator measures an outcome above the ceiling of accountability, so it can show contribution, not attribution"})
			}
		}
		return out
	})

	// ---- budget ------------------------------------------------------------
	budget := func(id string, sev validate.Severity, desc string, check func(*burden.Report, *impact.MeasurementBudget) []Finding) {
		register(id, sev, desc, func(c *Context) []Finding {
			if c.Doc.Budget == nil {
				return nil
			}
			rep, err := burden.Compute(c.Doc)
			if err != nil {
				return []Finding{{Path: "/indicators", Message: "collection cost cannot be totalled: " + err.Error()}}
			}
			return check(rep, c.Doc.Budget)
		})
	}
	budget("IMP-BUD-001", validate.SeverityWarning, "Selected indicators fit the staff time budget.", func(r *burden.Report, b *impact.MeasurementBudget) []Finding {
		if ch := r.Budget.StaffMinutesPerWeek; ch.Over {
			return []Finding{{Path: "/measurementBudget/staffMinutesPerWeek", Message: fmt.Sprintf("selected indicators take %.4g staff minutes a week; the budget is %.4g", ch.Actual, ch.Limit)}}
		}
		return nil
	})
	budget("IMP-BUD-002", validate.SeverityWarning, "Selected indicators fit the question budget per respondent.", func(r *burden.Report, b *impact.MeasurementBudget) []Finding {
		if ch := r.Budget.QuestionsPerYear; ch.Over {
			return []Finding{{Path: "/measurementBudget/maxQuestionsPerRespondentPerYear", Message: fmt.Sprintf("selected indicators ask each respondent %.4g questions a year; the budget is %.4g", ch.Actual, ch.Limit)}}
		}
		return nil
	})
	budget("IMP-BUD-003", validate.SeverityInfo, "The number of selected indicators is within the target range.", func(r *burden.Report, b *impact.MeasurementBudget) []Finding {
		if sm := r.Budget.SelectedMetrics; sm.Declared && !sm.Within {
			return []Finding{{Path: "/measurementBudget/targetSelectedMetrics", Message: fmt.Sprintf("%d indicators are selected; the target is %d to %d", sm.Actual, sm.Min, sm.Max)}}
		}
		return nil
	})

	register("IMP-BUD-004", validate.SeverityInfo, "Selected indicators say how their data is collected.", func(c *Context) []Finding {
		var out []Finding
		for _, ind := range activeIndicators(c.Doc) {
			if isSelected(ind.v) && (ind.v.Collection == nil || ind.v.Collection.Mode == "") {
				out = append(out, Finding{Path: at("indicators", ind.i) + "/collection", Element: ind.v.ID, Message: "this selected indicator does not say how its data is collected; prefer passive or derived data where it exists"})
			}
		}
		return out
	})

	// ---- assumptions -------------------------------------------------------
	critical := func(a impact.Assumption) bool {
		return a.Criticality == impact.CriticalityHigh || a.Criticality == impact.CriticalityCritical
	}
	asm := func(id string, sev validate.Severity, desc, msg string, bad func(*Context, impact.Assumption) bool) {
		register(id, sev, desc, func(c *Context) []Finding {
			var out []Finding
			for i, a := range c.Doc.Assumptions {
				if !a.Provenance.IsRejected() && bad(c, a) {
					out = append(out, Finding{Path: at("assumptions", i), Element: a.ID, Message: msg})
				}
			}
			return out
		})
	}
	asm("IMP-ASM-001", validate.SeverityWarning, "High and critical assumptions say how they will be tested.",
		"a high or critical assumption needs a description of how it will be tested",
		func(_ *Context, a impact.Assumption) bool { return critical(a) && a.Test == "" })
	asm("IMP-ASM-002", validate.SeverityWarning, "High and critical assumptions cite evidence.",
		"a high or critical assumption cites no evidence",
		func(_ *Context, a impact.Assumption) bool { return critical(a) && len(a.EvidenceIDs) == 0 })
	asm("IMP-ASM-004", validate.SeverityWarning, "Assumptions that need testing say how, and are watched or tied to a decision.",
		"this assumption needs testing but has no test, or no monitoring indicator or related decision",
		func(c *Context, a impact.Assumption) bool {
			if !a.NeedsTesting {
				return false
			}
			watched := len(a.IndicatorIDs) > 0
			for _, ind := range c.Doc.Indicators {
				watched = watched || ind.Measures == a.ID
			}
			for _, d := range c.Doc.Decisions {
				watched = watched || slices.Contains(d.RelatedIDs, a.ID)
			}
			return a.Test == "" || !watched
		})

	register("IMP-ASM-003", validate.SeverityInfo, "The impact model states at least one contextual or external assumption.", func(c *Context) []Finding {
		if len(c.Doc.Assumptions) == 0 {
			return nil // nothing is assumed yet; other rules say so
		}
		for _, a := range c.Doc.Assumptions {
			if !a.Provenance.IsRejected() && (a.Type == impact.AssumptionContextual || a.Type == impact.AssumptionExternal) {
				return nil
			}
		}
		return []Finding{{Path: "/assumptions", Message: "no assumption is about the context or outside forces; impact models that look only inward miss what can go wrong around them"}}
	})

	register("IMP-ASM-005", validate.SeverityInfo, "An impact model with shared or partner outcomes states a partnership assumption.", func(c *Context) []Finding {
		shared := false
		for _, o := range activeOutcomes(c.Doc) {
			r := o.v.EffectiveResponsibility()
			shared = shared || r == impact.ResponsibilityShared || r == impact.ResponsibilityPartner
		}
		if !shared {
			return nil
		}
		for _, a := range c.Doc.Assumptions {
			if !a.Provenance.IsRejected() && a.Type == impact.AssumptionPartnership {
				return nil
			}
		}
		return []Finding{{Path: "/assumptions", Message: "some outcomes depend on partners but no assumption says what the partners must do"}}
	})

	register("IMP-RSK-001", validate.SeverityInfo, "High and critical assumptions have a related risk with a mitigation.", func(c *Context) []Finding {
		var out []Finding
		for i, a := range c.Doc.Assumptions {
			if a.Provenance.IsRejected() || !critical(a) {
				continue
			}
			mitigated := false
			for _, r := range c.Doc.Risks {
				mitigated = mitigated || (slices.Contains(r.RelatedAssumptionIDs, a.ID) && r.Mitigation != "")
			}
			if !mitigated {
				out = append(out, Finding{Path: at("assumptions", i), Element: a.ID, Message: "no risk with a mitigation is tied to this high-stakes assumption"})
			}
		}
		return out
	})
}
