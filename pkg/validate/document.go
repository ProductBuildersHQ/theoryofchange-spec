package validate

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
	"github.com/ProductBuildersHQ/impact-model/pkg/ref"
)

// Option configures Document and Observations.
type Option func(*checker)

// WithActorRegistry supplies the shared actor registry the document
// references, so actor IDs defined there resolve.
func WithActorRegistry(r *impact.ActorRegistry) Option {
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
	doc      *impact.Model
	issues   []Issue
	source   string
	kinds    map[string]impact.ElementKind // first occurrence of each ID
	registry map[string]bool               // actor IDs from a supplied registry
}

func newChecker(doc *impact.Model, opts []Option) *checker {
	c := &checker{doc: doc, kinds: map[string]impact.ElementKind{}}
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

// Document checks an impact model for structural problems and returns
// the issues in document order. A nil result means the document is
// structurally sound.
func Document(doc *impact.Model, opts ...Option) []Issue {
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
	c.selection()
	c.audiencesAndThemes()
	c.causesAndGroups()
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
func (c *checker) ref(path, element, id string, kinds ...impact.ElementKind) {
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

func (c *checker) refs(path, element string, ids []string, kinds ...impact.ElementKind) {
	for i, id := range ids {
		c.ref(path+"/"+strconv.Itoa(i), element, id, kinds...)
	}
}

func kindList(kinds []impact.ElementKind) string {
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
		if k != impact.ElementActor {
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
	if d.Kind != impact.KindImpactModel {
		c.errorf("VAL-DOC-001", "/kind", "", "kind is %q; expected %q", string(d.Kind), string(impact.KindImpactModel))
	}
	if major, _, _ := strings.Cut(d.SpecVersion, "."); major != "0" {
		c.errorf("VAL-DOC-002", "/specVersion", "", "specVersion %q is not supported (this version reads 0.x)", d.SpecVersion)
	}
	if err := ref.ValidateSlug(d.ID); err != nil {
		c.errorf("VAL-DOC-003", "/id", "", "document id: %v", err)
	}
	if len(d.Outcomes) == 0 {
		c.errorf("VAL-DOC-004", "/outcomes", "", "An impact model needs at least one outcome")
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
	seenValue := map[string]bool{}
	for i, v := range d.Context.Values {
		p := fmt.Sprintf("/context/values/%d/name", i)
		name := strings.TrimSpace(v.Name)
		if name == "" {
			c.errorf("VAL-DOC-006", p, "", "value %d has no name", i)
			continue
		}
		if seenValue[name] {
			c.errorf("VAL-DOC-007", p, "", "value %q is listed more than once", name)
		}
		seenValue[name] = true
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
		c.refs("/ceilingOfAccountability/outcomeIds", "", cl.OutcomeIDs, impact.ElementOutcome)
		if cl.Level == "" && len(cl.OutcomeIDs) == 0 {
			c.errorf("VAL-CEIL-001", "/ceilingOfAccountability", "", "set level or outcomeIds")
		}
	}
	if m := d.Context.Market; m != nil {
		enum(c, "/context/market/framework", "", m.Framework)
		enum(c, "/context/market/relation", "", m.Relation)
	}
}

func (c *checker) logic(path, element string, l impact.Logic) {
	enum(c, path+"/type", element, l.Type)
	if l.Type == impact.LogicAtLeast && l.K < 1 {
		c.errorf("VAL-LOGIC-001", path+"/k", element, "at_least logic needs k of at least 1")
	}
	if l.Type != impact.LogicAtLeast && l.K != 0 {
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
		c.refs(p+"/audienceIds", o.ID, o.AudienceIDs, impact.ElementAudience)
		enum(c, p+"/responsibility", o.ID, o.Responsibility)
		c.actorRefs(p+"/deliveredBy", o.ID, o.DeliveredBy)
		c.refs(p+"/addressesCauses", o.ID, o.AddressesCauses, impact.ElementCause)
		c.refs(p+"/targetGroupIds", o.ID, o.TargetGroupIDs, impact.ElementTargetGroup)
		if o.Priority != nil {
			c.scores(p+"/priority/scores", o.ID, o.Priority.Scores, impact.SubjectOutcome)
		}
		c.contributesTo(p+"/contributesTo", o)
	}
}

func (c *checker) groups() {
	for i, g := range c.doc.Groups {
		p := "/preconditionGroups/" + strconv.Itoa(i)
		c.ref(p+"/target", g.ID, g.Target, impact.ElementOutcome)
		if len(g.Members) < 2 {
			c.errorf("VAL-GRP-001", p+"/members", g.ID, "a precondition group needs at least two members")
		}
		c.refs(p+"/members", g.ID, g.Members, impact.ElementOutcome, impact.ElementGroup)
		for j, m := range g.Members {
			if m == g.ID || (m == g.Target && m != "") {
				c.errorf("VAL-GRP-002", p+"/members/"+strconv.Itoa(j), g.ID, "group %q cannot contain %q", g.ID, m)
			}
		}
		if g.Logic.Type == "" {
			c.errorf("VAL-GRP-003", p+"/logic/type", g.ID, "group has no logic type")
		}
		c.logic(p+"/logic", g.ID, g.Logic)
		// A nested group combines preconditions of the same outcome as the
		// group that contains it.
		for j, m := range g.Members {
			if inner, ok := c.doc.Group(m); ok && inner.Target != g.Target {
				c.errorf("VAL-GRP-004", p+"/members/"+strconv.Itoa(j), g.ID,
					"nested group %q targets %q, not %q like %q", m, inner.Target, g.Target, g.ID)
			}
		}
		if g.Logic.Type == impact.LogicAtLeast && g.Logic.K > len(g.Members) {
			c.errorf("VAL-LOGIC-003", p+"/logic/k", g.ID, "k=%d exceeds the %d members", g.Logic.K, len(g.Members))
		}
	}
}

func (c *checker) links() {
	for i, l := range c.doc.Links {
		p := "/links/" + strconv.Itoa(i)
		c.ref(p+"/from", l.ID, l.From, impact.ElementOutcome, impact.ElementIntervention)
		c.ref(p+"/to", l.ID, l.To, impact.ElementOutcome)
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
		c.refs(p+"/assumptionIds", l.ID, l.AssumptionIDs, impact.ElementAssumption)
		c.refs(p+"/evidenceIds", l.ID, l.EvidenceIDs, impact.ElementEvidence)
		c.refs(p+"/interventionIds", l.ID, l.InterventionIDs, impact.ElementIntervention)
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
		c.refs(p+"/targets", in.ID, in.Targets, impact.ElementOutcome)
		c.actor(p+"/ownerActorId", in.ID, in.OwnerID)
		if in.Priority != nil {
			c.scores(p+"/priority/scores", in.ID, in.Priority.Scores, impact.SubjectIntervention)
		}
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

func (c *checker) externalRefs(path, element string, refs []impact.ExternalRef) {
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
		c.refs(p+"/appliesTo", a.ID, a.AppliesTo, impact.ElementLink, impact.ElementOutcome, impact.ElementIntervention)
		c.refs(p+"/indicatorIds", a.ID, a.IndicatorIDs, impact.ElementIndicator)
		c.refs(p+"/evidenceIds", a.ID, a.EvidenceIDs, impact.ElementEvidence)
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
		c.ref(p+"/measures", ind.ID, ind.Measures, impact.ElementOutcome, impact.ElementIntervention, impact.ElementAssumption, impact.ElementPathway)
		enum(c, p+"/measurementType", ind.ID, ind.MeasurementType)
		enum(c, p+"/direction", ind.ID, ind.Direction)
		c.duration(p+"/frequency", ind.ID, ind.Frequency)
		c.actor(p+"/ownerActorId", ind.ID, ind.OwnerID)
		c.refs(p+"/proxyFor", ind.ID, ind.ProxyFor, impact.ElementOutcome)
		c.externalRefs(p+"/externalRefs", ind.ID, ind.ExternalRefs)
		c.indicatorMetric(p, ind)
		c.collectionAndSelection(p, ind)
		c.refs(p+"/audienceIds", ind.ID, ind.AudienceIDs, impact.ElementAudience)
		c.refs(p+"/targetGroupIds", ind.ID, ind.TargetGroupIDs, impact.ElementTargetGroup)
		c.refs(p+"/themeIds", ind.ID, ind.ThemeIDs, impact.ElementTheme)
		switch {
		case ind.MeasurementType == impact.MeasurementThemeCount && len(ind.ThemeIDs) == 0:
			c.errorf("VAL-THM-003", p+"/themeIds", ind.ID, "a theme_count indicator needs themeIds, the themes it counts")
		case ind.MeasurementType != impact.MeasurementThemeCount && len(ind.ThemeIDs) > 0:
			c.errorf("VAL-THM-004", p+"/themeIds", ind.ID, "themeIds is only used by theme_count indicators")
		}
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
		c.refs(p+"/relatedAssumptionIds", r.ID, r.RelatedAssumptionIDs, impact.ElementAssumption)
	}
}

func (c *checker) alignmentsAndDecisions() {
	for i, a := range c.doc.Alignments {
		p := "/alignments/" + strconv.Itoa(i)
		if a.ElementID == "" {
			c.errorf("VAL-ALN-001", p+"/elementId", "", "alignment has no elementId")
		}
		c.ref(p+"/elementId", "", a.ElementID)
		c.externalRefs(p+"/target", "", []impact.ExternalRef{a.Target})
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
		c.refs(p+"/triggerEvidenceIds", element, r.TriggerEvidenceIDs, impact.ElementEvidence)
		if q := r.QualityReview; q != nil {
			enum(c, p+"/qualityReview/plausible", element, q.Plausible)
			enum(c, p+"/qualityReview/feasible", element, q.Feasible)
			enum(c, p+"/qualityReview/testable", element, q.Testable)
		}
	}
}

// docRef checks a reference to a whole document (model:<doc>).
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
// Impact model. Whether they resolve is checked by Family, which needs both
// documents.
func (c *checker) contributesTo(path string, o impact.Outcome) {
	for i, v := range o.ContributesTo {
		p := path + "/" + strconv.Itoa(i)
		r, err := ref.Parse(v)
		switch {
		case err != nil:
			c.errorf("VAL-PTH-002", p, o.ID, "%q is not a valid reference: %v", v, err)
		case !r.IsElement():
			c.errorf("VAL-PTH-002", p, o.ID, "%q must reference an outcome (model:<doc>#<outcome>)", v)
		case r.Doc == c.doc.ID:
			c.errorf("VAL-PTH-002", p, o.ID, "%q refers to this document; use a link", v)
		case c.doc.Parent != "":
			if parent, perr := ref.Parse(c.doc.Parent); perr == nil && r.Doc != parent.Doc {
				c.errorf("VAL-PTH-003", p, o.ID, "%q is not in the parent impact model %q", v, c.doc.Parent)
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
		c.refs(p+"/outcomeIds", pw.ID, pw.OutcomeIDs, impact.ElementOutcome)
		c.refs(p+"/interventionIds", pw.ID, pw.InterventionIDs, impact.ElementIntervention)
		c.docRef(p+"/childModel", pw.ID, pw.ChildModel)
		if f := pw.Focus; f != nil {
			c.ref(p+"/focus/northStarId", pw.ID, f.NorthStarID, impact.ElementIndicator)
			c.ref(p+"/focus/oneMetricId", pw.ID, f.OneMetricID, impact.ElementIndicator)
			c.refs(p+"/focus/guardrailIds", pw.ID, f.GuardrailIDs, impact.ElementIndicator)
			for j, id := range f.GuardrailIDs {
				if slices.Contains(f.GuardrailIDs[:j], id) {
					c.errorf("VAL-PTH-006", p+"/focus/guardrailIds/"+strconv.Itoa(j), pw.ID, "guardrail %q is listed more than once", id)
				}
			}
		}
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
		case m.Aggregation == impact.AggregationWeightedMean && m.WeightMetricID == "":
			c.errorf("VAL-MET-003", p+"/weightMetricId", m.ID, "a weighted_mean metric needs weightMetricId")
		case m.Aggregation != impact.AggregationWeightedMean && m.WeightMetricID != "":
			c.errorf("VAL-MET-004", p+"/weightMetricId", m.ID, "weightMetricId is only used with weighted_mean aggregation")
		case m.WeightMetricID == m.ID && m.ID != "":
			c.errorf("VAL-MET-004", p+"/weightMetricId", m.ID, "a metric cannot weight itself")
		}
		c.ref(p+"/weightMetricId", m.ID, m.WeightMetricID, impact.ElementMetric)
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
func (c *checker) indicatorMetric(p string, ind impact.Indicator) {
	enum(c, p+"/claim", ind.ID, ind.Claim)
	c.ref(p+"/metricId", ind.ID, ind.MetricID, impact.ElementMetric)
	c.refs(p+"/pathwayIds", ind.ID, ind.PathwayIDs, impact.ElementPathway)
	c.refs(p+"/rollupFrom", ind.ID, ind.RollupFrom, impact.ElementIndicator)
	c.composition(p, ind)
	if len(ind.RollupFrom) == 0 {
		return
	}
	if ind.MetricID == "" {
		c.errorf("VAL-MET-005", p+"/rollupFrom", ind.ID, "rollupFrom needs metricId, which defines how values combine")
		return
	}
	if m, ok := c.doc.Metric(ind.MetricID); ok && m.Aggregation == impact.AggregationNone {
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

// composition checks that an indicator's composition is well formed: a known
// operation with the right number of parts, parts that exist, and no part that
// leads back to the indicator itself.
func (c *checker) composition(p string, ind impact.Indicator) {
	comp := ind.ComposedOf
	if comp == nil {
		return
	}
	cp := p + "/composedOf"
	enum(c, cp+"/op", ind.ID, comp.Op)
	c.refs(cp+"/indicatorIds", ind.ID, comp.IndicatorIDs, impact.ElementIndicator)
	switch {
	case comp.Op == "":
		c.errorf("VAL-MET-007", cp+"/op", ind.ID, "composedOf needs an op")
	case comp.Op.Pair() && len(comp.IndicatorIDs) != 2:
		c.errorf("VAL-MET-007", cp+"/indicatorIds", ind.ID, "a %s takes exactly two indicators, in order; this has %d", comp.Op, len(comp.IndicatorIDs))
	case comp.Op.Valid() && len(comp.IndicatorIDs) < 2:
		c.errorf("VAL-MET-007", cp+"/indicatorIds", ind.ID, "a %s takes at least two indicators; this has %d", comp.Op, len(comp.IndicatorIDs))
	}
	if len(ind.RollupFrom) > 0 {
		c.errorf("VAL-MET-009", cp, ind.ID, "an indicator is either rolled up from others of its metric or composed of other metrics, not both")
	}
	for i, id := range comp.IndicatorIDs {
		if id == ind.ID {
			c.errorf("VAL-MET-008", cp+"/indicatorIds/"+strconv.Itoa(i), ind.ID, "an indicator cannot be composed of itself")
		} else if c.composesFrom(id, ind.ID, map[string]bool{}) {
			c.errorf("VAL-MET-008", cp+"/indicatorIds/"+strconv.Itoa(i), ind.ID, "%q is itself composed, directly or not, of this indicator, so the arithmetic is circular", id)
		}
	}
}

// composesFrom reports whether the indicator is composed, through any chain of
// compositions, of the target.
func (c *checker) composesFrom(id, target string, seen map[string]bool) bool {
	if seen[id] {
		return false
	}
	seen[id] = true
	ind, ok := c.doc.Indicator(id)
	if !ok || ind.ComposedOf == nil {
		return false
	}
	for _, part := range ind.ComposedOf.IndicatorIDs {
		if part == target || c.composesFrom(part, target, seen) {
			return true
		}
	}
	return false
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

func (c *checker) prov(array string, i int, element string, p *impact.Provenance) {
	if p == nil {
		return
	}
	path := "/" + array + "/" + strconv.Itoa(i) + "/provenance"
	if p.Basis == "" {
		c.errorf("VAL-PRV-001", path+"/basis", element, "provenance has no basis")
	}
	enum(c, path+"/basis", element, p.Basis)
	c.refs(path+"/evidenceIds", element, p.EvidenceIDs, impact.ElementEvidence)
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

// collectionAndSelection checks an indicator's collection cost and its
// metric-selection record.
func (c *checker) collectionAndSelection(p string, ind impact.Indicator) {
	if col := ind.Collection; col != nil {
		enum(c, p+"/collection/mode", ind.ID, col.Mode)
		c.duration(p+"/collection/period", ind.ID, col.Period)
		if col.MinutesPerPeriod > 0 && col.Period == "" {
			c.errorf("VAL-SEL-003", p+"/collection/period", ind.ID, "minutesPerPeriod needs period, which says how often the time is spent")
		}
		if col.QuestionsPerRespondent > 0 && col.Period == "" {
			c.errorf("VAL-SEL-003", p+"/collection/period", ind.ID, "questionsPerRespondent needs period, which says how often respondents are asked")
		}
	}
	sel := ind.Selection
	if sel == nil {
		return
	}
	enum(c, p+"/selection/status", ind.ID, sel.Status)
	c.scores(p+"/selection/scores", ind.ID, sel.Scores, impact.SubjectIndicator)
}

// scores checks criterion scores on a subject of the given kind: each score
// is 1 to 5, names a criterion that scores that kind of subject, and a
// criterion is scored once.
func (c *checker) scores(path, element string, scores []impact.CriterionScore, subject impact.SelectionSubject) {
	seen := map[string]bool{}
	for i, sc := range scores {
		sp := path + "/" + strconv.Itoa(i)
		if sc.Score < 1 || sc.Score > 5 {
			c.errorf("VAL-SEL-002", sp+"/score", element, "score %d is outside 1 to 5", sc.Score)
		}
		if seen[sc.CriterionID] {
			c.errorf("VAL-SEL-005", sp+"/criterionId", element, "criterion %q is scored more than once", sc.CriterionID)
		}
		seen[sc.CriterionID] = true
		c.ref(sp+"/criterionId", element, sc.CriterionID, impact.ElementCriterion)
		if cr, ok := c.doc.Criterion(sc.CriterionID); ok && cr.AppliesTo != subject {
			c.errorf("VAL-SEL-004", sp+"/criterionId", element, "criterion %q scores %ss, not %ss", sc.CriterionID, cr.AppliesTo, subject)
		}
	}
}

// selection checks the selection criteria and the measurement budget.
func (c *checker) selection() {
	for i, cr := range c.doc.SelectionCriteria {
		p := "/selectionCriteria/" + strconv.Itoa(i)
		if strings.TrimSpace(cr.Name) == "" {
			c.errorf("VAL-SEL-001", p+"/name", cr.ID, "selection criterion has no name")
		}
		enum(c, p+"/appliesTo", cr.ID, cr.AppliesTo)
	}
	if b := c.doc.Budget; b != nil && b.TargetSelectedMetrics != nil {
		if r := b.TargetSelectedMetrics; r.Max > 0 && r.Min > r.Max {
			c.errorf("VAL-SEL-006", "/measurementBudget/targetSelectedMetrics", "", "the minimum %d is above the maximum %d", r.Min, r.Max)
		}
	}
}

// audiencesAndThemes checks the audiences and the story codebook.
func (c *checker) audiencesAndThemes() {
	for i, a := range c.doc.Audiences {
		if strings.TrimSpace(a.Name) == "" {
			c.errorf("VAL-AUD-001", "/audiences/"+strconv.Itoa(i)+"/name", a.ID, "audience has no name")
		}
		for j, h := range a.Headline {
			hp := "/audiences/" + strconv.Itoa(i) + "/headline/" + strconv.Itoa(j)
			if strings.TrimSpace(h.IndicatorID) == "" {
				c.errorf("VAL-AUD-003", hp+"/indicatorId", a.ID, "headline entry names no indicator")
				continue
			}
			c.ref(hp+"/indicatorId", a.ID, h.IndicatorID, impact.ElementIndicator)
			for _, earlier := range a.Headline[:j] {
				if earlier.IndicatorID == h.IndicatorID {
					c.errorf("VAL-AUD-002", hp+"/indicatorId", a.ID, "indicator %q is in the headline list more than once", h.IndicatorID)
					break
				}
			}
		}
	}
	for i, t := range c.doc.Themes {
		p := "/themes/" + strconv.Itoa(i)
		for _, f := range []struct{ field, value string }{{"label", t.Label}, {"definition", t.Definition}} {
			if strings.TrimSpace(f.value) == "" {
				c.errorf("VAL-THM-001", p+"/"+f.field, t.ID, "theme has no %s", f.field)
			}
		}
		c.refs(p+"/outcomeIds", t.ID, t.OutcomeIDs, impact.ElementOutcome)
	}
}

// causesAndGroups checks the problem tree and the target groups.
func (c *checker) causesAndGroups() {
	for i, g := range c.doc.TargetGroups {
		p := "/targetGroups/" + strconv.Itoa(i)
		if strings.TrimSpace(g.Name) == "" {
			c.errorf("VAL-TGT-001", p+"/name", g.ID, "target group has no name")
		}
		c.actorRefs(p+"/actorIds", g.ID, g.ActorIDs)
	}
	for i, cause := range c.doc.Causes {
		p := "/causes/" + strconv.Itoa(i)
		if strings.TrimSpace(cause.Statement) == "" {
			c.errorf("VAL-CAU-001", p+"/statement", cause.ID, "cause has no statement")
		}
		if cause.Level == "" {
			c.errorf("VAL-CAU-002", p+"/level", cause.ID, "cause has no level (immediate, underlying, or structural)")
		}
		enum(c, p+"/level", cause.ID, cause.Level)
		c.refs(p+"/affectedGroupIds", cause.ID, cause.AffectedGroupIDs, impact.ElementTargetGroup)
		c.refs(p+"/evidenceIds", cause.ID, cause.EvidenceIDs, impact.ElementEvidence)
		if cause.OutOfScope != nil && strings.TrimSpace(cause.OutOfScope.Reason) == "" {
			c.errorf("VAL-CAU-005", p+"/outOfScope/reason", cause.ID, "a cause set aside needs a reason")
		}
		c.ref(p+"/parent", cause.ID, cause.Parent, impact.ElementCause)
		if cause.Parent == cause.ID && cause.ID != "" {
			c.errorf("VAL-CAU-003", p+"/parent", cause.ID, "a cause cannot be its own parent")
		} else if c.causeCycle(cause.ID) {
			c.errorf("VAL-CAU-004", p+"/parent", cause.ID, "the parent chain of cause %q loops back on itself", cause.ID)
		}
	}
}

// causeCycle reports whether following parents from id returns to id.
func (c *checker) causeCycle(id string) bool {
	seen := map[string]bool{id: true}
	cur, ok := c.doc.Cause(id)
	for ok && cur.Parent != "" {
		if cur.Parent == id {
			return cur.Parent != cur.ID
		}
		if seen[cur.Parent] {
			return false
		}
		seen[cur.Parent] = true
		cur, ok = c.doc.Cause(cur.Parent)
	}
	return false
}
