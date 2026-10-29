package validate

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/ref"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

// Option configures Document and Observations.
type Option func(*checker)

// WithActorRegistry supplies the shared actor registry the document
// references, so actor IDs defined there resolve.
func WithActorRegistry(r *toc.ActorRegistry) Option {
	return func(c *checker) {
		if r == nil {
			return
		}
		c.registry = map[string]bool{}
		for _, a := range r.Actors {
			c.registry[a.ID] = true
		}
	}
}

type checker struct {
	doc      *toc.Document
	issues   []Issue
	source   string
	kinds    map[string]toc.ElementKind // first occurrence of each ID
	registry map[string]bool            // actor IDs from a supplied registry
}

func newChecker(doc *toc.Document, opts []Option) *checker {
	c := &checker{doc: doc, kinds: map[string]toc.ElementKind{}}
	for _, o := range opts {
		o(c)
	}
	return c
}

func (c *checker) add(code string, sev Severity, path, element, format string, args ...any) {
	c.issues = append(c.issues, Issue{
		Code: code, Severity: sev, Path: path, Element: element,
		Message: fmt.Sprintf(format, args...), Source: c.source,
	})
}

func (c *checker) errorf(code, path, element, format string, args ...any) {
	c.add(code, SeverityError, path, element, format, args...)
}

// Document checks a Theory of Change for structural problems and returns
// the issues in document order. A nil result means the document is
// structurally sound.
func Document(doc *toc.Document, opts ...Option) []Issue {
	c := newChecker(doc, opts)
	c.header()
	c.identifiers()
	c.context()
	c.outcomes()
	c.groups()
	c.links()
	c.interventions()
	c.assumptions()
	c.indicators()
	c.actors()
	c.evidence()
	c.factorsAndRisks()
	c.alignmentsAndDecisions()
	c.revisions()
	c.pathways()
	c.metrics()
	c.glossary()
	c.provenance()
	c.consultations()
	return c.issues
}

var (
	durationRegex = regexp.MustCompile(`^P(\d+Y)?(\d+M)?(\d+W)?(\d+D)?(T(\d+H)?(\d+M)?(\d+(\.\d+)?S)?)?$`)
)

// ---- generic helpers ------------------------------------------------------

func validDate(s string) bool {
	_, err := time.Parse(time.DateOnly, s)
	return err == nil
}

func validDuration(s string) bool {
	return s != "P" && !strings.HasSuffix(s, "T") && durationRegex.MatchString(s)
}

type validator interface {
	~string
	Valid() bool
}

// enum reports an invalid value of an enumeration type. The empty value is
// skipped; required enums are checked for presence by their callers.
func enum[T validator](c *checker, path, element string, v T) {
	if v != "" && !v.Valid() {
		c.errorf("VAL-ENUM-001", path, element, "%q is not a valid value", string(v))
	}
}

func (c *checker) date(path, element, v string) {
	if v != "" && !validDate(v) {
		c.errorf("VAL-DATE-001", path, element, "%q is not a date (want YYYY-MM-DD)", v)
	}
}

func (c *checker) duration(path, element, v string) {
	if v != "" && !validDuration(v) {
		c.errorf("VAL-DUR-001", path, element, "%q is not an ISO 8601 duration (e.g. P6M, P1W, PT36H)", v)
	}
}

// ref checks that id resolves to an element of one of the given kinds (any
// kind when none are given). Empty IDs are skipped.
func (c *checker) ref(path, element, id string, kinds ...toc.ElementKind) {
	if id == "" {
		return
	}
	got, ok := c.kinds[id]
	if !ok {
		c.errorf("VAL-REF-001", path, element, "references unknown element %q", id)
		return
	}
	if len(kinds) > 0 && !slices.Contains(kinds, got) {
		c.errorf("VAL-REF-002", path, element, "references %q, which is a %s; expected %s", id, got, kindList(kinds))
	}
}

func (c *checker) refs(path, element string, ids []string, kinds ...toc.ElementKind) {
	for i, id := range ids {
		c.ref(path+"/"+strconv.Itoa(i), element, id, kinds...)
	}
}

func kindList(kinds []toc.ElementKind) string {
	names := make([]string, len(kinds))
	for i, k := range kinds {
		names[i] = string(k)
	}
	return strings.Join(names, " or ")
}

