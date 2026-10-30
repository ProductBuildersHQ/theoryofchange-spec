// Package tabular moves observations in from spreadsheets. A mapping names
// the columns; every row either becomes an observation or is reported as
// unmatched with its row number and the reason. No row is dropped silently.
package tabular

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

// DefaultDateFormat is the date layout used when a mapping gives none.
const DefaultDateFormat = "2006-01-02"

// ErrMapping is returned for a mapping that cannot be used with a sheet: a
// named column the sheet does not have, or no way to find the indicator.
var ErrMapping = errors.New("cannot apply the mapping")

// Unmatched is a row that could not be imported.
type Unmatched struct {
	// Row is the row's position in the file, counting the header as row 1.
	Row    int      `json:"row"`
	Reason string   `json:"reason"`
	Cells  []string `json:"cells"`
}

// Result is the outcome of an import.
type Result struct {
	Set *toc.ObservationSet `json:"set"`
	// Imported is the number of rows that became observations.
	Imported int `json:"imported"`
	// Blank is the number of empty rows skipped. They are not data.
	Blank     int         `json:"blank"`
	Unmatched []Unmatched `json:"unmatched"`
	// Header is the sheet's header row, for writing the unmatched rows back
	// out in the same shape.
	Header []string `json:"header"`
}

// Rows returns the number of data rows read: imported, unmatched, and blank.
func (r *Result) Rows() int { return r.Imported + len(r.Unmatched) + r.Blank }

// Import reads a CSV sheet into an observation set for doc. The first row is
// the header. Rows that cannot be resolved are returned in Unmatched, never
// dropped. A leading byte-order mark is ignored.
func Import(src io.Reader, m *toc.CSVMapping, doc *toc.Document) (*Result, error) {
	if m.TocID != "" && doc.ID != "" && m.TocID != doc.ID {
		return nil, fmt.Errorf("%w: the mapping is for theory %q, not %q", ErrMapping, m.TocID, doc.ID)
	}
	data, err := io.ReadAll(src)
	if err != nil {
		return nil, fmt.Errorf("tabular: read sheet: %w", err)
	}
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))
	r := csv.NewReader(bytes.NewReader(data))
	r.FieldsPerRecord = -1
	r.LazyQuotes = true

	header, err := r.Read()
	if errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("%w: the sheet is empty", ErrMapping)
	}
	if err != nil {
		return nil, fmt.Errorf("tabular: read header: %w", err)
	}
	cols, err := resolveColumns(header, m.Columns)
	if err != nil {
		return nil, err
	}
	imp := &importer{doc: doc, m: m, cols: cols, layout: m.DateFormat}
	if imp.layout == "" {
		imp.layout = DefaultDateFormat
	}
	imp.index()

	set := toc.NewObservationSet(doc.ID, m.Period)
	res := &Result{Set: set, Unmatched: []Unmatched{}, Header: header}
	for row := 2; ; row++ {
		rec, err := r.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			// A malformed line is reported with its row and goes on.
			res.Unmatched = append(res.Unmatched, Unmatched{Row: row, Reason: "unreadable row: " + err.Error()})
			continue
		}
		if blank(rec) {
			res.Blank++
			continue
		}
		ob, reason := imp.row(rec, len(header))
		if reason != "" {
			res.Unmatched = append(res.Unmatched, Unmatched{Row: row, Reason: reason, Cells: rec})
			continue
		}
		set.Observations = append(set.Observations, ob)
		res.Imported++
	}
	return res, nil
}

func blank(rec []string) bool {
	return !slices.ContainsFunc(rec, func(c string) bool { return strings.TrimSpace(c) != "" })
}

// columns holds the position of each mapped column, or -1.
type columns struct {
	indicator, date, value, level, pathway, quality, source, note int
}

