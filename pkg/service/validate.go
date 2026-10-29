// Package service holds one function per user operation. The CLI and the
// MCP server are thin adapters over it, so both behave identically.
package service

import (
	"fmt"
	"os"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/tocio"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/validate"
)

// Report is the validation result for one input.
type Report struct {
	// Source is the file path or another label for the input.
	Source string           `json:"source"`
	Format tocio.Format     `json:"format,omitempty"`
	Kind   toc.Kind         `json:"kind,omitempty"`
	Issues []validate.Issue `json:"issues"`
}

// Valid reports whether the input has no error-severity issues.
func (r Report) Valid() bool { return !validate.HasErrors(r.Issues) }

// ValidateBytes validates one input of any document kind. A Theory of
// Change is checked against its schema and structurally; other kinds are
// checked against their schema. When against is given, an observation set
// is also checked against that theory.
func ValidateBytes(source string, data []byte, f tocio.Format, against *toc.Document) Report {
	r := Report{Source: source, Format: f, Issues: []validate.Issue{}}
	if kind, err := tocio.DetectKind(data, f); err == nil {
		r.Kind = kind
	}
	_, issues := validate.Source(data, f)
	r.Issues = append(r.Issues, issues...)
	if r.Kind == toc.KindObservationSet && against != nil && !validate.HasErrors(issues) {
		set, err := tocio.ParseObservationSet(data, f)
		if err != nil {
			r.Issues = append(r.Issues, validate.Issue{Code: "VAL-SCH-002", Severity: validate.SeverityError, Message: err.Error()})
		} else {
			r.Issues = append(r.Issues, validate.Observations(against, set)...)
		}
	}
	return r
}

// ValidateFile reads and validates a file. The format comes from the
// extension.
func ValidateFile(path string, against *toc.Document) (Report, error) {
	data, f, err := readInput(path)
	if err != nil {
		return Report{}, err
	}
	return ValidateBytes(path, data, f, against), nil
}

func readInput(path string) ([]byte, tocio.Format, error) {
	f, err := tocio.FormatFromPath(path)
	if err != nil {
		return nil, "", err
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, "", fmt.Errorf("stat %s: %w", path, err)
	}
	if info.Size() > tocio.MaxFileBytes {
		return nil, "", fmt.Errorf("%w: %s is %d bytes (limit %d)", tocio.ErrTooLarge, path, info.Size(), tocio.MaxFileBytes)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", fmt.Errorf("read %s: %w", path, err)
	}
	return data, f, nil
}
