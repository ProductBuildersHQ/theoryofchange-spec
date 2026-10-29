package tocio

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Canonical serializes v (a document, observation set, status, or any other
// JSON-serializable value) in canonical form: struct field order, two-space
// indentation, no HTML escaping, and a trailing newline. Formatting canonical
// output again yields identical bytes.
func Canonical(v any, f Format) ([]byte, error) {
	j, err := canonicalJSON(v)
	if err != nil {
		return nil, err
	}
	switch f {
	case FormatJSON:
		return j, nil
	case FormatYAML:
		return jsonToYAML(j)
	default:
		return nil, fmt.Errorf("%w: %q", ErrUnsupportedFormat, string(f))
	}
}

// ContentHash returns the SHA-256 (hex) of v's canonical JSON. It is
// identical for the same content regardless of the input format, and is
// used to stamp exports and bind review responses to a version.
func ContentHash(v any) (string, error) {
	j, err := canonicalJSON(v)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(j)
	return hex.EncodeToString(sum[:]), nil
}

func canonicalJSON(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return nil, fmt.Errorf("encode canonical JSON: %w", err)
	}
	return buf.Bytes(), nil
}

// jsonToYAML renders JSON as block-style YAML, preserving key order. The
// JSON text is parsed as a YAML node tree (JSON is a YAML subset), then
// styles are cleared so the encoder picks plain scalars, quoting only where
// needed (for example the string "2026-01-01" or "true").
func jsonToYAML(j []byte) ([]byte, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(j, &doc); err != nil {
		return nil, fmt.Errorf("parse JSON as YAML: %w", err)
	}
	clearStyles(&doc)
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(&doc); err != nil {
		return nil, fmt.Errorf("encode YAML: %w", err)
	}
	if err := enc.Close(); err != nil {
		return nil, fmt.Errorf("encode YAML: %w", err)
	}
	return buf.Bytes(), nil
}

func clearStyles(n *yaml.Node) {
	n.Style = 0
	n.HeadComment, n.LineComment, n.FootComment = "", "", ""
	if n.Kind == yaml.ScalarNode && n.Tag == "!!str" && strings.Contains(n.Value, "\n") {
		n.Style = yaml.LiteralStyle
	}
	for _, c := range n.Content {
		clearStyles(c)
	}
}
