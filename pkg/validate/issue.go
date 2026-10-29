// Package validate checks Theory of Change documents and observation sets
// for structural problems: schema conformance, identifier uniqueness and
// format, resolution of every reference, and enumeration validity.
//
// Findings are returned as Issues with stable codes and JSON-pointer paths,
// in a deterministic order. Quality judgments (is this a good theory?) are
// the lint package's job, not this one's.
package validate

import (
	"fmt"
	"strings"
)

// Severity ranks an issue.
type Severity string

// Severities.
const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

// Issue is one finding.
type Issue struct {
	// Code is a stable identifier such as VAL-REF-001.
	Code     string   `json:"code"`
	Severity Severity `json:"severity"`
	// Path is the JSON pointer to the offending value, e.g. "/links/3/to".
	Path string `json:"path"`
	// Element is the ID of the element that contains the problem, if any.
	Element string `json:"element,omitempty"`
	Message string `json:"message"`
	// Source names the input when several are validated together, such as
	// "observations[1]".
	Source string `json:"source,omitempty"`
}

// String formats the issue for humans.
func (i Issue) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s %s", i.Severity, i.Code)
	if i.Source != "" {
		fmt.Fprintf(&b, " %s", i.Source)
	}
	if i.Path != "" {
		fmt.Fprintf(&b, " %s", i.Path)
	}
	fmt.Fprintf(&b, ": %s", i.Message)
	return b.String()
}

// HasErrors reports whether any issue has error severity.
func HasErrors(issues []Issue) bool {
	return CountSeverity(issues, SeverityError) > 0
}

// CountSeverity counts issues of one severity.
func CountSeverity(issues []Issue, s Severity) int {
	n := 0
	for _, i := range issues {
		if i.Severity == s {
			n++
		}
	}
	return n
}

// Errors returns only the error-severity issues.
func Errors(issues []Issue) []Issue {
	var out []Issue
	for _, i := range issues {
		if i.Severity == SeverityError {
			out = append(out, i)
		}
	}
	return out
}

// Delta compares two issue lists and returns the issues present only in
// after (introduced) and only in before (resolved). Issues are matched on
// code, element, and message — not path — so array indexes shifting does not
// make an unchanged problem look new. Duplicates are matched one-to-one.
func Delta(before, after []Issue) (introduced, resolved []Issue) {
	key := func(i Issue) string { return i.Code + "\x00" + i.Element + "\x00" + i.Source + "\x00" + i.Message }
	remaining := map[string][]Issue{}
	for _, i := range before {
		remaining[key(i)] = append(remaining[key(i)], i)
	}
	for _, i := range after {
		k := key(i)
		if list := remaining[k]; len(list) > 0 {
			remaining[k] = list[1:]
			continue
		}
		introduced = append(introduced, i)
	}
	for _, i := range before {
		k := key(i)
		if list := remaining[k]; len(list) > 0 {
			resolved = append(resolved, list[0])
			remaining[k] = list[1:]
		}
	}
	return introduced, resolved
}

// pointer escapes one JSON-pointer segment (RFC 6901).
func pointer(parts ...string) string {
	var b strings.Builder
	for _, p := range parts {
		p = strings.ReplaceAll(p, "~", "~0")
		p = strings.ReplaceAll(p, "/", "~1")
		b.WriteByte('/')
		b.WriteString(p)
	}
	return b.String()
}
