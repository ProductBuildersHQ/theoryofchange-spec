// Package judge prepares an impact model for review by an LLM-as-a-Judge
// and checks, stores, and combines what the judges return. It makes no model
// calls: a judge, human or agent, reads a bundle and returns a verdict.
//
// A bundle carries a hash of its content, and a verdict carries the hash of
// the bundle it judged, so a verdict can never be applied to an impact model that has
// since changed.
package judge

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/plexusone/structured-evaluation/rubric"
	"github.com/plexusone/structured-evaluation/summary"

	"github.com/ProductBuildersHQ/impact-model/pkg/graph"
	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
	"github.com/ProductBuildersHQ/impact-model/pkg/impactio"
	"github.com/ProductBuildersHQ/impact-model/pkg/lint"
	"github.com/ProductBuildersHQ/impact-model/pkg/report"
	"github.com/ProductBuildersHQ/impact-model/spec"
)

// BundleKind identifies a judge bundle.
const BundleKind = "JudgeBundle"

// Bundle is everything a judge needs, and nothing it should not see.
type Bundle struct {
	Kind    string `json:"kind"`
	ModelID string `json:"modelId"`
	// Profile is the rubric profile: core or development.
	Profile string `json:"profile"`
	// Rubric is what the judge scores against.
	Rubric *rubric.RubricSet `json:"rubric"`
	// Document is the impact model in canonical JSON. Evidence that is restricted,
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
func Build(doc *impact.Model, profile string) (*Bundle, error) {
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
	canon, err := impactio.Canonical(redact(doc), impactio.FormatJSON)
	if err != nil {
		return nil, err
	}
	b := &Bundle{
		Kind: BundleKind, ModelID: doc.ID, Profile: profile, Rubric: rs,
		Document: json.RawMessage(canon), Narrative: narrative,
		Lint:         res.SummaryReport(doc.ID, doc.Metadata.Version, time.Time{}),
		Instructions: spec.JudgeInstructions(),
	}
	if b.BundleHash, err = impactio.ContentHash(b); err != nil {
		return nil, fmt.Errorf("judge: hash bundle: %w", err)
	}
	return b, nil
}

func narrativeOf(doc *impact.Model, g *graph.Graph) (string, error) {
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
func redact(doc *impact.Model) *impact.Model {
	data, err := json.Marshal(doc)
	if err != nil {
		panic(fmt.Sprintf("judge: a document always encodes: %v", err))
	}
	var cp impact.Model
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
