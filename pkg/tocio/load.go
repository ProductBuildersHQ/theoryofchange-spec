package tocio

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
	"github.com/ProductBuildersHQ/theoryofchange-spec/schema"
)

// DetectKind reports the document kind declared by the input's "kind"
// field.
func DetectKind(data []byte, f Format) (toc.Kind, error) {
	j, err := ToJSON(data, f)
	if err != nil {
		return "", err
	}
	var head struct {
		Kind toc.Kind `json:"kind"`
	}
	if err := json.Unmarshal(j, &head); err != nil {
		return "", fmt.Errorf("%w: read kind: %w", ErrInvalidDocument, err)
	}
	if head.Kind == "" {
		return "", fmt.Errorf("%w: missing \"kind\"", ErrInvalidDocument)
	}
	return head.Kind, nil
}

// SchemaName returns the embedded schema that validates a document kind.
func SchemaName(k toc.Kind) (string, error) {
	switch k {
	case toc.KindTheoryOfChange:
		return schema.TheoryOfChange, nil
	case toc.KindObservationSet:
		return schema.ObservationSet, nil
	case toc.KindStatus:
		return schema.Status, nil
	case toc.KindActorRegistry:
		return schema.ActorRegistry, nil
	default:
		return "", fmt.Errorf("%w: unknown kind %q", ErrInvalidDocument, string(k))
	}
}

// decode validates JSON against the named schema, then decodes it strictly
// (unknown fields are errors) into v.
func decode[T any](j []byte, schemaName string) (*T, error) {
	if err := schema.Validate(schemaName, j); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidDocument, err)
	}
	dec := json.NewDecoder(bytes.NewReader(j))
	dec.DisallowUnknownFields()
	var v T
	if err := dec.Decode(&v); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidDocument, err)
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("%w: trailing data after document", ErrInvalidDocument)
	}
	return &v, nil
}

// ParseDocument parses a Theory of Change in the given format. The input is
// converted to JSON, validated against the embedded schema, and decoded
// strictly. A failure wraps ErrInvalidDocument; schema failures also wrap
// the validator's error, whose output lists JSON-pointer locations.
func ParseDocument(data []byte, f Format) (*toc.Document, error) {
	j, err := ToJSON(data, f)
	if err != nil {
		return nil, err
	}
	doc, err := decode[toc.Document](j, schema.TheoryOfChange)
	if err != nil {
		return nil, err
	}
	if n := len(doc.Elements()); n > MaxElements {
		return nil, fmt.Errorf("%w: %d elements exceeds %d", ErrTooLarge, n, MaxElements)
	}
	return doc, nil
}

// ParseObservationSet parses an observation set in the given format.
func ParseObservationSet(data []byte, f Format) (*toc.ObservationSet, error) {
	j, err := ToJSON(data, f)
	if err != nil {
		return nil, err
	}
	return decode[toc.ObservationSet](j, schema.ObservationSet)
}

// ParseActorRegistry parses an actor registry in the given format.
func ParseActorRegistry(data []byte, f Format) (*toc.ActorRegistry, error) {
	j, err := ToJSON(data, f)
	if err != nil {
		return nil, err
	}
	return decode[toc.ActorRegistry](j, schema.ActorRegistry)
}

// ParseStatus parses a status snapshot in the given format.
func ParseStatus(data []byte, f Format) (*toc.Status, error) {
	j, err := ToJSON(data, f)
	if err != nil {
		return nil, err
	}
	return decode[toc.Status](j, schema.Status)
}

func readFile(path string) ([]byte, Format, error) {
	f, err := FormatFromPath(path)
	if err != nil {
		return nil, "", err
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, "", fmt.Errorf("stat %s: %w", path, err)
	}
	if info.Size() > MaxFileBytes {
		return nil, "", fmt.Errorf("%w: %s is %d bytes (limit %d)", ErrTooLarge, path, info.Size(), MaxFileBytes)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", fmt.Errorf("read %s: %w", path, err)
	}
	return data, f, nil
}

