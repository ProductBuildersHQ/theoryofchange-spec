// Package slides renders Theory of Change slide decks as Marp Markdown.
//
// Decks are built from the same report.View as the Markdown reports, so a
// number or a label never differs between a deck and a report. This package
// only produces text. Rendering diagrams to images and exporting to PDF,
// PPTX, or HTML is done by external tools invoked from the command line,
// never from here.
package slides

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/logicmodel"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/report"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

// Options configures deck generation.
type Options struct {
	// Theme is the Marp theme name. The default is "default".
	Theme string
	// Stamp is the export stamp line shown on the title slide; empty for none.
	Stamp string
}

const (
	assumptionsPerSlide = 3
	metricsPerSlide     = 4
	termsPerSlide       = 8
	listedMarkedItems   = 10
)

// deck accumulates slides.
type deck struct {
	b      strings.Builder
	slides int
}

// slide starts a new slide. Marp separates slides with a line of "---".
func (d *deck) slide(class string) {
	if d.slides > 0 {
		d.b.WriteString("\n---\n\n")
	} else {
		d.b.WriteString("\n")
	}
	if class != "" {
		fmt.Fprintf(&d.b, "<!-- _class: %s -->\n\n", class)
	}
	d.slides++
}

func (d *deck) line(format string, args ...any) {
	fmt.Fprintf(&d.b, format, args...)
	d.b.WriteByte('\n')
}

func (d *deck) notes(text string) {
	d.b.WriteString("\n<!--\n" + strings.ReplaceAll(text, "-->", "--&gt;") + "\n-->\n")
}

// SlideCount returns the number of slides in Marp Markdown produced by this
// package: slides are separated by "---" lines after the front matter.
func SlideCount(markdown string) int {
	body := markdown
	if rest, ok := strings.CutPrefix(markdown, "---\n"); ok {
		if _, after, found := strings.Cut(rest, "\n---\n"); found {
			body = after
		}
	}
	return strings.Count(body, "\n---\n") + 1
}

// Workshop renders the workshop deck: the problem, the goal and what the
// organization is accountable for, the pathway diagram, a logic model per
// pathway, the key assumptions, the metric definitions, the glossary, how
// much of the theory is established versus inferred, and the open questions
// for the room. It needs no observations.
//
// whole is the view of the entire theory; pathways are the per-pathway
// views from report.PathwayViews (empty when the theory has none).
func Workshop(whole *report.View, pathways []*report.View, o Options) string {
	theme := o.Theme
	if theme == "" {
		theme = "default"
	}
	d := &deck{}
	d.line("---")
	d.line("marp: true")
	d.line("theme: %s", theme)
	d.line("paginate: true")
	d.line("title: %q", whole.Document.Title+": workshop draft")
	d.line("---")

	d.title(whole, o.Stamp)
	d.problem(whole)
	d.goal(whole)
	d.diagram("The theory on one page", report.Mermaid(whole, report.MermaidOptions{}),
		"Walk the arrows upward. Ask: is each arrow really a cause? What is missing between a box and the one above it?")
	for _, pv := range pathways {
		d.pathway(pv)
	}
	if len(pathways) == 0 {
		d.logicModel(whole)
		d.layeredLogicModel(whole)
	}
	d.assumptions(whole)
	d.metrics(whole)
	d.glossary(whole)
	d.basis(whole)
	d.openQuestions(whole)
	return d.b.String()
}

func (d *deck) title(v *report.View, stamp string) {
	d.slide("lead")
	d.line("# %s", v.Document.Title)
	d.line("")
	d.line("**Theory of Change: workshop draft**")
	d.line("")
	var facts []string
	if v.Document.Perspective != "" {
		facts = append(facts, "Perspective: "+v.Document.Perspective)
	}
	if v.Document.TimeStart != "" {
		facts = append(facts, fmt.Sprintf("%s to %s", v.Document.TimeStart, v.Document.TimeEnd))
	}
	if v.Document.Status != "" {
		facts = append(facts, "Status: "+report.Label(v.Document.Status))
	}
	if len(facts) > 0 {
		d.line("%s", strings.Join(facts, " · "))
	}
	if stamp != "" {
		d.line("")
		d.line("_%s_", stamp)
	}
	d.notes("Open by saying what this is: a draft to be corrected, not a finished plan. Anything marked inferred or hypothesis is our reading, and is yours to change.")
}

