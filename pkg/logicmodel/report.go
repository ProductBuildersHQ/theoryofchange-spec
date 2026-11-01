package logicmodel

import (
	"fmt"
	"strings"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/report"
)

// Format is how a diagram is written.
type Format string

// The formats a diagram can be written in.
const (
	FormatMermaid Format = "mermaid"
	FormatD2      Format = "d2"
)

// Diagrams renders the layered logic model diagram of each pathway, or of the
// whole theory when it has no pathways, each under a heading and in a fenced
// block that names its format.
func Diagrams(whole *report.View, pathways []*report.View, f Format) (string, error) {
	views := pathways
	if len(views) == 0 {
		views = []*report.View{whole}
	}
	var b strings.Builder
	for _, v := range views {
		g := Build(v)
		var text string
		switch f {
		case FormatMermaid:
			var err error
			if text, err = g.Mermaid(); err != nil {
				return "", err
			}
		case FormatD2:
			text = g.D2()
		default:
			return "", fmt.Errorf("logicmodel: unknown diagram format %q", f)
		}
		fmt.Fprintf(&b, "## Layered logic model: %s\n\n```%s\n%s```\n\n", strings.Join(strings.Fields(g.Title), " "), f, text)
		b.WriteString(feedbackNote(g))
	}
	return b.String(), nil
}

// feedbackNote lists the feedback loops the diagram leaves out.
func feedbackNote(g *Graph) string {
	if len(g.Feedback) == 0 {
		return ""
	}
	label := map[string]string{}
	for _, st := range g.Stages {
		for _, n := range st.Nodes {
			label[n.ID] = n.Label
		}
	}
	var b strings.Builder
	b.WriteString("Feedback, not drawn:\n\n")
	for _, e := range g.Feedback {
		fmt.Fprintf(&b, "- %s leads back to %s\n", label[e.From], label[e.To])
	}
	b.WriteString("\n")
	return b.String()
}
