package validate

import (
	"errors"
	"fmt"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
	"github.com/ProductBuildersHQ/impact-model/pkg/impactio"
	"github.com/ProductBuildersHQ/impact-model/schema"
)

// Schema validates raw input against the schema for its declared kind and
// returns one issue per violation, each with the JSON pointer of the
// offending value. Input that cannot be parsed, or whose kind is missing or
// unknown, yields a single issue. A nil result means the input conforms.
func Schema(data []byte, f impactio.Format) []Issue {
	j, err := impactio.ToJSON(data, f)
	if err != nil {
		return []Issue{{Code: "VAL-SCH-002", Severity: SeverityError, Path: "", Message: err.Error()}}
	}
	kind, err := impactio.DetectKind(data, f)
	if err != nil {
		return []Issue{{Code: "VAL-SCH-003", Severity: SeverityError, Path: "/kind", Message: err.Error()}}
	}
	name, err := impactio.SchemaName(kind)
	if err != nil {
		return []Issue{{Code: "VAL-SCH-003", Severity: SeverityError, Path: "/kind", Message: err.Error()}}
	}
	return schemaIssues(name, j)
}

func schemaIssues(name string, j []byte) []Issue {
	err := schema.Validate(name, j)
	if err == nil {
		return nil
	}
	var ve *jsonschema.ValidationError
	if !errors.As(err, &ve) {
		return []Issue{{Code: "VAL-SCH-002", Severity: SeverityError, Message: err.Error()}}
	}
	printer := message.NewPrinter(language.English)
	var out []Issue
	seen := map[string]bool{}
	var walk func(e *jsonschema.ValidationError)
	walk = func(e *jsonschema.ValidationError) {
		if len(e.Causes) > 0 {
			for _, c := range e.Causes {
				walk(c)
			}
			return
		}
		path := pointer(e.InstanceLocation...)
		msg := e.ErrorKind.LocalizedString(printer)
		key := path + "\x00" + msg
		if seen[key] {
			return
		}
		seen[key] = true
		out = append(out, Issue{Code: "VAL-SCH-001", Severity: SeverityError, Path: path, Message: msg})
	}
	walk(ve)
	return out
}

// Source validates raw input of any document kind and, for a impact model of
// Change that conforms to its schema, also runs the structural checks of
// Document. The returned document is nil unless the input passed the schema.
func Source(data []byte, f impactio.Format, opts ...Option) (*impact.Model, []Issue) {
	issues := Schema(data, f)
	if len(issues) > 0 {
		return nil, issues
	}
	kind, err := impactio.DetectKind(data, f)
	if err != nil || kind != impact.KindImpactModel {
		return nil, nil
	}
	doc, err := impactio.ParseDocument(data, f)
	if err != nil {
		return nil, []Issue{{Code: "VAL-SCH-002", Severity: SeverityError, Message: fmt.Sprint(err)}}
	}
	return doc, Document(doc, opts...)
}
