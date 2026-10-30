// Package report builds the view model that every report and slide deck
// renders from. One model feeds every output, so a number or a label never
// differs between a report and a deck.
//
// A View is derived from a Theory of Change document and its causal graph,
// and optionally from a computed status snapshot. Without a status it
// supports the definition reports (diagram, logic models, metric
// definitions, glossary) that need no data.
package report

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

// ViewVersion is the version of the View model. Templates written against a
// version keep working while it is unchanged; renaming or removing a field
// requires a new version.
const ViewVersion = "1"

// Options configures BuildView.
type Options struct {
	// Pathway restricts the view to one pathway: its outcomes and
	// interventions plus everything downstream of them, so the route to
	// impact stays visible. Empty means the whole theory.
	Pathway string
}

// View is the data every report and deck renders.
type View struct {
	Version string `json:"version"`
	// Filter is the pathway ID the view is restricted to, if any.
	Filter string `json:"filter,omitempty"`
	// AsOf is the status date; empty when the view has no status.
	AsOf string `json:"asOf,omitempty"`
	// ContentHash is the SHA-256 of the document's canonical JSON, when
	// the caller supplied it with WithContentHash.
	ContentHash string `json:"contentHash,omitempty"`

	Document      DocumentView       `json:"document"`
	Pathways      []PathwayView      `json:"pathways"`
	Outcomes      []OutcomeView      `json:"outcomes"`
	Links         []LinkView         `json:"links"`
	Groups        []GroupView        `json:"groups"`
	Interventions []InterventionView `json:"interventions"`
	Assumptions   []AssumptionView   `json:"assumptions"`
	Indicators    []IndicatorView    `json:"indicators"`
	Metrics       []MetricView       `json:"metrics"`
	Actors        []ActorView        `json:"actors"`
	Evidence      []EvidenceView     `json:"evidence"`
	Glossary      []toc.GlossaryTerm `json:"glossary"`
	// Rejected lists elements reviewers rejected. They are kept for
	// transparency and excluded from everything above.
	Rejected []RejectedView `json:"rejected"`
	// BasisCounts counts active elements by provenance basis; Unmarked
	// counts those with none.
	BasisCounts map[toc.Basis]int `json:"basisCounts"`
	Unmarked    int               `json:"unmarkedCount"`
}

// DocumentView is the theory's identity and context.
type DocumentView struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	Status           string `json:"status,omitempty"`
	Problem          string `json:"problem"`
	Narrative        string `json:"narrative,omitempty"`
	TargetPopulation string `json:"targetPopulation,omitempty"`
	Geography        string `json:"geography,omitempty"`
	TimeStart        string `json:"timeStart,omitempty"`
	TimeEnd          string `json:"timeEnd,omitempty"`
	// Perspective names the actor whose theory this is.
	Perspective string `json:"perspective,omitempty"`
	// Ceiling describes the ceiling of accountability in words.
	Ceiling          string   `json:"ceiling,omitempty"`
	CeilingRationale string   `json:"ceilingRationale,omitempty"`
	RootCauses       []string `json:"rootCauses,omitempty"`
}

// BasisView is an element's provenance in display form.
type BasisView struct {
	Basis toc.Basis `json:"basis,omitempty"`
	// Marked is true for inferred and hypothesis elements, which every
	// output must visibly mark so a draft is never read as the
	// organization's own claim.
	Marked      bool     `json:"marked,omitempty"`
	EvidenceIDs []string `json:"evidenceIds,omitempty"`
	Note        string   `json:"note,omitempty"`
}

// StatusView is an element's computed status.
type StatusView struct {
	State        string   `json:"state"`
	PathwayState string   `json:"pathwayState,omitempty"`
	Health       string   `json:"health,omitempty"`
	Confidence   *float64 `json:"confidence,omitempty"`
}

// PathwayView is a pathway with resolved member lists.
type PathwayView struct {
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	Kind            string    `json:"kind,omitempty"`
	Description     string    `json:"description,omitempty"`
	Owner           string    `json:"owner,omitempty"`
	OutcomeIDs      []string  `json:"outcomeIds"`
	InterventionIDs []string  `json:"interventionIds"`
	ChildToc        string    `json:"childToc,omitempty"`
	Basis           BasisView `json:"basis"`
}

