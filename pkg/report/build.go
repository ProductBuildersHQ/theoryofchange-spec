package report

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/ProductBuildersHQ/impact-model/pkg/graph"
	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
)

// builder holds the lookups shared by the section builders.
type builder struct {
	doc    *impact.Model
	g      *graph.Graph
	st     *impact.Status
	keep   map[string]bool // nil means keep everything
	filter string          // the pathway the view is restricted to, if any
	titles map[string]string

	indStatus map[string]impact.IndicatorStatus
	outStatus map[string]impact.OutcomeStatus
	lnStatus  map[string]impact.LinkStatus
	asmStatus map[string]impact.AssumptionState
}

// BuildView derives a View from a document and its causal graph (built from
// the document when g is nil) and, optionally, a status snapshot. Rejected
// elements are excluded from every section and listed in Rejected. The
// document, graph, and status are not modified.
func BuildView(doc *impact.Model, g *graph.Graph, st *impact.Status, o Options) (*View, error) {
	if doc == nil {
		return nil, fmt.Errorf("report: nil document")
	}
	if g == nil {
		var err error
		if g, err = graph.Build(doc); err != nil {
			return nil, err
		}
	}
	b := &builder{doc: doc, g: g, st: st, titles: map[string]string{}}
	b.indexTitles()
	b.indexStatus()
	if o.Pathway != "" {
		if err := b.restrictTo(o.Pathway); err != nil {
			return nil, err
		}
	}

	v := &View{
		Version:     ViewVersion,
		Filter:      o.Pathway,
		Document:    b.documentView(),
		BasisCounts: map[impact.Basis]int{},
	}
	if st != nil {
		v.AsOf = st.AsOf
	}
	v.Pathways = b.pathways()
	v.Outcomes = b.outcomes()
	v.Links = b.links()
	v.Groups = b.groups()
	v.Interventions = b.interventions()
	v.Assumptions = b.assumptions()
	v.Indicators = b.indicators()
	v.Metrics = b.metrics(v.Indicators)
	v.Actors = b.actors()
	v.Evidence = b.evidence()
	v.Glossary = slices.Clone(doc.Glossary)
	if v.Glossary == nil {
		v.Glossary = []impact.GlossaryTerm{}
	}
	v.Rejected = b.rejected()
	v.countBases()
	return v, nil
}

// kept reports whether the element survives the pathway filter.
func (b *builder) kept(id string) bool { return b.keep == nil || b.keep[id] }

func (b *builder) indexTitles() {
	for _, o := range b.doc.Outcomes {
		b.titles[o.ID] = o.Title
	}
	for _, i := range b.doc.Interventions {
		b.titles[i.ID] = i.Title
	}
	for _, a := range b.doc.Assumptions {
		b.titles[a.ID] = a.Statement
	}
	for _, a := range b.doc.Actors {
		b.titles[a.ID] = a.Name
	}
	for _, p := range b.doc.Pathways {
		b.titles[p.ID] = p.Title
	}
	for _, m := range b.doc.Metrics {
		b.titles[m.ID] = m.Name
	}
}

func (b *builder) indexStatus() {
	b.indStatus = map[string]impact.IndicatorStatus{}
	b.outStatus = map[string]impact.OutcomeStatus{}
	b.lnStatus = map[string]impact.LinkStatus{}
	b.asmStatus = map[string]impact.AssumptionState{}
	if b.st == nil {
		return
	}
	for _, s := range b.st.Indicators {
		b.indStatus[s.IndicatorID] = s
	}
	for _, s := range b.st.Outcomes {
		b.outStatus[s.OutcomeID] = s
	}
	for _, s := range b.st.Links {
		b.lnStatus[s.LinkID] = s
	}
	for _, s := range b.st.Assumptions {
		b.asmStatus[s.AssumptionID] = s
	}
}

// restrictTo limits the view to one pathway plus everything downstream.
func (b *builder) restrictTo(pathwayID string) error {
	pw, ok := b.doc.Pathway(pathwayID)
	if !ok {
		return fmt.Errorf("report: unknown pathway %q", pathwayID)
	}
	b.filter = pathwayID
	b.keep = map[string]bool{}
	for _, id := range slices.Concat(pw.OutcomeIDs, pw.InterventionIDs) {
		if _, active := b.g.Node(id); !active {
			continue
		}
		b.keep[id] = true
		for _, d := range b.g.Downstream(id) {
			b.keep[d] = true
		}
	}
	return nil
}

