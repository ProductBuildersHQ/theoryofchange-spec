package report

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/ProductBuildersHQ/impact-model/pkg/burden"
	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
)

// CSV tables are RFC 4180: comma separated, CRLF line ends, fields quoted
// only when needed. Multi-valued cells are joined with "; ". Rows follow the
// view's order, so output is deterministic.

// CSVTables lists the tables ExportCSV can write from an impact model alone.
var CSVTables = []string{"indicators", "metrics", "glossary", "metric-selection", "collection-plan"}

// CSV renders one named table. The selection and collection tables are built
// from the document, so doc must be set for them; the others use the view.
func CSV(table string, v *View, doc *impact.Model) (string, error) {
	if (table == "metric-selection" || table == "collection-plan") && doc == nil {
		return "", fmt.Errorf("table %q needs the document", table)
	}
	switch table {
	case "metric-selection":
		return metricSelectionCSV(doc)
	case "collection-plan":
		rep, err := burden.Compute(doc)
		if err != nil {
			return "", err
		}
		return collectionPlanCSV(doc, rep)
	case "indicators":
		return indicatorsCSV(v)
	case "metrics":
		return metricsCSV(v)
	case "glossary":
		return glossaryCSV(v)
	default:
		return "", fmt.Errorf("unknown table %q (want one of %s)", table, strings.Join(CSVTables, ", "))
	}
}

func writeCSV(header []string, rows [][]string) (string, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	w.UseCRLF = true
	if err := w.Write(header); err != nil {
		return "", err
	}
	if err := w.WriteAll(rows); err != nil {
		return "", err
	}
	return buf.String(), w.Error()
}

func join(items []string) string { return strings.Join(items, "; ") }

func basisCell(b BasisView) string {
	if !b.Marked {
		return ""
	}
	return string(b.Basis)
}

func indicatorsCSV(v *View) (string, error) {
	rows := make([][]string, 0, len(v.Indicators))
	for _, i := range v.Indicators {
		rows = append(rows, []string{
			i.ID, i.Name, i.Definition, i.Measures, i.Metric, i.Type, i.Unit, i.Direction, string(i.Claim),
			join(i.Pathways), i.Baseline, join(i.Targets), i.DataSource, i.Method, i.Frequency, i.Owner,
			i.DataAvailability, basisCell(i.Basis),
		})
	}
	return writeCSV([]string{
		"id", "name", "definition", "measures", "metric", "type", "unit", "direction", "claim",
		"pathways", "baseline", "targets", "data_source", "method", "frequency", "owner",
		"data_availability", "basis",
	}, rows)
}

func metricsCSV(v *View) (string, error) {
	rows := make([][]string, 0, len(v.Metrics))
	for _, m := range v.Metrics {
		rows = append(rows, []string{
			m.ID, m.Name, m.Definition, m.Formula, m.Unit, m.Direction, m.Aggregation, m.WeightedBy,
			m.Layer, join(m.Indicators), basisCell(m.Basis),
		})
	}
	return writeCSV([]string{
		"id", "name", "definition", "formula", "unit", "direction", "aggregation", "weighted_by",
		"layer", "indicators", "basis",
	}, rows)
}

func glossaryCSV(v *View) (string, error) {
	rows := make([][]string, 0, len(v.Glossary))
	for _, t := range v.Glossary {
		rows = append(rows, []string{t.ID, t.Term, t.Definition, join(t.Examples)})
	}
	return writeCSV([]string{"id", "term", "definition", "examples"}, rows)
}
