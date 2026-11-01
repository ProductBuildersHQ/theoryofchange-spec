// Package share prepares impact model artifacts for sharing outside the
// tools: the export stamp every artifact carries, and the Markdown profile
// that Pandoc converts to DOCX and PPTX. It only produces text. Rendering
// diagrams and running Pandoc are done by the command line, never here.
package share

import (
	"fmt"
	"strings"

	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
	"github.com/ProductBuildersHQ/impact-model/pkg/impactio"
)

// Stamp identifies exactly which version of which impact model an artifact was
// made from, so a reader can tell whether it is out of date.
type Stamp struct {
	ModelID string `json:"modelId"`
	Version string `json:"version"`
	// ContentHash is the SHA-256 (hex) of the document's canonical JSON.
	ContentHash string `json:"contentHash"`
	// AsOf is the date the artifact describes; empty if none applies.
	AsOf string `json:"asOf,omitempty"`
}

// unversioned labels a document that declares no version.
const unversioned = "unversioned"

// NewStamp stamps doc. The version is the document's metadata version, else
// that of its latest revision, else "unversioned". asOf is used as given; a
// stamp never reads the clock, so the same inputs give the same artifact.
func NewStamp(doc *impact.Model, asOf string) (Stamp, error) {
	hash, err := impactio.ContentHash(doc)
	if err != nil {
		return Stamp{}, fmt.Errorf("share: hash document: %w", err)
	}
	version := doc.Metadata.Version
	if version == "" && len(doc.Revisions) > 0 {
		version = doc.Revisions[len(doc.Revisions)-1].Version
	}
	if version == "" {
		version = unversioned
	}
	return Stamp{ModelID: doc.ID, Version: version, ContentHash: hash, AsOf: asOf}, nil
}

// shortHash is the leading characters of the hash, enough to compare by eye.
func (s Stamp) shortHash() string {
	if len(s.ContentHash) > 12 {
		return s.ContentHash[:12]
	}
	return s.ContentHash
}

// Line is the one-line form used in footers and on title slides.
func (s Stamp) Line() string {
	line := fmt.Sprintf("Impact model %s · version %s · hash %s", s.ModelID, s.Version, s.shortHash())
	if s.AsOf != "" {
		line += " · as of " + s.AsOf
	}
	return line
}

// YAMLBlock is the Markdown metadata block: the title and the stamp, in the
// form Pandoc reads.
func (s Stamp) YAMLBlock(title string) string {
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "title: %q\n", title)
	fmt.Fprintf(&b, "version: %q\n", s.Version)
	if s.AsOf != "" {
		fmt.Fprintf(&b, "date: %q\n", s.AsOf)
	}
	fmt.Fprintf(&b, "model-id: %q\n", s.ModelID)
	fmt.Fprintf(&b, "content-hash: %q\n", s.ContentHash)
	b.WriteString("---\n\n")
	return b.String()
}

// CSVRow is the leading comment row of a CSV export.
func (s Stamp) CSVRow() string {
	return fmt.Sprintf("# model:%s version:%s hash:%s\r\n", s.ModelID, s.Version, s.ContentHash)
}

// StampMarkdown prefixes a Markdown report with the metadata block.
func StampMarkdown(s Stamp, title, body string) string {
	return s.YAMLBlock(title) + body
}

// StampCSV prefixes a CSV export with the comment row.
func StampCSV(s Stamp, csv string) string { return s.CSVRow() + csv }
