package prismx

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/grokify/prism-roadmap/goals/dmaic"

	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
)

// ErrNotDMAICReady is returned, wrapped, when an impact model has nothing to
// draw as DMAIC: no pathway marked for it, or no indicator of such a pathway
// with a numeric baseline and target.
var ErrNotDMAICReady = errors.New("the impact model cannot be drawn as DMAIC")

// Omission is an indicator of a DMAIC pathway that the export leaves out,
// and why. DMAIC metrics are numbers with a baseline and a target, so an
// indicator without them cannot be one.
type Omission struct {
	IndicatorID string `json:"indicatorId"`
	Reason      string `json:"reason"`
}

// ToDMAIC exports the pathways marked `projections: [dmaic]` as a DMAIC
// metrics document.
//
//   - Each marked pathway is a category, and each of its indicators with a
//     numeric baseline and target is a metric. Others are returned as
//     omissions.
//   - A metric's data points are its undisaggregated observations, and its
//     current value is the latest of them, or the baseline before any. Control limits are computed from
//     the observations, as an individuals chart, only when there are at least
//     impact.DMAICMinPoints of them. Process capability is never filled: it
//     needs specification limits, which an impact model does not record.
//   - A metric's root causes are the causes its outcome addresses, and its
//     initiatives are the pathway's interventions aimed at that outcome.
//   - A metric's phase is derived, as a convention: Control once the target is
//     achieved; Measure while there are too few observations for control
//     limits; Improve while an intervention aimed at it is in progress;
//     otherwise Analyze.
//
// st and sets may be nil. Without them there are no data points, no control
// limits, and every metric is in the Measure phase.
func ToDMAIC(doc *impact.Model, st *impact.Status, sets []*impact.ObservationSet) (*dmaic.DMAICDocument, []Omission, error) {
	if doc == nil {
		return nil, nil, fmt.Errorf("prismx: nil document")
	}
	var pathways []impact.Pathway
	for _, p := range doc.Pathways {
		if !p.Rejected() && slices.Contains(p.Projections, impact.ProjectionDMAIC) {
			pathways = append(pathways, p)
		}
	}
	if len(pathways) == 0 {
		return nil, nil, fmt.Errorf("%w: mark a pathway with projections: [dmaic]", ErrNotDMAICReady)
	}

	statuses := map[string]impact.IndicatorStatus{}
	if st != nil {
		for _, s := range st.Indicators {
			statuses[s.IndicatorID] = s
		}
	}
	history := observationHistory(sets)

	out := &dmaic.DMAICDocument{
		Metadata: &dmaic.Metadata{
			ID:          doc.ID,
			Name:        doc.Metadata.Name,
			Description: doc.Metadata.Description,
			Owner:       doc.Metadata.Owner,
			Version:     doc.Metadata.Version,
			Status:      dmaicStatus(doc.Metadata.Status),
			CreatedAt:   doc.Metadata.CreatedAt,
			UpdatedAt:   doc.Metadata.UpdatedAt,
		},
	}
	if st != nil {
		out.Metadata.Period = st.AsOf
	}

	var omissions []Omission
	seen := map[string]bool{} // indicators already placed in a category
	initiatives := map[string][]string{}
	var initiativeOrder []string
	for _, p := range pathways {
		cat := dmaic.Category{ID: p.ID, Name: p.Title, Description: p.Description, Order: len(out.Categories) + 1, Owner: actorName(doc, p.OwnerID)}
		for _, ind := range doc.Indicators {
			if ind.Rejected() || seen[ind.ID] || !doc.IndicatorServes(ind, p.ID) {
				continue
			}
			seen[ind.ID] = true
			m, reason := metricFrom(doc, p, ind, statuses[ind.ID], history[ind.ID], st != nil)
			if reason != "" {
				omissions = append(omissions, Omission{IndicatorID: ind.ID, Reason: reason})
				continue
			}
			m.Order = len(cat.Metrics) + 1
			for _, id := range m.InitiativeIDs {
				if _, ok := initiatives[id]; !ok {
					initiativeOrder = append(initiativeOrder, id)
				}
				initiatives[id] = append(initiatives[id], m.ID)
			}
			cat.Metrics = append(cat.Metrics, m)
		}
		if len(cat.Metrics) > 0 {
			out.Categories = append(out.Categories, cat)
		}
	}
	if len(out.Categories) == 0 {
		return nil, omissions, fmt.Errorf("%w: no indicator of a DMAIC pathway has a numeric baseline and target", ErrNotDMAICReady)
	}
	for _, id := range initiativeOrder {
		in, ok := doc.Intervention(id)
		if !ok {
			continue // metricFrom only lists interventions it found
		}
		out.Initiatives = append(out.Initiatives, initiativeFrom(doc, *in, initiatives[id]))
	}
	return out, omissions, nil
}

