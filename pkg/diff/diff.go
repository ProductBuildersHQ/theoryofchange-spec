// Package diff compares two versions of an impact model by element. A
// change is keyed by the element's ID, so renaming a title is one changed
// element rather than a removed line and an added one, and a rejected element
// shows as the change that rejected it. The comparison is of meaning, not
// text: field order and formatting do not matter.
package diff

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/ProductBuildersHQ/impact-model/pkg/edit"
	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
)

// Op says what happened to an element.
type Op string

// Operations.
const (
	Added   Op = "added"
	Removed Op = "removed"
	Changed Op = "changed"
)

// FieldChange is one changed value, addressed by a dotted path such as
// "provenance.basis".
type FieldChange struct {
	Path string `json:"path"`
	From string `json:"from,omitempty"`
	To   string `json:"to,omitempty"`
}

// Change is what happened to one element.
type Change struct {
	ElementID string `json:"elementId"`
	Kind      string `json:"kind"`
	Title     string `json:"title,omitempty"`
	Op        Op     `json:"op"`
	// Rejected is true when the element is rejected in the newer version;
	// a rejection is a decision worth seeing apart from other edits.
	Rejected bool          `json:"rejected,omitempty"`
	Fields   []FieldChange `json:"fields,omitempty"`
}

// Result is the difference between two versions.
type Result struct {
	OldVersion string `json:"oldVersion,omitempty"`
	NewVersion string `json:"newVersion,omitempty"`
	// Document lists changes to the impact model itself: its context, ceiling,
	// perspective, metadata, and other fields that are not elements.
	Document []FieldChange `json:"document"`
	Changes  []Change      `json:"changes"`
}

// Empty reports whether the two versions are the same.
func (r *Result) Empty() bool { return len(r.Document) == 0 && len(r.Changes) == 0 }

// skipped are top-level fields that are not part of the impact model's meaning: the
// schema reference, and the revision log, which records changes rather than
// being one.
var skipped = map[string]bool{"$schema": true, "revisions": true}

// Compare returns what changed from old to cur. Both must be valid impact models
// of the same document ID; comparing different impact models is allowed but every
// element will differ.
func Compare(old, cur *impact.Model) (*Result, error) {
	a, err := tree(old)
	if err != nil {
		return nil, err
	}
	b, err := tree(cur)
	if err != nil {
		return nil, err
	}
	res := &Result{OldVersion: old.Metadata.Version, NewVersion: cur.Metadata.Version, Document: []FieldChange{}, Changes: []Change{}}

	arrays := map[string]string{} // field -> kind name
	var order []string
	for _, k := range edit.ElementKinds() {
		arrays[k.Field] = k.Name
		order = append(order, k.Field)
	}

	// Everything that is not an element array is part of the impact model itself.
	rest := func(m map[string]any) map[string]any {
		out := map[string]any{}
		for k, v := range m {
			if _, isElements := arrays[k]; !isElements && !skipped[k] {
				out[k] = v
			}
		}
		return out
	}
	res.Document = compareValues("", rest(a), rest(b))

	for _, field := range order {
		res.Changes = append(res.Changes, compareElements(arrays[field], asList(a[field]), asList(b[field]))...)
	}
	return res, nil
}

func tree(d *impact.Model) (map[string]any, error) {
	data, err := json.Marshal(d)
	if err != nil {
		return nil, fmt.Errorf("diff: encode document: %w", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("diff: decode document: %w", err)
	}
	return m, nil
}

func asList(v any) []map[string]any {
	items, _ := v.([]any)
	out := make([]map[string]any, 0, len(items))
	for _, it := range items {
		if m, ok := it.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}

func idOf(m map[string]any) string {
	id, _ := m["id"].(string)
	return id
}

// compareElements reports added and changed elements in the newer order,
// then removed ones in the older order.
func compareElements(kind string, old, cur []map[string]any) []Change {
	byID := map[string]map[string]any{}
	for _, m := range old {
		byID[idOf(m)] = m
	}
	seen := map[string]bool{}
	var out []Change
	for _, m := range cur {
		id := idOf(m)
		seen[id] = true
		prev, existed := byID[id]
		switch {
		case !existed:
			out = append(out, Change{ElementID: id, Kind: kind, Title: titleOf(m), Op: Added, Rejected: rejected(m)})
		default:
			if fields := compareValues("", prev, m); len(fields) > 0 {
				out = append(out, Change{ElementID: id, Kind: kind, Title: titleOf(m), Op: Changed, Rejected: rejected(m) && !rejected(prev), Fields: fields})
			}
		}
	}
	for _, m := range old {
		if id := idOf(m); !seen[id] {
			out = append(out, Change{ElementID: id, Kind: kind, Title: titleOf(m), Op: Removed})
		}
	}
	return out
}

func rejected(m map[string]any) bool {
	p, _ := m["provenance"].(map[string]any)
	return p["basis"] == string(impact.BasisRejected)
}

// titleOf finds the words that name an element.
func titleOf(m map[string]any) string {
	for _, k := range []string{"title", "name", "statement", "question", "term", "label", "description", "id"} {
		if s, ok := m[k].(string); ok && s != "" {
			return strings.Join(strings.Fields(s), " ")
		}
	}
	return ""
}

// compareValues compares two JSON values and lists the differences by dotted
// path, keys in sorted order. Objects are compared field by field; lists and
// scalars as wholes.
func compareValues(path string, a, b any) []FieldChange {
	am, aok := a.(map[string]any)
	bm, bok := b.(map[string]any)
	if aok && bok {
		keys := map[string]bool{}
		for k := range am {
			keys[k] = true
		}
		for k := range bm {
			keys[k] = true
		}
		sorted := make([]string, 0, len(keys))
		for k := range keys {
			sorted = append(sorted, k)
		}
		sort.Strings(sorted)
		var out []FieldChange
		for _, k := range sorted {
			child := k
			if path != "" {
				child = path + "." + k
			}
			out = append(out, compareValues(child, am[k], bm[k])...)
		}
		return out
	}
	if equal(a, b) {
		return nil
	}
	return []FieldChange{{Path: path, From: show(a), To: show(b)}}
}

func equal(a, b any) bool {
	ja, err1 := json.Marshal(a)
	jb, err2 := json.Marshal(b)
	return err1 == nil && err2 == nil && string(ja) == string(jb)
}

const maxShown = 240

// show renders a value for a person: text as it is, absence as empty, and
// anything structured as compact JSON, cut short.
func show(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return truncate(x)
	}
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return truncate(string(b))
}

func truncate(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if r := []rune(s); len(r) > maxShown {
		return string(r[:maxShown-1]) + "…"
	}
	return s
}