// actor checks an actor reference. IDs not defined in the document resolve
// against a supplied registry; if the document names a registry that was not
// supplied, the reference cannot be checked and is reported as a warning.
func (c *checker) actor(path, element, id string) {
	if id == "" {
		return
	}
	if k, ok := c.kinds[id]; ok {
		if k != toc.ElementActor {
			c.errorf("VAL-REF-002", path, element, "references %q, which is a %s; expected actor", id, k)
		}
		return
	}
	switch {
	case c.registry[id]:
	case c.doc.ActorRegistry != "" && c.registry == nil:
		c.add("VAL-REF-004", SeverityWarning, path, element,
			"actor %q is not defined here; it may be in registry %q, which was not supplied", id, c.doc.ActorRegistry)
	default:
		c.errorf("VAL-REF-001", path, element, "references unknown actor %q", id)
	}
}

func (c *checker) actorRefs(path, element string, ids []string) {
	for i, id := range ids {
		c.actor(path+"/"+strconv.Itoa(i), element, id)
	}
}

// ---- document sections ----------------------------------------------------

func (c *checker) header() {
	d := c.doc
	if d.Kind != toc.KindTheoryOfChange {
		c.errorf("VAL-DOC-001", "/kind", "", "kind is %q; expected %q", string(d.Kind), string(toc.KindTheoryOfChange))
	}
	if major, _, _ := strings.Cut(d.SpecVersion, "."); major != "0" {
		c.errorf("VAL-DOC-002", "/specVersion", "", "specVersion %q is not supported (this version reads 0.x)", d.SpecVersion)
	}
	if err := ref.ValidateSlug(d.ID); err != nil {
		c.errorf("VAL-DOC-003", "/id", "", "document id: %v", err)
	}
	if len(d.Outcomes) == 0 {
		c.errorf("VAL-DOC-004", "/outcomes", "", "a Theory of Change needs at least one outcome")
	}
}

// identifiers checks every element ID once and records its kind.
func (c *checker) identifiers() {
	firstPath := map[string]string{}
	for _, e := range c.doc.Elements() {
		switch {
		case e.ID == "":
			c.errorf("VAL-ID-001", e.Path+"/id", "", "%s has no id", e.Kind)
			continue
		case ref.ValidateSlug(e.ID) != nil:
			c.errorf("VAL-ID-003", e.Path+"/id", e.ID, "id %q is not valid: %v", e.ID, ref.ValidateSlug(e.ID))
		}
		if first, dup := firstPath[e.ID]; dup {
			c.errorf("VAL-ID-002", e.Path+"/id", e.ID, "duplicate id %q (first used at %s)", e.ID, first)
			continue
		}
		firstPath[e.ID] = e.Path
		c.kinds[e.ID] = e.Kind
	}
}

func (c *checker) context() {
	d := c.doc
	if strings.TrimSpace(d.Context.ProblemStatement) == "" {
		c.errorf("VAL-DOC-005", "/context/problemStatement", "", "problemStatement is empty")
	}
	if h := d.Context.TimeHorizon; h != nil {
		c.date("/context/timeHorizon/start", "", h.Start)
		c.date("/context/timeHorizon/end", "", h.End)
		if validDate(h.Start) && validDate(h.End) && h.End < h.Start {
			c.errorf("VAL-DATE-002", "/context/timeHorizon", "", "end %s is before start %s", h.End, h.Start)
		}
	}
	if p := d.Perspective; p != nil {
		c.actor("/perspective/actorId", "", p.ActorID)
		enum(c, "/perspective/role", "", p.Role)
	}
	if cl := d.Ceiling; cl != nil {
		enum(c, "/ceilingOfAccountability/level", "", cl.Level)
		c.refs("/ceilingOfAccountability/outcomeIds", "", cl.OutcomeIDs, toc.ElementOutcome)
		if cl.Level == "" && len(cl.OutcomeIDs) == 0 {
			c.errorf("VAL-CEIL-001", "/ceilingOfAccountability", "", "set level or outcomeIds")
		}
	}
	if m := d.Context.Market; m != nil {
		enum(c, "/context/market/framework", "", m.Framework)
		enum(c, "/context/market/relation", "", m.Relation)
	}
}

func (c *checker) logic(path, element string, l toc.Logic) {
	enum(c, path+"/type", element, l.Type)
	if l.Type == toc.LogicAtLeast && l.K < 1 {
		c.errorf("VAL-LOGIC-001", path+"/k", element, "at_least logic needs k of at least 1")
	}
	if l.Type != toc.LogicAtLeast && l.K != 0 {
		c.errorf("VAL-LOGIC-002", path+"/k", element, "k is only used with at_least logic")
	}
}