func (d *deck) problem(v *report.View) {
	d.slide("")
	d.line("# The problem")
	d.line("")
	d.line("%s", oneLine(v.Document.Problem))
	if len(v.Document.RootCauses) > 0 {
		d.line("")
		d.line("**Root causes we see**")
		d.line("")
		for _, c := range v.Document.RootCauses {
			d.line("- %s", oneLine(c))
		}
	}
	d.notes("Ask: would you describe the problem this way? Who is missing from this description?")
}

func (d *deck) goal(v *report.View) {
	d.slide("")
	d.line("# The change we are working toward")
	d.line("")
	for _, o := range v.Outcomes {
		if o.Level == toc.LevelImpact {
			d.line("- **%s**%s", oneLine(o.Title), report.BasisMark(o.Basis))
		}
	}
	if v.Document.Ceiling != "" {
		d.line("")
		d.line("**Ceiling of accountability:** %s", v.Document.Ceiling)
		if v.Document.CeilingRationale != "" {
			d.line("")
			d.line("%s", oneLine(v.Document.CeilingRationale))
		}
	}
	d.notes("Outcomes above the ceiling are tracked but not owned: we contribute to them, we do not claim them. Ask: is the line in the right place?")
}

func (d *deck) diagram(title, mermaid, notes string) {
	d.slide("")
	d.line("# %s", title)
	d.line("")
	d.line("```mermaid")
	d.b.WriteString(mermaid)
	d.line("```")
	d.notes(notes)
}

func (d *deck) pathway(v *report.View) {
	title := report.PathwayTitle(v)
	d.diagram("Pathway: "+title, report.Mermaid(v, report.MermaidOptions{}),
		"Ask the people who run this pathway to correct it. What do you do that is not on the page? What on the page do you not do?")
	d.logicModel(v)
	d.layeredLogicModel(v)
}

// layeredLogicModel draws the logic model as columns with an arrow wherever one
// item leads to another. It follows the table of the same logic model, which it
// does not replace.
func (d *deck) layeredLogicModel(v *report.View) {
	g := logicmodel.Build(v)
	if len(g.Stages) < 2 {
		return
	}
	text, err := g.Mermaid()
	if err != nil {
		// A graph built from a view only holds flowchart content, so this is a
		// fault in the generator and cannot be passed to the room as a slide.
		panic(fmt.Sprintf("slides: layered logic model: %v", err))
	}
	d.diagram("Logic model, step by step: "+report.PathwayTitle(v), text,
		"Follow one input to the change it is for. Is each arrow a thing that really leads to the next? Where does an arrow have no reason behind it?")
}

func (d *deck) logicModel(v *report.View) {
	d.slide("")
	d.line("# Logic model: %s", report.PathwayTitle(v))
	d.line("")
	d.line("| Stage | Elements |")
	d.line("|-------|----------|")
	for _, st := range report.LogicModelStages(v) {
		cell := "—"
		if len(st.Items) > 0 {
			cell = strings.Join(st.Items, "; ")
		}
		d.line("| %s | %s |", st.Name, cell)
	}
	d.notes("Read left to right: what goes in, what we do, what we produce, what changes. A blank row is a gap worth discussing.")
}

func (d *deck) assumptions(v *report.View) {
	var key []report.AssumptionView
	for _, a := range v.Assumptions {
		if a.Criticality == string(toc.CriticalityHigh) || a.Criticality == string(toc.CriticalityCritical) {
			key = append(key, a)
		}
	}
	for i, group := range chunk(key, assumptionsPerSlide) {
		d.slide("")
		d.line("# What we are assuming%s", pageSuffix(i, len(key), assumptionsPerSlide))
		d.line("")
		for _, a := range group {
			d.line("**%s**%s", oneLine(a.Statement), report.BasisMark(a.Basis))
			d.line("")
			if a.If != "" {
				d.line("- If %s, then %s%s", oneLine(a.If), oneLine(a.Then), becauseText(a.Because))
			}
			if a.Test != "" {
				d.line("- To test it: %s", oneLine(a.Test))
			}
			d.line("- Criticality %s · status %s", a.Criticality, report.Label(a.Status))
			d.line("")
		}
		d.notes("Ask: which of these do you believe least? What evidence would change your mind?")
	}
}

