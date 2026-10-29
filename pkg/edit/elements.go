package edit

import (
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

// ElementKind is a kind of identified element that can be added, changed,
// and removed.
type ElementKind struct {
	// Name is the command-line name, e.g. "outcome".
	Name string
	// Field is the document's array field, e.g. "outcomes".
	Field string
}

var elementKinds = []ElementKind{
	{"outcome", "outcomes"},
	{"group", "preconditionGroups"},
	{"link", "links"},
	{"intervention", "interventions"},
	{"assumption", "assumptions"},
	{"indicator", "indicators"},
	{"actor", "actors"},
	{"evidence", "evidence"},
	{"factor", "externalFactors"},
	{"risk", "risks"},
	{"decision", "decisions"},
	{"pathway", "pathways"},
	{"metric", "metrics"},
	{"term", "glossary"},
	{"cause", "causes"},
	{"target-group", "targetGroups"},
	{"audience", "audiences"},
	{"theme", "themes"},
	{"criterion", "selectionCriteria"},
	{"consultation", "consultations"},
}

// ElementKinds returns the editable element kinds in document order.
func ElementKinds() []ElementKind { return slices.Clone(elementKinds) }

// KindByName returns the element kind with the given name.
func KindByName(name string) (ElementKind, bool) {
	for _, k := range elementKinds {
		if k.Name == name {
			return k, true
		}
	}
	return ElementKind{}, false
}

// ParseValue interprets a command-line value. Text that is a JSON number,
// boolean, null, array, or object is used as that JSON value; anything else
// (including a quoted JSON string) is a string. Quote a value to force a
// string: '"2026"'.
func ParseValue(s string) any {
	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return s
	}
	return v
}

// ExpandDotted turns keys like "logic.type" into nested objects:
// {"logic.type": "all"} becomes {"logic": {"type": "all"}}.
func ExpandDotted(fields map[string]any) (map[string]any, error) {
	out := map[string]any{}
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		parts := strings.Split(key, ".")
		cur := out
		for i, p := range parts {
			if p == "" {
				return nil, fmt.Errorf("%w: empty segment in field %q", ErrPatch, key)
			}
			if i == len(parts)-1 {
				if _, exists := cur[p]; exists {
					return nil, fmt.Errorf("%w: field %q is set more than once", ErrPatch, key)
				}
				cur[p] = fields[key]
				break
			}
			next, ok := cur[p].(map[string]any)
			if !ok {
				if _, exists := cur[p]; exists {
					return nil, fmt.Errorf("%w: field %q conflicts with another field", ErrPatch, key)
				}
				next = map[string]any{}
				cur[p] = next
			}
			cur = next
		}
	}
	return out, nil
}

// AddElement returns a patch that appends a new element to doc. fields must
// include "id". Dotted keys become nested objects. When the document has no
// list of this kind yet, the patch creates it, since an append needs an
// existing array.
func AddElement(doc *toc.Document, k ElementKind, fields map[string]any) (Patch, error) {
	expanded, err := ExpandDotted(fields)
	if err != nil {
		return nil, err
	}
	if id, _ := expanded["id"].(string); id == "" {
		return nil, fmt.Errorf("%w: a new %s needs an id", ErrPatch, k.Name)
	}
	tree, err := treeOf(doc)
	if err != nil {
		return nil, err
	}
	var op Op
	if existing, _ := tree.(map[string]any)[k.Field].([]any); len(existing) > 0 {
		op, err = Add(Pointer(k.Field, "-"), expanded)
	} else {
		op, err = Add(Pointer(k.Field), []any{expanded})
	}
	if err != nil {
		return nil, err
	}
	return Patch{op}, nil
}

// SetFields returns a patch that sets and removes fields of an existing
// element of doc. Keys may be dotted ("provenance.basis") to reach nested
// objects, which are created when the element does not have them yet. An
// element's id cannot be changed.
func SetFields(doc *toc.Document, k ElementKind, id string, set map[string]any, unset []string) (Patch, error) {
	if _, ok := set["id"]; ok {
		return nil, fmt.Errorf("%w: an element's id cannot be changed; remove it and add a new one", ErrPatch)
	}
	tree, err := treeOf(doc)
	if err != nil {
		return nil, err
	}
	root, _ := tree.(map[string]any)
	idx, found := indexOfID(root, k.Field, id)
	if !found {
		return nil, fmt.Errorf("%w: no %s with id %q", ErrUnknownID, k.Name, id)
	}
	element := root[k.Field].([]any)[idx].(map[string]any)

	expanded, err := ExpandDotted(set)
	if err != nil {
		return nil, err
	}
	var p Patch
	if err := setOps(&p, []string{k.Field, "@" + id}, element, expanded); err != nil {
		return nil, err
	}
	for _, key := range unset {
		if key == "id" {
			return nil, fmt.Errorf("%w: an element's id cannot be removed", ErrPatch)
		}
		p = append(p, Remove(fieldPointer(k, id, key)))
	}
	if len(p) == 0 {
		return nil, fmt.Errorf("%w: nothing to change", ErrPatch)
	}
	return p, nil
}

