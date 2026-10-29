// Package measure turns observations into status: how far each indicator has
// moved, whether that is on track, and what that means for outcomes and
// pathways. Everything here is a pure function of the theory, the
// observations, and the as-of date; it never reads the clock.
package measure

import (
	"math"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"time"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

// DefaultTolerance is how far behind expected progress an indicator may fall
// and still be on track.
const DefaultTolerance = 0.10

// maxProgress caps displayed progress, so a target beaten many times over
// does not distort a chart.
const maxProgress = 1.5

// Options configure indicator evaluation.
type Options struct {
	// AsOf is the date the status describes.
	AsOf time.Time
	// Tolerance defaults to DefaultTolerance.
	Tolerance float64
}

func (o Options) tolerance() float64 {
	if o.Tolerance > 0 {
		return o.Tolerance
	}
	return DefaultTolerance
}

// point is a value on the indicator's scale: a number for quantitative
// indicators, or the position of a level for qualitative ones.
type point struct {
	value float64
	ok    bool
}

// scale reads values of one indicator.
type scale struct {
	levels []string
}

func (s scale) fromValue(v *float64, level string) point {
	if v != nil {
		return point{*v, true}
	}
	if level != "" {
		if i := slices.Index(s.levels, level); i >= 0 {
			return point{float64(i), true}
		}
	}
	return point{}
}

// EvaluateIndicator computes an indicator's status from its observations.
// Observations of other indicators, and observations after the as-of date,
// are ignored. The latest remaining observation is the current value.
//
// Progress is how far the current value has moved from the baseline toward
// the final target, direction-aware, capped at 1.5. Expected progress is the
// share of that movement the time-phased targets call for by the as-of date.
// The status follows the order of the table in the specification: no targets
// is baseline only; no observation is not started or no data; an old
// observation is stale; otherwise achieved, on track, at risk, or off track by
// how progress compares with expected.
func EvaluateIndicator(ind toc.Indicator, observations []toc.Observation, o Options) toc.IndicatorStatus {
	s := scale{levels: ind.Levels}
	latest, count := latestObservation(ind, s, observations, o.AsOf)
	st := toc.IndicatorStatus{IndicatorID: ind.ID, ObservationCount: count}
	if latest != nil {
		obs := *latest
		st.Latest = &obs
	}

	targets := sortedTargets(ind.Targets)
	base := point{}
	if ind.Baseline != nil {
		base = s.fromValue(ind.Baseline.Value, ind.Baseline.Level)
	}

	switch {
	case len(targets) == 0:
		st.State = baselineOnlyOrNoData(ind, latest)
	case latest == nil:
		st.State = toc.StateNoData
		if first, ok := parseDate(targets[0].Date); ok && o.AsOf.Before(first) {
			st.State = toc.StateNotStarted
		}
	case !base.ok:
		// Progress is measured from a baseline; without one only the
		// current value can be reported.
		st.State = toc.StateBaselineOnly
	default:
		evaluate(&st, ind, s, base, targets, latest, o)
	}
	st.Health = Health(st.State)
	return st
}

func baselineOnlyOrNoData(ind toc.Indicator, latest *toc.Observation) toc.State {
	if ind.Baseline != nil || latest != nil {
		return toc.StateBaselineOnly
	}
	return toc.StateNoData
}

// evaluate fills in progress, expected progress, and the state of an
// indicator that has a baseline, targets, and an observation.
func evaluate(st *toc.IndicatorStatus, ind toc.Indicator, s scale, base point, targets []toc.Target, latest *toc.Observation, o Options) {
	cur := s.fromValue(latest.Value, latest.Level)
	final := s.fromValue(targets[len(targets)-1].Value, targets[len(targets)-1].Level)
	if !cur.ok || !final.ok {
		st.State = toc.StateBaselineOnly
		return
	}
	p := progress(ind, base.value, cur.value, final.value)
	st.Progress = &p
	e := expected(s, ind.Baseline, base, targets, final, o.AsOf, ind.Direction)
	st.Expected = &e

	tol := o.tolerance()
	switch {
	case isStale(ind.Frequency, latest.Date, o.AsOf):
		st.State = toc.StateStale
	case p >= 1:
		st.State = toc.StateAchieved
	case p >= e-tol:
		st.State = toc.StateOnTrack
	case p >= e-2*tol:
		st.State = toc.StateAtRisk
	default:
		st.State = toc.StateOffTrack
	}
}

// direction returns the indicator's desired direction. When none is stated it
// follows from where the target lies relative to the baseline.
func direction(d toc.Direction, base, final float64) toc.Direction {
	if d != "" {
		return d
	}
	switch {
	case final > base:
		return toc.DirectionIncrease
	case final < base:
		return toc.DirectionDecrease
	default:
		return toc.DirectionMaintain
	}
}

// progress is the share of the way from baseline b to target t that the
// current value c has come, capped to [0, maxProgress].
func progress(ind toc.Indicator, b, c, t float64) float64 {
	dir := direction(ind.Direction, b, t)
	if dir == toc.DirectionMaintain {
		if math.Abs(c-t) <= ind.Tolerance {
			return 1
		}
		return 0
	}
	if t == b {
		// The target is the baseline: it holds when the value has not moved
		// the wrong way.
		if (dir == toc.DirectionIncrease && c >= t) || (dir == toc.DirectionDecrease && c <= t) {
			return 1
		}
		return 0
	}
	var p float64
	if dir == toc.DirectionDecrease {
		p = (b - c) / (b - t)
	} else {
		p = (c - b) / (t - b)
	}
	return math.Min(math.Max(p, 0), maxProgress)
}

// expected is the progress the targets call for at asOf: a straight line
// from the baseline date (progress 0) through each target's own share of the
// final change, flat before the first date and 1 after the last.
func expected(s scale, baseline *toc.Baseline, base point, targets []toc.Target, final point, asOf time.Time, dir toc.Direction) float64 {
	type step struct {
		at time.Time
		p  float64
	}
	var steps []step
	if baseline != nil {
		if at, ok := parseDate(baseline.Date); ok {
			steps = append(steps, step{at, 0})
		}
	}
	d := direction(dir, base.value, final.value)
	for _, t := range targets {
		at, ok := parseDate(t.Date)
		v := s.fromValue(t.Value, t.Level)
		if !ok || !v.ok {
			continue
		}
		steps = append(steps, step{at, share(d, base.value, v.value, final.value)})
	}
	if len(steps) == 0 {
		return 0
	}
	if !asOf.After(steps[0].at) {
		return steps[0].p
	}
	last := steps[len(steps)-1]
	if !asOf.Before(last.at) {
		return 1
	}
	for i := 1; i < len(steps); i++ {
		a, b := steps[i-1], steps[i]
		if asOf.Before(b.at) {
			span := b.at.Sub(a.at).Hours()
			if span <= 0 {
				return b.p
			}
			return a.p + (b.p-a.p)*asOf.Sub(a.at).Hours()/span
		}
	}
	return 1
}

// share is how much of the full baseline-to-final change a target value
// represents. A maintain indicator calls for the whole thing at its first
// target date.
func share(dir toc.Direction, b, v, final float64) float64 {
	if dir == toc.DirectionMaintain || final == b {
		return 1
	}
	return math.Min(math.Max((v-b)/(final-b), 0), maxProgress)
}

func sortedTargets(ts []toc.Target) []toc.Target {
	out := slices.Clone(ts)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Date < out[j].Date })
	return out
}