func resolveColumns(header []string, c toc.CSVColumns) (columns, error) {
	find := func(name, what string) (int, error) {
		if name == "" {
			return -1, nil
		}
		for i, h := range header {
			if strings.EqualFold(strings.TrimSpace(h), strings.TrimSpace(name)) {
				return i, nil
			}
		}
		return -1, fmt.Errorf("%w: the sheet has no %q column (mapped as %s); it has %q", ErrMapping, name, what, header)
	}
	var out columns
	var err error
	for _, f := range []struct {
		dst  *int
		name string
		what string
	}{
		{&out.indicator, c.Indicator, "indicator"}, {&out.date, c.Date, "date"}, {&out.value, c.Value, "value"},
		{&out.level, c.Level, "level"}, {&out.pathway, c.Pathway, "pathway"}, {&out.quality, c.Quality, "quality"},
		{&out.source, c.Source, "source"}, {&out.note, c.Note, "note"},
	} {
		if *f.dst, err = find(f.name, f.what); err != nil {
			return out, err
		}
	}
	switch {
	case out.date < 0:
		return out, fmt.Errorf("%w: the mapping names no date column", ErrMapping)
	case out.indicator < 0 && c.IndicatorFixed == "":
		return out, fmt.Errorf("%w: the mapping names neither an indicator column nor a fixed indicator", ErrMapping)
	case out.value < 0 && out.level < 0:
		return out, fmt.Errorf("%w: the mapping names neither a value column nor a level column", ErrMapping)
	}
	return out, nil
}

type importer struct {
	doc    *toc.Document
	m      *toc.CSVMapping
	cols   columns
	layout string

	indicators map[string]toc.Indicator // by ID
	byLabel    map[string]string        // lower-case alias or ID to indicator ID
	metrics    map[string]string        // lower-case alias, ID, or name to metric ID
	pathways   map[string]string        // lower-case ID or title to pathway ID
}

func (im *importer) index() {
	im.indicators = map[string]toc.Indicator{}
	im.byLabel = map[string]string{}
	im.metrics = map[string]string{}
	im.pathways = map[string]string{}
	for _, ind := range im.doc.Indicators {
		if ind.Provenance.IsRejected() {
			continue
		}
		im.indicators[ind.ID] = ind
		im.byLabel[key(ind.ID)] = ind.ID
	}
	for label, id := range im.m.IndicatorAliases {
		im.byLabel[key(label)] = id
	}
	for _, mt := range im.doc.Metrics {
		im.metrics[key(mt.ID)] = mt.ID
		im.metrics[key(mt.Name)] = mt.ID
	}
	for label, id := range im.m.MetricAliases {
		im.metrics[key(label)] = id
	}
	for _, p := range im.doc.Pathways {
		im.pathways[key(p.ID)] = p.ID
		im.pathways[key(p.Title)] = p.ID
	}
}

func key(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

// cell returns the trimmed cell at i, or "" when the column is not mapped or
// the row is short.
func cell(rec []string, i int) string {
	if i < 0 || i >= len(rec) {
		return ""
	}
	return strings.TrimSpace(rec[i])
}

// row turns one record into an observation, or says why it cannot.
func (im *importer) row(rec []string, width int) (toc.Observation, string) {
	if len(rec) != width {
		return toc.Observation{}, fmt.Sprintf("the row has %d cells and the header has %d", len(rec), width)
	}
	ind, reason := im.indicator(rec)
	if reason != "" {
		return toc.Observation{}, reason
	}
	if ind.MeasurementType == toc.MeasurementThemeCount {
		return toc.Observation{}, fmt.Sprintf("indicator %q counts story themes; its value comes from stories, not from a sheet", ind.ID)
	}
	raw := cell(rec, im.cols.date)
	at, err := time.Parse(im.layout, raw)
	if err != nil {
		return toc.Observation{}, fmt.Sprintf("the date %q does not match the format %q", raw, im.layout)
	}
	ob := toc.Observation{
		IndicatorID: ind.ID, Date: at.Format(time.DateOnly),
		Source: cell(rec, im.cols.source), Note: cell(rec, im.cols.note),
	}
	if reason := im.setValue(&ob, ind, rec); reason != "" {
		return toc.Observation{}, reason
	}
	if q := cell(rec, im.cols.quality); q != "" {
		quality := toc.DataQuality(strings.ToLower(q))
		if !quality.Valid() {
			return toc.Observation{}, fmt.Sprintf("the data quality %q is not one of verified, reported, estimated", q)
		}
		ob.Quality = quality
	}
	return ob, ""
}

// setValue fills in the observation's value or level.
func (im *importer) setValue(ob *toc.Observation, ind toc.Indicator, rec []string) string {
	value, level := cell(rec, im.cols.value), cell(rec, im.cols.level)
	if len(ind.Levels) > 0 || ind.MeasurementType == toc.MeasurementQualitative {
		if level == "" {
			level = value // a qualitative sheet may keep its levels in the value column
		}
		if level == "" {
			return fmt.Sprintf("indicator %q is qualitative and the row has no level", ind.ID)
		}
		if i := slices.IndexFunc(ind.Levels, func(l string) bool { return strings.EqualFold(l, level) }); i >= 0 {
			ob.Level = ind.Levels[i]
			return ""
		}
		return fmt.Sprintf("the level %q is not one of %q for indicator %q", level, ind.Levels, ind.ID)
	}
	if value == "" {
		return fmt.Sprintf("the row has no value for indicator %q", ind.ID)
	}
	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fmt.Sprintf("the value %q is not a number", value)
	}
	ob.Value = &v
	return ""
}

