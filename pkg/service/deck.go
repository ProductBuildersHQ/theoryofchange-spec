package service

import (
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/report"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/slides"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

// WorkshopDeck renders the workshop deck for a document as Marp Markdown.
// It needs no observations. Invalid documents are rejected by the graph
// build, as for every other read-only command.
func WorkshopDeck(doc *toc.Document, theme string) (string, error) {
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
	return slides.Workshop(whole, pathways, slides.Options{Theme: theme}), nil
}
