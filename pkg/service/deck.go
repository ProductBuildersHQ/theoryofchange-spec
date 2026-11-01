package service

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ProductBuildersHQ/impact-model/pkg/burden"
	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
	"github.com/ProductBuildersHQ/impact-model/pkg/logicmodel"
	"github.com/ProductBuildersHQ/impact-model/pkg/problemtree"
	"github.com/ProductBuildersHQ/impact-model/pkg/report"
	"github.com/ProductBuildersHQ/impact-model/pkg/slides"
)

// WorkshopDeck renders the workshop deck for a document as Marp Markdown.
// stamp, if not empty, is shown on the title slide.
// It needs no observations. Invalid documents are rejected by the graph
// build, as for every other read-only command.
func WorkshopDeck(doc *impact.Model, theme, stamp string) (string, error) {
	g, err := BuildGraph(doc)
	if err != nil {
		return "", err
	}
	whole, err := report.BuildView(doc, g, nil, report.Options{})
	if err != nil {
		return "", err
	}
	pathways, err := report.PathwayViews(doc, g, nil)
	if err != nil {
		return "", err
	}
	return slides.Workshop(whole, pathways, slides.Options{Theme: theme, Stamp: stamp}), nil
}

// DefinitionReport renders a definition report (see report.DefinitionKinds)
// as Markdown. pathway restricts the report to one pathway; it is empty for
// the whole impact model. Definition reports need no observations.
func DefinitionReport(doc *impact.Model, kind, pathway string) (string, error) {
	if !slices.Contains(report.DefinitionKinds, kind) {
		return "", fmt.Errorf("unknown definition report %q (want one of %s)", kind, strings.Join(report.DefinitionKinds, ", "))
	}
	g, err := BuildGraph(doc)
	if err != nil {
		return "", err
	}
	whole, err := report.BuildView(doc, g, nil, report.Options{Pathway: pathway})
	if err != nil {
		return "", err
	}
	var pathways []*report.View
	if pathway == "" {
		if pathways, err = report.PathwayViews(doc, g, nil); err != nil {
			return "", err
		}
	}
	switch kind {
	case "pathway":
		return report.PathwayDiagrams(whole, pathways), nil
	case "logic-model":
		return report.LogicModels(whole, pathways), nil
	case "logic-model-diagram":
		return report.LogicModelDiagrams(whole, pathways), nil
	case "logic-model-layered":
		return logicmodel.Diagrams(whole, pathways, logicmodel.FormatMermaid)
	case "logic-model-d2":
		return logicmodel.Diagrams(whole, pathways, logicmodel.FormatD2)
	case "problem-tree":
		return problemtree.Report(doc, problemtree.FormatMermaid)
	case "problem-tree-d2":
		return problemtree.Report(doc, problemtree.FormatD2)
	case "problem-tree-json":
		return problemtree.Build(doc).JSON()
	case "outcome-map":
		return report.OutcomeMaps(whole, pathways), nil
	case "logframe":
		return report.Logframes(whole, pathways), nil
	case "metrics":
		return report.MetricDefinitions(whole), nil
	case "glossary":
		return report.Glossary(whole), nil
	case "metric-selection":
		return report.MetricSelection(whole, doc), nil
	case "collection-plan":
		rep, err := burden.Compute(doc)
		if err != nil {
			return "", err
		}
		return report.CollectionPlan(doc, rep), nil
	case "assessment":
		return report.Assessment(whole, doc), nil
	default:
		return report.Definitions(whole, pathways, doc)
	}
}

// ExportCSV renders one CSV table (see report.CSVTables) from a document.
func ExportCSV(doc *impact.Model, table string) (string, error) {
	g, err := BuildGraph(doc)
	if err != nil {
		return "", err
	}
	v, err := report.BuildView(doc, g, nil, report.Options{})
	if err != nil {
		return "", err
	}
	return report.CSV(table, v, doc)
}
