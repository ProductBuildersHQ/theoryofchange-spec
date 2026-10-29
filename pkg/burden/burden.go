// Package burden totals what it costs to collect the indicators a theory
// selects: staff minutes per week and questions per respondent per year,
// overall and per pathway and collection mode, against the measurement
// budget. It reads only the document and never the clock, so the same theory
// always gives the same totals.
package burden

import (
	"fmt"
	"math"
	"regexp"
	"slices"
	"strconv"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

// Calendar. Collection periods are converted with a fixed 52-week year, a
// month of one twelfth of it, and a week of seven days, so a weekly period
// is exactly 52 collections a year and a monthly one exactly 12.
const (
	weeksPerYear  = 52.0
	monthsPerYear = 12.0
	daysPerWeek   = 7.0
	hoursPerDay   = 24.0
)

var durationRE = regexp.MustCompile(`^P(?:(\d+)Y)?(?:(\d+)M)?(?:(\d+)W)?(?:(\d+)D)?(?:T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?)?$`)

// CollectionsPerYear returns how many times a period of the given ISO 8601
// duration fits in a year.
func CollectionsPerYear(period string) (float64, error) {
	m := durationRE.FindStringSubmatch(period)
	if m == nil || period == "P" || period[len(period)-1] == 'T' {
		return 0, fmt.Errorf("burden: %q is not an ISO 8601 duration", period)
	}
	n := func(i int) float64 {
		if m[i] == "" {
			return 0
		}
		v, err := strconv.ParseFloat(m[i], 64)
		if err != nil {
			return 0
		}
		return v
	}
	weeks := n(1)*weeksPerYear + n(2)*weeksPerYear/monthsPerYear + n(3) + n(4)/daysPerWeek +
		n(5)/(hoursPerDay*daysPerWeek) + n(6)/(60*hoursPerDay*daysPerWeek) + n(7)/(3600*hoursPerDay*daysPerWeek)
	if weeks == 0 {
		return 0, fmt.Errorf("burden: duration %q is zero", period)
	}
	return weeksPerYear / weeks, nil
}

// Totals is the cost of a set of indicators.
type Totals struct {
	Indicators                    int     `json:"indicators"`
	MinutesPerWeek                float64 `json:"minutesPerWeek"`
	QuestionsPerRespondentPerYear float64 `json:"questionsPerRespondentPerYear"`
}

// Group is the totals for one pathway or one collection mode.
type Group struct {
	// Key is the pathway ID or the mode; empty for indicators with no
	// pathway or no mode.
	Key string `json:"key"`
	Totals
}

// Check compares one total with its declared limit.
type Check struct {
	// Declared is false when the budget sets no limit.
	Declared bool    `json:"declared"`
	Limit    float64 `json:"limit,omitempty"`
	Actual   float64 `json:"actual"`
	Over     bool    `json:"over"`
}

// Budget compares the totals with the declared measurement budget.
type Budget struct {
	StaffMinutesPerWeek Check `json:"staffMinutesPerWeek"`
	QuestionsPerYear    Check `json:"questionsPerRespondentPerYear"`
	// SelectedMetrics compares the number of selected indicators with the
	// target range: Min and Max are those declared, Within is true when the
	// count is inside it (or no range is declared).
	SelectedMetrics struct {
		Declared bool `json:"declared"`
		Min      int  `json:"min,omitempty"`
		Max      int  `json:"max,omitempty"`
		Actual   int  `json:"actual"`
		Within   bool `json:"within"`
	} `json:"selectedMetrics"`
}

// Report is the collection burden of a theory.
type Report struct {
	// Overall counts each selected indicator once.
	Overall Totals `json:"overall"`
	// ByPathway lists the pathways with selected indicators, in document
	// order, then indicators with no pathway under an empty key. An
	// indicator serving several pathways counts in each.
	ByPathway []Group `json:"byPathway"`
	// ByMode lists collection modes in the order of toc.AllCollectionModes,
	// then indicators with no collection under an empty key.
	ByMode []Group `json:"byMode"`
	// Uncosted lists selected indicators with no collection data, which the
	// totals therefore leave out.
	Uncosted []string `json:"uncosted"`
	// Budget is nil when the theory declares none.
	Budget *Budget `json:"budget,omitempty"`
}

// Compute totals the selected indicators of doc. Retired and candidate
// indicators, and indicators with no selection record, are not counted.
func Compute(doc *toc.Document) (*Report, error) {
	r := &Report{ByPathway: []Group{}, ByMode: []Group{}, Uncosted: []string{}}
	byPathway := map[string]*Totals{}
	byMode := map[string]*Totals{}
	for _, ind := range doc.Indicators {
		if ind.Selection == nil || ind.Selection.Status != toc.SelectionSelected || ind.Provenance.IsRejected() {
			continue
		}
		cost, err := costOf(ind)
		if err != nil {
			return nil, err
		}
		cost.Indicators = 1
		add(&r.Overall, cost)
		mode := ""
		if ind.Collection == nil {
			r.Uncosted = append(r.Uncosted, ind.ID)
		} else {
			mode = string(ind.Collection.Mode)
		}
		addTo(byMode, mode, cost)
		if len(ind.PathwayIDs) == 0 {
			addTo(byPathway, "", cost)
		}
		for _, id := range ind.PathwayIDs {
			addTo(byPathway, id, cost)
		}
	}
	var pathwayOrder []string
	for _, p := range doc.Pathways {
		pathwayOrder = append(pathwayOrder, p.ID)
	}
	r.ByPathway = groups(byPathway, pathwayOrder)
	var modeOrder []string
	for _, m := range toc.AllCollectionModes() {
		modeOrder = append(modeOrder, string(m))
	}
	r.ByMode = groups(byMode, modeOrder)
	r.Budget = compare(doc.Budget, r)
	return r, nil
}

func costOf(ind toc.Indicator) (Totals, error) {
	c := ind.Collection
	if c == nil || c.Period == "" {
		return Totals{}, nil
	}
	perYear, err := CollectionsPerYear(c.Period)
	if err != nil {
		return Totals{}, fmt.Errorf("indicator %s: %w", ind.ID, err)
	}
	sample := c.SampleRate
	if sample == 0 {
		sample = 1
	}
	return Totals{
		MinutesPerWeek:                round(c.MinutesPerPeriod * perYear / weeksPerYear),
		QuestionsPerRespondentPerYear: round(float64(c.QuestionsPerRespondent) * perYear * sample),
	}, nil
}

func add(t *Totals, c Totals) {
	t.Indicators += c.Indicators
	t.MinutesPerWeek = round(t.MinutesPerWeek + c.MinutesPerWeek)
	t.QuestionsPerRespondentPerYear = round(t.QuestionsPerRespondentPerYear + c.QuestionsPerRespondentPerYear)
}

func addTo(m map[string]*Totals, key string, c Totals) {
	t, ok := m[key]
	if !ok {
		t = &Totals{}
		m[key] = t
	}
	add(t, c)
}

// groups lists totals in the given key order, then any other keys sorted,
// with the empty key last.
func groups(m map[string]*Totals, order []string) []Group {
	out := []Group{}
	seen := map[string]bool{}
	emit := func(k string) {
		if t, ok := m[k]; ok && !seen[k] {
			seen[k] = true
			out = append(out, Group{Key: k, Totals: *t})
		}
	}
	for _, k := range order {
		emit(k)
	}
	var rest []string
	for k := range m {
		if !seen[k] && k != "" {
			rest = append(rest, k)
		}
	}
	slices.Sort(rest)
	for _, k := range rest {
		emit(k)
	}
	emit("")
	return out
}

func compare(b *toc.MeasurementBudget, r *Report) *Budget {
	if b == nil {
		return nil
	}
	out := &Budget{}
	out.StaffMinutesPerWeek = check(b.StaffMinutesPerWeek, r.Overall.MinutesPerWeek)
	out.QuestionsPerYear = check(float64(b.MaxQuestionsPerRespondentPerYear), r.Overall.QuestionsPerRespondentPerYear)
	sm := &out.SelectedMetrics
	sm.Actual, sm.Within = r.Overall.Indicators, true
	if t := b.TargetSelectedMetrics; t != nil {
		sm.Declared, sm.Min, sm.Max = true, t.Min, t.Max
		sm.Within = sm.Actual >= t.Min && (t.Max == 0 || sm.Actual <= t.Max)
	}
	return out
}

func check(limit, actual float64) Check {
	if limit == 0 {
		return Check{Actual: actual}
	}
	return Check{Declared: true, Limit: limit, Actual: actual, Over: actual > limit}
}

// round removes floating-point noise so totals print and compare cleanly.
func round(v float64) float64 { return math.Round(v*1e6) / 1e6 }
