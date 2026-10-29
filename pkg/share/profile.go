package share

import (
	"fmt"
	"regexp"
	"strings"
)

// Profile configures the Pandoc Markdown profile.
type Profile struct {
	// Diagram turns Mermaid source into the path of a rendered PNG, which
	// the profile references. Nil leaves Mermaid blocks as code.
	Diagram func(source string) (path string, err error)
}

var (
	mermaidBlock = regexp.MustCompile("(?s)```mermaid\n(.*?)```\n")
	basisBadge   = regexp.MustCompile(`_\((stated|inferred|hypothesis)\)_`)
	classComment = regexp.MustCompile(`(?m)^<!-- _class: [^>]*-->\n+`)
	notesComment = regexp.MustCompile(`(?s)\n*<!--\n(.*?)\n-->\n*$`)
)

// replaceDiagrams swaps each Mermaid block for an image reference.
func (p Profile) replaceDiagrams(md string) (string, error) {
	if p.Diagram == nil {
		return md, nil
	}
	var firstErr error
	out := mermaidBlock.ReplaceAllStringFunc(md, func(block string) string {
		if firstErr != nil {
			return block
		}
		path, err := p.Diagram(mermaidBlock.FindStringSubmatch(block)[1])
		if err != nil {
			firstErr = err
			return block
		}
		return fmt.Sprintf("![Diagram](%s)\n", path)
	})
	return out, firstErr
}

// badges writes the basis marker as plain text: [inferred].
func badges(md string) string { return basisBadge.ReplaceAllString(md, "[$1]") }

// Report converts a Markdown report to the Pandoc profile: the stamp as a
// metadata block, diagrams as PNG references, basis badges as text, a page
// break before each section after the first, and the stamp again as a
// closing line.
func (p Profile) Report(s Stamp, title, body string) (string, error) {
	body, err := p.replaceDiagrams(body)
	if err != nil {
		return "", err
	}
	body = badges(body)
	// A report that opens with its own "# Title" repeats the metadata title.
	if first, rest, ok := strings.Cut(body, "\n"); ok && strings.HasPrefix(first, "# ") {
		body = strings.TrimLeft(rest, "\n")
	}
	var b strings.Builder
	b.WriteString(s.YAMLBlock(title))
	sections := 0
	for _, line := range strings.SplitAfter(body, "\n") {
		if strings.HasPrefix(line, "## ") {
			if sections > 0 {
				b.WriteString("\\newpage\n\n")
			}
			sections++
		}
		b.WriteString(line)
	}
	fmt.Fprintf(&b, "\n*%s*\n", s.Line())
	return b.String(), nil
}

// Deck converts a workshop deck written for Marp to the Pandoc profile for
// PPTX: the stamp as metadata, no Marp front matter or directives, speaker
// notes as Pandoc notes, and a closing slide carrying the stamp line. The
// title slide is dropped because the metadata supplies it.
func (p Profile) Deck(s Stamp, title, marp string) (string, error) {
	md, err := p.replaceDiagrams(marp)
	if err != nil {
		return "", err
	}
	md = badges(md)
	if rest, ok := strings.CutPrefix(md, "---\n"); ok {
		if _, after, found := strings.Cut(rest, "\n---\n"); found {
			md = after
		}
	}
	slides := strings.Split(md, "\n---\n")
	var b strings.Builder
	b.WriteString(s.YAMLBlock(title))
	for i, slide := range slides {
		if i == 0 {
			continue
		}
		slide = classComment.ReplaceAllString(strings.TrimLeft(slide, "\n"), "")
		notes := ""
		if m := notesComment.FindStringSubmatch(slide); m != nil {
			notes = m[1]
			slide = notesComment.ReplaceAllString(slide, "")
		}
		b.WriteString(strings.TrimRight(slide, "\n"))
		b.WriteString("\n")
		if notes != "" {
			fmt.Fprintf(&b, "\n::: notes\n%s\n:::\n", notes)
		}
		b.WriteString("\n---\n\n")
	}
	fmt.Fprintf(&b, "# Export details\n\n%s\n", s.Line())
	return b.String(), nil
}
