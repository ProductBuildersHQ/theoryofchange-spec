package lint

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/validate"
)

// funcRule is a rule written as a function.
type funcRule struct {
	id       string
	severity validate.Severity
	desc     string
	check    func(*Context) []Finding
	from     Profile // lowest profile; empty for the default of the severity
}

func (r funcRule) MinProfile() Profile {
	if r.from != "" {
		return r.from
	}
	if r.severity == validate.SeverityError || Category(r.id) == CategorySTR {
		return ProfileDraft
	}
	if r.severity == validate.SeverityWarning {
		return ProfileStandard
	}
	return ProfileFull
}

func (r funcRule) ID() string                  { return r.id }
func (r funcRule) Severity() validate.Severity { return r.severity }
func (r funcRule) Description() string         { return r.desc }
func (r funcRule) Check(c *Context) []Finding  { return r.check(c) }

func register(id string, sev validate.Severity, desc string, check func(*Context) []Finding) {
	Register(funcRule{id: id, severity: sev, desc: desc, check: check})
}

// ---- helpers shared by the rule files --------------------------------------

func at(array string, i int) string { return "/" + array + "/" + strconv.Itoa(i) }

// activeOutcomes returns the outcomes that were not rejected, with their
// index in the document.
func activeOutcomes(d *toc.Document) []indexed[toc.Outcome] {
	var out []indexed[toc.Outcome]
	for i, o := range d.Outcomes {
		if !o.Rejected() {
			out = append(out, indexed[toc.Outcome]{i, o})
		}
	}
	return out
}

func activeLinks(d *toc.Document) []indexed[toc.CausalLink] {
	var out []indexed[toc.CausalLink]
	for i, l := range d.Links {
		if !l.Rejected() {
			out = append(out, indexed[toc.CausalLink]{i, l})
		}
	}
	return out
}

func activeGroups(d *toc.Document) []indexed[toc.PreconditionGroup] {
	var out []indexed[toc.PreconditionGroup]
	for i, g := range d.Groups {
		if !g.Rejected() {
			out = append(out, indexed[toc.PreconditionGroup]{i, g})
		}
	}
	return out
}

type indexed[T any] struct {
	i int
	v T
}

// outcomeLevel returns the level of an outcome ID, or false for any other ID.
func outcomeLevel(d *toc.Document, id string) (toc.Level, bool) {
	o, ok := d.Outcome(id)
	if !ok {
		return "", false
	}
	return o.Level, true
}

// ---- structure -------------------------------------------------------------

func init() {
	register("TOC-STR-001", validate.SeverityError, "The theory has at least one impact outcome.", func(c *Context) []Finding {
		for _, o := range activeOutcomes(c.Doc) {
			if o.v.Level == toc.LevelImpact {
				return nil
			}
		}
		return []Finding{{Path: "/outcomes", Message: "no outcome has level impact; say what the change is ultimately for"}}
	})

	register("TOC-STR-002", validate.SeverityError, "Causal links contain no cycle, unless marked as feedback.", func(c *Context) []Finding {
		var out []Finding
		for _, cyc := range c.Graph.Cycles() {
			out = append(out, Finding{Element: cyc[0], Message: fmt.Sprintf("these elements cause each other in a loop: %v; mark the link that closes the loop as feedback, or remove it", cyc)})
		}
		return out
	})

	register("TOC-STR-003", validate.SeverityError, "Every outcome below impact leads to an impact outcome.", func(c *Context) []Finding {
		if !hasImpact(c.Doc) {
			return nil // TOC-STR-001 reports the missing impact
		}
		dead := map[string]bool{}
		for _, id := range c.Graph.DeadEnds() {
			dead[id] = true
		}
		var out []Finding
		for _, o := range activeOutcomes(c.Doc) {
			if dead[o.v.ID] {
				out = append(out, Finding{Path: at("outcomes", o.i), Element: o.v.ID, Message: "no chain of links leads from this outcome to an impact outcome"})
			}
		}
		return out
	})

	register("TOC-STR-004", validate.SeverityWarning, "Every outcome has a precondition or an intervention that targets it.", func(c *Context) []Finding {
		orphan := map[string]bool{}
		for _, id := range c.Graph.Orphans() {
			orphan[id] = true
		}
		var out []Finding
		for _, o := range activeOutcomes(c.Doc) {
			if orphan[o.v.ID] {
				out = append(out, Finding{Path: at("outcomes", o.i), Element: o.v.ID, Message: "nothing leads to this outcome: add a causal link into it or an intervention that targets it"})
			}
		}
		return out
	})

	register("TOC-STR-005", validate.SeverityWarning, "An at_least rule on an outcome asks for between one and all of its preconditions.", func(c *Context) []Finding {
		var out []Finding
		for _, o := range activeOutcomes(c.Doc) {
			if o.v.Logic == nil || o.v.Logic.Type != toc.LogicAtLeast {
				continue
			}
			if n := preconditionCount(c, o.v.ID); o.v.Logic.K < 1 || o.v.Logic.K > n {
				out = append(out, Finding{Path: at("outcomes", o.i) + "/logic/k", Element: o.v.ID,
					Message: fmt.Sprintf("at_least %d of %d preconditions; k must be between 1 and %d", o.v.Logic.K, n, n)})
			}
		}
		return out
	})

	register("TOC-STR-006", validate.SeverityWarning, "Links run from lower levels to higher ones, unless they are feedback.", func(c *Context) []Finding {
		var out []Finding
		for _, l := range activeLinks(c.Doc) {
			from, ok1 := outcomeLevel(c.Doc, l.v.From)
			to, ok2 := outcomeLevel(c.Doc, l.v.To)
			if !ok1 || !ok2 || l.v.Feedback || from.Rank() <= to.Rank() {
				continue
			}
			out = append(out, Finding{Path: at("links", l.i), Element: l.v.ID,
				Message: fmt.Sprintf("this link runs from a %s outcome down to a %s outcome; reverse it, or mark it as feedback", from, to)})
		}
		return out
	})

	register("TOC-STR-007", validate.SeverityWarning, "Precondition groups are complete: each member leads into the group's target, and groups do not contain each other in a loop.", checkGroups)
}

