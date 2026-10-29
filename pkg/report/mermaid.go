package report

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

// MermaidOptions configures the pathway diagram.
type MermaidOptions struct {
	// HideInterventions leaves interventions out of the diagram.
	HideInterventions bool
	// ShowIDs adds each element's ID to its label.
	ShowIDs bool
}

// Mermaid renders the view as a Mermaid flowchart. Arrows point upward from
// preconditions to the outcomes they lead to, outcomes are grouped by level,
// and the ceiling of accountability appears as a boundary: outcomes above it
// are tracked but not owned. Precondition groups appear as small AND / OR
// nodes, feedback links as dashed arrows, and inferred or hypothesis
// elements are visibly marked. When the view carries status, nodes are
// colored by health. Output is deterministic.
func Mermaid(v *View, o MermaidOptions) string {
	r := &mermaidRenderer{v: v, o: o, ids: map[string]string{}}
	return r.render()
}

type mermaidRenderer struct {
	v   *View
	o   MermaidOptions
	ids map[string]string
	b   strings.Builder

	classes map[string][]string // class name -> node ids, in order
}

func (r *mermaidRenderer) render() string {
	r.classes = map[string][]string{}
	r.line(0, "flowchart BT")
	r.line(1, "classDef green fill:#d9f2d9,stroke:#2e7d32,color:#111")
	r.line(1, "classDef yellow fill:#fff4cc,stroke:#b8860b,color:#111")
	r.line(1, "classDef red fill:#f8d7d7,stroke:#c62828,color:#111")
	r.line(1, "classDef grey fill:#eeeeee,stroke:#757575,color:#111")
	r.line(1, "classDef inferred stroke-dasharray: 5 3")
	r.line(1, "classDef hypothesis stroke-dasharray: 2 3")
	r.line(1, "classDef intervention fill:#e8eefc,stroke:#3f51b5,color:#111")
	r.line(1, "classDef logic fill:#ffffff,stroke:#555555,color:#111")

	owned, above := r.splitByCeiling()
	if r.v.Document.Ceiling != "" && len(above) > 0 {
		r.line(1, `subgraph above["Above the ceiling of accountability: tracked, not owned"]`)
		r.levels(2, above)
		r.line(1, "end")
		r.line(1, `subgraph within["Within accountability: `+mermaidText(r.v.Document.Ceiling)+`"]`)
		r.levels(2, owned)
		r.line(1, "end")
	} else {
		r.levels(1, slices.Concat(above, owned))
	}

	if !r.o.HideInterventions && len(r.v.Interventions) > 0 {
		r.line(1, `subgraph interventions["Interventions"]`)
		for _, in := range r.v.Interventions {
			label := mermaidText(in.Title) + r.suffix(in.ID, in.Basis, "")
			r.line(2, fmt.Sprintf(`%s(["%s"])`, r.id(in.ID), label))
			r.classes["intervention"] = append(r.classes["intervention"], r.id(in.ID))
			r.markBasis(in.ID, in.Basis)
		}
		r.line(1, "end")
	}

	r.groups()
	r.edges()
	r.assignClasses()
	return r.b.String()
}

func (r *mermaidRenderer) line(indent int, s string) {
	r.b.WriteString(strings.Repeat("  ", indent))
	r.b.WriteString(s)
	r.b.WriteByte('\n')
}

// id returns a Mermaid-safe node ID for an element ID.
func (r *mermaidRenderer) id(elementID string) string {
	if id, ok := r.ids[elementID]; ok {
		return id
	}
	id := "n_" + strings.Map(func(c rune) rune {
		if c >= 'a' && c <= 'z' || c >= '0' && c <= '9' {
			return c
		}
		return '_'
	}, elementID)
	r.ids[elementID] = id
	return id
}

func (r *mermaidRenderer) splitByCeiling() (owned, above []OutcomeView) {
	for _, o := range r.v.Outcomes {
		if o.Owned {
			owned = append(owned, o)
		} else {
			above = append(above, o)
		}
	}
	return owned, above
}

