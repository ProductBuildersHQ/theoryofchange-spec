package report

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"text/template"
)

// Limits on custom templates. A template is code supplied by a user, so it is
// bounded: it cannot be huge, and it cannot produce unbounded output.
const (
	MaxTemplateBytes = 64 << 10
	MaxOutputBytes   = 8 << 20
)

// ErrTemplate is returned for a template that cannot be used.
var ErrTemplate = errors.New("custom report template")

// templateFuncs are the only functions a custom template may call. None reads
// a file, the environment, or the network; they only format text.
func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"label": Label,
		"cell":  MDCell,
		"text":  mdText,
		"lower": strings.ToLower,
		"upper": strings.ToUpper,
		"join":  strings.Join,
		"dash":  orDash,
		"json": func(v any) (string, error) {
			b, err := json.MarshalIndent(v, "", "  ")
			return string(b), err
		},
	}
}

// RenderTemplate renders a user's text/template over the view. The template
// sees the View exactly as reports do, so anything a report withholds,
// such as the excerpt, summary, and URL of restricted evidence, is already
// absent from the data. It can call only the functions in templateFuncs. A
// reference to a field the view does not have is an error, not an empty
// string.
func RenderTemplate(tmpl string, v *View) (string, error) {
	if len(tmpl) > MaxTemplateBytes {
		return "", fmt.Errorf("%w: the template is %d bytes; the limit is %d", ErrTemplate, len(tmpl), MaxTemplateBytes)
	}
	t, err := template.New("report").Funcs(templateFuncs()).Option("missingkey=error").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrTemplate, err)
	}
	w := &limitedBuffer{limit: MaxOutputBytes}
	if err := t.Execute(w, v); err != nil {
		return "", fmt.Errorf("%w: %w", ErrTemplate, err)
	}
	return w.buf.String(), nil
}

// limitedBuffer fails a write that would take the output past the limit.
type limitedBuffer struct {
	buf   bytes.Buffer
	limit int
}

var errOutputTooLarge = errors.New("the output is too large")

func (l *limitedBuffer) Write(p []byte) (int, error) {
	if l.buf.Len()+len(p) > l.limit {
		return 0, fmt.Errorf("%w: more than %d bytes", errOutputTooLarge, l.limit)
	}
	return l.buf.Write(p)
}