func (b *builder) name(id string) string {
	if t, ok := b.titles[id]; ok {
		return t
	}
	return id
}

func (b *builder) documentView() DocumentView {
	d := b.doc
	dv := DocumentView{
		ID: d.ID, Title: d.Metadata.Name, Status: d.Metadata.Status,
		Problem: d.Context.ProblemStatement, Narrative: d.Context.Narrative,
		TargetPopulation: d.Context.TargetPopulation, Geography: d.Context.Geography,
		RootCauses: strs(d.Context.RootCauses),
	}
	if dv.Title == "" {
		dv.Title = d.ID
	}
	if h := d.Context.TimeHorizon; h != nil {
		dv.TimeStart, dv.TimeEnd = h.Start, h.End
	}
	if d.Perspective != nil {
		dv.Perspective = b.name(d.Perspective.ActorID)
	}
	if c := d.Ceiling; c != nil {
		switch {
		case c.Level != "" && len(c.OutcomeIDs) > 0:
			dv.Ceiling = fmt.Sprintf("%s and below, plus %s", LevelLabel(c.Level), b.names(c.OutcomeIDs))
		case c.Level != "":
			dv.Ceiling = fmt.Sprintf("%s and below", LevelLabel(c.Level))
		default:
			dv.Ceiling = b.names(c.OutcomeIDs) + " and everything that leads to it"
		}
		dv.CeilingRationale = c.Rationale
	}
	return dv
}

func (b *builder) names(ids []string) string {
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = b.name(id)
	}
	return strings.Join(out, ", ")
}

func (b *builder) pathways() []PathwayView {
	out := []PathwayView{}
	for _, p := range b.doc.Pathways {
		if p.Rejected() || (b.filter != "" && p.ID != b.filter) {
			continue
		}
		pv := PathwayView{
			ID: p.ID, Title: p.Title, Kind: string(p.Kind), Description: p.Description,
			ChildModel: p.ChildModel, Basis: basisOf(p.Provenance),
			OutcomeIDs: []string{}, InterventionIDs: []string{},
		}
		if p.OwnerID != "" {
			pv.Owner = b.name(p.OwnerID)
		}
		for _, id := range p.OutcomeIDs {
			if _, active := b.g.Node(id); active && b.kept(id) {
				pv.OutcomeIDs = append(pv.OutcomeIDs, id)
			}
		}
		for _, id := range p.InterventionIDs {
			if _, active := b.g.Node(id); active && b.kept(id) {
				pv.InterventionIDs = append(pv.InterventionIDs, id)
			}
		}
		out = append(out, pv)
	}
	return out
}

func (b *builder) pathwaysOf(id string) []string {
	out := []string{}
	for _, p := range b.doc.Pathways {
		if p.Rejected() {
			continue
		}
		if slices.Contains(p.OutcomeIDs, id) || slices.Contains(p.InterventionIDs, id) {
			out = append(out, p.ID)
		}
	}
	return out
}

func (b *builder) outcomes() []OutcomeView {
	out := []OutcomeView{}
	for _, o := range b.doc.Outcomes {
		if _, active := b.g.Node(o.ID); !active || !b.kept(o.ID) {
			continue
		}
		ov := OutcomeView{
			ID: o.ID, Title: o.Title, Description: o.Description, Level: o.Level, Sphere: o.Sphere,
			Category: o.Category, ChangeType: o.ChangeType, Population: o.Population,
			Owned:          b.g.BelowCeiling(o.ID),
			Responsibility: string(o.EffectiveResponsibility()),
			Accountable:    b.g.BelowCeiling(o.ID) && o.Accountable(),
			Logic:          logicText(o.EffectiveLogic()),
			Preconditions:  strs(b.g.Preconditions(o.ID)),
			Pathways:       b.pathwaysOf(o.ID),
			Indicators:     b.activeIndicatorIDs(o.ID),
			Basis:          basisOf(o.Provenance),
		}
		for _, id := range o.ActorIDs {
			ov.Actors = append(ov.Actors, b.name(id))
		}
		for _, id := range o.DeliveredBy {
			ov.DeliveredBy = append(ov.DeliveredBy, b.name(id))
		}
		ov.AddressesCauses = slices.Clone(o.AddressesCauses)
		if s, ok := b.outStatus[o.ID]; ok {
			ov.Status = &StatusView{State: string(s.State), PathwayState: string(s.PathwayState), Health: s.Health, Confidence: s.Confidence}
		}
		out = append(out, ov)
	}
	return out
}

