package report

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/ProductBuildersHQ/impact-model/pkg/burden"
	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
)

// invertedScore flips a 1 to 5 score, so a criterion where low is better
// (such as cost to collect) counts in the same direction as the others.
const invertedScore = 6

// WeightedScore combines scores into one number from 1 to 5, rounded to two
// places. Each criterion counts by its weight (a criterion with no weight
// counts as 1), and an inverted criterion counts as 6 minus its score. It
// reports false when nothing could be scored, and ignores scores that name
// an unknown criterion, which validation reports separately.
func WeightedScore(doc *impact.Model, scores []impact.CriterionScore) (float64, bool) {
	var sum, weights float64
	for _, s := range scores {
		cr, ok := doc.Criterion(s.CriterionID)
		if !ok {
			continue
		}
		w := cr.Weight
		if w == 0 {
			w = 1
		}
		v := float64(s.Score)
		if cr.Inverted {
			v = invertedScore - v
		}
		sum += w * v
		weights += w
	}
	if weights == 0 {
		return 0, false
	}
	return math.Round(sum/weights*100) / 100, true
}

func scoreText(score float64, ok bool) string {
	if !ok {
		return "—"
	}
	return strconv.FormatFloat(score, 'f', 2, 64)
}

// scored is one subject with its scores, ready to tabulate.
type scored struct {
	id, name, status, rationale string
	scores                      []impact.CriterionScore
	weighted                    float64
	ok                          bool
}

// byScore sorts the best score first and unscored subjects last, breaking
// ties by ID so the order is stable.
func byScore(rows []scored) {
	slices.SortStableFunc(rows, func(a, b scored) int {
		switch {
		case a.ok != b.ok:
			if a.ok {
				return -1
			}
			return 1
		case a.weighted != b.weighted:
			if a.weighted > b.weighted {
				return -1
			}
			return 1
		default:
			return strings.Compare(a.id, b.id)
		}
	})
}

func criteriaFor(doc *impact.Model, subject impact.SelectionSubject) []impact.SelectionCriterion {
	var out []impact.SelectionCriterion
	for _, c := range doc.SelectionCriteria {
		if c.AppliesTo == subject {
			out = append(out, c)
		}
	}
	return out
}

func criterionHeader(c impact.SelectionCriterion) string {
	h := MDCell(c.Name)
	if c.Inverted {
		h += " (lower is better)"
	}
	return h
}

func scoreCell(scores []impact.CriterionScore, id string) string {
	for _, s := range scores {
		if s.CriterionID == id {
			return strconv.Itoa(s.Score)
		}
	}
	return "—"
}

func selectionRows(doc *impact.Model) []scored {
	var rows []scored
	for _, ind := range doc.Indicators {
		if ind.Selection == nil || ind.Provenance.IsRejected() {
			continue
		}
		w, ok := WeightedScore(doc, ind.Selection.Scores)
		rows = append(rows, scored{ind.ID, ind.Name, string(ind.Selection.Status), ind.Selection.Rationale, ind.Selection.Scores, w, ok})
	}
	byScore(rows)
	return rows
}

func priorityRows(doc *impact.Model, subject impact.SelectionSubject) []scored {
	var rows []scored
	add := func(id, name string, p *impact.Priority, rejected bool) {
		if p == nil || rejected {
			return
		}
		w, ok := WeightedScore(doc, p.Scores)
		rows = append(rows, scored{id, name, "", p.Rationale, p.Scores, w, ok})
	}
	switch subject {
	case impact.SubjectOutcome:
		for _, o := range doc.Outcomes {
			add(o.ID, o.Title, o.Priority, o.Provenance.IsRejected())
		}
	case impact.SubjectIntervention:
		for _, in := range doc.Interventions {
			add(in.ID, in.Title, in.Priority, in.Provenance.IsRejected())
		}
	case impact.SubjectIndicator:
	}
	byScore(rows)
	return rows
}

// scoreTable renders subjects against their criteria. The status column is
// shown only for indicators.
func scoreTable(b *strings.Builder, heading string, criteria []impact.SelectionCriterion, rows []scored, withStatus bool) {
	cols := []string{heading}
	if withStatus {
		cols = append(cols, "Status")
	}
	for _, c := range criteria {
		cols = append(cols, criterionHeader(c))
	}
	cols = append(cols, "Weighted score", "Rationale")
	b.WriteString("| " + strings.Join(cols, " | ") + " |\n")
	b.WriteString("|" + strings.Repeat("---|", len(cols)) + "\n")
	for _, r := range rows {
		cells := []string{MDCell(r.name) + " (`" + r.id + "`)"}
		if withStatus {
			cells = append(cells, Label(r.status))
		}
		for _, c := range criteria {
			cells = append(cells, scoreCell(r.scores, c.ID))
		}
		cells = append(cells, scoreText(r.weighted, r.ok), orDash(MDCell(r.rationale)))
		b.WriteString("| " + strings.Join(cells, " | ") + " |\n")
	}
	b.WriteString("\n")
}