// LoadDocument reads and parses a Theory of Change file (.json, .yaml, or
// .yml).
func LoadDocument(path string) (*toc.Document, error) {
	data, f, err := readFile(path)
	if err != nil {
		return nil, err
	}
	doc, err := ParseDocument(data, f)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return doc, nil
}

// LoadObservations reads and parses observation set files, in the order
// given.
func LoadObservations(paths ...string) ([]*toc.ObservationSet, error) {
	sets := make([]*toc.ObservationSet, 0, len(paths))
	for _, p := range paths {
		data, f, err := readFile(p)
		if err != nil {
			return nil, err
		}
		set, err := ParseObservationSet(data, f)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", p, err)
		}
		sets = append(sets, set)
	}
	return sets, nil
}

// MergeObservations combines observation sets for one Theory of Change into
// a single set. Observations, assessments, and notes are concatenated in
// input order; the merged Period is the common period if all sets share one
// and empty otherwise. Sets for different documents cannot be merged.
func MergeObservations(sets ...*toc.ObservationSet) (*toc.ObservationSet, error) {
	if len(sets) == 0 {
		return nil, errors.New("merge observations: no observation sets")
	}
	merged := toc.NewObservationSet(sets[0].TocID, sets[0].Period)
	for i, s := range sets {
		if s.TocID != merged.TocID {
			return nil, fmt.Errorf("merge observations: set %d measures %q, expected %q", i, s.TocID, merged.TocID)
		}
		if s.Period != merged.Period {
			merged.Period = ""
		}
		merged.Observations = append(merged.Observations, s.Observations...)
		merged.Assessments = append(merged.Assessments, s.Assessments...)
		merged.Notes = append(merged.Notes, s.Notes...)
	}
	return merged, nil
}

// Reformat rewrites data in canonical form in the same format, by decoding
// into the typed model for the declared kind and re-encoding it. Field order
// follows the model, so reformatting YAML never reorders by key name.
func Reformat(data []byte, f Format) ([]byte, error) {
	kind, err := DetectKind(data, f)
	if err != nil {
		return nil, err
	}
	var v any
	switch kind {
	case toc.KindTheoryOfChange:
		v, err = ParseDocument(data, f)
	case toc.KindObservationSet:
		v, err = ParseObservationSet(data, f)
	case toc.KindStatus:
		v, err = ParseStatus(data, f)
	case toc.KindActorRegistry:
		v, err = ParseActorRegistry(data, f)
	default:
		return nil, fmt.Errorf("%w: unknown kind %q", ErrInvalidDocument, string(kind))
	}
	if err != nil {
		return nil, err
	}
	return Canonical(v, f)
}

// WriteFile writes data to path atomically: it writes a temporary file in
// the same directory and renames it into place, so a failure never leaves a
// half-written document. An existing file keeps its permissions; new files
// are created with mode 0o644 because specifications are meant to be shared.
func WriteFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	mode := os.FileMode(0o644)
	if info, err := os.Stat(path); err == nil {
		mode = info.Mode().Perm()
	}
	tmp, err := os.CreateTemp(dir, ".tocio-*")
	if err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	tmpName := tmp.Name()
	// fail removes the temporary file and reports why the write failed,
	// including any failure to remove it.
	fail := func(err error) error {
		if rmErr := os.Remove(tmpName); rmErr != nil && !errors.Is(rmErr, os.ErrNotExist) {
			err = errors.Join(err, rmErr)
		}
		return fmt.Errorf("write %s: %w", path, err)
	}
	if _, err := tmp.Write(data); err != nil {
		return fail(errors.Join(err, tmp.Close()))
	}
	if err := tmp.Close(); err != nil {
		return fail(err)
	}
	if err := os.Chmod(tmpName, mode); err != nil {
		return fail(err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fail(err)
	}
	return nil
}
