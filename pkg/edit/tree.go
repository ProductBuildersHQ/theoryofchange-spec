package edit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// The patch engine edits a generic JSON tree: map[string]any, []any,
// json.Number, string, bool, nil. Numbers stay json.Number so integers and
// decimals survive a patch exactly.

func decodeTree(data []byte) (any, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	return v, nil
}

// parsePointer splits an RFC 6901 pointer into unescaped segments.
func parsePointer(p string) ([]string, error) {
	if !strings.HasPrefix(p, "/") {
		return nil, fmt.Errorf("%w: %q must start with /", ErrPath, p)
	}
	parts := strings.Split(p[1:], "/")
	for i, s := range parts {
		if strings.Contains(s, "~") {
			s = strings.ReplaceAll(s, "~1", "/")
			s = strings.ReplaceAll(s, "~0", "~")
		}
		parts[i] = s
	}
	return parts, nil
}

// segment is a resolved step into a container.
type segment struct {
	key   string // object key
	index int    // array index; -1 when the container is an object
}

// resolve finds seg within container. For arrays, seg is a canonical index
// or @id. When appendOK, "-" resolves to len(array) (one past the end);
// when insertOK, an index equal to len(array) is allowed.
func resolve(container any, seg string, appendOK bool) (segment, error) {
	switch c := container.(type) {
	case map[string]any:
		return segment{key: seg, index: -1}, nil
	case []any:
		switch {
		case seg == "-":
			if !appendOK {
				return segment{}, fmt.Errorf("%w: \"-\" is only valid when adding", ErrPath)
			}
			return segment{index: len(c)}, nil
		case strings.HasPrefix(seg, "@"):
			id := seg[1:]
			for i, el := range c {
				if m, ok := el.(map[string]any); ok && m["id"] == id {
					return segment{index: i}, nil
				}
			}
			return segment{}, fmt.Errorf("%w: %q", ErrUnknownID, id)
		default:
			i, err := strconv.Atoi(seg)
			if err != nil || i < 0 || strconv.Itoa(i) != seg {
				return segment{}, fmt.Errorf("%w: %q is not an array index", ErrPath, seg)
			}
			if i > len(c) || (i == len(c) && !appendOK) {
				return segment{}, fmt.Errorf("%w: index %d out of range (length %d)", ErrPath, i, len(c))
			}
			return segment{index: i}, nil
		}
	default:
		return segment{}, fmt.Errorf("%w: cannot step into a %T with %q", ErrPath, container, seg)
	}
}

// lookup returns the value at segs.
func lookup(root any, segs []string) (any, error) {
	cur := root
	for _, s := range segs {
		seg, err := resolve(cur, s, false)
		if err != nil {
			return nil, err
		}
		switch c := cur.(type) {
		case map[string]any:
			v, ok := c[seg.key]
			if !ok {
				return nil, fmt.Errorf("%w: no member %q", ErrPath, seg.key)
			}
			cur = v
		case []any:
			cur = c[seg.index]
		}
	}
	return cur, nil
}

// modify walks to the container holding the last segment and applies f to
// it, writing back any replacement along the way (slices may be reallocated).
func modify(node any, segs []string, f func(container any, last string) (any, error)) (any, error) {
	if len(segs) == 1 {
		return f(node, segs[0])
	}
	seg, err := resolve(node, segs[0], false)
	if err != nil {
		return nil, err
	}
	switch c := node.(type) {
	case map[string]any:
		child, ok := c[seg.key]
		if !ok {
			return nil, fmt.Errorf("%w: no member %q", ErrPath, seg.key)
		}
		updated, err := modify(child, segs[1:], f)
		if err != nil {
			return nil, err
		}
		c[seg.key] = updated
		return c, nil
	case []any:
		updated, err := modify(c[seg.index], segs[1:], f)
		if err != nil {
			return nil, err
		}
		c[seg.index] = updated
		return c, nil
	default:
		return nil, fmt.Errorf("%w: cannot step into a %T", ErrPath, node)
	}
}

func addValue(root any, path string, value any) (any, error) {
	segs, err := parsePointer(path)
	if err != nil {
		return nil, err
	}
	return modify(root, segs, func(container any, last string) (any, error) {
		switch c := container.(type) {
		case map[string]any:
			c[last] = value
			return c, nil
		case []any:
			if strings.HasPrefix(last, "@") {
				return nil, fmt.Errorf("%w: cannot add at %q; use an index or \"-\"", ErrPath, last)
			}
			seg, err := resolve(c, last, true)
			if err != nil {
				return nil, err
			}
			c = append(c, nil)
			copy(c[seg.index+1:], c[seg.index:])
			c[seg.index] = value
			return c, nil
		default:
			return nil, fmt.Errorf("%w: cannot add into a %T", ErrPath, container)
		}
	})
}