// MetricSelection renders the candidate indicators against the selection
// criteria, best weighted score first, with status and rationale. Scores
// run from 1 (poor) to 5 (strong); a criterion marked "lower is better" is
// flipped before it is weighted. Outcome and intervention priorities follow
// when they have been scored.
func MetricSelection(v *View, doc *impact.Model) string {
	var b strings.Builder
	fmt.Fprintf(&b, "## Metric selection\n\n")
	rows := selectionRows(doc)
	if len(rows) == 0 {
		b.WriteString("No indicator has a selection record yet.\n\n")
		return b.String()
	}
	counts := map[string]int{}
	for _, r := range rows {
		counts[r.status]++
	}
	var parts []string
	for _, st := range impact.AllSelectionStatuses() {
		if n := counts[string(st)]; n > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", n, st))
		}
	}
	fmt.Fprintf(&b, "%s. Scores run from 1 (poor) to 5 (strong); the weighted score is the average of the scores by criterion weight.\n\n", strings.Join(parts, ", "))
	if bg := doc.Budget; bg != nil && bg.TargetSelectedMetrics != nil {
		fmt.Fprintf(&b, "Target: between %d and %d selected metrics.\n\n", bg.TargetSelectedMetrics.Min, bg.TargetSelectedMetrics.Max)
	}
	b.WriteString("### Indicators\n\n")
	scoreTable(&b, "Indicator", criteriaFor(doc, impact.SubjectIndicator), rows, true)
	for _, s := range []struct {
		subject impact.SelectionSubject
		heading string
	}{{impact.SubjectOutcome, "Outcome"}, {impact.SubjectIntervention, "Intervention"}} {
		if pr := priorityRows(doc, s.subject); len(pr) > 0 {
			fmt.Fprintf(&b, "### %s priorities\n\n", s.heading)
			scoreTable(&b, s.heading, criteriaFor(doc, s.subject), pr, false)
		}
	}
	return b.String()
}

func pathwayName(doc *impact.Model, id string) string {
	if id == "" {
		return "No pathway"
	}
	if p, ok := doc.Pathway(id); ok {
		return MDCell(p.Title)
	}
	return id
}

func modeName(m string) string {
	if m == "" {
		return "No collection data"
	}
	return Label(m)
}

func num(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }

func rounded(v float64) string { return strconv.FormatFloat(math.Round(v*100)/100, 'f', -1, 64) }