func fieldPointer(k ElementKind, id, dotted string) string {
	segs := append([]string{k.Field, "@" + id}, strings.Split(dotted, ".")...)
	return Pointer(segs...)
}

// ---- removal ---------------------------------------------------------------

// action says what to do when a reference to a removed element is found.
type action int

const (
	dropValue   action = iota // remove the scalar field
	dropItem                  // remove the entry from the array field
	dropElement               // remove the whole element that holds the reference
	dropMetric                // remove an indicator's metricId and the roll-up that depends on it
)

// refSite describes a field that references other elements.
type refSite struct {
	array  string   // top-level array holding the elements
	field  []string // path to the reference within an element
	action action
	min    int // for dropItem: fewer remaining items drops the element
}

// refSites lists every field that references an element ID. A test removes
// every element of the full fixture with cascade, so a missing entry makes
// that test fail rather than leaving a dangling reference unnoticed.
// Revision changedIds are deliberately absent: they record history, and a
// reference to a later-removed element is only a warning.
var refSites = []refSite{
	{"outcomes", []string{"actorIds"}, dropItem, 0},
	{"outcomes", []string{"audienceIds"}, dropItem, 0},
	{"outcomes", []string{"deliveredBy"}, dropItem, 0},
	{"outcomes", []string{"addressesCauses"}, dropItem, 0},
	{"outcomes", []string{"targetGroupIds"}, dropItem, 0},
	{"indicators", []string{"targetGroupIds"}, dropItem, 0},
	{"causes", []string{"parent"}, dropValue, 0},
	{"causes", []string{"affectedGroupIds"}, dropItem, 0},
	{"causes", []string{"evidenceIds"}, dropItem, 0},
	{"targetGroups", []string{"actorIds"}, dropItem, 0},
	{"indicators", []string{"audienceIds"}, dropItem, 0},
	{"indicators", []string{"themeIds"}, dropItem, 1},
	{"themes", []string{"outcomeIds"}, dropItem, 0},
	{"preconditionGroups", []string{"target"}, dropElement, 0},
	{"preconditionGroups", []string{"members"}, dropItem, 2},
	{"links", []string{"from"}, dropElement, 0},
	{"links", []string{"to"}, dropElement, 0},
	{"links", []string{"assumptionIds"}, dropItem, 0},
	{"links", []string{"evidenceIds"}, dropItem, 0},
	{"links", []string{"interventionIds"}, dropItem, 0},
	{"interventions", []string{"targets"}, dropItem, 1},
	{"interventions", []string{"ownerActorId"}, dropValue, 0},
	{"assumptions", []string{"appliesTo"}, dropItem, 0},
	{"assumptions", []string{"indicatorIds"}, dropItem, 0},
	{"assumptions", []string{"evidenceIds"}, dropItem, 0},
	{"assumptions", []string{"ownerActorId"}, dropValue, 0},
	{"indicators", []string{"measures"}, dropElement, 0},
	{"indicators", []string{"ownerActorId"}, dropValue, 0},
	{"indicators", []string{"proxyFor"}, dropItem, 0},
	{"indicators", []string{"metricId"}, dropMetric, 0},
	{"indicators", []string{"pathwayIds"}, dropItem, 0},
	{"indicators", []string{"rollupFrom"}, dropItem, 0},
	{"metrics", []string{"weightMetricId"}, dropElement, 0},
	{"glossary", []string{"relatedIds"}, dropItem, 0},
	{"indicators", []string{"population", "actorIds"}, dropItem, 0},
	{"externalFactors", []string{"affects"}, dropItem, 0},
	{"risks", []string{"appliesTo"}, dropItem, 0},
	{"risks", []string{"relatedAssumptionIds"}, dropItem, 0},
	{"pathways", []string{"outcomeIds"}, dropItem, 0},
	{"pathways", []string{"interventionIds"}, dropItem, 0},
	{"pathways", []string{"ownerActorId"}, dropValue, 0},
	{"alignments", []string{"elementId"}, dropElement, 0},
	{"decisions", []string{"ownerActorId"}, dropValue, 0},
	{"decisions", []string{"relatedIds"}, dropItem, 0},
	{"revisions", []string{"triggerEvidenceIds"}, dropItem, 0},
}

