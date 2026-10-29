package schema

import (
	"errors"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func TestNamesAndGet(t *testing.T) {
	want := []string{ActorRegistry, ObservationSet, StatusDiff, TheoryOfChange, Status}
	slices.Sort(want)
	if got := Names(); !slices.Equal(got, want) {
		t.Fatalf("Names() = %v, want %v", got, want)
	}
	for _, n := range Names() {
		if _, err := Get(n); err != nil {
			t.Fatalf("Get(%q): %v", n, err)
		}
		if _, err := Compiled(n); err != nil {
			t.Fatalf("Compiled(%q): %v", n, err)
		}
	}
	if _, err := Get("nope"); !errors.Is(err, ErrUnknownSchema) {
		t.Fatalf("Get(nope) error = %v", err)
	}
	if err := Validate("nope", []byte(`{}`)); !errors.Is(err, ErrUnknownSchema) {
		t.Fatalf("Validate(nope) error = %v", err)
	}
}

func TestFixturesValidate(t *testing.T) {
	tests := []struct{ schema, file string }{
		{TheoryOfChange, "minimal.json"},
		{TheoryOfChange, "full.json"},
		{ObservationSet, "observations.json"},
		{Status, "status.json"},
		{StatusDiff, "diff.json"},
		{ActorRegistry, "registry.json"},
	}
	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			data, err := os.ReadFile("../pkg/toc/testdata/" + tt.file)
			if err != nil {
				t.Fatalf("read: %v", err)
			}
			if err := Validate(tt.schema, data); err != nil {
				t.Fatalf("Validate(%s, %s): %v", tt.schema, tt.file, err)
			}
		})
	}
}

const minimal = `{
  "kind": "TheoryOfChange", "specVersion": "0.1", "id": "m",
  "metadata": {}, "context": {"problemStatement": "p"},
  "outcomes": [{"id": "oc-1", "title": "t", "level": "impact"}]
}`

func TestInvalidDocuments(t *testing.T) {
	if err := Validate(TheoryOfChange, []byte(minimal)); err != nil {
		t.Fatalf("baseline minimal document invalid: %v", err)
	}
	tests := map[string]struct {
		old, new string
		want     string
	}{
		"missing outcomes": {`"outcomes": [{"id": "oc-1", "title": "t", "level": "impact"}]`, `"outcomes": []`, "/outcomes"},
		"bad level":        {`"level": "impact"`, `"level": "huge"`, "/outcomes/0/level"},
		"unknown property": {`"id": "m",`, `"id": "m", "bogus": 1,`, "bogus"},
		"wrong kind":       {`"kind": "TheoryOfChange"`, `"kind": "ObservationSet"`, "/kind"},
		"missing problem":  {`{"problemStatement": "p"}`, `{}`, "/context"},
		"bad date":         {`"context": {"problemStatement": "p"}`, `"context": {"problemStatement": "p", "timeHorizon": {"start": "soon", "end": "2027-01-01"}}`, "/context/timeHorizon/start"},
		"missing title":    {`"title": "t", `, ``, "/outcomes/0"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			doc := strings.Replace(minimal, tt.old, tt.new, 1)
			if doc == minimal {
				t.Fatalf("test edit %q did not apply", tt.old)
			}
			err := Validate(TheoryOfChange, []byte(doc))
			var verr *jsonschema.ValidationError
			if !errors.As(err, &verr) {
				t.Fatalf("expected a validation error, got %v", err)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error does not mention %q:\n%v", tt.want, err)
			}
		})
	}
}

func TestValidateRejectsMalformedJSON(t *testing.T) {
	err := Validate(TheoryOfChange, []byte(`{"kind":`))
	if err == nil {
		t.Fatal("expected a parse error")
	}
	var verr *jsonschema.ValidationError
	if errors.As(err, &verr) {
		t.Fatal("malformed JSON should be a parse error, not a validation error")
	}
}
