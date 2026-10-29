// Package tocio loads and writes Theory of Change documents.
//
// Documents may be authored as JSON or YAML. YAML is converted to JSON
// before schema validation and decoding, so both formats validate and
// decode identically. Output is canonical: stable field order, two-space
// indentation, and a trailing newline, so diffs stay readable and
// formatting is idempotent.
package tocio

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Format is a serialization format.
type Format string

// Supported formats.
const (
	FormatJSON Format = "json"
	FormatYAML Format = "yaml"
)

// Limits protect against oversized or hostile input.
const (
	// MaxFileBytes is the maximum accepted input size.
	MaxFileBytes = 10 << 20
	// MaxElements is the maximum number of identified elements in a
	// document.
	MaxElements = 10000
)

// Sentinel errors, wrapped with detail by the functions below.
var (
	ErrUnsupportedFormat = errors.New("unsupported format")
	ErrInvalidDocument   = errors.New("invalid document")
	ErrTooLarge          = errors.New("document too large")
)

// FormatFromPath returns the format implied by a file extension (.json,
// .yaml, .yml).
func FormatFromPath(path string) (Format, error) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		return FormatJSON, nil
	case ".yaml", ".yml":
		return FormatYAML, nil
	default:
		return "", fmt.Errorf("%w: %q (want .json, .yaml, or .yml)", ErrUnsupportedFormat, path)
	}
}

// Valid reports whether f is a supported format.
func (f Format) Valid() bool { return f == FormatJSON || f == FormatYAML }

// ToJSON converts input in the given format to JSON bytes. JSON input is
// checked for well-formedness and returned unchanged. Unquoted YAML dates
// become date strings so authors need not quote them.
func ToJSON(data []byte, f Format) ([]byte, error) {
	if len(data) > MaxFileBytes {
		return nil, fmt.Errorf("%w: %d bytes exceeds %d", ErrTooLarge, len(data), MaxFileBytes)
	}
	switch f {
	case FormatJSON:
		if !json.Valid(data) {
			var v any
			err := json.Unmarshal(data, &v)
			return nil, fmt.Errorf("%w: invalid JSON: %w", ErrInvalidDocument, err)
		}
		return data, nil
	case FormatYAML:
		return yamlToJSON(data)
	default:
		return nil, fmt.Errorf("%w: %q", ErrUnsupportedFormat, string(f))
	}
}

func yamlToJSON(data []byte) ([]byte, error) {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, fmt.Errorf("%w: invalid YAML: %w", ErrInvalidDocument, err)
	}
	var extra any
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("%w: YAML input must contain exactly one document", ErrInvalidDocument)
	}
	norm, err := normalizeYAML(v, "")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidDocument, err)
	}
	out, err := json.Marshal(norm)
	if err != nil {
		return nil, fmt.Errorf("%w: encode YAML as JSON: %w", ErrInvalidDocument, err)
	}
	return out, nil
}

// normalizeYAML converts decoded YAML into values encoding/json can emit:
// string keys only, and timestamps as strings.
func normalizeYAML(v any, path string) (any, error) {
	switch x := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, child := range x {
			n, err := normalizeYAML(child, path+"/"+k)
			if err != nil {
				return nil, err
			}
			out[k] = n
		}
		return out, nil
	case map[any]any:
		return nil, fmt.Errorf("non-string map key near %q", path)
	case []any:
		out := make([]any, len(x))
		for i, child := range x {
			n, err := normalizeYAML(child, fmt.Sprintf("%s/%d", path, i))
			if err != nil {
				return nil, err
			}
			out[i] = n
		}
		return out, nil
	case time.Time:
		if x.Hour() == 0 && x.Minute() == 0 && x.Second() == 0 && x.Nanosecond() == 0 {
			return x.Format(time.DateOnly), nil
		}
		return x.Format(time.RFC3339Nano), nil
	case nil, string, bool, int, int64, uint64, float64:
		return x, nil
	default:
		return nil, fmt.Errorf("unsupported YAML value %T near %q", v, path)
	}
}