func (c *checker) outcomes() {
	for i, o := range c.doc.Outcomes {
		p := "/outcomes/" + strconv.Itoa(i)
		if strings.TrimSpace(o.Title) == "" {
			c.errorf("VAL-OUT-001", p+"/title", o.ID, "outcome has no title")
		}
		if o.Level == "" {
			c.errorf("VAL-OUT-002", p+"/level", o.ID, "outcome has no level")
		}
		enum(c, p+"/level", o.ID, o.Level)
		enum(c, p+"/sphere", o.ID, o.Sphere)
		enum(c, p+"/changeType", o.ID, o.ChangeType)
		if o.Logic != nil {
			c.logic(p+"/logic", o.ID, *o.Logic)
		}
		c.actorRefs(p+"/actorIds", o.ID, o.ActorIDs)
		c.contributesTo(p+"/contributesTo", o)
	}
}

func (c *checker) groups() {
	for i, g := range c.doc.Groups {
		p := "/preconditionGroups/" + strconv.Itoa(i)
		c.ref(p+"/target", g.ID, g.Target, toc.ElementOutcome)
		if len(g.Members) < 2 {
			c.errorf("VAL-GRP-001", p+"/members", g.ID, "a precondition group needs at least two members")
		}
		c.refs(p+"/members", g.ID, g.Members, toc.ElementOutcome, toc.ElementGroup)
		for j, m := range g.Members {
			if m == g.ID || (m == g.Target && m != "") {
				c.errorf("VAL-GRP-002", p+"/members/"+strconv.Itoa(j), g.ID, "group %q cannot contain %q", g.ID, m)
			}
		}
		if g.Logic.Type == "" {
			c.errorf("VAL-GRP-003", p+"/logic/type", g.ID, "group has no logic type")
		}
		c.logic(p+"/logic", g.ID, g.Logic)
		if g.Logic.Type == toc.LogicAtLeast && g.Logic.K > len(g.Members) {
			c.errorf("VAL-LOGIC-003", p+"/logic/k", g.ID, "k=%d exceeds the %d members", g.Logic.K, len(g.Members))
		}
	}
}

func (c *checker) links() {
	for i, l := range c.doc.Links {
		p := "/links/" + strconv.Itoa(i)
		c.ref(p+"/from", l.ID, l.From, toc.ElementOutcome, toc.ElementIntervention)
		c.ref(p+"/to", l.ID, l.To, toc.ElementOutcome)
		if l.From != "" && l.From == l.To {
			c.errorf("VAL-LINK-001", p, l.ID, "link connects %q to itself", l.From)
		}
		enum(c, p+"/type", l.ID, l.Type)
		enum(c, p+"/polarity", l.ID, l.Polarity)
		enum(c, p+"/loopType", l.ID, l.LoopType)
		if l.LoopType != "" && !l.Feedback {
			c.errorf("VAL-LINK-002", p+"/loopType", l.ID, "loopType requires feedback: true")
		}
		c.duration(p+"/lag", l.ID, l.Lag)
		c.refs(p+"/assumptionIds", l.ID, l.AssumptionIDs, toc.ElementAssumption)
		c.refs(p+"/evidenceIds", l.ID, l.EvidenceIDs, toc.ElementEvidence)
		c.refs(p+"/interventionIds", l.ID, l.InterventionIDs, toc.ElementIntervention)
	}
}

func (c *checker) interventions() {
	for i, in := range c.doc.Interventions {
		p := "/interventions/" + strconv.Itoa(i)
		if strings.TrimSpace(in.Title) == "" {
			c.errorf("VAL-INT-002", p+"/title", in.ID, "intervention has no title")
		}
		if len(in.Targets) == 0 {
			c.errorf("VAL-INT-001", p+"/targets", in.ID, "intervention targets no outcomes")
		}
		c.refs(p+"/targets", in.ID, in.Targets, toc.ElementOutcome)
		c.actor(p+"/ownerActorId", in.ID, in.OwnerID)
		enum(c, p+"/type", in.ID, in.Type)
		enum(c, p+"/status", in.ID, in.Status)
		for j, cr := range in.Capabilities {
			if cr.CapabilityID == "" {
				c.errorf("VAL-INT-003", p+"/capabilities/"+strconv.Itoa(j)+"/capabilityId", in.ID, "capability reference has no capabilityId")
			}
		}
		c.externalRefs(p+"/initiativeRefs", in.ID, in.InitiativeRefs)
	}
}