func init() {
	// Every element that carries provenance can cite evidence.
	for _, array := range []string{
		"outcomes", "preconditionGroups", "links", "interventions", "assumptions",
		"indicators", "actors", "externalFactors", "risks", "metrics", "pathways",
	} {
		refSites = append(refSites, refSite{array, []string{"provenance", "evidenceIds"}, dropItem, 0})
	}
}

// RemoveElement returns a patch that removes an element. Without cascade it
// is a single remove; Apply then rejects it if anything still references the
// element, and reports those references. With cascade, the patch also
// removes or rewrites every reference, and removes any element that cannot
// survive losing one (a link whose end is gone, an indicator whose subject
// is gone, a group left with fewer than two members), repeating for those.
func RemoveElement(doc *toc.Document, k ElementKind, id string, cascade bool) (Patch, error) {
	tree, err := treeOf(doc)
	if err != nil {
		return nil, err
	}
	root, ok := tree.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%w: document is not an object", ErrPatch)
	}
	if _, found := indexOfID(root, k.Field, id); !found {
		return nil, fmt.Errorf("%w: no %s with id %q", ErrUnknownID, k.Name, id)
	}
	first := Remove(Pointer(k.Field, "@"+id))
	if !cascade {
		return Patch{first}, nil
	}

	patch := Patch{first}
	tree, err = applyOp(tree, first)
	if err != nil {
		return nil, err
	}
	queue := []string{id}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for {
			site, ok := findReference(tree.(map[string]any), cur)
			if !ok {
				break
			}
			ops, dropped, err := site.remedy()
			if err != nil {
				return nil, err
			}
			for _, op := range ops {
				tree, err = applyOp(tree, op)
				if err != nil {
					return nil, fmt.Errorf("cascade removal of %q: %w", cur, err)
				}
				patch = append(patch, op)
			}
			if dropped != "" {
				queue = append(queue, dropped)
			}
			if len(patch) > 100000 {
				return nil, fmt.Errorf("%w: cascade did not terminate", ErrPatch)
			}
		}
	}
	return patch, nil
}

func treeOf(doc *toc.Document) (any, error) {
	data, err := json.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("encode document: %w", err)
	}
	return decodeTree(data)
}

func indexOfID(root map[string]any, array, id string) (int, bool) {
	items, _ := root[array].([]any)
	for i, it := range items {
		if m, ok := it.(map[string]any); ok && m["id"] == id {
			return i, true
		}
	}
	return 0, false
}

// found is a located reference and how to remove it.
type found struct {
	site    refSite
	element map[string]any
	index   int // element index within its array (-1 for singletons)
	item    int // index within the reference array (dropItem), else -1
	holder  map[string]any
	left    int // items that would remain (dropItem)
}

// findReference returns the first reference to id in a fixed scan order.
func findReference(root map[string]any, id string) (found, bool) {
	// Singletons first: the perspective and the accountability ceiling.
	if p, ok := root["perspective"].(map[string]any); ok && p["actorId"] == id {
		return found{site: refSite{array: "perspective", action: dropElement}, element: p, index: -1, item: -1}, true
	}
	if c, ok := root["ceilingOfAccountability"].(map[string]any); ok {
		if ids, _ := c["outcomeIds"].([]any); len(ids) > 0 {
			for i, v := range ids {
				if v == id {
					return found{
						site:    refSite{array: "ceilingOfAccountability", field: []string{"outcomeIds"}, action: dropItem},
						element: c, index: -1, item: i, holder: c, left: len(ids) - 1,
					}, true
				}
			}
		}
	}
	if f, ok := findScore(root, id); ok {
		return f, true
	}
	for _, site := range refSites {
		items, _ := root[site.array].([]any)
		for ei, raw := range items {
			el, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			holder, last, ok := walkTo(el, site.field)
			if !ok {
				continue
			}
			switch v := holder[last].(type) {
			case string:
				if v == id {
					return found{site: site, element: el, index: ei, item: -1, holder: holder}, true
				}
			case []any:
				for i, x := range v {
					if x == id {
						return found{site: site, element: el, index: ei, item: i, holder: holder, left: len(v) - 1}, true
					}
				}
			}
		}
	}
	return found{}, false
}

