package share

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/ProductBuildersHQ/impact-model/pkg/report"
)

// ReviewRow is one element put to a reviewer. The response columns are not
// part of the row: they are empty in every export and filled in by the
// reviewer.
type ReviewRow struct {
	// Section is the packet section: outcomes, links, assumptions,
	// interventions, indicators, metrics, or actors.
	Section string
	// Kind is the element kind within the section (a link section also holds
	// precondition groups).
	Kind      string
	ID        string
	Statement string
	// Basis is the element's provenance basis, or empty if none is recorded.
	Basis string
	// Sources names the evidence the element cites, as "id: title". Only the
	// title is shown, never an excerpt or URL, so restricted evidence cannot
	// leak through a document sent for review.
	Sources []string
}

// Decisions are the values a reviewer may enter in the decision column.
var Decisions = []string{"agree", "change", "remove", "unsure"}

// sections lists the packet sections in order, with their headings.
var sections = []struct{ key, heading string }{
	{"outcomes", "Outcomes"},
	{"links", "Causal links"},
	{"assumptions", "Assumptions"},
	{"interventions", "Interventions"},
	{"indicators", "Indicators"},
	{"metrics", "Metrics"},
	{"actors", "Actors"},
}

// ReviewRows lists every active element of the view, grouped by section in
// packet order and otherwise in view order. Rejected elements are not
// included: they have already been decided.
func ReviewRows(v *report.View) []ReviewRow {
	titles := map[string]string{}
	for _, o := range v.Outcomes {
		titles[o.ID] = o.Title
	}
	for _, in := range v.Interventions {
		titles[in.ID] = in.Title
	}
	name := func(id string) string {
		if t, ok := titles[id]; ok {
			return t
		}
		return id
	}
	evidence := map[string]string{}
	for _, ev := range v.Evidence {
		evidence[ev.ID] = ev.Title
	}
	sources := func(b report.BasisView) []string {
		out := make([]string, 0, len(b.EvidenceIDs))
		for _, id := range b.EvidenceIDs {
			if t, ok := evidence[id]; ok {
				out = append(out, id+": "+oneLine(t))
			} else {
				out = append(out, id)
			}
		}
		return out
	}
	var rows []ReviewRow
	add := func(section, kind, id, statement string, b report.BasisView) {
		rows = append(rows, ReviewRow{Section: section, Kind: kind, ID: id, Statement: oneLine(statement), Basis: string(b.Basis), Sources: sources(b)})
	}
	for _, o := range v.Outcomes {
		add("outcomes", "outcome", o.ID, fmt.Sprintf("%s: %s", report.LevelLabel(o.Level), o.Title), o.Basis)
	}
	for _, l := range v.Links {
		s := name(l.From) + " → " + name(l.To)
		if l.Rationale != "" {
			s += ". " + l.Rationale
		}
		add("links", "link", l.ID, s, l.Basis)
	}
	for _, g := range v.Groups {
		add("links", "group", g.ID, fmt.Sprintf("Precondition group (%s of its members) leading to %s", g.Logic, name(g.Target)), g.Basis)
	}
	for _, a := range v.Assumptions {
		s := a.Statement
		if a.If != "" {
			s += fmt.Sprintf(" (If %s, then %s", oneLine(a.If), oneLine(a.Then))
			if a.Because != "" {
				s += ", because " + oneLine(a.Because)
			}
			s += ")"
		}
		add("assumptions", "assumption", a.ID, s, a.Basis)
	}
	for _, in := range v.Interventions {
		add("interventions", "intervention", in.ID, in.Title, in.Basis)
	}
	for _, i := range v.Indicators {
		add("indicators", "indicator", i.ID, fmt.Sprintf("%s, measuring %s", i.Name, name(i.Measures)), i.Basis)
	}
	for _, m := range v.Metrics {
		add("metrics", "metric", m.ID, m.Name+": "+m.Definition, m.Basis)
	}
	for _, a := range v.Actors {
		s := a.Name
		if a.Type != "" {
			s += " (" + report.Label(a.Type) + ")"
		}
		add("actors", "actor", a.ID, s, a.Basis)
	}
	return rows
}

func oneLine(s string) string { return strings.Join(strings.Fields(s), " ") }

func cell(s string) string {
	s = strings.ReplaceAll(oneLine(s), "|", `\|`)
	if s == "" {
		return "—"
	}
	return s
}

// ReviewPacket renders the review packet: instructions, then one table per
// element kind with ID, statement, basis, and sources, and empty decision,
// suggested change, and comment columns.
func ReviewPacket(v *report.View) string {
	rows := ReviewRows(v)
	var b strings.Builder
	fmt.Fprintf(&b, "# %s: review packet\n\n", oneLine(v.Document.Title))
	b.WriteString("## How to review\n\n")
	b.WriteString("This is a draft for you to correct. For each row, write one of " +
		strings.Join(Decisions, ", ") + " in the **Decision** column. " +
		"If you choose change, say what it should be in **Suggested change**. " +
		"Use **Comment** for anything else. Please leave the ID column as it is: it is how your answers are matched to the impact model.\n\n")
	b.WriteString("Rows marked inferred or hypothesis are our reading, not something the organization said, and are the most important to check.\n\n")
	for _, sec := range sections {
		var table strings.Builder
		for _, r := range rows {
			if r.Section != sec.key {
				continue
			}
			basis := r.Basis
			if basis == "" {
				basis = "—"
			}
			fmt.Fprintf(&table, "| %s | %s | %s | %s |  |  |  |\n", r.ID, cell(r.Statement), basis, cell(strings.Join(r.Sources, "; ")))
		}
		if table.Len() == 0 {
			continue
		}
		fmt.Fprintf(&b, "## %s\n\n", sec.heading)
		b.WriteString("| ID | Statement | Basis | Sources | Decision | Suggested change | Comment |\n")
		b.WriteString("|----|-----------|-------|---------|----------|------------------|---------|\n")
		b.WriteString(table.String())
		b.WriteString("\n")
	}
	return b.String()
}

// ReviewSheetHeader is the column order of the review sheet.
var ReviewSheetHeader = []string{"kind", "id", "statement", "basis", "sources", "decision", "suggested_change", "comment"}

// ReviewSheet renders the same rows as an RFC 4180 CSV, one row per element,
// with a kind column and empty response columns.
func ReviewSheet(v *report.View) (string, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	w.UseCRLF = true
	if err := w.Write(ReviewSheetHeader); err != nil {
		return "", err
	}
	for _, r := range ReviewRows(v) {
		if err := w.Write([]string{r.Kind, r.ID, r.Statement, r.Basis, strings.Join(r.Sources, "; "), "", "", ""}); err != nil {
			return "", err
		}
	}
	w.Flush()
	return buf.String(), w.Error()
}
