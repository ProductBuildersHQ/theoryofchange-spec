package judge

import (
	"fmt"
	"slices"

	"github.com/plexusone/structured-evaluation/rubric"
)

// Aggregations lists the ways judges' scores combine.
func Aggregations() []rubric.AggregationMethod {
	return []rubric.AggregationMethod{
		rubric.AggregationMajority, rubric.AggregationConservative,
		rubric.AggregationOptimistic, rubric.AggregationUnanimous,
	}
}

// CategorySummary is how the judges scored one category.
type CategorySummary struct {
	Category string `json:"category"`
	Name     string `json:"name"`
	Required bool   `json:"required,omitempty"`
	Blocking bool   `json:"blocking,omitempty"`
	// Scores counts the judges' scores: pass, partial, fail.
	Scores map[string]int `json:"scores"`
	// Result is the combined score under the chosen aggregation.
	Result rubric.ScoreValue `json:"result"`
	// Agreed is true when every judge gave the same score.
	Agreed bool `json:"agreed"`
}

// Summary combines the verdicts for a bundle.
type Summary struct {
	BundleHash  string                   `json:"bundleHash"`
	Profile     string                   `json:"profile"`
	Aggregation rubric.AggregationMethod `json:"aggregation"`
	Judges      []string                 `json:"judges"`
	Categories  []CategorySummary        `json:"categories"`
	// Agreement is the share of categories on which the judges agree exactly,
	// from 0 to 1; with one judge it is 1.
	Agreement float64 `json:"agreement"`
	// Disagreements lists the categories where the judges differ.
	Disagreements []string `json:"disagreements"`
	// Pass is true when every required category passes, no blocking category
	// fails, and the findings stay within the rubric's limits.
	Pass bool `json:"pass"`
	// Reasons say why the impact model does not pass.
	Reasons []string `json:"reasons"`
}

// Summarize combines verdicts, given in judge-name order, under an
// aggregation. It needs at least one verdict. The categories follow the
// rubric's order, so the result is the same on every run.
func Summarize(b *Bundle, names []string, verdicts map[string]*Verdict, method rubric.AggregationMethod) (*Summary, error) {
	if len(names) == 0 {
		return nil, fmt.Errorf("judge: no verdicts recorded for this version of the impact model")
	}
	if !slices.Contains(Aggregations(), method) {
		return nil, fmt.Errorf("judge: unknown aggregation %q (want one of %v)", method, Aggregations())
	}
	reports := make([]*rubric.Rubric, len(names))
	for i, n := range names {
		reports[i] = &verdicts[n].Rubric
	}
	multi := rubric.AggregateEvaluations(reports, method)
	combined := map[string]rubric.ScoreValue{}
	for _, c := range multi.AggregatedCategories {
		combined[c.Category] = c.Score
	}

	s := &Summary{
		BundleHash: b.BundleHash, Profile: b.Profile, Aggregation: method, Judges: slices.Clone(names),
		Categories: []CategorySummary{}, Disagreements: []string{}, Reasons: []string{}, Pass: true,
	}
	for _, rc := range b.Rubric.Categories {
		cs := CategorySummary{Category: rc.ID, Name: rc.Name, Required: rc.Required, Blocking: rc.Blocking,
			Scores: map[string]int{"pass": 0, "partial": 0, "fail": 0}}
		for _, r := range reports {
			for _, c := range r.Categories {
				if c.Category == rc.ID {
					cs.Scores[string(c.Score)]++
				}
			}
		}
		cs.Result = combined[rc.ID]
		cs.Agreed = distinct(cs.Scores) <= 1
		s.Categories = append(s.Categories, cs)
		if !cs.Agreed {
			s.Disagreements = append(s.Disagreements, rc.ID)
		}
		if cs.Required && cs.Result != rubric.ScorePass {
			s.fail("required category %q is %s", rc.ID, cs.Result)
		}
		if cs.Blocking && cs.Result == rubric.ScoreFail {
			s.fail("blocking category %q fails", rc.ID)
		}
	}
	s.Agreement = 1
	if n := len(s.Categories); n > 0 {
		s.Agreement = float64(n-len(s.Disagreements)) / float64(n)
	}
	s.checkFindings(b, multi.ConsolidatedFindings)
	return s, nil
}

func distinct(scores map[string]int) int {
	n := 0
	for _, c := range scores {
		if c > 0 {
			n++
		}
	}
	return n
}

func (s *Summary) fail(format string, args ...any) {
	s.Pass = false
	s.Reasons = append(s.Reasons, fmt.Sprintf(format, args...))
}

// checkFindings holds the consolidated findings to the rubric's limits; a
// limit of -1 means any number.
func (s *Summary) checkFindings(b *Bundle, findings []rubric.Finding) {
	lim := b.Rubric.PassCriteria.MaxFindings
	if lim == nil {
		return
	}
	count := map[rubric.Severity]int{}
	for _, f := range findings {
		count[f.Severity]++
	}
	for _, c := range []struct {
		sev   rubric.Severity
		limit int
	}{{rubric.SeverityCritical, lim.Critical}, {rubric.SeverityHigh, lim.High}, {rubric.SeverityMedium, lim.Medium}, {rubric.SeverityLow, lim.Low}} {
		if c.limit >= 0 && count[c.sev] > c.limit {
			s.fail("%d %s finding(s); the rubric allows %d", count[c.sev], c.sev, c.limit)
		}
	}
}