func (c *checker) externalRefs(path, element string, refs []toc.ExternalRef) {
	for i, r := range refs {
		p := path + "/" + strconv.Itoa(i)
		enum(c, p+"/framework", element, r.Framework)
		enum(c, p+"/relation", element, r.Relation)
		if r.Ref == "" {
			c.errorf("VAL-EXT-001", p+"/ref", element, "external reference has no ref")
		}
	}
}

func (c *checker) assumptions() {
	for i, a := range c.doc.Assumptions {
		p := "/assumptions/" + strconv.Itoa(i)
		if strings.TrimSpace(a.Statement) == "" {
			c.errorf("VAL-ASM-001", p+"/statement", a.ID, "assumption has no statement")
		}
		enum(c, p+"/type", a.ID, a.Type)
		enum(c, p+"/criticality", a.ID, a.Criticality)
		enum(c, p+"/plausibility", a.ID, a.Plausibility)
		enum(c, p+"/status", a.ID, a.Status)
		c.refs(p+"/appliesTo", a.ID, a.AppliesTo, toc.ElementLink, toc.ElementOutcome, toc.ElementIntervention)
		c.refs(p+"/indicatorIds", a.ID, a.IndicatorIDs, toc.ElementIndicator)
		c.refs(p+"/evidenceIds", a.ID, a.EvidenceIDs, toc.ElementEvidence)
		c.actor(p+"/ownerActorId", a.ID, a.OwnerID)
	}
}

func (c *checker) indicators() {
	for i, ind := range c.doc.Indicators {
		p := "/indicators/" + strconv.Itoa(i)
		if strings.TrimSpace(ind.Name) == "" {
			c.errorf("VAL-IND-002", p+"/name", ind.ID, "indicator has no name")
		}
		if ind.Measures == "" {
			c.errorf("VAL-IND-003", p+"/measures", ind.ID, "indicator measures nothing")
		}
		c.ref(p+"/measures", ind.ID, ind.Measures, toc.ElementOutcome, toc.ElementIntervention, toc.ElementAssumption)
		enum(c, p+"/measurementType", ind.ID, ind.MeasurementType)
		enum(c, p+"/direction", ind.ID, ind.Direction)
		c.duration(p+"/frequency", ind.ID, ind.Frequency)
		c.actor(p+"/ownerActorId", ind.ID, ind.OwnerID)
		c.refs(p+"/proxyFor", ind.ID, ind.ProxyFor, toc.ElementOutcome)
		c.externalRefs(p+"/externalRefs", ind.ID, ind.ExternalRefs)
		c.indicatorMetric(p, ind)
		if pop := ind.Population; pop != nil {
			c.actorRefs(p+"/population/actorIds", ind.ID, pop.ActorIDs)
		}
		levelOK := func(path, level string) {
			if level != "" && len(ind.Levels) > 0 && !slices.Contains(ind.Levels, level) {
				c.errorf("VAL-IND-001", path, ind.ID, "level %q is not one of the indicator's levels %v", level, ind.Levels)
			}
		}
		if b := ind.Baseline; b != nil {
			c.date(p+"/baseline/date", ind.ID, b.Date)
			levelOK(p+"/baseline/level", b.Level)
		}
		for j, t := range ind.Targets {
			tp := p + "/targets/" + strconv.Itoa(j)
			if t.Date == "" {
				c.errorf("VAL-DATE-003", tp+"/date", ind.ID, "target has no date")
			}
			c.date(tp+"/date", ind.ID, t.Date)
			levelOK(tp+"/level", t.Level)
		}
	}
}

func (c *checker) actors() {
	for i, a := range c.doc.Actors {
		p := "/actors/" + strconv.Itoa(i)
		if strings.TrimSpace(a.Name) == "" {
			c.errorf("VAL-ACT-001", p+"/name", a.ID, "actor has no name")
		}
		enum(c, p+"/type", a.ID, a.Type)
		enum(c, p+"/stance", a.ID, a.Stance)
		enum(c, p+"/influence", a.ID, a.Influence)
		enum(c, p+"/interest", a.ID, a.Interest)
		for j, r := range a.Roles {
			enum(c, p+"/roles/"+strconv.Itoa(j), a.ID, r)
		}
		c.externalRefs(p+"/externalRefs", a.ID, a.ExternalRefs)
	}
}

