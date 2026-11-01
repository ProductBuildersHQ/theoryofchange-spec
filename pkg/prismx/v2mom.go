package prismx

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/grokify/prism-roadmap/goals/v2mom"

	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
)

// ErrNotV2MOMReady is returned, wrapped, when an impact model lacks what a
// V2MOM requires: a vision, at least one value, and at least one pathway.
var ErrNotV2MOMReady = errors.New("the impact model cannot be drawn as a V2MOM")

// ToV2MOM exports the impact model as a V2MOM.
//
//   - Vision is context.vision, and Values are context.values in order.
//   - Each pathway is a Method. Its measures are the indicators of its focus:
//     the North Star, the one metric that matters, then the guardrails. Its
//     projects are its interventions.
//   - Obstacles are risks and the assumptions that could still break the plan:
//     high or critical ones that are untested, at risk, or invalidated, by
//     their assessed status when st has one. An
//     obstacle that applies to a pathway's outcomes, interventions, or links is
//     listed under that method; any other is listed for the whole V2MOM.
//
// st may be nil; the measures then have no current value, progress, or status.
// ToV2MOM refuses a model without a vision, values, or pathways, so it never
// writes a V2MOM that fails validation for a missing section.
func ToV2MOM(doc *impact.Model, st *impact.Status) (*v2mom.V2MOM, error) {
	if doc == nil {
		return nil, fmt.Errorf("prismx: nil document")
	}
	var missing []string
	if doc.Context.Vision == "" {
		missing = append(missing, "context.vision")
	}
	if len(doc.Context.Values) == 0 {
		missing = append(missing, "context.values")
	}
	var pathways []impact.Pathway
	for _, p := range doc.Pathways {
		if !p.Rejected() {
			pathways = append(pathways, p)
		}
	}
	if len(pathways) == 0 {
		missing = append(missing, "a pathway")
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("%w: add %s", ErrNotV2MOMReady, joinAnd(missing))
	}

	statuses := map[string]impact.IndicatorStatus{}
	assessed := map[string]impact.AssumptionStatus{}
	if st != nil {
		for _, s := range st.Indicators {
			statuses[s.IndicatorID] = s
		}
		for _, a := range st.Assumptions {
			assessed[a.AssumptionID] = a.Status
		}
	}

	out := &v2mom.V2MOM{
		Metadata: &v2mom.Metadata{
			ID:          doc.ID,
			Name:        doc.Metadata.Name,
			Author:      doc.Metadata.Owner,
			Version:     doc.Metadata.Version,
			Status:      v2momStatus(doc.Metadata.Status),
			ParentID:    doc.Parent,
			CreatedAt:   doc.Metadata.CreatedAt,
			UpdatedAt:   doc.Metadata.UpdatedAt,
			Structure:   v2mom.StructureNested,
			Terminology: v2mom.TerminologyV2MOM,
		},
		Vision: doc.Context.Vision,
	}
	for i, v := range doc.Context.Values {
		out.Values = append(out.Values, v2mom.Value{Name: v.Name, Description: v.Description, Priority: i + 1})
	}

	all := obstacles(doc, assessed)
	placed := map[string]bool{} // obstacles already listed under a method
	for _, p := range pathways {
		scope := pathwayScope(doc, p)
		m := v2mom.Method{ID: p.ID, Name: p.Title, Description: p.Description, Owner: actorName(doc, p.OwnerID)}
		if f := p.Focus; f != nil {
			for _, r := range focusMeasures(f) {
				ind, ok := doc.Indicator(r.id)
				if !ok || ind.Rejected() {
					continue
				}
				m.Measures = append(m.Measures, measureFrom(*ind, r.role, statuses[ind.ID], st != nil))
			}
		}
		for _, id := range p.InterventionIDs {
			in, ok := doc.Intervention(id)
			if !ok || in.Rejected() {
				continue
			}
			m.Projects = append(m.Projects, in.ID)
			out.Projects = append(out.Projects, projectFrom(*in, p.ID))
		}
		for _, o := range all {
			if o.appliesWithin(scope) {
				m.Obstacles = append(m.Obstacles, o.obstacle)
				placed[o.obstacle.ID] = true
			}
		}
		out.Methods = append(out.Methods, m)
	}
	for _, o := range all {
		if !placed[o.obstacle.ID] {
			out.Obstacles = append(out.Obstacles, o.obstacle)
		}
	}
	return out, nil
}