func (b *builder) activeIndicatorIDs(measures string) []string {
	out := []string{}
	for _, ind := range b.doc.Indicators {
		if ind.Measures == measures && !ind.Rejected() {
			out = append(out, ind.ID)
		}
	}
	return out
}

func (b *builder) links() []LinkView {
	out := []LinkView{}
	active := map[string]bool{}
	for _, e := range slices.Concat(b.g.Edges(), b.g.FeedbackEdges()) {
		if e.LinkID != "" {
			active[e.LinkID] = true
		}
	}
	for _, l := range b.doc.Links {
		if !active[l.ID] || !b.kept(l.From) || !b.kept(l.To) {
			continue
		}
		lv := LinkView{
			ID: l.ID, From: l.From, To: l.To, Type: string(l.Type), Rationale: l.Rationale,
			Feedback: l.Feedback, LoopType: string(l.LoopType), Polarity: string(l.Polarity), Lag: l.Lag,
			Assumptions: slices.Clone(l.AssumptionIDs), Interventions: slices.Clone(l.InterventionIDs),
			Basis: basisOf(l.Provenance),
		}
		if s, ok := b.lnStatus[l.ID]; ok {
			lv.State = string(s.State)
		}
		out = append(out, lv)
	}
	return out
}

func (b *builder) groups() []GroupView {
	out := []GroupView{}
	for _, o := range b.doc.Outcomes {
		if !b.kept(o.ID) {
			continue
		}
		for _, grp := range b.g.GroupsFor(o.ID) {
			gv := GroupView{
				ID: grp.ID, Target: grp.Target, Members: strs(grp.Members),
				Logic: logicText(grp.Logic), Sufficiency: grp.Sufficiency, Basis: basisOf(grp.Provenance),
			}
			out = append(out, gv)
		}
	}
	return out
}

func (b *builder) interventions() []InterventionView {
	out := []InterventionView{}
	for _, in := range b.doc.Interventions {
		if _, active := b.g.Node(in.ID); !active || !b.kept(in.ID) {
			continue
		}
		iv := InterventionView{
			ID: in.ID, Title: in.Title, Description: in.Description, Type: string(in.Type),
			Status: string(in.Status), Outputs: slices.Clone(in.Outputs), Resources: in.Resources,
			Targets: []string{}, Pathways: b.pathwaysOf(in.ID), Basis: basisOf(in.Provenance),
		}
		if in.OwnerID != "" {
			iv.Owner = b.name(in.OwnerID)
		}
		for _, t := range in.Targets {
			if _, active := b.g.Node(t); active {
				iv.Targets = append(iv.Targets, t)
			}
		}
		out = append(out, iv)
	}
	return out
}

func (b *builder) assumptions() []AssumptionView {
	out := []AssumptionView{}
	for _, a := range b.doc.Assumptions {
		if a.Rejected() {
			continue
		}
		if b.keep != nil && !slices.ContainsFunc(a.AppliesTo, b.keptElement) {
			continue
		}
		av := AssumptionView{
			ID: a.ID, Statement: a.Statement, Type: string(a.Type), Criticality: string(a.Criticality),
			Plausibility: string(a.Plausibility), NeedsTesting: a.NeedsTesting, Test: a.Test,
			AppliesTo: slices.Clone(a.AppliesTo), Indicators: slices.Clone(a.IndicatorIDs),
			Status: string(a.Status), StatusSource: string(impact.SourceDocument), Basis: basisOf(a.Provenance),
		}
		if a.IfThenBecause != nil {
			av.If, av.Then, av.Because = a.IfThenBecause.If, a.IfThenBecause.Then, a.IfThenBecause.Because
		}
		if s, ok := b.asmStatus[a.ID]; ok {
			av.Status, av.StatusSource = string(s.Status), string(s.Source)
		}
		out = append(out, av)
	}
	return out
}

