package skill_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ProductBuildersHQ/impact-model/internal/cli"
)

type source struct {
	File     string `json:"file"`
	URL      string `json:"url"`
	Title    string `json:"title"`
	Excerpt  string `json:"excerpt"`
	Accessed string `json:"accessed"`
}

func run(t *testing.T, args ...string) string {
	t.Helper()
	var out, errOut bytes.Buffer
	if code := cli.Execute(args, strings.NewReader(""), &out, &errOut); code != cli.ExitOK {
		t.Fatalf("impact-model %s: exit %d: %s", strings.Join(args, " "), code, errOut.String())
	}
	return out.String()
}

// draft runs the skill's command sequence over the saved fixture pages and
// returns the resulting document.
func draft(t *testing.T) []byte {
	t.Helper()
	raw, err := os.ReadFile("testdata/sources.json")
	if err != nil {
		t.Fatal(err)
	}
	var sources []source
	if err := json.Unmarshal(raw, &sources); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(t.TempDir(), "impact-model.json")
	run(t, "init", "--id", "lakeside", "--name", "Lakeside Community Alliance", "--out", file)
	for _, s := range sources {
		page, err := os.ReadFile(filepath.Join("testdata", s.File))
		if err != nil {
			t.Fatal(err)
		}
		// An excerpt that is not in the page would be an invented citation.
		if !strings.Contains(strings.Join(strings.Fields(string(page)), " "), s.Excerpt) {
			t.Fatalf("excerpt %q is not verbatim in %s", s.Excerpt, s.File)
		}
		run(t, "evidence", "add", file, "--title", s.Title, "--url", s.URL,
			"--accessed", s.Accessed, "--sensitivity", "public", "--excerpt", s.Excerpt)
	}
	run(t, "patch", file, "templates/draft.patch.json")
	run(t, "validate", file)
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

type element struct {
	ID         string `json:"id"`
	Provenance *struct {
		Basis       string   `json:"basis"`
		EvidenceIDs []string `json:"evidenceIds"`
	} `json:"provenance"`
}

func TestDraftFromSourcesIsHonest(t *testing.T) {
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(draft(t), &doc); err != nil {
		t.Fatal(err)
	}
	list := func(kind string) []element {
		raw, ok := doc[kind]
		if !ok {
			return nil
		}
		var els []element
		if err := json.Unmarshal(raw, &els); err != nil {
			t.Fatalf("decode %s: %v", kind, err)
		}
		return els
	}

	known := map[string]bool{}
	var evidence []struct{ ID, URL, Excerpt string }
	if err := json.Unmarshal(doc["evidence"], &evidence); err != nil {
		t.Fatal(err)
	}
	if len(evidence) != 2 {
		t.Fatalf("want 2 evidence records, got %d", len(evidence))
	}
	for _, ev := range evidence {
		known[ev.ID] = true
		if ev.URL == "" || ev.Excerpt == "" {
			t.Errorf("evidence %s lacks a URL or excerpt", ev.ID)
		}
	}

	checked := 0
	for _, kind := range []string{"outcomes", "links", "indicators", "interventions", "assumptions"} {
		for _, el := range list(kind) {
			checked++
			if el.Provenance == nil || el.Provenance.Basis == "" {
				t.Errorf("%s %s has no basis", kind, el.ID)
				continue
			}
			if el.Provenance.Basis != "stated" {
				continue
			}
			if len(el.Provenance.EvidenceIDs) == 0 {
				t.Errorf("stated %s %s cites no source", kind, el.ID)
			}
			for _, id := range el.Provenance.EvidenceIDs {
				if !known[id] {
					t.Errorf("%s %s cites unknown evidence %s", kind, el.ID, id)
				}
			}
		}
	}
	if checked == 0 {
		t.Fatal("no elements were checked")
	}
}

func TestDraftFromSourcesIsRepeatable(t *testing.T) {
	if a, b := draft(t), draft(t); !bytes.Equal(a, b) {
		t.Error("running the skill's commands twice gave different documents")
	}
}

func TestDraftRendersTheWorkshopPack(t *testing.T) {
	data := draft(t)
	file := filepath.Join(t.TempDir(), "impact-model.json")
	if err := os.WriteFile(file, data, 0o600); err != nil {
		t.Fatal(err)
	}
	out := run(t, "report", "definitions", file)
	if !strings.Contains(out, "Member groups run joint events") {
		t.Errorf("definitions pack does not show the drafted outcomes:\n%s", out)
	}
}