type focusRole struct{ id, role string }

// focusMeasures lists a pathway's focus indicators in order, each once.
func focusMeasures(f *impact.Focus) []focusRole {
	var out []focusRole
	seen := map[string]bool{}
	add := func(id, role string) {
		if id != "" && !seen[id] {
			seen[id] = true
			out = append(out, focusRole{id, role})
		}
	}
	add(f.NorthStarID, "North Star")
	add(f.OneMetricID, "One metric that matters")
	for _, id := range f.GuardrailIDs {
		add(id, "Guardrail")
	}
	return out
}

func measureFrom(ind impact.Indicator, role string, s impact.IndicatorStatus, withStatus bool) v2mom.Measure {
	m := v2mom.Measure{ID: ind.ID, Name: ind.Name, Description: role, Unit: ind.Unit}
	if ind.Definition != "" {
		m.Description = role + ": " + ind.Definition
	}
	if ind.Baseline != nil {
		m.Baseline = measureValue(ind.Baseline.Value, ind.Baseline.Level)
	}
	if len(ind.Targets) > 0 {
		final := ind.Targets[0]
		for _, t := range ind.Targets[1:] {
			if t.Date > final.Date {
				final = t
			}
		}
		m.Target = measureValue(final.Value, final.Level)
		m.Timeline = final.Date
	}
	if !withStatus {
		return m
	}
	if s.Latest != nil {
		m.Current = measureValue(s.Latest.Value, s.Latest.Level)
	}
	if s.Progress != nil {
		m.Progress = math.Round(math.Min(*s.Progress, 1)*100) / 100
	}
	m.Status = measureStatus(s.State)
	return m
}

// measureValue is the number, or the qualitative level; the unit is a field
// of its own in a V2MOM measure.
func measureValue(v *float64, level string) string {
	if v != nil {
		return strconv.FormatFloat(*v, 'f', -1, 64)
	}
	return level
}

func measureStatus(s impact.State) string {
	switch s {
	case impact.StateAchieved:
		return "Achieved"
	case impact.StateOnTrack:
		return "On Track"
	case impact.StateAtRisk, impact.StateStale:
		return "At Risk"
	case impact.StateOffTrack:
		return "Behind"
	}
	return ""
}

func projectFrom(in impact.Intervention, methodID string) v2mom.Project {
	p := v2mom.Project{ID: in.ID, Name: in.Title, Description: in.Description, MethodID: methodID, Status: projectStatus(in.Status)}
	for _, r := range in.InitiativeRefs {
		if p.ExternalLinks == nil {
			p.ExternalLinks = map[string]string{}
		}
		key, link := string(r.Framework), r.URL
		if link == "" {
			link = r.Ref
		}
		for n := 2; p.ExternalLinks[key] != ""; n++ {
			key = string(r.Framework) + "-" + strconv.Itoa(n)
		}
		p.ExternalLinks[key] = link
	}
	return p
}

// projectStatus maps an intervention's status; a paused intervention has no
// V2MOM project status and is left blank.
func projectStatus(s impact.InterventionStatus) string {
	switch s {
	case impact.InterventionPlanned:
		return "Approved"
	case impact.InterventionInProgress:
		return "In Progress"
	case impact.InterventionCompleted:
		return "Completed"
	case impact.InterventionCancelled:
		return "Cancelled"
	}
	return ""
}

