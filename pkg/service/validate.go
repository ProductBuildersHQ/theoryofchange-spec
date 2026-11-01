// Package service holds one function per user operation. The CLI and the
// MCP server are thin adapters over it, so both behave identically.
package service

import (
	"fmt"
	"os"

	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
	"github.com/ProductBuildersHQ/impact-model/pkg/impactio"
	"github.com/ProductBuildersHQ/impact-model/pkg/validate"
)

// Report is the validation result for one input.
type Report struct {
	// Source is the file path or another label for the input.
	Source string           `json:"source"`
	Format impactio.Format  `json:"format,omitempty"`
	Kind   impact.Kind      `json:"kind,omitempty"`
	Issues []validate.Issue `json:"issues"`
}

// Valid reports whether the input has no error-severity issues.
func (r Report) Valid() bool { return !validate.HasErrors(r.Issues) }

// ValidateBytes validates one input of any document kind. A impact model of
// Change is checked against its schema and structurally; other kinds are
// checked against their schema. When against is given, an observation set
// is also checked against that impact model.
func ValidateBytes(source string, data []byte, f impactio.Format, against *impact.Model, opts ...validate.Option) Report {
	r := Report{Source: source, Format: f, Issues: []validate.Issue{}}
	if kind, err := impactio.DetectKind(data, f); err == nil {
		r.Kind = kind
	}
	_, issues := validate.Source(data, f, opts...)
	r.Issues = append(r.Issues, issues...)
	if r.Kind == impact.KindObservationSet && against != nil && !validate.HasErrors(issues) {
		set, err := impactio.ParseObservationSet(data, f)
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
func ValidateFile(path string, against *impact.Model, opts ...validate.Option) (Report, error) {
	data, f, err := readInput(path)
	if err != nil {
		return Report{}, err
	}
	return ValidateBytes(path, data, f, against, opts...), nil
}

// ValidateFamily checks that child impact models resolve against their parent:
// each names the parent, and every outcome reference into the parent
// exists there. Only impact model files can be children.
func ValidateFamily(parentPath string, childPaths ...string) (Report, error) {
	parent, err := impactio.LoadDocument(parentPath)
	if err != nil {
		return Report{}, err
	}
	children := make([]*impact.Model, 0, len(childPaths))
	for _, p := range childPaths {
		child, err := impactio.LoadDocument(p)
		if err != nil {
			return Report{}, err
		}
		children = append(children, child)
	}
	return Report{
		Source: parentPath,
		Kind:   impact.KindImpactModel,
		Issues: append([]validate.Issue{}, validate.Family(parent, children...)...),
	}, nil
}

// ValidateFutures checks impact models that describe competing futures of one
// market against one another (see validate.Futures). Each must also be valid
// alone; use ValidateFile with the shared actor registry for that.
func ValidateFutures(paths ...string) (Report, error) {
	docs := make([]*impact.Model, 0, len(paths))
	for _, p := range paths {
		d, err := impactio.LoadDocument(p)
		if err != nil {
			return Report{}, err
		}
		docs = append(docs, d)
	}
	return Report{Source: "competing futures", Kind: impact.KindImpactModel, Issues: append([]validate.Issue{}, validate.Futures(docs...)...)}, nil
}

// LoadActorRegistry reads an actor registry file.
func LoadActorRegistry(path string) (*impact.ActorRegistry, error) {
	data, f, err := readInput(path)
	if err != nil {
		return nil, err
	}
	return impactio.ParseActorRegistry(data, f)
}

func readInput(path string) ([]byte, impactio.Format, error) {
	f, err := impactio.FormatFromPath(path)
	if err != nil {
		return nil, "", err
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, "", err
	}
	if info.Size() > impactio.MaxFileBytes {
		return nil, "", fmt.Errorf("%w: %s is %d bytes (limit %d)", impactio.ErrTooLarge, path, info.Size(), impactio.MaxFileBytes)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	return data, f, nil
}