// indicator finds the indicator a row is about.
func (im *importer) indicator(rec []string) (toc.Indicator, string) {
	if fixed := im.m.Columns.IndicatorFixed; fixed != "" && im.cols.indicator < 0 {
		return im.lookup(fixed)
	}
	label := cell(rec, im.cols.indicator)
	if label == "" {
		return toc.Indicator{}, "the row names no indicator"
	}
	if id, ok := im.byLabel[key(label)]; ok {
		return im.lookup(id)
	}
	metric, ok := im.metrics[key(label)]
	if !ok {
		return toc.Indicator{}, fmt.Sprintf("unknown indicator %q", label)
	}
	return im.byMetric(metric, label, cell(rec, im.cols.pathway))
}

func (im *importer) lookup(id string) (toc.Indicator, string) {
	ind, ok := im.indicators[id]
	if !ok {
		return toc.Indicator{}, fmt.Sprintf("indicator %q is not in the theory", id)
	}
	return ind, ""
}

// byMetric chooses, among the indicators of a metric, the one for the
// pathway named in the row.
func (im *importer) byMetric(metric, label, pathway string) (toc.Indicator, string) {
	var of []toc.Indicator
	for _, ind := range im.doc.Indicators {
		if ind.MetricID == metric && !ind.Provenance.IsRejected() {
			of = append(of, ind)
		}
	}
	switch {
	case len(of) == 0:
		return toc.Indicator{}, fmt.Sprintf("%q is a metric with no indicator", label)
	case len(of) == 1 && pathway == "":
		return of[0], ""
	case pathway == "":
		return toc.Indicator{}, fmt.Sprintf("%q is a metric with %d indicators; the row needs a pathway to choose one", label, len(of))
	}
	pw, ok := im.pathways[key(pathway)]
	if !ok {
		return toc.Indicator{}, fmt.Sprintf("unknown pathway %q", pathway)
	}
	var match []toc.Indicator
	for _, ind := range of {
		if slices.Contains(ind.PathwayIDs, pw) {
			match = append(match, ind)
		}
	}
	if len(match) > 1 {
		// An indicator that rolls several pathways up also serves this one;
		// the sheet means the indicator for the pathway alone.
		match = slices.DeleteFunc(slices.Clone(match), func(i toc.Indicator) bool { return len(i.PathwayIDs) != 1 })
	}
	switch len(match) {
	case 1:
		return match[0], ""
	case 0:
		return toc.Indicator{}, fmt.Sprintf("no indicator of metric %q serves pathway %q", label, pathway)
	default:
		return toc.Indicator{}, fmt.Sprintf("%d indicators of metric %q serve pathway %q", len(match), label, pathway)
	}
}

// UnmatchedCSV writes the unmatched rows as a sheet the author can fix and
// import again: the original header plus row and reason columns.
func UnmatchedCSV(res *Result) (string, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	w.UseCRLF = true
	header := append(slices.Clone(res.Header), "row", "reason")
	if err := w.Write(header); err != nil {
		return "", fmt.Errorf("tabular: write unmatched: %w", err)
	}
	for _, u := range res.Unmatched {
		cells := slices.Clone(u.Cells)
		for len(cells) < len(res.Header) {
			cells = append(cells, "")
		}
		if err := w.Write(append(cells[:len(res.Header):len(res.Header)], strconv.Itoa(u.Row), u.Reason)); err != nil {
			return "", fmt.Errorf("tabular: write unmatched: %w", err)
		}
	}
	w.Flush()
	return buf.String(), w.Error()
}
