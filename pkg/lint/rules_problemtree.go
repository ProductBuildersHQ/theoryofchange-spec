package lint

import (
	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
	"github.com/ProductBuildersHQ/impact-model/pkg/validate"
)

// The rules below check the problem tree's branches: that each effect of the
// core problem is relieved by some outcome or set aside with a reason, and
// that the outcomes relieving effects reach far enough up the chain.
func init() {
	register("IMP-EFF-001", validate.SeverityWarning, "Each effect is relieved by an outcome or set aside with a reason.", func(c *Context) []Finding {
		relieved := relievers(c.Doc)
		var out []Finding
		for i, e := range c.Doc.Effects {
			if !e.Rejected() && len(relieved[e.ID]) == 0 && e.OutOfScope == nil {
				out = append(out, Finding{Path: at("effects", i), Element: e.ID,
					Message: "no outcome relieves this effect; name it in an outcome's addressesEffects, or set it aside with a reason"})
			}
		}
		return out
	})

	register("IMP-EFF-002", validate.SeverityInfo, "Effects are relieved by long-term or impact outcomes, not only by earlier ones.", func(c *Context) []Finding {
		relieved := relievers(c.Doc)
		if len(relieved) == 0 {
			return nil
		}
		for _, levels := range relieved {
			for _, l := range levels {
				if l == impact.LevelLongTerm || l == impact.LevelImpact {
					return nil
				}
			}
		}
		return []Finding{{Path: "/effects", Message: "effects are relieved only by outcomes below the long-term level; effects are what the program ultimately reduces, so a long-term or impact outcome should name them"}}
	})
}

// relievers maps each effect to the levels of the active outcomes that
// relieve it.
func relievers(d *impact.Model) map[string][]impact.Level {
	out := map[string][]impact.Level{}
	for _, o := range activeOutcomes(d) {
		for _, id := range o.v.AddressesEffects {
			out[id] = append(out[id], o.v.Level)
		}
	}
	return out
}