// metricFrom builds the metric for an indicator, or returns why it cannot be
// one.
func metricFrom(doc *impact.Model, p impact.Pathway, ind impact.Indicator, s impact.IndicatorStatus, points []impact.Observation, withStatus bool) (dmaic.Metric, string) {
	if ind.Baseline == nil || ind.Baseline.Value == nil {
		return dmaic.Metric{}, "it has no numeric baseline"
	}
	final, ok := finalTarget(ind)
	if !ok || final.Value == nil {
		return dmaic.Metric{}, "it has no numeric target"
	}
	m := dmaic.Metric{
		ID: ind.ID, Name: ind.Name, Description: ind.Definition, Unit: ind.Unit,
		Baseline: *ind.Baseline.Value, Current: *ind.Baseline.Value, Target: *final.Value,
		TrendDirection: trendDirection(ind.Direction),
		Frequency:      frequencyName(ind.Frequency),
		DataSource:     ind.DataSource,
		Owner:          actorName(doc, ind.OwnerID),
	}
	if m.DataSource == "" && ind.Collection != nil {
		m.DataSource = ind.Collection.System
	}
	if withStatus && s.Health != "" {
		m.Status = capitalize(s.Health)
	}
	for _, o := range points {
		ts, err := time.Parse(time.DateOnly, o.Date)
		if err != nil {
			continue // validation rejects an observation with a bad date
		}
		m.DataPoints = append(m.DataPoints, dmaic.DataPoint{Timestamp: ts, Value: *o.Value, Notes: o.Note})
	}
	// The current value is the last point on the chart, so the two agree; a
	// status without points (such as one from a disaggregated observation)
	// is used only when there are none.
	switch {
	case len(m.DataPoints) > 0:
		m.Current = m.DataPoints[len(m.DataPoints)-1].Value
	case withStatus && s.Latest != nil && s.Latest.Value != nil && len(s.Latest.Disaggregation) == 0:
		m.Current = *s.Latest.Value
	}
	if len(m.DataPoints) >= impact.DMAICMinPoints {
		m.ControlLimits = individualsChart(m.DataPoints)
	}

	targetOutcome := ind.Measures
	if o, ok := doc.Outcome(targetOutcome); ok {
		for _, id := range o.AddressesCauses {
			if c, ok := doc.Cause(id); ok {
				m.RootCauses = append(m.RootCauses, dmaic.RootCause{
					ID: c.ID, Description: c.Statement, Category: capitalize(string(c.Level)), Validated: len(c.EvidenceIDs) > 0,
				})
			}
		}
	}
	active := false
	for _, id := range p.InterventionIDs {
		in, ok := doc.Intervention(id)
		if !ok || in.Rejected() || in.Status == impact.InterventionCancelled {
			continue
		}
		if in.ID == ind.Measures || slices.Contains(in.Targets, targetOutcome) {
			m.InitiativeIDs = append(m.InitiativeIDs, in.ID)
			active = active || in.Status == impact.InterventionInProgress
		}
	}
	m.Phase = metricPhase(s.State, len(m.DataPoints), active)
	return m, ""
}

// metricPhase derives a DMAIC phase from the indicator's state, as a
// convention the DMAIC guide documents.
func metricPhase(state impact.State, points int, improving bool) string {
	switch {
	case state == impact.StateAchieved:
		return dmaic.PhaseControl
	case points < impact.DMAICMinPoints:
		return dmaic.PhaseMeasure
	case improving:
		return dmaic.PhaseImprove
	default:
		return dmaic.PhaseAnalyze
	}
}

