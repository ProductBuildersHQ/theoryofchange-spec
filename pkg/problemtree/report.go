package problemtree

import (
	"fmt"
	"strings"

	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
)

// Format is how the diagram in a report is written.
type Format string

// The diagram formats.
const (
	FormatMermaid Format = "mermaid"
	FormatD2      Format = "d2"
)

// Report renders the problem tree as Markdown: the diagram in format f, then a
// table of the causes and one of the effects, each with what it leads to and
// the outcomes that address it, or why it is set aside.
func Report(doc *impact.Model, f Format) (string, error) {
	t := Build(doc)
	var b strings.Builder
	b.WriteString("## Problem tree\n\n")
	if len(t.Rows) == 1 {
		fmt.Fprintf(&b, "No causes or effects have been recorded for the core problem: %s\n\n", oneLine(doc.Context.ProblemStatement))
		return b.String(), nil
	}
	var diagram string
	switch f {
	case FormatMermaid:
		var err error
		if diagram, err = t.Mermaid(); err != nil {
			return "", err
		}
	case FormatD2:
		diagram = t.D2()
	default:
		return "", fmt.Errorf("problemtree: unknown diagram format %q", f)
	}
	fmt.Fprintf(&b, "Causes (roots) below lead up to the core problem, which leads up to its effects (branches).\n\n```%s\n%s```\n\n", f, diagram)

	addressedBy := map[string][]string{}
	for _, o := range doc.Outcomes {
		if o.Rejected() {
			continue
		}
		name := fmt.Sprintf("%s (`%s`)", cell(o.Title), o.ID)
		for _, id := range o.AddressesCauses {
			addressedBy[id] = append(addressedBy[id], name)
		}
		for _, id := range o.AddressesEffects {
			addressedBy[id] = append(addressedBy[id], name)
		}
	}
	statement := map[string]string{ProblemID: "the core problem"}
	for _, c := range doc.Causes {
		statement[c.ID] = c.Statement
	}
	for _, e := range doc.Effects {
		statement[e.ID] = e.Statement
	}
	targets := func(ids []string) string {
		if len(ids) == 0 {
			return "the core problem"
		}
		var out []string
		for _, id := range ids {
			out = append(out, fmt.Sprintf("%s (`%s`)", cell(statement[id]), id))
		}
		return strings.Join(out, "; ")
	}
	answer := func(id string, aside *impact.OutOfScope) string {
		switch {
		case aside != nil:
			return "set aside: " + cell(aside.Reason)
		case len(addressedBy[id]) == 0:
			return "—"
		}
		return strings.Join(addressedBy[id], "; ")
	}

	var causes, effects strings.Builder
	for _, c := range doc.Causes {
		if !c.Rejected() {
			fmt.Fprintf(&causes, "| %s (`%s`) | %s | %s | %s |\n", cell(c.Statement), c.ID, c.Level, targets(c.Parents), answer(c.ID, c.OutOfScope))
		}
	}
	for _, e := range doc.Effects {
		if !e.Rejected() {
			fmt.Fprintf(&effects, "| %s (`%s`) | %s | %s |\n", cell(e.Statement), e.ID, targets(e.Parents), answer(e.ID, e.OutOfScope))
		}
	}
	if causes.Len() > 0 {
		b.WriteString("### Causes\n\n| Cause | Level | Leads to | Addressed by |\n|-------|-------|----------|--------------|\n")
		b.WriteString(causes.String() + "\n")
	}
	if effects.Len() > 0 {
		b.WriteString("### Effects\n\n| Effect | Follows from | Relieved by |\n|--------|--------------|-------------|\n")
		b.WriteString(effects.String() + "\n")
	}
	return b.String(), nil
}

// cell makes text safe for a Markdown table cell.
func cell(s string) string { return strings.ReplaceAll(oneLine(s), "|", `\|`) }