// levels writes one subgraph per outcome level, highest first.
func (r *mermaidRenderer) levels(indent int, outcomes []OutcomeView) {
	for _, level := range toc.AllLevels() {
		var members []OutcomeView
		for _, o := range outcomes {
			if o.Level == level {
				members = append(members, o)
			}
		}
		if len(members) == 0 {
			continue
		}
		r.line(indent, fmt.Sprintf(`subgraph lv_%s["%s"]`, level, mermaidText(LevelLabel(level))+"s"))
		for _, o := range members {
			status := ""
			if o.Status != nil {
				status = o.Status.PathwayState
			}
			r.line(indent+1, fmt.Sprintf(`%s["%s"]`, r.id(o.ID), mermaidText(o.Title)+r.suffix(o.ID, o.Basis, status)))
			r.markBasis(o.ID, o.Basis)
			if o.Status != nil {
				r.classes[healthClass(o.Status.Health)] = append(r.classes[healthClass(o.Status.Health)], r.id(o.ID))
			}
		}
		r.line(indent, "end")
	}
}

// suffix builds the label lines after the title: the ID (optionally), the
// basis marker, and the status.
func (r *mermaidRenderer) suffix(id string, b BasisView, status string) string {
	var parts []string
	if r.o.ShowIDs {
		parts = append(parts, mermaidText(id))
	}
	if b.Marked {
		parts = append(parts, "«"+string(b.Basis)+"»")
	}
	if status != "" {
		parts = append(parts, mermaidText(Label(status)))
	}
	if len(parts) == 0 {
		return ""
	}
	return "<br/><i>" + strings.Join(parts, " · ") + "</i>"
}

func (r *mermaidRenderer) markBasis(id string, b BasisView) {
	if b.Marked {
		r.classes[string(b.Basis)] = append(r.classes[string(b.Basis)], r.id(id))
	}
}

func healthClass(h string) string {
	switch h {
	case "green", "yellow", "red":
		return h
	default:
		return "grey"
	}
}

// groups writes a small logic node for each precondition group.
func (r *mermaidRenderer) groups() {
	for _, g := range r.v.Groups {
		label := strings.ToUpper(g.Logic)
		switch g.Logic {
		case "all":
			label = "AND"
		case "any":
			label = "OR"
		}
		r.line(1, fmt.Sprintf(`%s{"%s"}`, r.id(g.ID), mermaidText(label)))
		r.classes["logic"] = append(r.classes["logic"], r.id(g.ID))
	}
}

// edges writes links, group wiring, and the implicit edges from
// interventions to the outcomes they target.
func (r *mermaidRenderer) edges() {
	grouped := map[[2]string]bool{} // member -> target edges carried by a group
	for _, g := range r.v.Groups {
		for _, m := range g.Members {
			grouped[[2]string{m, g.Target}] = true
			r.line(1, fmt.Sprintf("%s --> %s", r.id(m), r.id(g.ID)))
		}
		r.line(1, fmt.Sprintf("%s --> %s", r.id(g.ID), r.id(g.Target)))
	}

	explicit := map[[2]string]bool{}
	for _, l := range r.v.Links {
		explicit[[2]string{l.From, l.To}] = true
		if grouped[[2]string{l.From, l.To}] && !l.Feedback {
			continue
		}
		r.line(1, r.linkLine(l))
	}
	if r.o.HideInterventions {
		return
	}
	for _, in := range r.v.Interventions {
		for _, t := range in.Targets {
			if explicit[[2]string{in.ID, t}] {
				continue
			}
			r.line(1, fmt.Sprintf("%s --> %s", r.id(in.ID), r.id(t)))
		}
	}
}

func (r *mermaidRenderer) linkLine(l LinkView) string {
	from, to := r.id(l.From), r.id(l.To)
	var label string
	switch {
	case l.Feedback && l.LoopType != "":
		label = l.LoopType + " loop"
	case l.Feedback:
		label = "feedback"
	case l.Polarity == string(toc.PolarityNegative):
		label = "reduces"
	}
	arrow := "-->"
	if l.Feedback {
		arrow = "-.->"
	}
	if label == "" {
		return fmt.Sprintf("%s %s %s", from, arrow, to)
	}
	return fmt.Sprintf(`%s %s|"%s"| %s`, from, arrow, mermaidText(label), to)
}

func (r *mermaidRenderer) assignClasses() {
	names := make([]string, 0, len(r.classes))
	for name := range r.classes {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		r.line(1, fmt.Sprintf("class %s %s", strings.Join(r.classes[name], ","), name))
	}
}

// mermaidText escapes text for use inside a quoted Mermaid label.
func mermaidText(s string) string {
	s = strings.ReplaceAll(s, "&", "#amp;")
	s = strings.ReplaceAll(s, `"`, "#quot;")
	s = strings.ReplaceAll(s, "<", "#lt;")
	s = strings.ReplaceAll(s, ">", "#gt;")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}