func v2momStatus(s string) string {
	switch s {
	case "active":
		return v2mom.StatusActive
	case "approved":
		return v2mom.StatusApproved
	case "in_review":
		return v2mom.StatusInReview
	case "completed", "deprecated", "archived":
		return v2mom.StatusCompleted
	default:
		return v2mom.StatusDraft
	}
}

// pathwayScope is the set of element IDs that belong to a pathway: its
// outcomes, its interventions, and the links between them.
func pathwayScope(doc *impact.Model, p impact.Pathway) map[string]bool {
	scope := map[string]bool{}
	for _, id := range p.OutcomeIDs {
		scope[id] = true
	}
	for _, id := range p.InterventionIDs {
		scope[id] = true
	}
	for _, l := range doc.Links {
		if scope[l.From] && scope[l.To] {
			scope[l.ID] = true
		}
	}
	return scope
}

type obstacleOf struct {
	obstacle  v2mom.Obstacle
	appliesTo []string
}

func (o obstacleOf) appliesWithin(scope map[string]bool) bool {
	return slices.ContainsFunc(o.appliesTo, func(id string) bool { return scope[id] })
}

// obstacles lists, in file order, the risks and then the assumptions that
// could still break the plan. assessed holds assumption statuses from the
// status computation, which take precedence over the model's own.
func obstacles(doc *impact.Model, assessed map[string]impact.AssumptionStatus) []obstacleOf {
	var out []obstacleOf
	for _, r := range doc.Risks {
		if r.Rejected() {
			continue
		}
		out = append(out, obstacleOf{
			obstacle: v2mom.Obstacle{
				ID: r.ID, Name: r.Description, Severity: capitalize(r.Impact), Likelihood: capitalize(r.Probability),
				Mitigation: r.Mitigation, Status: capitalize(r.Status),
			},
			appliesTo: r.AppliesTo,
		})
	}
	for _, a := range doc.Assumptions {
		status := assumptionStatus(a, assessed)
		if a.Rejected() || !openAssumption(a.Criticality, status) {
			continue
		}
		desc := "An assumption the plan depends on, " + strings.ReplaceAll(string(status), "_", " ") + "."
		if a.Test != "" {
			desc += " Test: " + a.Test
		}
		out = append(out, obstacleOf{
			obstacle: v2mom.Obstacle{
				ID: a.ID, Name: a.Statement, Description: desc, Severity: capitalize(string(a.Criticality)), Status: "Identified",
			},
			appliesTo: a.AppliesTo,
		})
	}
	return out
}

// openAssumption reports whether an assumption is an obstacle: high or
// critical, and not yet shown to hold.
func openAssumption(c impact.Criticality, status impact.AssumptionStatus) bool {
	if c != impact.CriticalityHigh && c != impact.CriticalityCritical {
		return false
	}
	switch status {
	case impact.AssumptionUntested, impact.AssumptionAtRisk, impact.AssumptionInvalidated:
		return true
	}
	return false
}

// assumptionStatus is the assessed status if there is one, otherwise the
// model's own; an assumption with neither counts as untested.
func assumptionStatus(a impact.Assumption, assessed map[string]impact.AssumptionStatus) impact.AssumptionStatus {
	if s, ok := assessed[a.ID]; ok && s != "" {
		return s
	}
	if a.Status == "" {
		return impact.AssumptionUntested
	}
	return a.Status
}

func actorName(doc *impact.Model, id string) string {
	if a, ok := doc.Actor(id); ok {
		return a.Name
	}
	return ""
}

// joinAnd joins words as "a", "a and b", or "a, b, and c".
func joinAnd(words []string) string {
	switch len(words) {
	case 1:
		return words[0]
	case 2:
		return words[0] + " and " + words[1]
	}
	out := ""
	for i, w := range words {
		switch {
		case i == len(words)-1:
			out += ", and " + w
		case i > 0:
			out += ", " + w
		default:
			out = w
		}
	}
	return out
}
