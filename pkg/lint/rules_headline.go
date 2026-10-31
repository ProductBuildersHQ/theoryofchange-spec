package lint

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/validate"
)

const (
	// maxHeadline is the longest headline list that is still a headline list.
	// Beyond it the audience is being handed a dashboard.
	maxHeadline = 7
	// curateAbove is the number of indicators reported to an audience above
	// which its headline list should be chosen and not left to file order.
	curateAbove = 3
)

func init() {
	// ---- audiences and their headline lists -------------------------------
	register("TOC-AUD-002", validate.SeverityWarning, "An audience that outcomes are reported to also has indicators reported to it.", func(c *Context) []Finding {
		var out []Finding
		for i, a := range c.Doc.Audiences {
			var outcomes, numbers int
			for _, o := range activeOutcomes(c.Doc) {
				if slices.Contains(o.v.AudienceIDs, a.ID) {
					outcomes++
				}
			}
			for _, ind := range activeIndicators(c.Doc) {
				if slices.Contains(ind.v.AudienceIDs, a.ID) {
					numbers++
				}
			}
			if outcomes > 0 && numbers == 0 {
				out = append(out, Finding{Path: at("audiences", i), Element: a.ID,
					Message: fmt.Sprintf("%d outcome(s) are reported to this audience but no indicator is, so its story has no numbers", outcomes)})
			}
		}
		return out
	})

	register("TOC-HLN-001", validate.SeverityWarning, "A headline list is short.", func(c *Context) []Finding {
		var out []Finding
		for i, a := range c.Doc.Audiences {
			if len(a.Headline) > maxHeadline {
				out = append(out, Finding{Path: at("audiences", i) + "/headline", Element: a.ID,
					Message: fmt.Sprintf("the headline list has %d entries; more than %d is a dashboard, not a headline", len(a.Headline), maxHeadline)})
			}
		}
		return out
	})

	register("TOC-HLN-002", validate.SeverityWarning, "Every headline indicator is reported to its audience.", func(c *Context) []Finding {
		var out []Finding
		for i, a := range c.Doc.Audiences {
			for j, h := range a.Headline {
				ind, ok := c.Doc.Indicator(h.IndicatorID)
				if ok && !slices.Contains(ind.AudienceIDs, a.ID) {
					out = append(out, Finding{Path: at("audiences", i) + "/headline/" + strconv.Itoa(j), Element: a.ID,
						Message: fmt.Sprintf("indicator %q is in the headline list but does not name this audience in audienceIds", h.IndicatorID)})
				}
			}
		}
		return out
	})

	register("TOC-HLN-003", validate.SeverityInfo, "Every headline indicator can be produced.", func(c *Context) []Finding {
		var out []Finding
		for i, a := range c.Doc.Audiences {
			for j, h := range a.Headline {
				ind, ok := c.Doc.Indicator(h.IndicatorID)
				if ok && (ind.DataAvailability == "" || ind.DataAvailability == toc.DataUnknown) {
					out = append(out, Finding{Path: at("audiences", i) + "/headline/" + strconv.Itoa(j), Element: a.ID,
						Message: fmt.Sprintf("it is not known whether the data for %q exists, and an audience cannot be told a number that cannot be produced", h.IndicatorID)})
				}
			}
		}
		return out
	})

	register("TOC-HLN-004", validate.SeverityInfo, "An audience with several indicators has a chosen headline list.", func(c *Context) []Finding {
		var out []Finding
		for i, a := range c.Doc.Audiences {
			var reported int
			for _, ind := range activeIndicators(c.Doc) {
				if slices.Contains(ind.v.AudienceIDs, a.ID) {
					reported++
				}
			}
			if reported > curateAbove && len(a.Headline) == 0 {
				out = append(out, Finding{Path: at("audiences", i), Element: a.ID,
					Message: fmt.Sprintf("%d indicators are reported to this audience and none is chosen as the headline, so file order decides what leads", reported)})
			}
		}
		return out
	})

	// ---- the North Star and what keeps it honest --------------------------
	northStars := func(c *Context, fn func(p indexed[toc.Pathway], ns toc.Indicator) []Finding) []Finding {
		var out []Finding
		for _, p := range activePathways(c.Doc) {
			if p.v.Focus == nil || p.v.Focus.NorthStarID == "" {
				continue
			}
			ns, ok := c.Doc.Indicator(p.v.Focus.NorthStarID)
			if !ok {
				continue // validation reports the unknown reference
			}
			out = append(out, fn(p, *ns)...)
		}
		return out
	}
	nsPath := func(p indexed[toc.Pathway]) string { return at("pathways", p.i) + "/focus/northStarId" }

	register("TOC-NSM-001", validate.SeverityWarning, "A North Star measures an outcome, never an output or an activity.", func(c *Context) []Finding {
		return northStars(c, func(p indexed[toc.Pathway], ns toc.Indicator) []Finding {
			var what string
			switch c.Doc.IndicatorTier(ns) {
			case toc.TierOutcome, toc.TierImpact:
				return nil
			case toc.TierOutput:
				what = "an output, which counts what was delivered and not the value to the people served"
			case toc.TierActivity:
				what = "an activity, which counts what was done and not what changed"
			default:
				what = "something that is not an outcome"
			}
			return []Finding{{Path: nsPath(p), Element: p.v.ID,
				Message: fmt.Sprintf("the North Star %q measures %s; choose an indicator of an outcome", ns.ID, what)}}
		})
	})

	register("TOC-NSM-002", validate.SeverityWarning, "A North Star sits at or below the ceiling of accountability.", func(c *Context) []Finding {
		return northStars(c, func(p indexed[toc.Pathway], ns toc.Indicator) []Finding {
			switch c.Doc.IndicatorTier(ns) {
			case toc.TierOutcome, toc.TierImpact:
				if !c.Graph.BelowCeiling(ns.Measures) {
					return []Finding{{Path: nsPath(p), Element: p.v.ID,
						Message: fmt.Sprintf("the North Star %q measures %q, which is above the ceiling of accountability; steer by something the pathway answers for", ns.ID, ns.Measures)}}
				}
			}
			return nil
		})
	})

	register("TOC-NSM-003", validate.SeverityWarning, "A North Star has a baseline and a target.", func(c *Context) []Finding {
		return northStars(c, func(p indexed[toc.Pathway], ns toc.Indicator) []Finding {
			var missing []string
			if ns.Baseline == nil {
				missing = append(missing, "baseline")
			}
			if len(ns.Targets) == 0 {
				missing = append(missing, "target")
			}
			if len(missing) == 0 {
				return nil
			}
			return []Finding{{Path: nsPath(p), Element: p.v.ID,
				Message: fmt.Sprintf("the North Star %q has no %s, so progress toward it cannot be told", ns.ID, joinWords(missing))}}
		})
	})

	register("TOC-NSM-004", validate.SeverityWarning, "A pathway with a North Star has a guardrail.", func(c *Context) []Finding {
		return northStars(c, func(p indexed[toc.Pathway], ns toc.Indicator) []Finding {
			if len(p.v.Focus.GuardrailIDs) > 0 {
				return nil
			}
			return []Finding{{Path: at("pathways", p.i) + "/focus/guardrailIds", Element: p.v.ID,
				Message: fmt.Sprintf("the North Star %q has no guardrail, so nothing watches what pursuing it could harm", ns.ID)}}
		})
	})

	register("TOC-NSM-005", validate.SeverityWarning, "A North Star is an indicator of its own pathway.", func(c *Context) []Finding {
		return northStars(c, func(p indexed[toc.Pathway], ns toc.Indicator) []Finding {
			if c.Doc.IndicatorServes(ns, p.v.ID) {
				return nil
			}
			return []Finding{{Path: nsPath(p), Element: p.v.ID,
				Message: fmt.Sprintf("the North Star %q does not serve this pathway: it measures %q, which the pathway does not list", ns.ID, ns.Measures)}}
		})
	})

	register("TOC-NSM-006", validate.SeverityWarning, "A guardrail is not the North Star.", func(c *Context) []Finding {
		return northStars(c, func(p indexed[toc.Pathway], ns toc.Indicator) []Finding {
			if !slices.Contains(p.v.Focus.GuardrailIDs, ns.ID) {
				return nil
			}
			return []Finding{{Path: at("pathways", p.i) + "/focus/guardrailIds", Element: p.v.ID,
				Message: fmt.Sprintf("%q is both the North Star and a guardrail; a guardrail watches something else", ns.ID)}}
		})
	})

	register("TOC-NSM-007", validate.SeverityInfo, "A program or product has a North Star.", func(c *Context) []Finding {
		var out []Finding
		for _, p := range activePathways(c.Doc) {
			if p.v.Kind != toc.PathwayProgram && p.v.Kind != toc.PathwayProduct {
				continue
			}
			if p.v.Focus == nil || p.v.Focus.NorthStarID == "" {
				out = append(out, Finding{Path: at("pathways", p.i), Element: p.v.ID,
					Message: "this " + string(p.v.Kind) + " has no North Star, so there is no one number that says whether it delivers its value"})
			}
		}
		return out
	})

	// ---- framing ----------------------------------------------------------
	register("TOC-CTX-001", validate.SeverityInfo, "The theory states its mission or its vision.", func(c *Context) []Finding {
		if c.Doc.Context.Mission != "" || c.Doc.Context.Vision != "" {
			return nil
		}
		return []Finding{{Path: "/context", Message: "neither a mission nor a vision is stated, so the theory does not say why the organization exists or what it works toward"}}
	})
}

// joinWords joins the words as "a", "a or b", or "a, b, or c".
func joinWords(words []string) string {
	switch len(words) {
	case 0:
		return ""
	case 1:
		return words[0]
	case 2:
		return words[0] + " or " + words[1]
	}
	return strings.Join(words[:len(words)-1], ", ") + ", or " + words[len(words)-1]
}