// CollectionPlan renders what it costs to collect the selected indicators:
// each one's mode, time, and questions, the totals overall, per pathway and
// per mode, and the comparison with the measurement budget. The numbers are
// those of the burden report, which the caller computes once.
func CollectionPlan(doc *impact.Model, rep *burden.Report) string {
	var b strings.Builder
	b.WriteString("## Collection plan\n\n")
	if len(rep.Lines) == 0 {
		b.WriteString("No indicator is selected yet.\n\n")
		return b.String()
	}
	b.WriteString("| Indicator | Pathways | How collected | Where | Staff minutes each time | How often | Staff minutes per week | Questions per respondent per year |\n")
	b.WriteString("|-----------|----------|---------------|-------|-------------------------|-----------|------------------------|-----------------------------------|\n")
	for _, l := range rep.Lines {
		ind, _ := doc.Indicator(l.ID)
		var pathways []string
		for _, id := range ind.PathwayIDs {
			pathways = append(pathways, pathwayName(doc, id))
		}
		how, system, minutes, period := "—", "—", "—", "—"
		if c := ind.Collection; c != nil {
			how = Label(string(c.Mode))
			system = orDash(MDCell(c.System))
			if c.MinutesPerPeriod > 0 {
				minutes = num(c.MinutesPerPeriod)
			}
			period = orDash(c.Period)
		}
		fmt.Fprintf(&b, "| %s (`%s`) | %s | %s | %s | %s | %s | %s | %s |\n", MDCell(ind.Name), ind.ID,
			orDash(strings.Join(pathways, "; ")), how, system, minutes, period, rounded(l.MinutesPerWeek), rounded(l.QuestionsPerRespondentPerYear))
	}
	b.WriteString("\n")
	if len(rep.Uncosted) > 0 {
		fmt.Fprintf(&b, "No collection cost is recorded for %s, so the totals leave them out.\n\n", strings.Join(rep.Uncosted, ", "))
	}

	totals := func(heading, first string, groups []burden.Group, name func(string) string) {
		fmt.Fprintf(&b, "### %s\n\n| %s | Indicators | Staff minutes per week | Questions per respondent per year |\n|---|---|---|---|\n", heading, first)
		for _, g := range groups {
			fmt.Fprintf(&b, "| %s | %d | %s | %s |\n", name(g.Key), g.Indicators, rounded(g.MinutesPerWeek), rounded(g.QuestionsPerRespondentPerYear))
		}
		b.WriteString("\n")
	}
	fmt.Fprintf(&b, "### Totals\n\nThe %d selected indicator(s) take %s staff minutes a week and ask each respondent %s questions a year. An indicator that serves several pathways counts in each pathway below, and once in the totals.\n\n",
		rep.Overall.Indicators, rounded(rep.Overall.MinutesPerWeek), rounded(rep.Overall.QuestionsPerRespondentPerYear))
	totals("By pathway", "Pathway", rep.ByPathway, func(id string) string { return pathwayName(doc, id) })
	totals("By collection mode", "Mode", rep.ByMode, modeName)

	if bg := rep.Budget; bg != nil {
		b.WriteString("### Against the budget\n\n| Measure | Budget | Planned | Within budget |\n|---|---|---|---|\n")
		row := func(name string, c burden.Check) {
			if !c.Declared {
				fmt.Fprintf(&b, "| %s | not set | %s | — |\n", name, rounded(c.Actual))
				return
			}
			fmt.Fprintf(&b, "| %s | %s | %s | %s |\n", name, rounded(c.Limit), rounded(c.Actual), yesNo(!c.Over))
		}
		row("Staff minutes per week", bg.StaffMinutesPerWeek)
		row("Questions per respondent per year", bg.QuestionsPerYear)
		sm := bg.SelectedMetrics
		if sm.Declared {
			fmt.Fprintf(&b, "| Selected metrics | %d to %d | %d | %s |\n", sm.Min, sm.Max, sm.Actual, yesNo(sm.Within))
		}
		b.WriteString("\n")
	}
	return b.String()
}

func yesNo(v bool) string {
	if v {
		return "Yes"
	}
	return "No"
}

// ---- CSV ------------------------------------------------------------------

func metricSelectionCSV(doc *impact.Model) (string, error) {
	criteria := criteriaFor(doc, impact.SubjectIndicator)
	header := []string{"id", "name", "status"}
	for _, c := range criteria {
		header = append(header, c.ID)
	}
	header = append(header, "weighted_score", "rationale")
	var rows [][]string
	for _, r := range selectionRows(doc) {
		row := []string{r.id, r.name, r.status}
		for _, c := range criteria {
			cell := scoreCell(r.scores, c.ID)
			if cell == "—" {
				cell = ""
			}
			row = append(row, cell)
		}
		score := ""
		if r.ok {
			score = strconv.FormatFloat(r.weighted, 'f', 2, 64)
		}
		rows = append(rows, append(row, score, r.rationale))
	}
	return writeCSV(header, rows)
}

func collectionPlanCSV(doc *impact.Model, rep *burden.Report) (string, error) {
	rows := make([][]string, 0, len(rep.Lines))
	for _, l := range rep.Lines {
		ind, _ := doc.Indicator(l.ID)
		row := []string{ind.ID, ind.Name, join(ind.PathwayIDs), "", "", "", "", "", "", rounded(l.MinutesPerWeek), rounded(l.QuestionsPerRespondentPerYear)}
		if c := ind.Collection; c != nil {
			row[3], row[4] = string(c.Mode), c.System
			if c.MinutesPerPeriod > 0 {
				row[5] = num(c.MinutesPerPeriod)
			}
			row[6] = c.Period
			if c.QuestionsPerRespondent > 0 {
				row[7] = strconv.Itoa(c.QuestionsPerRespondent)
			}
			if c.SampleRate > 0 {
				row[8] = num(c.SampleRate)
			}
		}
		rows = append(rows, row)
	}
	return writeCSV([]string{
		"id", "name", "pathways", "mode", "system", "minutes_per_period", "period",
		"questions_per_respondent", "sample_rate", "minutes_per_week", "questions_per_year",
	}, rows)
}
