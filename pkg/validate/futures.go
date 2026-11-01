package validate

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
	"github.com/ProductBuildersHQ/impact-model/pkg/ref"
)

// futureRef is the shape of a competing-future reference: ufm:<slug>.
var futureRef = regexp.MustCompile(`^ufm:[a-z][a-z0-9]*(-[a-z0-9]+)*$`)

// Futures checks impact models that describe competing futures of one market. They
// should share one actor registry, so the same actor means the same thing in
// each; each should name its future and its perspective; no two should speak
// for the same perspective in the same future; and any reference from one
// Impact model to an element of another, in the set, must resolve. The impact models are
// each valid on their own (see Document); this checks them against one
// another. Issues carry Source "future[<id>]".
func Futures(docs ...*impact.Model) []Issue {
	c := newChecker(&impact.Model{}, nil)
	type key struct{ future, actor string }
	spoken := map[key]string{}
	byID := map[string]*impact.Model{}
	for _, d := range docs {
		byID[d.ID] = d
	}
	registries := map[string][]string{} // registry ID -> impact models naming it
	var market, marketOwner string

	for _, d := range docs {
		c.source = "future[" + d.ID + "]"
		if d.ActorRegistry != "" {
			registries[d.ActorRegistry] = append(registries[d.ActorRegistry], d.ID)
		} else {
			registries[""] = append(registries[""], d.ID)
		}
		p := d.Perspective
		switch {
		case p == nil:
			c.add("VAL-FUT-001", SeverityWarning, "/perspective", "", "Impact model %q names no perspective, so it cannot be placed in a future", d.ID)
		case p.FutureRef == "":
			c.add("VAL-FUT-001", SeverityWarning, "/perspective/futureRef", "", "Impact model %q names no future", d.ID)
		case !futureRef.MatchString(p.FutureRef):
			c.errorf("VAL-FUT-002", "/perspective/futureRef", "", "future %q is not of the form ufm:<slug>", p.FutureRef)
		default:
			k := key{p.FutureRef, p.ActorID}
			if other, dup := spoken[k]; dup {
				c.errorf("VAL-FUT-004", "/perspective", "", "Impact models %q and %q both speak for %q in %s", other, d.ID, p.ActorID, p.FutureRef)
			}
			spoken[k] = d.ID
		}
		if m := d.Context.Market; m != nil {
			if market != "" && m.Ref != market {
				c.add("VAL-FUT-007", SeverityWarning, "/context/market/ref", "", "Impact model %q is about market %q, but %q is about %q; competing futures describe one market", d.ID, m.Ref, marketOwner, market)
			} else if market == "" {
				market, marketOwner = m.Ref, d.ID
			}
		}
		crossRefs(c, d, byID)
	}

	c.source = "futures"
	if named := registryNames(registries); len(named) > 1 {
		c.errorf("VAL-FUT-003", "/actorRegistry", "", "Impact models name different actor registries (%s); competing futures share one", strings.Join(named, ", "))
	} else if len(named) == 1 && len(registries[""]) > 0 {
		c.add("VAL-FUT-003", SeverityWarning, "/actorRegistry", "", "Impact models %s name no actor registry, but the others share %q", strings.Join(registries[""], ", "), named[0])
	}
	return c.issues
}

func registryNames(m map[string][]string) []string {
	var out []string
	for name := range m {
		if name != "" {
			out = append(out, name)
		}
	}
	slices.Sort(out)
	return out
}

// crossRefs checks references from d into other impact models of the set: an
// alignment or the external references of an element that point at
// model:<doc>#<element>.
func crossRefs(c *checker, d *impact.Model, set map[string]*impact.Model) {
	check := func(path, element string, ext impact.ExternalRef) {
		if !strings.HasPrefix(ext.Ref, ref.Prefix) {
			return
		}
		r, err := ref.Parse(ext.Ref)
		if err != nil {
			c.errorf("VAL-FUT-005", path, element, "%q is not a valid reference: %v", ext.Ref, err)
			return
		}
		target, inSet := set[r.Doc]
		switch {
		case !inSet:
			// Points at an impact model that was not supplied; nothing to check.
		case !r.IsElement():
		default:
			if _, ok := target.Element(r.Element); !ok {
				c.errorf("VAL-FUT-006", path, element, "%q: Impact model %q has no element %q", ext.Ref, r.Doc, r.Element)
			}
		}
	}
	for i, a := range d.Alignments {
		check(fmt.Sprintf("/alignments/%d/target", i), a.ElementID, a.Target)
	}
	for i, ind := range d.Indicators {
		for j, e := range ind.ExternalRefs {
			check(fmt.Sprintf("/indicators/%d/externalRefs/%d", i, j), ind.ID, e)
		}
	}
}
