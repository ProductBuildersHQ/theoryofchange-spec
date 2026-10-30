package report

import (
	"fmt"
	"strings"
)

// DefinitionKinds lists the definition report kinds that need no
// observations, in the order the definitions pack presents them.
var DefinitionKinds = []string{"pathway", "logic-model", "logic-model-diagram", "outcome-map", "logframe", "metrics", "metric-selection", "collection-plan", "glossary", "assessment", "definitions"}

// PathwayDiagrams renders the causal-pathway diagram of the whole theory,
// then one per pathway, each as a titled Mermaid block.
func PathwayDiagrams(whole *View, pathways []*View) string {
	var b strings.Builder
	b.WriteString("## Causal pathways\n\n")
	diagram := func(title string, v *View) {
		fmt.Fprintf(&b, "### %s\n\n```mermaid\n%s```\n\n", mdText(title), Mermaid(v, MermaidOptions{}))
	}
	diagram("The theory on one page", whole)
	for _, v := range pathways {
		diagram(PathwayTitle(v), v)
	}
	return b.String()
}

// Definitions renders every definition report as one document, the pack a
// workshop reviews before any data exists.
func Definitions(whole *View, pathways []*View) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s: definitions\n\n", mdText(whole.Document.Title))
	b.WriteString(PathwayDiagrams(whole, pathways))
	b.WriteString(LogicModels(whole, pathways))
	b.WriteString(OutcomeMaps(whole, pathways))
	b.WriteString(Logframes(whole, pathways))
	b.WriteString(MetricDefinitions(whole))
	b.WriteString(Glossary(whole))
	return b.String()
}