// OutcomeView is an outcome with its place in the graph and its status.
type OutcomeView struct {
	ID          string         `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description,omitempty"`
	Level       toc.Level      `json:"level"`
	Sphere      toc.Sphere     `json:"sphere,omitempty"`
	Category    string         `json:"category,omitempty"`
	ChangeType  toc.ChangeType `json:"changeType,omitempty"`
	Population  string         `json:"population,omitempty"`
	// Owned reports whether the outcome is within the ceiling of
	// accountability.
	Owned bool `json:"owned"`
	// Responsibility says how far the perspective actor answers for the
	// outcome: own, shared, partner, or unaddressed.
	Responsibility string `json:"responsibility"`
	// Accountable is true for an outcome the perspective actor answers
	// for: it is within the ceiling and owned or shared. Every other
	// outcome is monitored only.
	Accountable bool `json:"accountable"`
	// DeliveredBy names the actors who deliver the outcome.
	DeliveredBy []string `json:"deliveredBy,omitempty"`
	// AddressesCauses lists the IDs of the causes the outcome answers.
	AddressesCauses []string `json:"addressesCauses,omitempty"`
	// Logic says how incoming preconditions combine: "all", "any", or
	// "at least k".
	Logic         string      `json:"logic"`
	Preconditions []string    `json:"preconditions"`
	Pathways      []string    `json:"pathways"`
	Indicators    []string    `json:"indicators"`
	Actors        []string    `json:"actors,omitempty"`
	Basis         BasisView   `json:"basis"`
	Status        *StatusView `json:"status,omitempty"`
}

// LinkView is a causal link in display form.
type LinkView struct {
	ID            string    `json:"id"`
	From          string    `json:"from"`
	To            string    `json:"to"`
	Type          string    `json:"type,omitempty"`
	Rationale     string    `json:"rationale,omitempty"`
	Feedback      bool      `json:"feedback,omitempty"`
	LoopType      string    `json:"loopType,omitempty"`
	Polarity      string    `json:"polarity,omitempty"`
	Lag           string    `json:"lag,omitempty"`
	Assumptions   []string  `json:"assumptions,omitempty"`
	Interventions []string  `json:"interventions,omitempty"`
	Basis         BasisView `json:"basis"`
	State         string    `json:"state,omitempty"`
}

// GroupView is a precondition group.
type GroupView struct {
	ID          string    `json:"id"`
	Target      string    `json:"target"`
	Members     []string  `json:"members"`
	Logic       string    `json:"logic"`
	Sufficiency string    `json:"sufficiency,omitempty"`
	Basis       BasisView `json:"basis"`
}

// InterventionView is an intervention with its targets and pathways.
type InterventionView struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Type        string    `json:"type,omitempty"`
	Status      string    `json:"status,omitempty"`
	Owner       string    `json:"owner,omitempty"`
	Targets     []string  `json:"targets"`
	Outputs     []string  `json:"outputs,omitempty"`
	Resources   string    `json:"resources,omitempty"`
	Pathways    []string  `json:"pathways,omitempty"`
	Basis       BasisView `json:"basis"`
}

// AssumptionView is an assumption in display form.
type AssumptionView struct {
	ID           string    `json:"id"`
	Statement    string    `json:"statement"`
	If           string    `json:"if,omitempty"`
	Then         string    `json:"then,omitempty"`
	Because      string    `json:"because,omitempty"`
	Type         string    `json:"type,omitempty"`
	Criticality  string    `json:"criticality,omitempty"`
	Plausibility string    `json:"plausibility,omitempty"`
	NeedsTesting bool      `json:"needsTesting,omitempty"`
	Test         string    `json:"test,omitempty"`
	AppliesTo    []string  `json:"appliesTo,omitempty"`
	Indicators   []string  `json:"indicators,omitempty"`
	Status       string    `json:"status,omitempty"`
	StatusSource string    `json:"statusSource,omitempty"`
	Basis        BasisView `json:"basis"`
}

