// Package lint checks the quality of a Theory of Change with deterministic
// rules, each with a stable ID and a severity. Validation answers "is this a
// well-formed document?"; lint answers "is this a sound theory?".
//
// Rules are registered in a Registry. A profile picks how much to report:
// draft runs the errors and the structural rules, so a small valid document
// is GO while it is being written; standard adds the warnings; full adds the
// advice. Results render as a structured-evaluation SummaryReport.
package lint

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/plexusone/structured-evaluation/summary"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/graph"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/validate"
)

// Profile selects which rules run.
type Profile string

// Profiles, from least to most.
const (
	// ProfileDraft runs error rules and structural rules.
	ProfileDraft Profile = "draft"
	// ProfileStandard also runs warning rules. It is the default.
	ProfileStandard Profile = "standard"
	// ProfileFull also runs informational rules.
	ProfileFull Profile = "full"
)

// Profiles lists the profiles in order.
func Profiles() []Profile { return []Profile{ProfileDraft, ProfileStandard, ProfileFull} }

// Valid reports whether p is a known profile.
func (p Profile) Valid() bool { return slices.Contains(Profiles(), p) }

// Categories group rules in the report. A rule's category is the middle part
// of its ID, such as STR in TOC-STR-001.
const (
	// CategorySTR is the structural category, which runs in every profile.
	CategorySTR = "STR"
)

// Errors returned by Run and Registry.
var (
	ErrUnknownProfile = errors.New("unknown lint profile")
	ErrUnknownRule    = errors.New("unknown lint rule")
	ErrDuplicateRule  = errors.New("duplicate lint rule")
)

// Finding is one problem a rule found.
type Finding struct {
	Rule     string            `json:"rule"`
	Severity validate.Severity `json:"severity"`
	// Path is the JSON pointer to the element or field, when known.
	Path string `json:"path,omitempty"`
	// Element is the ID of the element concerned, when there is one.
	Element string `json:"element,omitempty"`
	Message string `json:"message"`
}

// Context is what a rule sees. Rules must not modify it.
type Context struct {
	Doc   *toc.Document
	Graph *graph.Graph
	// AsOf is the date time-based rules measure from (YYYY-MM-DD); empty
	// when none was given, in which case those rules do not report.
	AsOf string
}

// Rule is one check.
type Rule interface {
	// ID is the stable identifier, such as TOC-STR-001.
	ID() string
	// Severity is the severity of every finding the rule reports.
	Severity() validate.Severity
	// Description says what the rule checks, in one sentence.
	Description() string
	// Check returns the rule's findings; each carries the rule's ID and
	// severity, which the engine fills in.
	Check(*Context) []Finding
}

// Category returns the category of a rule ID: TOC-STR-001 is STR.
func Category(id string) string {
	parts := strings.Split(id, "-")
	if len(parts) != 3 {
		return ""
	}
	return parts[1]
}

// minProfile is the lowest profile in which a rule runs.
func minProfile(r Rule) Profile {
	switch {
	case r.Severity() == validate.SeverityError || Category(r.ID()) == CategorySTR:
		return ProfileDraft
	case r.Severity() == validate.SeverityWarning:
		return ProfileStandard
	default:
		return ProfileFull
	}
}

// Runs reports whether the rule runs in the profile.
func Runs(r Rule, p Profile) bool {
	return slices.Index(Profiles(), minProfile(r)) <= slices.Index(Profiles(), p)
}

