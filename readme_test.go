package impactmodel

import (
	"os"
	"strings"
	"testing"
)

// The README and the documentation site's home page share their overview and
// their PRISM section. The site's copy links within docs/, so its links drop
// the "docs/" prefix; otherwise the shared sections must match, so that a
// change to one is made to both.
func TestReadmeAndSiteHomeShareTheirOverview(t *testing.T) {
	readme := readText(t, "README.md")
	home := readText(t, "docs/index.md")
	shared := []struct {
		name               string
		start, readmeEnd   string
		homeEnd            string
		rewriteReadmeLinks bool
	}{
		{"overview", "**Impact Model** is", "\n## Quick start", "\n## Start here", true},
		{"PRISM ecosystem", "## Part of the PRISM ecosystem", "\n## Credits", "\n## License", false},
	}
	for _, s := range shared {
		r := section(t, "README.md", readme, s.start, s.readmeEnd)
		if s.rewriteReadmeLinks {
			r = strings.ReplaceAll(r, "](docs/", "](")
		}
		h := section(t, "docs/index.md", home, s.start, s.homeEnd)
		if r != h {
			t.Errorf("the %s section of README.md and docs/index.md differ; make the same change to both\n%s", s.name, firstDifference(r, h))
		}
	}
}

func readText(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return strings.ReplaceAll(string(b), "\r\n", "\n")
}

// section returns the text from start up to, not including, end.
func section(t *testing.T, name, text, start, end string) string {
	t.Helper()
	i := strings.Index(text, start)
	if i < 0 {
		t.Fatalf("%s: missing %q", name, start)
	}
	j := strings.Index(text[i:], end)
	if j < 0 {
		t.Fatalf("%s: missing %q after %q", name, end, start)
	}
	return strings.TrimSpace(text[i : i+j])
}

// firstDifference shows the first line on which a and b differ.
func firstDifference(a, b string) string {
	al, bl := strings.Split(a, "\n"), strings.Split(b, "\n")
	for i := range max(len(al), len(bl)) {
		var x, y string
		if i < len(al) {
			x = al[i]
		}
		if i < len(bl) {
			y = bl[i]
		}
		if x != y {
			return "README.md:     " + x + "\ndocs/index.md: " + y
		}
	}
	return ""
}