func removeValue(root any, path string) (any, error) {
	segs, err := parsePointer(path)
	if err != nil {
		return nil, err
	}
	return modify(root, segs, func(container any, last string) (any, error) {
		switch c := container.(type) {
		case map[string]any:
			if _, ok := c[last]; !ok {
				return nil, fmt.Errorf("%w: no member %q", ErrPath, last)
			}
			delete(c, last)
			return c, nil
		case []any:
			seg, err := resolve(c, last, false)
			if err != nil {
				return nil, err
			}
			return append(c[:seg.index], c[seg.index+1:]...), nil
		default:
			return nil, fmt.Errorf("%w: cannot remove from a %T", ErrPath, container)
		}
	})
}

func replaceValue(root any, path string, value any) (any, error) {
	segs, err := parsePointer(path)
	if err != nil {
		return nil, err
	}
	return modify(root, segs, func(container any, last string) (any, error) {
		switch c := container.(type) {
		case map[string]any:
			if _, ok := c[last]; !ok {
				return nil, fmt.Errorf("%w: no member %q to replace", ErrPath, last)
			}
			c[last] = value
			return c, nil
		case []any:
			seg, err := resolve(c, last, false)
			if err != nil {
				return nil, err
			}
			c[seg.index] = value
			return c, nil
		default:
			return nil, fmt.Errorf("%w: cannot replace in a %T", ErrPath, container)
		}
	})
}

// clone deep-copies a tree value.
func clone(v any) any {
	switch x := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, c := range x {
			out[k] = clone(c)
		}
		return out
	case []any:
		out := make([]any, len(x))
		for i, c := range x {
			out[i] = clone(c)
		}
		return out
	default:
		return v
	}
}

// equal compares tree values; numbers compare by numeric value.
func equal(a, b any) bool {
	switch x := a.(type) {
	case map[string]any:
		y, ok := b.(map[string]any)
		if !ok || len(x) != len(y) {
			return false
		}
		for k, v := range x {
			w, ok := y[k]
			if !ok || !equal(v, w) {
				return false
			}
		}
		return true
	case []any:
		y, ok := b.([]any)
		if !ok || len(x) != len(y) {
			return false
		}
		for i := range x {
			if !equal(x[i], y[i]) {
				return false
			}
		}
		return true
	case json.Number:
		y, ok := b.(json.Number)
		if !ok {
			return false
		}
		if x == y {
			return true
		}
		fx, errX := x.Float64()
		fy, errY := y.Float64()
		return errX == nil && errY == nil && fx == fy
	default:
		return a == b
	}
}

// isPrefix reports whether prefix is a proper prefix of path (segment-wise).
func isPrefix(prefix, path []string) bool {
	if len(prefix) >= len(path) {
		return false
	}
	for i := range prefix {
		if prefix[i] != path[i] {
			return false
		}
	}
	return true
}

// applyOp applies one operation to the tree and returns the new root.
func applyOp(root any, op Op) (any, error) {
	var value any
	if op.Value != nil {
		v, err := decodeTree(op.Value)
		if err != nil {
			return nil, fmt.Errorf("%w: value is not valid JSON: %w", ErrPatch, err)
		}
		value = v
	}
	switch op.Op {
	case OpAdd:
		return addValue(root, op.Path, value)
	case OpReplace:
		return replaceValue(root, op.Path, value)
	case OpRemove:
		return removeValue(root, op.Path)
	case OpTest:
		segs, err := parsePointer(op.Path)
		if err != nil {
			return nil, err
		}
		got, err := lookup(root, segs)
		if err != nil {
			return nil, err
		}
		if !equal(got, value) {
			return nil, fmt.Errorf("%w: value at %s differs", ErrTestFailed, op.Path)
		}
		return root, nil
	case OpMove, OpCopy:
		from, err := parsePointer(op.From)
		if err != nil {
			return nil, err
		}
		to, err := parsePointer(op.Path)
		if err != nil {
			return nil, err
		}
		got, err := lookup(root, from)
		if err != nil {
			return nil, err
		}
		got = clone(got)
		if op.Op == OpCopy {
			return addValue(root, op.Path, got)
		}
		if isPrefix(from, to) {
			return nil, fmt.Errorf("%w: cannot move %s into itself", ErrPath, op.From)
		}
		root, err = removeValue(root, op.From)
		if err != nil {
			return nil, err
		}
		return addValue(root, op.Path, got)
	default:
		return nil, fmt.Errorf("%w %q", ErrUnknownOp, string(op.Op))
	}
}
