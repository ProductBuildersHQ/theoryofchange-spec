package lint

import (
	"fmt"
	"slices"

	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
	"github.com/ProductBuildersHQ/impact-model/pkg/validate"
)

// activeInterventions returns the interventions that were not rejected, with
// their index in the document.
func activeInterventions(d *impact.Model) []indexed[impact.Intervention] {
	var out []indexed[impact.Intervention]
	for i, in := range d.Interventions {
		if !in.Rejected() {
			out = append(out, indexed[impact.Intervention]{i, in})
		}
	}
	return out
}

// The rules below check the chain from what is done to what changes: that each
// activity says what it produces, that it leads somewhere, that the links and
// the interventions that carry them agree, and that what was done is measured.
func init() {
	register("IMP-INT-003", validate.SeverityWarning, "Every intervention says what it produces.", func(c *Context) []Finding {
		var out []Finding
		for _, in := range activeInterventions(c.Doc) {
			if len(in.v.Outputs) > 0 {
				continue
			}
			producesOutput := false
			for _, t := range in.v.Targets {
				if lv, ok := outcomeLevel(c.Doc, t); ok && lv == impact.LevelOutput {
					producesOutput = true
				}
			}
			if !producesOutput {
				out = append(out, Finding{Path: at("interventions", in.i), Element: in.v.ID,
					Message: "this intervention lists no outputs and targets no output-level outcome, so it does not say what it produces"})
			}
		}
		return out
	})

	register("IMP-INT-004", validate.SeverityWarning, "Every intervention leads, through the outcomes it targets, to an impact outcome.", func(c *Context) []Finding {
		var out []Finding
		for _, in := range activeInterventions(c.Doc) {
			leads := false
			for _, t := range in.v.Targets {
				o, ok := c.Doc.Outcome(t)
				if !ok || o.Rejected() {
					continue
				}
				leads = leads || o.Level == impact.LevelImpact || c.Graph.HasPathToImpact(t)
			}
			if !leads {
				out = append(out, Finding{Path: at("interventions", in.i) + "/targets", Element: in.v.ID,
					Message: "no outcome this intervention targets leads to an impact outcome, so the work does not connect to what the impact model is for"})
			}
		}
		return out
	})

	register("IMP-INT-005", validate.SeverityWarning, "A link that names an intervention is carried out by an intervention that targets one end of the link.", func(c *Context) []Finding {
		var out []Finding
		for _, l := range activeLinks(c.Doc) {
			for _, id := range l.v.InterventionIDs {
				in, ok := c.Doc.Intervention(id)
				if !ok || in.Rejected() {
					continue
				}
				if !slices.Contains(in.Targets, l.v.From) && !slices.Contains(in.Targets, l.v.To) {
					out = append(out, Finding{Path: at("links", l.i) + "/interventionIds", Element: l.v.ID,
						Message: fmt.Sprintf("it names intervention %q, which targets neither %q nor %q", id, l.v.From, l.v.To)})
				}
			}
		}
		return out
	})

	register("IMP-INT-006", validate.SeverityInfo, "Every intervention has an indicator of what was done.", func(c *Context) []Finding {
		measured := map[string]bool{}
		for _, ind := range activeIndicators(c.Doc) {
			measured[ind.v.Measures] = true
		}
		var out []Finding
		for _, in := range activeInterventions(c.Doc) {
			if !measured[in.v.ID] {
				out = append(out, Finding{Path: at("interventions", in.i), Element: in.v.ID,
					Message: "nothing measures what was done or the resources it used, so cost per outcome cannot be worked out"})
			}
		}
		return out
	})

	register("IMP-INT-007", validate.SeverityInfo, "Every intervention states its resources.", func(c *Context) []Finding {
		var out []Finding
		for _, in := range activeInterventions(c.Doc) {
			if in.v.Resources == "" {
				out = append(out, Finding{Path: at("interventions", in.i), Element: in.v.ID,
					Message: "no resources are stated, so the logic model shows no inputs for it"})
			}
		}
		return out
	})
}