// IndicatorView is an indicator with its metric and latest status.
type IndicatorView struct {
	ID               string      `json:"id"`
	Name             string      `json:"name"`
	Definition       string      `json:"definition,omitempty"`
	Measures         string      `json:"measures"`
	MeasuresTitle    string      `json:"measuresTitle,omitempty"`
	Metric           string      `json:"metric,omitempty"`
	Type             string      `json:"type,omitempty"`
	Unit             string      `json:"unit,omitempty"`
	Direction        string      `json:"direction,omitempty"`
	Claim            toc.Claim   `json:"claim"`
	Caveats          []string    `json:"caveats,omitempty"`
	Pathways         []string    `json:"pathways,omitempty"`
	Levels           []string    `json:"levels,omitempty"`
	Baseline         string      `json:"baseline,omitempty"`
	Targets          []string    `json:"targets,omitempty"`
	DataSource       string      `json:"dataSource,omitempty"`
	Method           string      `json:"method,omitempty"`
	Frequency        string      `json:"frequency,omitempty"`
	Owner            string      `json:"owner,omitempty"`
	DataAvailability string      `json:"dataAvailability,omitempty"`
	RollupFrom       []string    `json:"rollupFrom,omitempty"`
	Basis            BasisView   `json:"basis"`
	Status           *StatusView `json:"status,omitempty"`
	Latest           string      `json:"latest,omitempty"`
}

// MetricView is a shared metric with the indicators that instantiate it.
type MetricView struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Definition  string    `json:"definition"`
	Formula     string    `json:"formula"`
	Unit        string    `json:"unit,omitempty"`
	Direction   string    `json:"direction,omitempty"`
	Aggregation string    `json:"aggregation,omitempty"`
	WeightedBy  string    `json:"weightedBy,omitempty"`
	Layer       string    `json:"layer,omitempty"`
	Indicators  []string  `json:"indicators"`
	Basis       BasisView `json:"basis"`
}

// ActorView is an actor in display form.
type ActorView struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type,omitempty"`
	Roles       []string  `json:"roles,omitempty"`
	Stance      string    `json:"stance,omitempty"`
	Influence   string    `json:"influence,omitempty"`
	Interest    string    `json:"interest,omitempty"`
	Description string    `json:"description,omitempty"`
	Basis       BasisView `json:"basis"`
}

// EvidenceView is a source. Restricted evidence has its excerpt withheld so
// no renderer can quote it by accident.
type EvidenceView struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Type        string `json:"type,omitempty"`
	Strength    string `json:"strength,omitempty"`
	Direction   string `json:"direction,omitempty"`
	Source      string `json:"source,omitempty"`
	URL         string `json:"url,omitempty"`
	Summary     string `json:"summary,omitempty"`
	Excerpt     string `json:"excerpt,omitempty"`
	Sensitivity string `json:"sensitivity"`
	// Withheld is true when the excerpt, summary, or URL exists but must not
	// be shown.
	Withheld   bool   `json:"withheld,omitempty"`
	AccessedAt string `json:"accessedAt,omitempty"`
}

// RejectedView names an element reviewers rejected.
type RejectedView struct {
	ID    string          `json:"id"`
	Kind  toc.ElementKind `json:"kind"`
	Title string          `json:"title"`
	Note  string          `json:"note,omitempty"`
}

// Label turns an enumeration value like "long_term" into "Long term".
func Label(v string) string {
	if v == "" {
		return ""
	}
	s := strings.ReplaceAll(v, "_", " ")
	return strings.ToUpper(s[:1]) + s[1:]
}

// LevelLabel is the display name of an outcome level.
func LevelLabel(l toc.Level) string {
	switch l {
	case toc.LevelImpact:
		return "Impact"
	case toc.LevelLongTerm:
		return "Long-term outcome"
	case toc.LevelIntermediate:
		return "Intermediate outcome"
	case toc.LevelShortTerm:
		return "Short-term outcome"
	case toc.LevelOutput:
		return "Output"
	default:
		return Label(string(l))
	}
}

func logicText(l toc.Logic) string {
	if l.Type == toc.LogicAtLeast {
		return fmt.Sprintf("at least %d", l.K)
	}
	return string(l.Type)
}

func basisOf(p *toc.Provenance) BasisView {
	if p == nil {
		return BasisView{}
	}
	return BasisView{
		Basis:       p.Basis,
		Marked:      p.Basis == toc.BasisInferred || p.Basis == toc.BasisHypothesis,
		EvidenceIDs: slices.Clone(p.EvidenceIDs),
		Note:        p.Note,
	}
}

// strs returns a non-nil copy so JSON renders [] rather than null.
func strs(s []string) []string {
	if s == nil {
		return []string{}
	}
	return slices.Clone(s)
}