// individualsChart computes the limits of an individuals (I-MR) chart: the
// mean, and three estimated standard deviations either side, the deviation
// estimated from the average moving range divided by d2 = 1.128.
func individualsChart(points []dmaic.DataPoint) *dmaic.ControlLimits {
	sum, mr := 0.0, 0.0
	for i, p := range points {
		sum += p.Value
		if i > 0 {
			mr += math.Abs(p.Value - points[i-1].Value)
		}
	}
	mean := sum / float64(len(points))
	sigma := mr / float64(len(points)-1) / 1.128
	r := func(v float64) float64 { return math.Round(v*1000) / 1000 }
	return &dmaic.ControlLimits{CenterLine: r(mean), UCL: r(mean + 3*sigma), LCL: r(mean - 3*sigma), Sigma: r(sigma)}
}

// observationHistory groups the numeric, undisaggregated observations by
// indicator, each in date order. A disaggregated observation is one slice of
// a total, so it is not a point on the total's chart.
func observationHistory(sets []*impact.ObservationSet) map[string][]impact.Observation {
	out := map[string][]impact.Observation{}
	for _, set := range sets {
		for _, o := range set.Observations {
			if o.Value != nil && len(o.Disaggregation) == 0 {
				out[o.IndicatorID] = append(out[o.IndicatorID], o)
			}
		}
	}
	for id := range out {
		slices.SortStableFunc(out[id], func(a, b impact.Observation) int { return strings.Compare(a.Date, b.Date) })
	}
	return out
}

func finalTarget(ind impact.Indicator) (impact.Target, bool) {
	if len(ind.Targets) == 0 {
		return impact.Target{}, false
	}
	final := ind.Targets[0]
	for _, t := range ind.Targets[1:] {
		if t.Date > final.Date {
			final = t
		}
	}
	return final, true
}

func trendDirection(d impact.Direction) string {
	switch d {
	case impact.DirectionIncrease:
		return dmaic.TrendHigherBetter
	case impact.DirectionDecrease:
		return dmaic.TrendLowerBetter
	case impact.DirectionMaintain:
		return dmaic.TrendTargetValue
	}
	return ""
}

// frequencyName names the common ISO 8601 cadences in words and passes any
// other through as written.
func frequencyName(iso string) string {
	names := map[string]string{
		"P1D": "Daily", "P1W": "Weekly", "P2W": "Every two weeks", "P1M": "Monthly",
		"P3M": "Quarterly", "P6M": "Every six months", "P1Y": "Yearly",
	}
	if n, ok := names[iso]; ok {
		return n
	}
	return iso
}

func initiativeFrom(doc *impact.Model, in impact.Intervention, metricIDs []string) dmaic.Initiative {
	out := dmaic.Initiative{
		ID: in.ID, Name: in.Title, Description: in.Description, Owner: actorName(doc, in.OwnerID),
		Status: initiativeStatus(in.Status), MetricIDs: metricIDs,
	}
	var targets []string
	for _, id := range in.Targets {
		if o, ok := doc.Outcome(id); ok {
			targets = append(targets, o.Title)
		}
	}
	if len(targets) > 0 {
		out.ExpectedImpact = strings.Join(targets, "; ")
	}
	for _, r := range in.InitiativeRefs {
		if out.ExternalLinks == nil {
			out.ExternalLinks = map[string]string{}
		}
		key, link := string(r.Framework), r.URL
		if link == "" {
			link = r.Ref
		}
		for n := 2; out.ExternalLinks[key] != ""; n++ {
			key = fmt.Sprintf("%s-%d", r.Framework, n)
		}
		out.ExternalLinks[key] = link
	}
	return out
}

// dmaicStatus maps a prism-core document status to a DMAIC document status.
func dmaicStatus(s string) string {
	switch s {
	case "active", "approved":
		return "Active"
	case "completed":
		return "Completed"
	case "deprecated", "archived":
		return "Archived"
	default:
		return "Draft"
	}
}

// initiativeStatus maps an intervention's status; a paused intervention has no
// DMAIC initiative status and is left blank.
func initiativeStatus(s impact.InterventionStatus) string {
	switch s {
	case impact.InterventionPlanned:
		return "Planned"
	case impact.InterventionInProgress:
		return "In Progress"
	case impact.InterventionCompleted:
		return "Completed"
	case impact.InterventionCancelled:
		return "Cancelled"
	}
	return ""
}