func (c *checker) evidence() {
	for i, e := range c.doc.Evidence {
		p := "/evidence/" + strconv.Itoa(i)
		if strings.TrimSpace(e.Title) == "" {
			c.errorf("VAL-EVD-001", p+"/title", e.ID, "evidence has no title")
		}
		enum(c, p+"/type", e.ID, e.Type)
		enum(c, p+"/strength", e.ID, e.Strength)
		enum(c, p+"/direction", e.ID, e.Direction)
		enum(c, p+"/processTracingTest", e.ID, e.ProcessTracingTest)
		enum(c, p+"/sensitivity", e.ID, e.Sensitivity)
		c.date(p+"/accessedAt", e.ID, e.AccessedAt)
	}
}

func (c *checker) factorsAndRisks() {
	for i, f := range c.doc.ExternalFactors {
		p := "/externalFactors/" + strconv.Itoa(i)
		enum(c, p+"/kind", f.ID, f.Kind)
		enum(c, p+"/likelihood", f.ID, f.Likelihood)
		c.refs(p+"/affects", f.ID, f.Affects)
	}
	for i, r := range c.doc.Risks {
		p := "/risks/" + strconv.Itoa(i)
		if strings.TrimSpace(r.Description) == "" {
			c.errorf("VAL-RSK-001", p+"/description", r.ID, "risk has no description")
		}
		enum(c, p+"/category", r.ID, r.Category)
		c.refs(p+"/appliesTo", r.ID, r.AppliesTo)
		c.refs(p+"/relatedAssumptionIds", r.ID, r.RelatedAssumptionIDs, toc.ElementAssumption)
	}
}

func (c *checker) alignmentsAndDecisions() {
	for i, a := range c.doc.Alignments {
		p := "/alignments/" + strconv.Itoa(i)
		if a.ElementID == "" {
			c.errorf("VAL-ALN-001", p+"/elementId", "", "alignment has no elementId")
		}
		c.ref(p+"/elementId", "", a.ElementID)
		c.externalRefs(p+"/target", "", []toc.ExternalRef{a.Target})
	}
	for i, d := range c.doc.Decisions {
		p := "/decisions/" + strconv.Itoa(i)
		if strings.TrimSpace(d.Question) == "" {
			c.errorf("VAL-DEC-001", p+"/question", d.ID, "decision has no question")
		}
		enum(c, p+"/status", d.ID, d.Status)
		c.date(p+"/dueDate", d.ID, d.DueDate)
		c.actor(p+"/ownerActorId", d.ID, d.OwnerID)
		c.refs(p+"/relatedIds", d.ID, d.RelatedIDs)
	}
}

func (c *checker) revisions() {
	for i, r := range c.doc.Revisions {
		p := "/revisions/" + strconv.Itoa(i)
		element := r.Version
		if r.Date == "" {
			c.errorf("VAL-DATE-003", p+"/date", element, "revision has no date")
		}
		c.date(p+"/date", element, r.Date)
		enum(c, p+"/trigger", element, r.Trigger)
		// Changed elements may have been removed by the revision itself.
		for j, id := range r.ChangedIDs {
			if _, ok := c.kinds[id]; !ok && id != "" {
				c.add("VAL-REF-003", SeverityWarning, p+"/changedIds/"+strconv.Itoa(j), element,
					"changed element %q no longer exists", id)
			}
		}
		c.refs(p+"/triggerEvidenceIds", element, r.TriggerEvidenceIDs, toc.ElementEvidence)
		if q := r.QualityReview; q != nil {
			enum(c, p+"/qualityReview/plausible", element, q.Plausible)
			enum(c, p+"/qualityReview/feasible", element, q.Feasible)
			enum(c, p+"/qualityReview/testable", element, q.Testable)
		}
	}
}

// docRef checks a reference to a whole document (toc:<doc>).
func (c *checker) docRef(path, element, v string) {
	if v == "" {
		return
	}
	r, err := ref.Parse(v)
	switch {
	case err != nil:
		c.errorf("VAL-PTH-001", path, element, "%q is not a valid reference: %v", v, err)
	case r.IsElement():
		c.errorf("VAL-PTH-001", path, element, "%q must reference a whole document, not an element", v)
	case r.Doc == c.doc.ID:
		c.errorf("VAL-PTH-001", path, element, "%q refers to this document itself", v)
	}
}

