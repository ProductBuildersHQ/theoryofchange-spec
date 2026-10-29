// Package schema embeds the JSON Schemas generated from pkg/toc and
// validates documents against them.
//
// Never edit the *.schema.json files by hand; regenerate them with:
//
//	go generate ./schema/...
package schema

//go:generate go run ../cmd/generate-schema -out .

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"slices"
	"strings"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

//go:embed *.schema.json
var files embed.FS

const suffix = ".schema.json"

// Schema names.
const (
	TheoryOfChange = "theory-of-change"
	ObservationSet = "observation-set"
	Status         = "theory-of-change-status"
	StatusDiff     = "status-diff"
	ActorRegistry  = "actor-registry"
	ReviewResponse = "review-response"
)

// ErrUnknownSchema is returned for a name that is not embedded.
var ErrUnknownSchema = errors.New("unknown schema")

// Names returns the embedded schema names, sorted.
func Names() []string {
	entries, err := fs.Glob(files, "*"+suffix)
	if err != nil {
		// The pattern is a constant; a failure is a programming error.
		panic(err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, strings.TrimSuffix(e, suffix))
	}
	slices.Sort(names)
	return names
}

// Get returns the raw JSON of the named schema.
func Get(name string) ([]byte, error) {
	data, err := files.ReadFile(name + suffix)
	if err != nil {
		return nil, fmt.Errorf("%w: %q", ErrUnknownSchema, name)
	}
	return data, nil
}

var (
	compileMu sync.Mutex
	compiled  = map[string]*jsonschema.Schema{}
)

// Compiled returns the compiled named schema. Compilation is cached.
func Compiled(name string) (*jsonschema.Schema, error) {
	compileMu.Lock()
	defer compileMu.Unlock()
	if s, ok := compiled[name]; ok {
		return s, nil
	}
	data, err := Get(name)
	if err != nil {
		return nil, err
	}
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("parse schema %q: %w", name, err)
	}
	url := "file:///" + name + suffix
	c := jsonschema.NewCompiler()
	c.AssertFormat()
	if err := c.AddResource(url, doc); err != nil {
		return nil, fmt.Errorf("add schema %q: %w", name, err)
	}
	s, err := c.Compile(url)
	if err != nil {
		return nil, fmt.Errorf("compile schema %q: %w", name, err)
	}
	compiled[name] = s
	return s, nil
}

// Validate validates a JSON document against the named schema. A failed
// validation returns a *jsonschema.ValidationError, whose BasicOutput lists
// each error with its JSON-pointer instance location.
func Validate(name string, data []byte) error {
	s, err := Compiled(name)
	if err != nil {
		return err
	}
	inst, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("parse document: %w", err)
	}
	return s.Validate(inst)
}
