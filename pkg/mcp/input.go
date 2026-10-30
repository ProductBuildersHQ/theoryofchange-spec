package mcp

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/measure"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/tocio"
)

// maxInline bounds an inline document or observation set, the same limit as a
// file.
const maxInline = tocio.MaxFileBytes

// Source names a theory: inline as the document argument, or by path inside
// the server root. Inline is preferred for agents, since it needs no file.
type Source struct {
	// Document is the theory itself: a JSON object, or a string of JSON or YAML.
	Document any `json:"document,omitempty" jsonschema:"the theory, as a JSON object or a string of JSON or YAML"`
	// Path is a theory file inside the server root.
	Path string `json:"path,omitempty" jsonschema:"a theory file inside the server root, used when document is not given"`
}

// Observations names observation sets the same two ways.
type Observations struct {
	Observations     []any    `json:"observations,omitempty" jsonschema:"observation sets given inline, as objects or strings"`
	ObservationPaths []string `json:"observationPaths,omitempty" jsonschema:"observation set files inside the server root"`
}

// text returns the raw bytes and format of an inline value.
func inlineBytes(v any) ([]byte, tocio.Format, error) {
	switch x := v.(type) {
	case string:
		if len(x) > maxInline {
			return nil, "", fmt.Errorf("the inline document is %d bytes; the limit is %d", len(x), maxInline)
		}
		return []byte(x), tocio.FormatYAML, nil // YAML reads JSON too
	default:
		b, err := json.Marshal(x)
		if err != nil {
			return nil, "", fmt.Errorf("the inline document cannot be encoded: %w", err)
		}
		if len(b) > maxInline {
			return nil, "", fmt.Errorf("the inline document is %d bytes; the limit is %d", len(b), maxInline)
		}
		return b, tocio.FormatJSON, nil
	}
}

// load reads the theory a Source names. It returns the resolved path when
// the theory came from a file, else "".
func (s *Server) load(src Source) (*toc.Document, string, error) {
	switch {
	case src.Document != nil:
		data, f, err := inlineBytes(src.Document)
		if err != nil {
			return nil, "", err
		}
		doc, err := tocio.ParseDocument(data, f)
		return doc, "", err
	case src.Path != "":
		path, err := s.root.Resolve(src.Path)
		if err != nil {
			return nil, "", err
		}
		doc, err := tocio.LoadDocument(path)
		return doc, path, err
	}
	return nil, "", errors.New("give a document or a path")
}

// loadObservations reads the observation sets in o.
func (s *Server) loadObservations(o Observations) ([]*toc.ObservationSet, error) {
	var sets []*toc.ObservationSet
	for i, v := range o.Observations {
		data, f, err := inlineBytes(v)
		if err != nil {
			return nil, fmt.Errorf("observations[%d]: %w", i, err)
		}
		set, err := tocio.ParseObservationSet(data, f)
		if err != nil {
			return nil, fmt.Errorf("observations[%d]: %w", i, err)
		}
		sets = append(sets, set)
	}
	if len(o.ObservationPaths) > 0 {
		paths := make([]string, len(o.ObservationPaths))
		for i, p := range o.ObservationPaths {
			rp, err := s.root.Resolve(p)
			if err != nil {
				return nil, err
			}
			paths[i] = rp
		}
		more, err := tocio.LoadObservations(paths...)
		if err != nil {
			return nil, err
		}
		sets = append(sets, more...)
	}
	return sets, nil
}

// asOfDate returns the status date: the one given, else the latest in the
// observations.
func asOfDate(asOf string, sets []*toc.ObservationSet) (time.Time, error) {
	if asOf != "" {
		at, err := time.Parse(time.DateOnly, asOf)
		if err != nil {
			return time.Time{}, fmt.Errorf("asOf must be a date like 2026-09-30: %w", err)
		}
		return at, nil
	}
	if at, ok := measure.LatestDate(sets); ok {
		return at, nil
	}
	return time.Time{}, errors.New("give asOf: there are no observations to take a date from")
}