// scoreSites are the places criterion scores live: an indicator's
// selection, and the priority of an outcome or an intervention.
var scoreSites = []refSite{
	{array: "indicators", field: []string{"selection", "scores"}, action: dropItem},
	{array: "outcomes", field: []string{"priority", "scores"}, action: dropItem},
	{array: "interventions", field: []string{"priority", "scores"}, action: dropItem},
}

// findScore finds a score that names the criterion id. Scores are objects
// inside their subject, so they are scanned apart from the scalar reference
// fields.
func findScore(root map[string]any, id string) (found, bool) {
	for _, site := range scoreSites {
		items, _ := root[site.array].([]any)
		for ei, raw := range items {
			el, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			holder, last, ok := walkTo(el, site.field)
			if !ok {
				continue
			}
			scores, _ := holder[last].([]any)
			for i, sc := range scores {
				if m, ok := sc.(map[string]any); ok && m["criterionId"] == id {
					return found{site: site, element: el, index: ei, item: i, holder: holder, left: len(scores) - 1}, true
				}
			}
		}
	}
	return found{}, false
}

// walkTo follows all but the last element of path and returns the object
// holding the last key.
func walkTo(el map[string]any, path []string) (map[string]any, string, bool) {
	cur := el
	for _, p := range path[:len(path)-1] {
		next, ok := cur[p].(map[string]any)
		if !ok {
			return nil, "", false
		}
		cur = next
	}
	return cur, path[len(path)-1], true
}

// remedy returns the operations that remove this reference, and the ID of
// any element they remove (so its own references get cleaned up).
func (f found) remedy() ([]Op, string, error) {
	one := func(op Op) ([]Op, string, error) { return []Op{op}, "", nil }
	switch f.site.array {
	case "perspective":
		return one(Remove(Pointer("perspective")))
	case "ceilingOfAccountability":
		if f.left == 0 && f.holder["level"] == nil {
			return one(Remove(Pointer("ceilingOfAccountability")))
		}
		return one(f.removeItem())
	}
	switch f.site.action {
	case dropElement:
		return f.removeElement()
	case dropValue:
		return one(Remove(f.fieldPointer()))
	case dropMetric:
		ops := []Op{Remove(f.fieldPointer())}
		if _, has := f.element["rollupFrom"]; has {
			ops = append(ops, Remove(f.elementPointer()+Pointer("rollupFrom")))
		}
		return ops, "", nil
	default: // dropItem
		if f.left < f.site.min {
			return f.removeElement()
		}
		if f.left == 0 {
			return one(Remove(f.fieldPointer()))
		}
		return one(f.removeItem())
	}
}

func (f found) elementPointer() string {
	if id, ok := f.element["id"].(string); ok && id != "" {
		return Pointer(f.site.array, "@"+id)
	}
	return Pointer(f.site.array, strconv.Itoa(f.index))
}

func (f found) fieldPointer() string {
	return f.elementPointer() + Pointer(f.site.field...)
}

func (f found) removeItem() Op {
	base := f.fieldPointer()
	if f.site.array == "ceilingOfAccountability" {
		base = Pointer(f.site.array) + Pointer(f.site.field...)
	}
	return Remove(base + "/" + strconv.Itoa(f.item))
}

func (f found) removeElement() ([]Op, string, error) {
	id, _ := f.element["id"].(string)
	return []Op{Remove(f.elementPointer())}, id, nil
}

// setOps appends operations that set each key of values on the object at
// base, descending into nested objects that exist and adding whole nested
// objects where they do not.
func setOps(p *Patch, base []string, existing map[string]any, values map[string]any) error {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		path := append(slices.Clone(base), key)
		if nested, ok := values[key].(map[string]any); ok {
			if current, ok := existing[key].(map[string]any); ok {
				if err := setOps(p, path, current, nested); err != nil {
					return err
				}
				continue
			}
		}
		op, err := Add(Pointer(path...), values[key])
		if err != nil {
			return err
		}
		*p = append(*p, op)
	}
	return nil
}