// contributesTo checks an outcome's references to outcomes in another
// theory. Whether they resolve is checked by Family, which needs both
// documents.
func (c *checker) contributesTo(path string, o toc.Outcome) {
	for i, v := range o.ContributesTo {
		p := path + "/" + strconv.Itoa(i)
		r, err := ref.Parse(v)
		switch {
		case err != nil:
			c.errorf("VAL-PTH-002", p, o.ID, "%q is not a valid reference: %v", v, err)
		case !r.IsElement():
			c.errorf("VAL-PTH-002", p, o.ID, "%q must reference an outcome (toc:<doc>#<outcome>)", v)
		case r.Doc == c.doc.ID:
			c.errorf("VAL-PTH-002", p, o.ID, "%q refers to this document; use a link", v)
		case c.doc.Parent != "":
			if parent, perr := ref.Parse(c.doc.Parent); perr == nil && r.Doc != parent.Doc {
				c.errorf("VAL-PTH-003", p, o.ID, "%q is not in the parent theory %q", v, c.doc.Parent)
			}
		}
	}
}

func (c *checker) pathways() {
	c.docRef("/parent", "", c.doc.Parent)
	for i, pw := range c.doc.Pathways {
		p := "/pathways/" + strconv.Itoa(i)
		if strings.TrimSpace(pw.Title) == "" {
			c.errorf("VAL-PTH-005", p+"/title", pw.ID, "pathway has no title")
		}
		enum(c, p+"/kind", pw.ID, pw.Kind)
		c.actor(p+"/ownerActorId", pw.ID, pw.OwnerID)
		c.refs(p+"/outcomeIds", pw.ID, pw.OutcomeIDs, toc.ElementOutcome)
		c.refs(p+"/interventionIds", pw.ID, pw.InterventionIDs, toc.ElementIntervention)
		c.docRef(p+"/childToc", pw.ID, pw.ChildToc)
	}
}

func (c *checker) metrics() {
	for i, m := range c.doc.Metrics {
		p := "/metrics/" + strconv.Itoa(i)
		for _, f := range []struct{ field, value string }{
			{"name", m.Name}, {"definition", m.Definition}, {"formula", m.Formula},
		} {
			if strings.TrimSpace(f.value) == "" {
				c.errorf("VAL-MET-001", p+"/"+f.field, m.ID, "metric has no %s", f.field)
			}
		}
		enum(c, p+"/direction", m.ID, m.Direction)
		enum(c, p+"/aggregation", m.ID, m.Aggregation)
		enum(c, p+"/layer", m.ID, m.Layer)
		switch {
		case m.Aggregation == toc.AggregationWeightedMean && m.WeightMetricID == "":
			c.errorf("VAL-MET-003", p+"/weightMetricId", m.ID, "a weighted_mean metric needs weightMetricId")
		case m.Aggregation != toc.AggregationWeightedMean && m.WeightMetricID != "":
			c.errorf("VAL-MET-004", p+"/weightMetricId", m.ID, "weightMetricId is only used with weighted_mean aggregation")
		case m.WeightMetricID == m.ID && m.ID != "":
			c.errorf("VAL-MET-004", p+"/weightMetricId", m.ID, "a metric cannot weight itself")
		}
		c.ref(p+"/weightMetricId", m.ID, m.WeightMetricID, toc.ElementMetric)
	}
}

func (c *checker) glossary() {
	for i, t := range c.doc.Glossary {
		p := "/glossary/" + strconv.Itoa(i)
		if strings.TrimSpace(t.Term) == "" {
			c.errorf("VAL-GLO-001", p+"/term", t.ID, "glossary entry has no term")
		}
		if strings.TrimSpace(t.Definition) == "" {
			c.errorf("VAL-GLO-001", p+"/definition", t.ID, "glossary entry has no definition")
		}
		c.refs(p+"/relatedIds", t.ID, t.RelatedIDs)
	}
}

