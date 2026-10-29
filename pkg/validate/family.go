package validate

import (
	"fmt"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/ref"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

// Family checks theories that elaborate a parent theory. Every child must
// name the parent, every outcome reference into the parent must resolve to
// an outcome there, and a parent pathway that names a child must name one
// that was supplied. Issues carry Source "child[<id>]" or "parent".
func Family(parent *toc.Document, children ...*toc.Document) []Issue {
	c := newChecker(parent, nil)
	supplied := map[string]bool{}
	for _, ch := range children {
		supplied[ch.ID] = true
	}

	c.source = "parent"
	for i, pw := range parent.Pathways {
		if pw.ChildToc == "" {
			continue
		}
		r, err := ref.Parse(pw.ChildToc)
		if err != nil {
			continue // reported by Document
		}
		if !supplied[r.Doc] {
			c.add("VAL-FAM-005", SeverityWarning, fmt.Sprintf("/pathways/%d/childToc", i), pw.ID,
				"pathway names child theory %q, which was not supplied", r.Doc)
		}
	}

	named := map[string]bool{}
	for _, pw := range parent.Pathways {
		if r, err := ref.Parse(pw.ChildToc); err == nil {
			named[r.Doc] = true
		}
	}
	seen := map[string]bool{}
	for _, ch := range children {
		c.source = "child[" + ch.ID + "]"
		if seen[ch.ID] {
			c.errorf("VAL-FAM-006", "/id", "", "child theory %q supplied more than once", ch.ID)
			continue
		}
		seen[ch.ID] = true
		c.child(parent, ch, named[ch.ID])
	}
	return c.issues
}

func (c *checker) child(parent, ch *toc.Document, named bool) {
	pr, err := ref.Parse(ch.Parent)
	switch {
	case ch.Parent == "":
		c.errorf("VAL-FAM-001", "/parent", "", "theory %q does not name a parent", ch.ID)
	case err != nil:
		c.errorf("VAL-FAM-001", "/parent", "", "parent %q is not a valid reference", ch.Parent)
	case pr.Doc != parent.ID:
		c.errorf("VAL-FAM-001", "/parent", "", "theory %q names parent %q, not %q", ch.ID, ch.Parent, ref.MustNew(parent.ID, ""))
	}
	if !named {
		c.add("VAL-FAM-004", SeverityWarning, "/parent", "", "no pathway in %q names %q as its child theory", parent.ID, ch.ID)
	}
	for i, o := range ch.Outcomes {
		for j, v := range o.ContributesTo {
			r, err := ref.Parse(v)
			if err != nil || !r.IsElement() || r.Doc != parent.ID {
				continue // reported by Document, or points elsewhere
			}
			p := fmt.Sprintf("/outcomes/%d/contributesTo/%d", i, j)
			target, ok := parent.Outcome(r.Element)
			switch {
			case ok && target != nil:
			default:
				if _, exists := parent.Element(r.Element); exists {
					c.errorf("VAL-FAM-003", p, o.ID, "%q is not an outcome in %q", v, parent.ID)
				} else {
					c.errorf("VAL-FAM-002", p, o.ID, "%q does not exist in %q", v, parent.ID)
				}
			}
		}
	}
}
