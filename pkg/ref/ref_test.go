package ref

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestParseRoundTrip(t *testing.T) {
	tests := []struct {
		in      string
		doc     string
		element string
	}{
		{"model:agent-auth-open-standard", "agent-auth-open-standard", ""},
		{"model:agent-auth-open-standard#oc-idp-support", "agent-auth-open-standard", "oc-idp-support"},
		{"model:a#b", "a", "b"},
		{"model:network2026#ind-attendee-hours", "network2026", "ind-attendee-hours"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			r, err := Parse(tt.in)
			if err != nil {
				t.Fatalf("Parse(%q): %v", tt.in, err)
			}
			if r.Doc != tt.doc || r.Element != tt.element {
				t.Fatalf("Parse(%q) = %+v, want doc=%q element=%q", tt.in, r, tt.doc, tt.element)
			}
			if got := r.String(); got != tt.in {
				t.Fatalf("String() = %q, want %q", got, tt.in)
			}
			if r.IsElement() != (tt.element != "") {
				t.Fatalf("IsElement() = %v", r.IsElement())
			}
		})
	}
}

func TestParseInvalid(t *testing.T) {
	tests := []struct {
		in   string
		want error
	}{
		{"", ErrInvalidPrefix},
		{"agent-auth", ErrInvalidPrefix},
		{"MODEL:agent-auth", ErrInvalidPrefix},
		{"model:", ErrEmptySlug},
		{"model:Agent-Auth", ErrInvalidSlug},
		{"model:1abc", ErrInvalidSlug},
		{"model:abc-", ErrInvalidSlug},
		{"model:a--b", ErrInvalidSlug},
		{"model:a_b", ErrInvalidSlug},
		{"model:doc#", ErrEmptySlug},
		{"model:doc#El", ErrInvalidSlug},
		{"model:doc#a#b", ErrInvalidSlug},
		{"model:" + strings.Repeat("a", MaxSlugLength+1), ErrSlugTooLong},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			_, err := Parse(tt.in)
			if err == nil {
				t.Fatalf("Parse(%q): expected error", tt.in)
			}
			if !errors.Is(err, tt.want) {
				t.Fatalf("Parse(%q) error = %v, want %v", tt.in, err, tt.want)
			}
		})
	}
}

func TestValidateSlugBoundary(t *testing.T) {
	if err := ValidateSlug(strings.Repeat("a", MaxSlugLength)); err != nil {
		t.Fatalf("max-length slug rejected: %v", err)
	}
	if !IsSlug("oc-1") || IsSlug("-oc") {
		t.Fatal("IsSlug mismatch")
	}
}

func TestNewAndMustNew(t *testing.T) {
	r, err := New("doc", "")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if r.IsElement() || r.String() != "model:doc" {
		t.Fatalf("unexpected ref %+v", r)
	}
	if _, err := New("doc", "Bad"); err == nil {
		t.Fatal("New accepted invalid element")
	}
	el := MustNew("doc", "oc-1")
	if el.Document() != r {
		t.Fatalf("Document() = %+v", el.Document())
	}
	if r.WithElement("oc-1") != el {
		t.Fatalf("WithElement() = %+v", r.WithElement("oc-1"))
	}
	defer func() {
		if recover() == nil {
			t.Fatal("MustNew did not panic on invalid input")
		}
	}()
	MustNew("Bad", "")
}

type holder struct {
	Target Ref  `json:"target"`
	Parent *Ref `json:"parent,omitempty"`
}

func TestJSON(t *testing.T) {
	h := holder{Target: MustNew("doc", "oc-1")}
	b, err := json.Marshal(h)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if string(b) != `{"target":"model:doc#oc-1"}` {
		t.Fatalf("Marshal = %s", b)
	}
	var got holder
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.Target != h.Target {
		t.Fatalf("round trip = %+v", got)
	}

	var empty holder
	if err := json.Unmarshal([]byte(`{"target":""}`), &empty); err != nil {
		t.Fatalf("Unmarshal empty: %v", err)
	}
	if !empty.Target.IsZero() {
		t.Fatalf("empty target = %+v", empty.Target)
	}
	b, err = json.Marshal(empty)
	if err != nil {
		t.Fatalf("Marshal zero: %v", err)
	}
	if string(b) != `{"target":""}` {
		t.Fatalf("Marshal zero = %s", b)
	}

	if err := json.Unmarshal([]byte(`{"target":"nope"}`), &got); err == nil {
		t.Fatal("Unmarshal accepted invalid ref")
	}
	if _, err := json.Marshal(holder{Target: Ref{Doc: "Bad"}}); err == nil {
		t.Fatal("Marshal accepted invalid ref")
	}
}