func (d *deck) metrics(v *report.View) {
	for i, group := range chunk(v.Metrics, metricsPerSlide) {
		d.slide("")
		d.line("# How we will measure%s", pageSuffix(i, len(v.Metrics), metricsPerSlide))
		d.line("")
		d.line("| Metric | Definition | Formula | Combines across pathways |")
		d.line("|--------|------------|---------|--------------------------|")
		for _, m := range group {
			d.line("| %s | %s | %s | %s |", report.MDCell(m.Name)+report.BasisMark(m.Basis),
				report.MDCell(m.Definition), report.MDCell(m.Formula), report.AggregationText(m))
		}
		d.notes("Ask: could you collect this with the time you have? Is the formula what you would count?")
	}
}

func (d *deck) glossary(v *report.View) {
	for i, group := range chunk(v.Glossary, termsPerSlide) {
		d.slide("")
		d.line("# Words we are using%s", pageSuffix(i, len(v.Glossary), termsPerSlide))
		d.line("")
		d.line("| Term | Meaning here |")
		d.line("|------|--------------|")
		for _, t := range group {
			d.line("| %s | %s |", report.MDCell(t.Term), report.MDCell(t.Definition))
		}
		d.notes("Agree the words before you debate the content. Which of these would your funders define differently?")
	}
}

// basis shows how much of the theory is established, the heart of an honest
// draft.
func (d *deck) basis(v *report.View) {
	if len(v.BasisCounts) == 0 {
		return
	}
	d.slide("")
	d.line("# What is established, and what is our reading")
	d.line("")
	d.line("| Basis | Elements |")
	d.line("|-------|----------|")
	for _, b := range toc.AllBases() {
		if b == toc.BasisRejected {
			continue
		}
		if n := v.BasisCounts[b]; n > 0 {
			d.line("| %s | %d |", report.Label(string(b)), n)
		}
	}
	if v.Unmarked > 0 {
		d.line("| No basis recorded | %d |", v.Unmarked)
	}
	if n := len(v.Rejected); n > 0 {
		d.line("")
		d.line("%d element(s) were considered and rejected; they are kept on record but left out of the diagrams.", n)
	}
	d.notes("Stated means a source says it. Inferred and hypothesis are our reading and are the first things to challenge. Confirmed means the people it concerns agreed.")
}

// openQuestions lists what the room should settle.
func (d *deck) openQuestions(v *report.View) {
	var marked []string
	add := func(title string, b report.BasisView) {
		if b.Marked {
			marked = append(marked, oneLine(title)+" ("+string(b.Basis)+")")
		}
	}
	for _, o := range v.Outcomes {
		add(o.Title, o.Basis)
	}
	for _, in := range v.Interventions {
		add(in.Title, in.Basis)
	}
	for _, a := range v.Assumptions {
		add(a.Statement, a.Basis)
	}
	if len(marked) == 0 {
		return
	}
	d.slide("")
	d.line("# For the room to settle")
	d.line("")
	d.line("**Our reading, to confirm or correct**")
	d.line("")
	for _, m := range marked[:min(len(marked), listedMarkedItems)] {
		d.line("- %s", m)
	}
	if more := len(marked) - listedMarkedItems; more > 0 {
		d.line("- …and %d more, listed in the assessment", more)
	}
	d.notes("Work down this list. For each: agree, change, remove, or not sure. Record what the room decides.")
}

// ---- helpers ----------------------------------------------------------------

func oneLine(s string) string { return strings.Join(strings.Fields(s), " ") }

func becauseText(because string) string {
	if because == "" {
		return ""
	}
	return ", because " + oneLine(because)
}

func pageSuffix(i, total, per int) string {
	if total <= per {
		return ""
	}
	return fmt.Sprintf(" (%d of %d)", i+1, (total+per-1)/per)
}

func chunk[T any](items []T, size int) [][]T {
	var out [][]T
	for chunked := range slices.Chunk(items, size) {
		out = append(out, chunked)
	}
	return out
}