// keptElement reports whether an assumption target (a link, outcome, or
// intervention) survives the pathway filter.
func (b *builder) keptElement(id string) bool {
	if l, ok := b.doc.Link(id); ok {
		return b.kept(l.From) && b.kept(l.To)
	}
	return b.kept(id)
}

func (b *builder) indicators() []IndicatorView {
	out := []IndicatorView{}
	for _, ind := range b.doc.Indicators {
		if ind.Rejected() {
			continue
		}
		if b.keep != nil && !b.keptSubject(ind.Measures) {
			continue
		}
		iv := IndicatorView{
			ID: ind.ID, Name: ind.Name, Definition: ind.Definition, Measures: ind.Measures,
			MeasuresTitle: b.name(ind.Measures), Metric: ind.MetricID, Type: string(ind.MeasurementType),
			Unit: ind.Unit, Direction: string(ind.Direction), Claim: ind.EffectiveClaim(),
			Caveats: slices.Clone(ind.Caveats), Pathways: slices.Clone(ind.PathwayIDs),
			Levels: slices.Clone(ind.Levels), DataSource: ind.DataSource, Method: ind.Method,
			Frequency: ind.Frequency, DataAvailability: string(ind.DataAvailability),
			RollupFrom: slices.Clone(ind.RollupFrom), Basis: basisOf(ind.Provenance),
		}
		if ind.OwnerID != "" {
			iv.Owner = b.name(ind.OwnerID)
		}
		if bl := ind.Baseline; bl != nil {
			iv.Baseline = withDate(valueText(bl.Value, bl.Level, ind.Unit), bl.Date, "as of ")
		}
		for _, t := range ind.Targets {
			text := withDate(valueText(t.Value, t.Level, ind.Unit), t.Date, "by ")
			if t.Label != "" {
				text += " — " + t.Label
			}
			if t.Threshold != "" {
				text += " (" + t.Threshold + ")"
			}
			iv.Targets = append(iv.Targets, text)
		}
		if s, ok := b.indStatus[ind.ID]; ok {
			iv.Status = &StatusView{State: string(s.State), Health: s.Health}
			if s.Latest != nil {
				// An estimate reads as an estimate, so a rounded or partial figure
				// is not mistaken for a count.
				prefix := "on "
				if s.Latest.Quality == impact.QualityEstimated {
					prefix = "estimate, on "
				}
				iv.Latest = withDate(valueText(s.Latest.Value, s.Latest.Level, ind.Unit), s.Latest.Date, prefix)
				iv.LatestNote = s.Latest.Note
			}
		}
		out = append(out, iv)
	}
	return out
}

// keptSubject reports whether the thing an indicator measures survives the
// pathway filter. Assumption-monitoring indicators follow their assumption.
func (b *builder) keptSubject(id string) bool {
	if _, ok := b.doc.Pathway(id); ok {
		return b.filter == "" || id == b.filter
	}
	if a, ok := b.doc.Assumption(id); ok {
		return slices.ContainsFunc(a.AppliesTo, b.keptElement)
	}
	return b.kept(id)
}

func (b *builder) metrics(indicators []IndicatorView) []MetricView {
	out := []MetricView{}
	byMetric := map[string][]string{}
	for _, ind := range indicators {
		if ind.Metric != "" {
			byMetric[ind.Metric] = append(byMetric[ind.Metric], ind.ID)
		}
	}
	for _, m := range b.doc.Metrics {
		if m.Rejected() {
			continue
		}
		ids := byMetric[m.ID]
		if b.keep != nil && len(ids) == 0 {
			continue
		}
		mv := MetricView{
			ID: m.ID, Name: m.Name, Definition: m.Definition, Formula: m.Formula, Unit: m.Unit,
			Direction: string(m.Direction), Aggregation: string(m.Aggregation), Layer: string(m.Layer),
			Indicators: strs(ids), Basis: basisOf(m.Provenance),
		}
		if m.WeightMetricID != "" {
			mv.WeightedBy = b.name(m.WeightMetricID)
		}
		out = append(out, mv)
	}
	return out
}

func (b *builder) actors() []ActorView {
	out := []ActorView{}
	for _, a := range b.doc.Actors {
		if a.Rejected() {
			continue
		}
		av := ActorView{
			ID: a.ID, Name: a.Name, Type: string(a.Type), Stance: string(a.Stance),
			Influence: string(a.Influence), Interest: string(a.Interest), Description: a.Description,
			Basis: basisOf(a.Provenance),
		}
		for _, r := range a.Roles {
			av.Roles = append(av.Roles, string(r))
		}
		out = append(out, av)
	}
	return out
}