// latestObservation returns the most recent usable observation of the
// indicator at or before asOf, and how many such observations there are. An
// observation is usable when it has a valid date and a value the indicator's
// scale understands. Of two on the same day, the later one in the input wins.
func latestObservation(ind toc.Indicator, s scale, observations []toc.Observation, asOf time.Time) (*toc.Observation, int) {
	var latest *toc.Observation
	var latestAt time.Time
	count := 0
	for i := range observations {
		ob := &observations[i]
		if ob.IndicatorID != ind.ID {
			continue
		}
		at, ok := parseDate(ob.Date)
		if !ok || at.After(asOf) || !s.fromValue(ob.Value, ob.Level).ok {
			continue
		}
		count++
		if latest == nil || !at.Before(latestAt) {
			latest, latestAt = ob, at
		}
	}
	return latest, count
}

// Health maps a state to a health color: achieved and on track are green, at
// risk and stale are yellow, off track is red, and states with no judgment
// have none.
func Health(s toc.State) string {
	switch s {
	case toc.StateAchieved, toc.StateOnTrack:
		return "green"
	case toc.StateAtRisk, toc.StateStale:
		return "yellow"
	case toc.StateOffTrack:
		return "red"
	case toc.StateNotStarted, toc.StateNoData, toc.StateBaselineOnly, toc.StateUnmeasured:
		return ""
	}
	return ""
}

func parseDate(s string) (time.Time, bool) {
	t, err := time.Parse(time.DateOnly, s)
	return t, err == nil
}

var durationRE = regexp.MustCompile(`^P(?:(\d+)Y)?(?:(\d+)M)?(?:(\d+)W)?(?:(\d+)D)?(?:T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?)?$`)

// addDuration returns t plus n times an ISO 8601 duration, using calendar
// months and years. It reports false if the duration is not valid or is
// zero.
func addDuration(t time.Time, iso string, n int) (time.Time, bool) {
	m := durationRE.FindStringSubmatch(iso)
	if m == nil || iso == "P" {
		return t, false
	}
	num := func(i int) int {
		v, err := strconv.Atoi(m[i])
		if err != nil {
			return 0
		}
		return v
	}
	years, months, days := num(1), num(2), num(3)*7+num(4)
	clock := time.Duration(num(5))*time.Hour + time.Duration(num(6))*time.Minute + time.Duration(num(7))*time.Second
	if years == 0 && months == 0 && days == 0 && clock == 0 {
		return t, false
	}
	return t.AddDate(n*years, n*months, n*days).Add(time.Duration(n) * clock), true
}

// isStale reports whether the last observation is older than twice the
// indicator's frequency. An indicator with no valid frequency is never stale.
func isStale(frequency, lastDate string, asOf time.Time) bool {
	last, ok := parseDate(lastDate)
	if !ok {
		return false
	}
	limit, ok := addDuration(last, frequency, 2)
	return ok && asOf.After(limit)
}
