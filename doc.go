// Package impactmodel is a Go-first specification for representing,
// measuring, and reporting on an impact: why a program expects to bring about
// change, what it does, and what it measures. The impact model, the logic
// model, the outcome map, and the logframe are views of one impact model.
//
// Go structs are the source of truth; JSON Schemas are generated from them
// and embedded. Tooling built on the model validates documents, analyzes
// their causal graphs, edits them through validated patches, rolls up
// indicator observations into pathway status, and renders reports and
// slide decks.
//
// Key packages:
//   - pkg/ref: typed references to impact models and their elements
//   - pkg/impact: the impact model
//   - schema: JSON Schemas generated from pkg/impact, embedded for runtime use
//
// See docs/specs/initiatives/INIT-THEORYOFCHANGESPEC-001 for the PRD, TRD,
// plan, and roadmap, and docs/research for the methodology background.
package impactmodel