func (b *builder) evidence() []EvidenceView {
	out := []EvidenceView{}
	for _, e := range b.doc.Evidence {
		ev := EvidenceView{
			ID: e.ID, Title: e.Title, Type: string(e.Type), Strength: string(e.Strength),
			Direction: string(e.Direction), Source: e.Source, URL: e.URL, Summary: e.Summary,
			Excerpt: e.Excerpt, Sensitivity: string(e.Sensitivity), AccessedAt: e.AccessedAt,
		}
		if ev.Sensitivity == "" {
			ev.Sensitivity = string(impact.SensitivityRestricted) // unset is treated as restricted
		}
		if !e.Sensitivity.Renderable() {
			ev.Withheld = e.Excerpt != "" || e.Summary != ""
			ev.Withheld = ev.Withheld || e.URL != ""
			ev.Excerpt, ev.Summary, ev.URL = "", "", ""
		}
		out = append(out, ev)
	}
	return out
}

func (b *builder) rejected() []RejectedView {
	out := []RejectedView{}
	add := func(id string, kind impact.ElementKind, title string, p *impact.Provenance) {
		if p.IsRejected() {
			out = append(out, RejectedView{ID: id, Kind: kind, Title: title, Note: p.Note})
		}
	}
	d := b.doc
	for _, v := range d.Outcomes {
		add(v.ID, impact.ElementOutcome, v.Title, v.Provenance)
	}
	for _, v := range d.Groups {
		add(v.ID, impact.ElementGroup, "Precondition group for "+b.name(v.Target), v.Provenance)
	}
	for _, v := range d.Links {
		add(v.ID, impact.ElementLink, b.name(v.From)+" → "+b.name(v.To), v.Provenance)
	}
	for _, v := range d.Interventions {
		add(v.ID, impact.ElementIntervention, v.Title, v.Provenance)
	}
	for _, v := range d.Assumptions {
		add(v.ID, impact.ElementAssumption, v.Statement, v.Provenance)
	}
	for _, v := range d.Indicators {
		add(v.ID, impact.ElementIndicator, v.Name, v.Provenance)
	}
	for _, v := range d.Actors {
		add(v.ID, impact.ElementActor, v.Name, v.Provenance)
	}
	for _, v := range d.ExternalFactors {
		add(v.ID, impact.ElementExternalFactor, v.Description, v.Provenance)
	}
	for _, v := range d.Risks {
		add(v.ID, impact.ElementRisk, v.Description, v.Provenance)
	}
	for _, v := range d.Metrics {
		add(v.ID, impact.ElementMetric, v.Name, v.Provenance)
	}
	for _, v := range d.Pathways {
		add(v.ID, impact.ElementPathway, v.Title, v.Provenance)
	}
	return out
}

func (v *View) countBases() {
	add := func(b BasisView) {
		if b.Basis == "" {
			v.Unmarked++
			return
		}
		v.BasisCounts[b.Basis]++
	}
	for _, x := range v.Pathways {
		add(x.Basis)
	}
	for _, x := range v.Outcomes {
		add(x.Basis)
	}
	for _, x := range v.Links {
		add(x.Basis)
	}
	for _, x := range v.Groups {
		add(x.Basis)
	}
	for _, x := range v.Interventions {
		add(x.Basis)
	}
	for _, x := range v.Assumptions {
		add(x.Basis)
	}
	for _, x := range v.Indicators {
		add(x.Basis)
	}
	for _, x := range v.Metrics {
		add(x.Basis)
	}
	for _, x := range v.Actors {
		add(x.Basis)
	}
}

// ---- value formatting -------------------------------------------------------

func valueText(v *float64, level, unit string) string {
	switch {
	case level != "":
		return level
	case v != nil:
		s := strconv.FormatFloat(*v, 'f', -1, 64)
		if unit != "" {
			s += " " + unit
		}
		return s
	default:
		return ""
	}
}

func withDate(text, date, prefix string) string {
	switch {
	case text == "" && date == "":
		return ""
	case date == "":
		return text
	case text == "":
		return prefix + date
	default:
		return text + " (" + prefix + date + ")"
	}
}