// Registry holds rules by ID.
type Registry struct {
	rules map[string]Rule
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry { return &Registry{rules: map[string]Rule{}} }

// Add registers a rule. An empty or repeated ID is an error.
func (g *Registry) Add(r Rule) error {
	id := r.ID()
	if Category(id) == "" || !strings.HasPrefix(id, "TOC-") {
		return fmt.Errorf("lint: rule ID %q is not of the form TOC-<CAT>-<NNN>", id)
	}
	if _, dup := g.rules[id]; dup {
		return fmt.Errorf("%w: %s", ErrDuplicateRule, id)
	}
	g.rules[id] = r
	return nil
}

// Rules returns the registered rules sorted by ID.
func (g *Registry) Rules() []Rule {
	out := make([]Rule, 0, len(g.rules))
	for _, r := range g.rules {
		out = append(out, r)
	}
	slices.SortFunc(out, func(a, b Rule) int { return strings.Compare(a.ID(), b.ID()) })
	return out
}

// Rule returns the rule with the given ID.
func (g *Registry) Rule(id string) (Rule, bool) {
	r, ok := g.rules[id]
	return r, ok
}

// Default is the registry of the rules this module ships. The rule files
// register into it from init.
var Default = NewRegistry()

// Register adds a rule to Default. It panics on a bad or repeated ID, which
// is a programming error found the first time the program runs.
func Register(r Rule) {
	if err := Default.Add(r); err != nil {
		panic(err)
	}
}

// Options configure a run.
type Options struct {
	// Profile defaults to standard.
	Profile Profile
	// Disable lists rule IDs to skip.
	Disable []string
	// AsOf is the date for time-based rules (YYYY-MM-DD).
	AsOf string
	// Registry defaults to Default.
	Registry *Registry
}

// Result is the outcome of a run.
type Result struct {
	Profile  Profile   `json:"profile"`
	Findings []Finding `json:"findings"`
	// Ran lists the rules that ran, by ID.
	Ran []string `json:"ran"`
	// Disabled lists the rules skipped because they were disabled.
	Disabled []string `json:"disabled"`

	rules map[string]Rule
}

// Count returns the number of findings of a severity.
func (r *Result) Count(s validate.Severity) int {
	n := 0
	for _, f := range r.Findings {
		if f.Severity == s {
			n++
		}
	}
	return n
}

// Go reports whether the document passes: no finding is an error.
func (r *Result) Go() bool { return r.Count(validate.SeverityError) == 0 }

// Run checks doc with the rules of the profile. g is the document's graph.
// The result is the same for the same inputs.
func Run(doc *toc.Document, g *graph.Graph, o Options) (*Result, error) {
	if o.Profile == "" {
		o.Profile = ProfileStandard
	}
	if !o.Profile.Valid() {
		return nil, fmt.Errorf("%w: %q (want draft, standard, or full)", ErrUnknownProfile, o.Profile)
	}
	reg := o.Registry
	if reg == nil {
		reg = Default
	}
	disabled := map[string]bool{}
	for _, id := range o.Disable {
		if _, ok := reg.Rule(id); !ok {
			return nil, fmt.Errorf("%w: %q", ErrUnknownRule, id)
		}
		disabled[id] = true
	}
	res := &Result{Profile: o.Profile, Findings: []Finding{}, Ran: []string{}, Disabled: []string{}, rules: map[string]Rule{}}
	ctx := &Context{Doc: doc, Graph: g, AsOf: o.AsOf}
	for _, rule := range reg.Rules() {
		if !Runs(rule, o.Profile) {
			continue
		}
		res.rules[rule.ID()] = rule
		if disabled[rule.ID()] {
			res.Disabled = append(res.Disabled, rule.ID())
			continue
		}
		res.Ran = append(res.Ran, rule.ID())
		for _, f := range rule.Check(ctx) {
			f.Rule, f.Severity = rule.ID(), rule.Severity()
			res.Findings = append(res.Findings, f)
		}
	}
	slices.SortStableFunc(res.Findings, func(a, b Finding) int {
		return cmpStrings(a.Rule, b.Rule, a.Path, b.Path, a.Element, b.Element, a.Message, b.Message)
	})
	return res, nil
}

func cmpStrings(pairs ...string) int {
	for i := 0; i+1 < len(pairs); i += 2 {
		if c := strings.Compare(pairs[i], pairs[i+1]); c != 0 {
			return c
		}
	}
	return 0
}

// categoryNames title the report sections.
var categoryNames = map[string]string{
	"STR": "Structure", "LOG": "Causal logic", "ASM": "Assumptions", "MEA": "Measurement",
	"INT": "Interventions", "ACC": "Accountability", "ACT": "Actors", "EVD": "Evidence",
	"PTH": "Pathways", "MET": "Metrics", "CLM": "Claims", "BUD": "Budget", "AUD": "Audiences",
	"CAU": "Problem tree", "RSP": "Responsibility", "RSK": "Risks", "EQT": "Equity",
	"VLD": "Validation", "PRV": "Provenance",
}

// SummaryReport renders the result as a structured-evaluation summary: one
// team per rule category and one task per rule that ran. A rule with an error
// finding is NO-GO, a rule with warnings is WARN, and a rule with only
// advice is GO with the advice counted in its detail. generatedAt is stamped
// on the report as given, so the report is reproducible.
func (r *Result) SummaryReport(project, version string, generatedAt time.Time) *summary.SummaryReport {
	rep := &summary.SummaryReport{
		Project: project, Version: version, Phase: "THEORY OF CHANGE LINT (" + strings.ToUpper(string(r.Profile)) + ")",
		Teams: []summary.TeamSection{}, GeneratedAt: generatedAt, GeneratedBy: "theoryofchange lint",
	}
	byCat := map[string][]summary.TaskResult{}
	var cats []string
	add := func(id string, t summary.TaskResult) {
		c := Category(id)
		if _, ok := byCat[c]; !ok {
			cats = append(cats, c)
		}
		byCat[c] = append(byCat[c], t)
	}
	for _, id := range r.sortedRuleIDs() {
		rule := r.rules[id]
		var own []Finding
		for _, f := range r.Findings {
			if f.Rule == id {
				own = append(own, f)
			}
		}
		add(id, task(rule, own, slices.Contains(r.Disabled, id)))
	}
	slices.Sort(cats)
	for _, c := range cats {
		name := categoryNames[c]
		if name == "" {
			name = c
		}
		team := summary.TeamSection{ID: "lint-" + strings.ToLower(c), Name: name, Tasks: byCat[c]}
		team.ComputeStatus()
		rep.AddTeam(team)
	}
	rep.ComputeOverallStatus()
	return rep
}

func (r *Result) sortedRuleIDs() []string {
	ids := make([]string, 0, len(r.rules))
	for id := range r.rules {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	return ids
}

func task(rule Rule, findings []Finding, disabled bool) summary.TaskResult {
	t := summary.TaskResult{ID: rule.ID(), Status: summary.StatusGo, Detail: rule.Description()}
	switch {
	case disabled:
		t.Status, t.Detail = summary.StatusSkip, "disabled: "+rule.Description()
	case len(findings) > 0:
		switch rule.Severity() {
		case validate.SeverityError:
			t.Status = summary.StatusNoGo
		case validate.SeverityWarning:
			t.Status = summary.StatusWarn
		case validate.SeverityInfo:
		}
		t.Detail = fmt.Sprintf("%d finding(s): %s", len(findings), findings[0].Message)
		t.Metadata = map[string]any{"findings": findings}
	}
	return t
}