// indicatorMetric checks an indicator's metric, pathway, claim, and
// roll-up fields.
func (c *checker) indicatorMetric(p string, ind toc.Indicator) {
	enum(c, p+"/claim", ind.ID, ind.Claim)
	c.ref(p+"/metricId", ind.ID, ind.MetricID, toc.ElementMetric)
	c.refs(p+"/pathwayIds", ind.ID, ind.PathwayIDs, toc.ElementPathway)
	c.refs(p+"/rollupFrom", ind.ID, ind.RollupFrom, toc.ElementIndicator)
	if len(ind.RollupFrom) == 0 {
		return
	}
	if ind.MetricID == "" {
		c.errorf("VAL-MET-005", p+"/rollupFrom", ind.ID, "rollupFrom needs metricId, which defines how values combine")
		return
	}
	if m, ok := c.doc.Metric(ind.MetricID); ok && m.Aggregation == toc.AggregationNone {
		c.errorf("VAL-MET-002", p+"/rollupFrom", ind.ID, "metric %q has aggregation none; its values cannot be rolled up", ind.MetricID)
	}
	for i, srcID := range ind.RollupFrom {
		sp := p + "/rollupFrom/" + strconv.Itoa(i)
		if srcID == ind.ID {
			c.errorf("VAL-MET-006", sp, ind.ID, "an indicator cannot roll up from itself")
			continue
		}
		if src, ok := c.doc.Indicator(srcID); ok && src.MetricID != ind.MetricID {
			c.errorf("VAL-MET-002", sp, ind.ID, "%q instantiates metric %q, not %q", srcID, src.MetricID, ind.MetricID)
		}
	}
}

// provenance checks every element's provenance block.
func (c *checker) provenance() {
	d := c.doc
	for i, v := range d.Outcomes {
		c.prov("outcomes", i, v.ID, v.Provenance)
	}
	for i, v := range d.Groups {
		c.prov("preconditionGroups", i, v.ID, v.Provenance)
	}
	for i, v := range d.Links {
		c.prov("links", i, v.ID, v.Provenance)
	}
	for i, v := range d.Interventions {
		c.prov("interventions", i, v.ID, v.Provenance)
	}
	for i, v := range d.Assumptions {
		c.prov("assumptions", i, v.ID, v.Provenance)
	}
	for i, v := range d.Indicators {
		c.prov("indicators", i, v.ID, v.Provenance)
		enum(c, "/indicators/"+strconv.Itoa(i)+"/dataAvailability", v.ID, v.DataAvailability)
	}
	for i, v := range d.Actors {
		c.prov("actors", i, v.ID, v.Provenance)
	}
	for i, v := range d.ExternalFactors {
		c.prov("externalFactors", i, v.ID, v.Provenance)
	}
	for i, v := range d.Risks {
		c.prov("risks", i, v.ID, v.Provenance)
	}
	for i, v := range d.Metrics {
		c.prov("metrics", i, v.ID, v.Provenance)
	}
	for i, v := range d.Pathways {
		c.prov("pathways", i, v.ID, v.Provenance)
	}
}

func (c *checker) prov(array string, i int, element string, p *toc.Provenance) {
	if p == nil {
		return
	}
	path := "/" + array + "/" + strconv.Itoa(i) + "/provenance"
	if p.Basis == "" {
		c.errorf("VAL-PRV-001", path+"/basis", element, "provenance has no basis")
	}
	enum(c, path+"/basis", element, p.Basis)
	c.refs(path+"/evidenceIds", element, p.EvidenceIDs, toc.ElementEvidence)
	c.date(path+"/decidedAt", element, p.DecidedAt)
}

func (c *checker) consultations() {
	for i, k := range c.doc.Consultations {
		p := "/consultations/" + strconv.Itoa(i)
		if k.Date == "" {
			c.errorf("VAL-DATE-003", p+"/date", k.ID, "consultation has no date")
		}
		c.date(p+"/date", k.ID, k.Date)
		if k.Step == "" {
			c.errorf("VAL-CON-001", p+"/step", k.ID, "consultation has no step")
		}
		enum(c, p+"/step", k.ID, k.Step)
		enum(c, p+"/method", k.ID, k.Method)
		if k.Count < 0 {
			c.errorf("VAL-CON-002", p+"/count", k.ID, "count cannot be negative")
		}
		// Changed elements may have been removed by the consultation's outcome.
		for j, id := range k.ChangedIDs {
			if _, ok := c.kinds[id]; !ok && id != "" {
				c.add("VAL-REF-003", SeverityWarning, p+"/changedIds/"+strconv.Itoa(j), k.ID,
					"changed element %q no longer exists", id)
			}
		}
	}
}
