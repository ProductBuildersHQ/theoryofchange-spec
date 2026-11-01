package schemagen

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
	"github.com/ProductBuildersHQ/impact-model/schema"
)

// TestEmbeddedSchemasAreCurrent fails when schema/*.schema.json no longer
// matches the Go types. Fix by running: go generate ./schema/...
func TestEmbeddedSchemasAreCurrent(t *testing.T) {
	files, err := Generate()
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if len(files) != len(schema.Names()) {
		t.Fatalf("generated %d schemas, embedded %d; run go generate ./schema/...", len(files), len(schema.Names()))
	}
	for _, f := range files {
		name := f.Name[:len(f.Name)-len(".schema.json")]
		embedded, err := schema.Get(name)
		if err != nil {
			t.Fatalf("embedded schema %s missing: %v; run go generate ./schema/...", name, err)
		}
		if !bytes.Equal(embedded, f.Data) {
			t.Errorf("schema %s is stale; run go generate ./schema/...", name)
		}
	}
}

func TestGenerateIsDeterministic(t *testing.T) {
	a, err := Generate()
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	b, err := Generate()
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Fatal("Generate output differs between runs")
	}
}

func TestSchemaHeaders(t *testing.T) {
	for _, d := range Definitions() {
		data, err := GenerateOne(d)
		if err != nil {
			t.Fatalf("GenerateOne(%s): %v", d.Name, err)
		}
		var head struct {
			ID    string `json:"$id"`
			Title string `json:"title"`
		}
		if err := json.Unmarshal(data, &head); err != nil {
			t.Fatalf("unmarshal %s: %v", d.Name, err)
		}
		if head.ID != ID(d.Name) || head.Title != d.Title {
			t.Errorf("%s header = %+v", d.Name, head)
		}
		if data[len(data)-1] != '\n' {
			t.Errorf("%s lacks a trailing newline", d.Name)
		}
	}
}

// TestEveryEnumTypeIsMapped walks every type reachable from the published
// documents and requires each named string type from pkg/impact (other than
// the per-document Kind) to have enum values.
func TestEveryEnumTypeIsMapped(t *testing.T) {
	enums := EnumValues()
	tocPkg := reflect.TypeFor[impact.Model]().PkgPath()
	kindType := reflect.TypeFor[impact.Kind]()
	seen := map[reflect.Type]bool{}
	var walk func(reflect.Type)
	walk = func(rt reflect.Type) {
		if seen[rt] {
			return
		}
		seen[rt] = true
		switch rt.Kind() {
		case reflect.Pointer, reflect.Slice, reflect.Array, reflect.Map:
			walk(rt.Elem())
		case reflect.Struct:
			for i := range rt.NumField() {
				walk(rt.Field(i).Type)
			}
		case reflect.String:
			if rt.PkgPath() == tocPkg && rt.Name() != "" && rt != kindType {
				if _, ok := enums[rt]; !ok {
					t.Errorf("enum type toc.%s has no entry in EnumValues", rt.Name())
				}
			}
		}
	}
	for _, d := range Definitions() {
		walk(reflect.TypeOf(d.Value))
	}
	for rt, values := range enums {
		if len(values) == 0 {
			t.Errorf("enum type %s has no values", rt)
		}
	}
}
