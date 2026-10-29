// Package edit is the only way a Theory of Change document is modified.
//
// Changes are expressed as RFC 6902 JSON Patches, extended with ID
// addressing: a pointer segment of the form @<id> selects the array
// element whose "id" equals <id>, so /outcomes/@oc-adoption/title keeps
// working when elements are added or removed before it.
//
// Apply is atomic and pure. It works on a copy; if any operation fails, the
// result no longer conforms to the schema, or the patch would leave the
// document with more error-severity validation issues than it started with,
// nothing changes and the original document is untouched.
package edit

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// OpName is a JSON Patch operation.
type OpName string

// Operations.
const (
	OpAdd     OpName = "add"
	OpRemove  OpName = "remove"
	OpReplace OpName = "replace"
	OpMove    OpName = "move"
	OpCopy    OpName = "copy"
	OpTest    OpName = "test"
)

// Op is one JSON Patch operation.
type Op struct {
	Op   OpName `json:"op"`
	Path string `json:"path"`
	From string `json:"from,omitempty"`
	// Value is the JSON value for add, replace, and test. A nil Value means
	// absent; the JSON literal null is a present value.
	Value json.RawMessage `json:"value,omitempty"`
}

// Patch is an ordered list of operations applied as one unit.
type Patch []Op

// Errors returned (wrapped) by ParsePatch and Apply.
var (
	// ErrPatch is the root of every error caused by a malformed or
	// inapplicable patch.
	ErrPatch = errors.New("invalid patch")
	// ErrUnknownOp is an operation name outside RFC 6902.
	ErrUnknownOp = fmt.Errorf("%w: unknown operation", ErrPatch)
	// ErrPath is a pointer that is malformed or does not resolve.
	ErrPath = fmt.Errorf("%w: bad path", ErrPatch)
	// ErrUnknownID is an @id segment matching no element.
	ErrUnknownID = fmt.Errorf("%w: unknown element id", ErrPatch)
	// ErrTestFailed is a failed "test" operation.
	ErrTestFailed = fmt.Errorf("%w: test failed", ErrPatch)
)

// OpError locates a failure within a patch.
type OpError struct {
	Index int
	Op    Op
	Err   error
}

func (e *OpError) Error() string {
	return fmt.Sprintf("patch op %d (%s %s): %v", e.Index, e.Op.Op, e.Op.Path, e.Err)
}

func (e *OpError) Unwrap() error { return e.Err }

// ParsePatch decodes a JSON Patch document and checks each operation is
// well-formed: a known name, a path, a from for move and copy, and a value
// for add, replace, and test.
func ParsePatch(data []byte) (Patch, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var p Patch
	if err := dec.Decode(&p); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrPatch, err)
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

// Validate checks each operation is well-formed.
func (p Patch) Validate() error {
	for i, op := range p {
		if err := op.validate(); err != nil {
			return &OpError{Index: i, Op: op, Err: err}
		}
	}
	return nil
}

func (op Op) validate() error {
	switch op.Op {
	case OpAdd, OpReplace, OpTest:
		if op.Value == nil {
			return fmt.Errorf("%w: %s requires a value", ErrPatch, op.Op)
		}
	case OpRemove:
	case OpMove, OpCopy:
		if op.From == "" {
			return fmt.Errorf("%w: %s requires from", ErrPatch, op.Op)
		}
	default:
		return fmt.Errorf("%w %q", ErrUnknownOp, string(op.Op))
	}
	if op.Path == "" {
		return fmt.Errorf("%w: the document root cannot be the target of %s", ErrPath, op.Op)
	}
	if !strings.HasPrefix(op.Path, "/") {
		return fmt.Errorf("%w: %q must start with /", ErrPath, op.Path)
	}
	if op.From != "" && !strings.HasPrefix(op.From, "/") {
		return fmt.Errorf("%w: from %q must start with /", ErrPath, op.From)
	}
	return nil
}

func valueOf(v any) (json.RawMessage, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("%w: encode value: %w", ErrPatch, err)
	}
	return b, nil
}

// Add returns an add operation. Use "-" as the last array segment to append.
func Add(path string, value any) (Op, error) {
	v, err := valueOf(value)
	return Op{Op: OpAdd, Path: path, Value: v}, err
}

// Replace returns a replace operation.
func Replace(path string, value any) (Op, error) {
	v, err := valueOf(value)
	return Op{Op: OpReplace, Path: path, Value: v}, err
}

// Test returns a test operation, which fails the whole patch unless the
// value at path equals value.
func Test(path string, value any) (Op, error) {
	v, err := valueOf(value)
	return Op{Op: OpTest, Path: path, Value: v}, err
}

// Remove returns a remove operation.
func Remove(path string) Op { return Op{Op: OpRemove, Path: path} }

// Move returns a move operation.
func Move(from, path string) Op { return Op{Op: OpMove, From: from, Path: path} }

// Copy returns a copy operation.
func Copy(from, path string) Op { return Op{Op: OpCopy, From: from, Path: path} }

// Pointer builds a JSON pointer from segments, escaping "~" and "/". Use
// "@id" segments for ID addressing and "-" to append to an array.
func Pointer(segments ...string) string {
	var b strings.Builder
	for _, s := range segments {
		s = strings.ReplaceAll(s, "~", "~0")
		s = strings.ReplaceAll(s, "/", "~1")
		b.WriteByte('/')
		b.WriteString(s)
	}
	return b.String()
}
