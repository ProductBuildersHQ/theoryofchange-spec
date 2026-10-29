// Package theoryofchange is a Go-first specification for representing,
// measuring, and reporting on Theories of Change.
//
// Go structs are the source of truth; JSON Schemas are generated from them
// and embedded. Tooling built on the model validates documents, analyzes
// their causal graphs, edits them through validated patches, rolls up
// indicator observations into pathway status, and renders reports and
// slide decks.
//
// Key packages:
//   - pkg/ref: typed references to ToC documents and their elements
//   - pkg/toc: the Theory of Change document model
//   - schema: JSON Schemas generated from pkg/toc, embedded for runtime use
//
// See docs/specs/initiatives/INIT-THEORYOFCHANGESPEC-001 for the PRD, TRD,
// plan, and roadmap, and docs/research for the methodology background.
package theoryofchange
