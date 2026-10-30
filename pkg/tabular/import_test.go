package tabular

import (
	"errors"
	"strings"
	"testing"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/tocio"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/validate"
)

func load(t *testing.T) *toc.Document {
	t.Helper()
	doc, err := tocio.LoadDocument("../../examples/intermediary/theory.yaml")
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

func mapping() *toc.CSVMapping {
	return &toc.CSVMapping{
		TocID: "riverbend-network", Period: "2026-Q4",
		Columns:          toc.CSVColumns{Indicator: "Measure", Date: "Month end", Value: "Value", Pathway: "Program", Quality: "Quality", Source: "Source"},
		IndicatorAliases: map[string]string{"Participant retention (Learning)": "ind-retention-learning"},
		MetricAliases:    map[string]string{"Hours": "met-attendee-hours"},
		DateFormat:       "02/01/2006",
	}
}

func run(t *testing.T, sheet string, m *toc.CSVMapping) *Result {
	t.Helper()
	res, err := Import(strings.NewReader(sheet), m, load(t))
	if err != nil {
		t.Fatal(err)
	}
	return res
}

const header = "Measure,Month end,Value,Program,Quality,Source\r\n"

func TestImportsRowsInAllTheWaysAnIndicatorCanBeNamed(t *testing.T) {
	res := run(t, header+
		"ind-hours-learning,31/10/2026,410,,verified,Sign-in sheets\r\n"+ // by ID
		"Participant retention (Learning),30/11/2026,0.82,,,\r\n"+ // by alias
		"Hours,31/10/2026,120,Shared Services,reported,\r\n"+ // by metric and pathway title
		"attendee hours,31/10/2026,90,pw-civic,,\r\n", // by metric name and pathway ID
		mapping())
	if res.Imported != 4 || len(res.Unmatched) != 0 {
		t.Fatalf("imported %d, unmatched %+v", res.Imported, res.Unmatched)
	}
	var got []string
	for _, o := range res.Set.Observations {
		got = append(got, o.IndicatorID+"@"+o.Date)
	}
	want := []string{"ind-hours-learning@2026-10-31", "ind-retention-learning@2026-11-30", "ind-hours-shared@2026-10-31", "ind-hours-civic@2026-10-31"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("observations = %v, want %v", got, want)
	}
	first := res.Set.Observations[0]
	if *first.Value != 410 || first.Quality != toc.QualityVerified || first.Source != "Sign-in sheets" {
		t.Errorf("first = %+v", first)
	}
	if res.Set.TocID != "riverbend-network" || res.Set.Period != "2026-Q4" {
		t.Errorf("set = %+v", res.Set)
	}
	if issues := validate.Observations(load(t), res.Set); validate.HasErrors(issues) {
		t.Errorf("the imported set does not validate: %+v", issues)
	}
}

func TestEverySeededBadRowIsReportedAndNoneIsDropped(t *testing.T) {
	sheet := header +
		"ind-hours-learning,31/10/2026,410,,,\r\n" + // good
		"ind-ghost,31/10/2026,5,,,\r\n" + // unknown indicator
		"ind-hours-learning,2026-10-31,5,,,\r\n" + // wrong date format
		"ind-hours-learning,31/10/2026,lots,,,\r\n" + // value not a number
		"ind-hours-learning,31/10/2026,,,,\r\n" + // no value
		"ind-hours-learning,31/10/2026,5,,great,\r\n" + // bad quality
		",31/10/2026,5,,,\r\n" + // no indicator
		"Hours,31/10/2026,5,,,\r\n" + // metric with several indicators and no pathway
		"Hours,31/10/2026,5,Nowhere,,\r\n" + // unknown pathway
		"ind-story-peer-help,31/10/2026,5,,,\r\n" + // derived from stories
		"ind-hours-learning,31/10/2026,5\r\n" + // short row
		",,,,,\r\n" + // blank
		"ind-practice-maturity,31/10/2026,Mythical,,,\r\n" + // level the indicator lacks
		"ind-practice-maturity,31/10/2026,established,,,\r\n" // good, qualitative, case-insensitive
	res := run(t, sheet, mapping())

	wantReasons := map[int]string{
		3: "unknown indicator", 4: "does not match the format", 5: "is not a number", 6: "has no value",
		7: "data quality", 8: "names no indicator", 9: "needs a pathway", 10: "unknown pathway",
		11: "story themes", 12: "has 3 cells", 14: "is not one of",
	}
	if len(res.Unmatched) != len(wantReasons) {
		t.Fatalf("%d unmatched, want %d: %+v", len(res.Unmatched), len(wantReasons), res.Unmatched)
	}
	for _, u := range res.Unmatched {
		want, ok := wantReasons[u.Row]
		if !ok || !strings.Contains(u.Reason, want) {
			t.Errorf("row %d: reason %q, want it to contain %q", u.Row, u.Reason, want)
		}
	}
	if res.Imported != 2 || res.Blank != 1 {
		t.Errorf("imported %d blank %d, want 2 and 1", res.Imported, res.Blank)
	}
	if res.Rows() != 14 {
		t.Errorf("rows accounted for = %d, want 14 data rows (0 silent drops)", res.Rows())
	}
	last := res.Set.Observations[1]
	if last.IndicatorID != "ind-practice-maturity" || last.Level != "Established" {
		t.Errorf("qualitative row = %+v", last)
	}
}

func TestFixedIndicatorSheet(t *testing.T) {
	m := &toc.CSVMapping{TocID: "riverbend-network", Columns: toc.CSVColumns{IndicatorFixed: "ind-hours-learning", Date: "d", Value: "v"}}
	res := run(t, "d,v\r\n2026-10-31,10\r\n2026-11-30,20\r\n", m)
	if res.Imported != 2 || res.Set.Observations[1].IndicatorID != "ind-hours-learning" {
		t.Errorf("result = %+v", res)
	}
	bad := run(t, "d,v\r\n2026-10-31,10\r\n", &toc.CSVMapping{Columns: toc.CSVColumns{IndicatorFixed: "ind-ghost", Date: "d", Value: "v"}})
	if len(bad.Unmatched) != 1 || !strings.Contains(bad.Unmatched[0].Reason, "not in the theory") {
		t.Errorf("unknown fixed indicator: %+v", bad.Unmatched)
	}
}

func TestLevelColumn(t *testing.T) {
	m := &toc.CSVMapping{Columns: toc.CSVColumns{Indicator: "i", Date: "d", Level: "l"}}
	res := run(t, "i,d,l\r\nind-practice-maturity,2026-10-31,Emerging\r\nind-practice-maturity,2026-11-30,\r\n", m)
	if res.Imported != 1 || len(res.Unmatched) != 1 || !strings.Contains(res.Unmatched[0].Reason, "no level") {
		t.Errorf("result = %+v", res)
	}
}

func TestMappingErrorsAreWholeFileErrors(t *testing.T) {
	doc := load(t)
	tests := map[string]*toc.CSVMapping{
		"missing column": {Columns: toc.CSVColumns{Indicator: "Nope", Date: "Month end", Value: "Value"}},
		"no date":        {Columns: toc.CSVColumns{Indicator: "Measure", Value: "Value"}},
		"no indicator":   {Columns: toc.CSVColumns{Date: "Month end", Value: "Value"}},
		"no value":       {Columns: toc.CSVColumns{Indicator: "Measure", Date: "Month end"}},
		"wrong theory":   {TocID: "another", Columns: toc.CSVColumns{Indicator: "Measure", Date: "Month end", Value: "Value"}},
	}
	for name, m := range tests {
		if _, err := Import(strings.NewReader(header), m, doc); !errors.Is(err, ErrMapping) {
			t.Errorf("%s: error = %v", name, err)
		}
	}
	if _, err := Import(strings.NewReader(""), mapping(), doc); !errors.Is(err, ErrMapping) {
		t.Errorf("empty sheet: %v", err)
	}
}

func TestHeaderMatchingIgnoresCaseAndAByteOrderMark(t *testing.T) {
	res := run(t, "\xef\xbb\xbfMEASURE,month end,VALUE\r\nind-hours-learning,31/10/2026,5\r\n", &toc.CSVMapping{
		Columns: toc.CSVColumns{Indicator: "measure", Date: "Month End", Value: "value"}, DateFormat: "02/01/2006",
	})
	if res.Imported != 1 {
		t.Errorf("result = %+v", res)
	}
}

func TestUnmatchedRowsRoundTripAsASheet(t *testing.T) {
	res := run(t, header+"ind-ghost,31/10/2026,5,,,\r\n"+"ind-hours-learning,31/10/2026,5\r\n", mapping())
	out, err := UnmatchedCSV(res)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(out), "\r\n")
	if len(lines) != 3 || lines[0] != "Measure,Month end,Value,Program,Quality,Source,row,reason" {
		t.Fatalf("sheet = %q", out)
	}
	if !strings.HasPrefix(lines[1], "ind-ghost,31/10/2026,5,,,,2,") || !strings.Contains(lines[2], ",3,") {
		t.Errorf("rows = %q", lines[1:])
	}
}

func TestImportIsDeterministic(t *testing.T) {
	sheet := header + "ind-hours-learning,31/10/2026,410,,,\r\nHours,30/11/2026,1,pw-civic,,\r\n"
	a, b := run(t, sheet, mapping()), run(t, sheet, mapping())
	ja, _ := tocio.Canonical(a.Set, tocio.FormatJSON)
	jb, _ := tocio.Canonical(b.Set, tocio.FormatJSON)
	if string(ja) != string(jb) {
		t.Error("two imports differ")
	}
}
