// Package judge prepares a Theory of Change for review by an LLM-as-a-Judge
// and checks, stores, and combines what the judges return. It makes no model
// calls: a judge, human or agent, reads a bundle and returns a verdict.
//
// A bundle carries a hash of its content, and a verdict carries the hash of
// the bundle it judged, so a verdict can never be applied to a theory that has
// since changed.
package judge

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/plexusone/structured-evaluation/rubric"
	"github.com/plexusone/structured-evaluation/summary"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/graph"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/lint"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/report"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/tocio"
	"github.com/ProductBuildersHQ/theoryofchange-spec/spec"
)

// BundleKind identifies a judge bundle.
const BundleKind = "JudgeBundle"

// Bundle is everything a judge needs, and nothing it should not see.
type Bundle struct {
	Kind  string `json:"kind"`
	TocID string `json:"tocId"`
	// Profile is the rubric profile: core or development.
	Profile string `json:"profile"`
	// Rubric is what the judge scores against.
	Rubric *rubric.RubricSet `json:"rubric"`
	// Document is the theory in canonical JSON. Evidence that is restricted,
	// or has no sensitivity set, is stripped of its excerpt, summary, and URL.
	Document json.RawMessage `json:"document"`
	// Narrative is the definitions pack and the assessment, in Markdown.
	Narrative string `json:"narrative"`
	// Lint is the result of the deterministic checks.
	Lint *summary.SummaryReport `json:"lint"`
	// Instructions tell the judge how to judge.
	Instructions string `json:"instructions"`
	// BundleHash is the SHA-256 of the bundle with this field empty.
	BundleHash string `json:"bundleHash"`
}

// Build assembles the bundle for doc. The document must be valid. The bundle
// is the same for the same document and profile.
func Build(doc *toc.Document, profile string) (*Bundle, error) {
	rs, err := spec.Rubric(profile)
	if err != nil {
		return nil, err
	}
	g, err := graph.Build(doc)
	if err != nil {
		return nil, err
	}
	res, err := lint.Run(doc, g, lint.Options{Profile: lint.ProfileStandard})
	if err != nil {
		return nil, err
	}
	narrative, err := narrativeOf(doc, g)
	if err != nil {
		return nil, err
	}
	canon, err := tocio.Canonical(redact(doc), tocio.FormatJSON)
	if err != nil {
		return nil, err
	}
	b := &Bundle{
		Kind: BundleKind, TocID: doc.ID, Profile: profile, Rubric: rs,
		Document: json.RawMessage(canon), Narrative: narrative,
		Lint:         res.SummaryReport(doc.ID, doc.Metadata.Version, time.Time{}),
		Instructions: spec.JudgeInstructions(),
	}
	if b.BundleHash, err = tocio.ContentHash(b); err != nil {
		return nil, fmt.Errorf("judge: hash bundle: %w", err)
	}
	return b, nil
}

func narrativeOf(doc *toc.Document, g *graph.Graph) (string, error) {
	whole, err := report.BuildView(doc, g, nil, report.Options{})
	if err != nil {
		return "", err
	}
	pathways, err := report.PathwayViews(doc, g, nil)
	if err != nil {
		return "", err
	}
	return report.Definitions(whole, pathways) + "\n" + report.Assessment(whole, doc), nil
}

// redact returns a copy of doc with the restricted parts of evidence removed,
// so what a judge reads matches what a report may show.
func redact(doc *toc.Document) *toc.Document {
	data, err := json.Marshal(doc)
	if err != nil {
		panic(fmt.Sprintf("judge: a document always encodes: %v", err))
	}
	var cp toc.Document
	if err := json.Unmarshal(data, &cp); err != nil {
		panic(fmt.Sprintf("judge: a document always decodes: %v", err))
	}
	for i := range cp.Evidence {
		if e := &cp.Evidence[i]; !e.Sensitivity.Renderable() {
			e.Excerpt, e.Summary, e.URL = "", "", ""
		}
	}
	return &cp
}