func hasImpact(d *toc.Document) bool {
	for _, o := range activeOutcomes(d) {
		if o.v.Level == toc.LevelImpact {
			return true
		}
	}
	return false
}

// incomingSources returns the distinct sources of the active, non-feedback
// causal links into an outcome. Interventions that target the outcome are
// not counted: they bring it about, but outcomes do not combine them.
func incomingSources(d *toc.Document, id string) []string {
	var out []string
	for _, l := range activeLinks(d) {
		if l.v.To == id && !l.v.Feedback && !slices.Contains(out, l.v.From) {
			out = append(out, l.v.From)
		}
	}
	return out
}

// preconditionCount counts what an outcome's own logic combines: the
// incoming sources that belong to no group, plus each group aimed at it.
func preconditionCount(c *Context, id string) int {
	grouped := map[string]bool{}
	groups := c.Graph.GroupsFor(id)
	for _, g := range groups {
		for _, m := range g.Members {
			grouped[m] = true
		}
	}
	n := len(groups)
	for _, src := range incomingSources(c.Doc, id) {
		if !grouped[src] {
			n++
		}
	}
	return n
}

func checkGroups(c *Context) []Finding {
	var out []Finding
	active := map[string]toc.PreconditionGroup{}
	for _, g := range activeGroups(c.Doc) {
		active[g.v.ID] = g.v
	}
	linked := map[[2]string]bool{}
	for _, l := range activeLinks(c.Doc) {
		linked[[2]string{l.v.From, l.v.To}] = true
	}
	for _, g := range activeGroups(c.Doc) {
		p := at("preconditionGroups", g.i)
		if len(g.v.Members) < 2 {
			out = append(out, Finding{Path: p + "/members", Element: g.v.ID, Message: "a group needs at least two members"})
		}
		for j, m := range g.v.Members {
			if _, isGroup := active[m]; isGroup {
				continue // a nested group is checked as a group of its own
			}
			if !linked[[2]string{m, g.v.Target}] {
				out = append(out, Finding{Path: p + "/members/" + strconv.Itoa(j), Element: g.v.ID,
					Message: fmt.Sprintf("member %q has no causal link into the group's target %q", m, g.v.Target)})
			}
		}
		if groupLoops(active, g.v.ID) {
			out = append(out, Finding{Path: p + "/members", Element: g.v.ID, Message: "this group contains itself through its members"})
		}
	}
	return out
}

// groupLoops reports whether following group members from id returns to id.
func groupLoops(groups map[string]toc.PreconditionGroup, id string) bool {
	seen := map[string]bool{}
	var walk func(string) bool
	walk = func(cur string) bool {
		g, ok := groups[cur]
		if !ok {
			return false
		}
		for _, m := range g.Members {
			if m == id {
				return true
			}
			if _, isGroup := groups[m]; isGroup && !seen[m] {
				seen[m] = true
				if walk(m) {
					return true
				}
			}
		}
		return false
	}
	return walk(id)
}

// ---- causal logic ----------------------------------------------------------

func init() {
	aboveShortTerm := func(l toc.Level) bool { return l.Rank() > toc.LevelShortTerm.Rank() }
	linkRule := func(id, desc, msg string, ok func(toc.CausalLink) bool) {
		register(id, validate.SeverityWarning, desc, func(c *Context) []Finding {
			var out []Finding
			for _, l := range activeLinks(c.Doc) {
				if to, found := outcomeLevel(c.Doc, l.v.To); found && aboveShortTerm(to) && !ok(l.v) {
					out = append(out, Finding{Path: at("links", l.i), Element: l.v.ID, Message: msg})
				}
			}
			return out
		})
	}
	linkRule("TOC-LOG-001", "Every link into an intermediate, long-term, or impact outcome says why it holds.",
		"this link has no rationale; say why the first outcome makes the second more likely",
		func(l toc.CausalLink) bool { return l.Rationale != "" })
	linkRule("TOC-LOG-002", "Every link into an intermediate, long-term, or impact outcome names an assumption.",
		"this link has no assumption; say what must be true for it to hold",
		func(l toc.CausalLink) bool { return len(l.AssumptionIDs) > 0 })

	register("TOC-LOG-003", validate.SeverityInfo, "Impact outcomes have at least two preconditions.", func(c *Context) []Finding {
		var out []Finding
		for _, o := range activeOutcomes(c.Doc) {
			if o.v.Level == toc.LevelImpact && len(incomingSources(c.Doc, o.v.ID)) < 2 {
				out = append(out, Finding{Path: at("outcomes", o.i), Element: o.v.ID, Message: "a single chain leads to this impact; theories with one chain are usually incomplete"})
			}
		}
		return out
	})

	register("TOC-LOG-004", validate.SeverityInfo, "Outcomes with several incoming links say whether all, any, or some are needed.", func(c *Context) []Finding {
		var out []Finding
		for _, o := range activeOutcomes(c.Doc) {
			if len(incomingSources(c.Doc, o.v.ID)) < 2 || o.v.Logic != nil || len(c.Graph.GroupsFor(o.v.ID)) > 0 {
				continue
			}
			out = append(out, Finding{Path: at("outcomes", o.i), Element: o.v.ID, Message: "several things lead here; state whether all, any, or some of them are needed"})
		}
		return out
	})
}
