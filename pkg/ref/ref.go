// Package ref provides typed references to Theory of Change documents and
// the elements inside them.
//
// The reference grammar is
//
//	ref       = "toc:" docSlug [ "#" elementId ]
//	docSlug   = slug
//	elementId = slug
//	slug      = [a-z][a-z0-9]*(-[a-z0-9]+)*   (at most 128 characters)
//
// Examples:
//
//	toc:agent-auth-open-standard
//	toc:agent-auth-open-standard#oc-idp-support
//
// References are stable across repositories, so other models (for example a
// strategy or futures model) can point at a ToC or one of its elements
// without importing it.
package ref

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Prefix is the scheme prefix of every Theory of Change reference.
const Prefix = "toc:"

// MaxSlugLength is the maximum length of a document slug or element ID.
const MaxSlugLength = 128

var slugRegexp = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)

// Sentinel errors returned (wrapped) by Parse and ValidateSlug.
var (
	ErrEmptySlug     = errors.New("slug is empty")
	ErrSlugTooLong   = fmt.Errorf("slug exceeds %d characters", MaxSlugLength)
	ErrInvalidSlug   = errors.New("slug must be lowercase kebab-case starting with a letter")
	ErrInvalidPrefix = fmt.Errorf("reference must start with %q", Prefix)
)

// ValidateSlug reports whether s is a valid document slug or element ID.
func ValidateSlug(s string) error {
	switch {
	case s == "":
		return ErrEmptySlug
	case len(s) > MaxSlugLength:
		return fmt.Errorf("%w: %q", ErrSlugTooLong, s)
	case !slugRegexp.MatchString(s):
		return fmt.Errorf("%w: %q", ErrInvalidSlug, s)
	}
	return nil
}

// IsSlug reports whether s is a valid slug.
func IsSlug(s string) bool {
	return ValidateSlug(s) == nil
}

// Ref identifies a Theory of Change document and, optionally, one element
// within it.
type Ref struct {
	// Doc is the document slug (the document's id).
	Doc string
	// Element is the element ID within the document; empty for a reference
	// to the whole document.
	Element string
}

// New returns a validated Ref. element may be empty.
func New(doc, element string) (Ref, error) {
	r := Ref{Doc: doc, Element: element}
	if err := r.Validate(); err != nil {
		return Ref{}, err
	}
	return r, nil
}

// MustNew is like New but panics on invalid input. Use it only with
// compile-time constants.
func MustNew(doc, element string) Ref {
	r, err := New(doc, element)
	if err != nil {
		panic(err)
	}
	return r
}

// Parse parses a string of the form toc:<doc>[#<element>].
func Parse(s string) (Ref, error) {
	rest, ok := strings.CutPrefix(s, Prefix)
	if !ok {
		return Ref{}, fmt.Errorf("%w: %q", ErrInvalidPrefix, s)
	}
	doc, element, hasElement := strings.Cut(rest, "#")
	if hasElement && element == "" {
		return Ref{}, fmt.Errorf("parse %q: element: %w", s, ErrEmptySlug)
	}
	r := Ref{Doc: doc, Element: element}
	if err := r.Validate(); err != nil {
		return Ref{}, fmt.Errorf("parse %q: %w", s, err)
	}
	return r, nil
}

// Validate checks the document slug and, when present, the element ID.
func (r Ref) Validate() error {
	if err := ValidateSlug(r.Doc); err != nil {
		return fmt.Errorf("doc: %w", err)
	}
	if r.Element != "" {
		if err := ValidateSlug(r.Element); err != nil {
			return fmt.Errorf("element: %w", err)
		}
	}
	return nil
}

// IsZero reports whether r is the zero Ref.
func (r Ref) IsZero() bool {
	return r.Doc == "" && r.Element == ""
}

// IsElement reports whether r points at an element rather than a document.
func (r Ref) IsElement() bool {
	return r.Element != ""
}

// Document returns the reference to r's document (dropping any element).
func (r Ref) Document() Ref {
	return Ref{Doc: r.Doc}
}

// WithElement returns a reference to the given element of r's document.
func (r Ref) WithElement(element string) Ref {
	return Ref{Doc: r.Doc, Element: element}
}

// String formats r as toc:<doc>[#<element>]. The zero Ref formats as "".
func (r Ref) String() string {
	if r.IsZero() {
		return ""
	}
	if r.Element == "" {
		return Prefix + r.Doc
	}
	return Prefix + r.Doc + "#" + r.Element
}

// MarshalText implements encoding.TextMarshaler, so a Ref is encoded as a
// JSON string. Invalid references are rejected.
func (r Ref) MarshalText() ([]byte, error) {
	if r.IsZero() {
		return []byte{}, nil
	}
	if err := r.Validate(); err != nil {
		return nil, err
	}
	return []byte(r.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler. An empty string
// decodes to the zero Ref.
func (r *Ref) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*r = Ref{}
		return nil
	}
	parsed, err := Parse(string(text))
	if err != nil {
		return err
	}
	*r = parsed
	return nil
}
