package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/edit"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/service"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/tocio"
)

type flagType int

const (
	typeString flagType = iota
	typeStrings
	typeInt
	typeBool
)

// elementFlag is a typed convenience flag for one JSON field. Field may be
// dotted to reach a nested object ("logic.type").
type elementFlag struct {
	name     string
	field    string
	typ      flagType
	required bool // required by add
	help     string
}

func s(name, field, help string) elementFlag {
	return elementFlag{name, field, typeString, false, help}
}
func sr(name, field, help string) elementFlag {
	return elementFlag{name, field, typeString, true, help}
}
func ss(name, field, help string) elementFlag {
	return elementFlag{name, field, typeStrings, false, help}
}
func ssr(name, field, help string) elementFlag {
	return elementFlag{name, field, typeStrings, true, help}
}
func b(name, field, help string) elementFlag { return elementFlag{name, field, typeBool, false, help} }
func n(name, field, help string) elementFlag { return elementFlag{name, field, typeInt, false, help} }

// elementFlags lists the typed flags per element kind. Any other field is
// reachable with --set.
var elementFlags = map[string][]elementFlag{
	"outcome": {
		sr("title", "title", "short title (a state, not an activity)"),
		sr("level", "level", "impact, long_term, intermediate, short_term, or output"),
		s("description", "description", "longer description"),
		s("sphere", "sphere", "control, influence, or interest"),
		s("category", "category", "taxonomy category"),
		s("change-type", "changeType", "knowledge, attitude, skill, behavior, practice, policy, norm, condition, investment, or adoption"),
		s("population", "population", "who or what changes"),
		ss("actors", "actorIds", "actor IDs who change or co-own the outcome"),
	},
	"group": {
		sr("target", "target", "outcome the group combines preconditions for"),
		ssr("members", "members", "outcome or group IDs (at least two)"),
		sr("logic", "logic.type", "all, any, or at_least"),
		n("k", "logic.k", "how many members are needed with --logic at_least"),
		s("sufficiency", "sufficiency", "why the members are jointly sufficient"),
	},
	"link": {
		sr("from", "from", "outcome or intervention ID"),
		sr("to", "to", "outcome ID"),
		s("type", "type", "causes, enables, contributes_to, or necessary_condition"),
		s("rationale", "rationale", "why the link holds"),
		ss("assumptions", "assumptionIds", "assumption IDs underpinning the link"),
		ss("evidence", "evidenceIds", "evidence IDs"),
		ss("interventions", "interventionIds", "interventions placed on the link"),
		b("feedback", "feedback", "mark as an intentional feedback loop"),
		s("loop-type", "loopType", "reinforcing or balancing (with --feedback)"),
		s("lag", "lag", "ISO 8601 duration, e.g. P6M"),
	},
	"intervention": {
		sr("title", "title", "short title"),
		ssr("targets", "targets", "outcome IDs the intervention targets"),
		s("description", "description", "longer description"),
		s("type", "type", "program, product, policy, advocacy, partnership, standardization, investment, communication, research, or other"),
		s("owner", "ownerActorId", "owning actor ID"),
		s("status", "status", "planned, in_progress, completed, paused, or cancelled"),
		ss("outputs", "outputs", "direct products"),
		s("resources", "resources", "resources required"),
	},
	"assumption": {
		sr("statement", "statement", "the assumption, ideally if X then Y because Z"),
		s("type", "type", "causal, contextual, implementation, behavioral, external, or partnership"),
		ss("applies-to", "appliesTo", "link, outcome, or intervention IDs"),
		s("criticality", "criticality", "low, medium, high, or critical"),
		s("plausibility", "plausibility", "low, medium, or high"),
		b("needs-testing", "needsTesting", "the assumption must be tested"),
		s("test", "test", "how it will be tested"),
		s("status", "status", "untested, holding, at_risk, invalidated, or validated"),
		s("owner", "ownerActorId", "owning actor ID"),
	},
	"indicator": {
		sr("measures", "measures", "outcome, intervention, or assumption ID"),
		sr("name", "name", "what is measured"),
		s("definition", "definition", "precise definition"),
		s("type", "measurementType", "quantitative, qualitative, or hybrid"),
		s("unit", "unit", "unit of measure"),
		s("direction", "direction", "increase, decrease, or maintain"),
		s("frequency", "frequency", "ISO 8601 duration, e.g. P1M"),
		s("data-source", "dataSource", "where the data comes from"),
		s("data-availability", "dataAvailability", "public, internal_existing, internal_new, partner_held, or unknown"),
		s("method", "method", "how it is collected"),
		s("owner", "ownerActorId", "owning actor ID"),
		s("metric", "metricId", "shared metric this indicator instantiates"),
		ss("pathways", "pathwayIds", "pathway IDs the indicator serves"),
		s("claim", "claim", "attribution or contribution (default contribution)"),
		ss("caveats", "caveats", "caveats to show wherever it is reported"),
		ss("rollup-from", "rollupFrom", "indicator IDs of the same metric that combine into this one"),
	},
	"actor": {
		sr("name", "name", "display name"),
		s("type", "type", "organization, person, group, community, customer_segment, standards_body, regulator, vendor, platform, investor, or other"),
		ss("roles", "roles", "implementer, beneficiary, boundary_partner, funder, champion, influencer, competitor, adopter, or gatekeeper"),
		s("stance", "stance", "champion, supporter, neutral, skeptic, opponent, or unknown"),
		s("influence", "influence", "low, medium, or high"),
		s("interest", "interest", "low, medium, or high"),
		s("description", "description", "longer description"),
	},
	"evidence": {
		sr("title", "title", "source title"),
		s("type", "type", "research, evaluation, monitoring_data, expert_opinion, stakeholder_testimony, case_study, internal_metric, market_signal, outcome_harvest, or anecdote"),
		s("strength", "strength", "strong, moderate, weak, or anecdotal"),
		s("direction", "direction", "supports or contradicts"),
		s("url", "url", "where to find it"),
		s("source", "source", "author or publisher"),
		s("summary", "summary", "what it says"),
		s("excerpt", "excerpt", "a short quotation"),
		s("sensitivity", "sensitivity", "public, internal, or restricted"),
		s("accessed", "accessedAt", "date accessed (YYYY-MM-DD)"),
	},
	"factor": {
		sr("description", "description", "the external condition"),
		s("kind", "kind", "enabler, constraint, trend, or shock"),
		ss("affects", "affects", "IDs of affected elements"),
		s("likelihood", "likelihood", "low, medium, or high"),
	},
	"risk": {
		sr("description", "description", "the risk"),
		s("impact", "impact", "low, medium, high, or critical"),
		s("probability", "probability", "low, medium, or high"),
		s("mitigation", "mitigation", "how it is managed"),
		s("status", "status", "open, mitigated, accepted, or closed"),
		s("owner", "owner", "who owns the risk (free text)"),
		s("category", "category", "environmental_political, opportunity, design, partnership, market, technical, or operational"),
		ss("applies-to", "appliesTo", "IDs of affected elements"),
		ss("related-assumptions", "relatedAssumptionIds", "assumptions this risk mirrors"),
	},
	"metric": {
		sr("name", "name", "metric name"),
		sr("definition", "definition", "what it measures"),
		sr("formula", "formula", "exactly how it is computed"),
		s("unit", "unit", "unit of measure"),
		s("direction", "direction", "increase, decrease, or maintain"),
		s("aggregation", "aggregation", "sum, mean, weighted_mean, min, max, latest, or none"),
		s("weight-metric", "weightMetricId", "metric that weights a weighted_mean"),
		s("layer", "layer", "common or pathway"),
	},
	"term": {
		sr("term", "term", "the word or phrase"),
		sr("definition", "definition", "what it means here"),
		ss("examples", "examples", "examples from this organization"),
		ss("related", "relatedIds", "related element IDs"),
	},
	"consultation": {
		sr("date", "date", "date held (YYYY-MM-DD)"),
		sr("step", "step", "focus, change_analysis, assumptions_risks, actors, measurement, or whole"),
		s("method", "method", "workshop, interview, survey, document_review, evidence_review, or meeting"),
		ss("groups", "stakeholderGroups", "stakeholder groups or roles that took part (not individuals)"),
		n("count", "count", "how many people took part"),
		ss("findings", "findings", "what was learned"),
		ss("changed", "changedIds", "IDs of elements changed as a result"),
	},
	"pathway": {
		sr("title", "title", "pathway name, e.g. a program area"),
		s("kind", "kind", "program, strategy, theme, geography, or segment"),
		s("description", "description", "longer description"),
		s("owner", "ownerActorId", "owning actor ID"),
		ss("outcomes", "outcomeIds", "outcome IDs in the pathway"),
		ss("interventions", "interventionIds", "intervention IDs in the pathway"),
		s("child", "childToc", "theory that elaborates this pathway, as toc:<doc>"),
	},
	"decision": {
		sr("question", "question", "what needs deciding"),
		ss("options", "options", "the options"),
		s("due", "dueDate", "due date (YYYY-MM-DD)"),
		s("owner", "ownerActorId", "deciding actor ID"),
		s("status", "status", "open, decided, or deferred"),
		ss("related", "relatedIds", "related element IDs"),
	},
}

// provenanceKinds are the element kinds that carry a provenance block.
var provenanceKinds = []string{
	"outcome", "group", "link", "intervention", "assumption", "indicator",
	"actor", "factor", "risk", "metric", "pathway",
}

func init() {
	for _, kind := range provenanceKinds {
		elementFlags[kind] = append(elementFlags[kind],
			s("basis", "provenance.basis", "why it is in the theory: stated, inferred, hypothesis, confirmed, or rejected"),
			ss("sources", "provenance.evidenceIds", "evidence IDs that support the basis"),
			s("basis-note", "provenance.note", "note on the basis"),
			s("decided-at", "provenance.decidedAt", "date a confirmed or rejected basis was decided (YYYY-MM-DD)"),
			s("decided-by", "provenance.decidedBy", "group or role that decided it"),
		)
	}
}

func newElementCmds(e *env) []*cobra.Command {
	var cmds []*cobra.Command
	for _, k := range edit.ElementKinds() {
		cmds = append(cmds, newKindCmd(e, k))
	}
	return cmds
}

func newKindCmd(e *env, k edit.ElementKind) *cobra.Command {
	flags := elementFlags[k.Name]
	parent := &cobra.Command{
		Use:   k.Name,
		Short: fmt.Sprintf("Add, change, or remove %s", article(k.Name)),
		Long: fmt.Sprintf(`Add, change, or remove %s.

Each command is compiled to a JSON Patch and applied through the same
validated path as "theoryofchange patch": it is atomic, and nothing is
written if the result would have more structural errors than before.`, article(k.Name)),
	}
	parent.AddCommand(newAddCmd(e, k, flags), newSetCmd(e, k, flags), newRmCmd(e, k))
	return parent
}

func addCommonFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("dry-run", false, "do not write; print the resulting document")
	cmd.Flags().Bool("allow-errors", false, "accept a change that adds structural errors (never a schema violation)")
}

func addFieldFlags(cmd *cobra.Command, flags []elementFlag, forAdd bool) {
	for _, f := range flags {
		help := f.help
		if forAdd && f.required {
			help += " (required)"
		}
		switch f.typ {
		case typeString:
			cmd.Flags().String(f.name, "", help)
		case typeStrings:
			cmd.Flags().StringSlice(f.name, nil, help)
		case typeInt:
			cmd.Flags().Int(f.name, 0, help)
		case typeBool:
			cmd.Flags().Bool(f.name, false, help)
		}
		if forAdd && f.required {
			if err := cmd.MarkFlagRequired(f.name); err != nil {
				panic(fmt.Sprintf("cli: mark --%s required: %v", f.name, err))
			}
		}
	}
	cmd.Flags().StringArray("set", nil, "set any field by its JSON name, as key=value; dotted keys reach nested objects; JSON values (numbers, true, [..], {..}) are parsed, quote to force a string")
}

// collectFields gathers the typed flags that were given and any --set
// pairs into one field map.
func collectFields(cmd *cobra.Command, flags []elementFlag) (map[string]any, error) {
	fields := map[string]any{}
	put := func(key string, v any) error {
		if _, dup := fields[key]; dup {
			return usageErr("field %q is set more than once", key)
		}
		fields[key] = v
		return nil
	}
	for _, f := range flags {
		if !cmd.Flags().Changed(f.name) {
			continue
		}
		var v any
		switch f.typ {
		case typeString:
			v = flagString(cmd, f.name)
		case typeStrings:
			sl, err := cmd.Flags().GetStringSlice(f.name)
			if err != nil {
				panic(fmt.Sprintf("cli: flag --%s: %v", f.name, err))
			}
			v = sl
		case typeInt:
			v = flagInt(cmd, f.name)
		case typeBool:
			v = flagBool(cmd, f.name)
		}
		if err := put(f.field, v); err != nil {
			return nil, err
		}
	}
	pairs, err := cmd.Flags().GetStringArray("set")
	if err != nil {
		panic(fmt.Sprintf("cli: flag --set: %v", err))
	}
	for _, kv := range pairs {
		key, val, ok := strings.Cut(kv, "=")
		if !ok || key == "" {
			return nil, usageErr("--set wants key=value, got %q", kv)
		}
		if err := put(key, edit.ParseValue(val)); err != nil {
			return nil, err
		}
	}
	return fields, nil
}

func (e *env) applyElementPatch(cmd *cobra.Command, path string, p edit.Patch) error {
	format, err := outputFormat(cmd)
	if err != nil {
		return err
	}
	res, err := service.PatchFile(path, p, service.PatchOptions{
		DryRun:      flagBool(cmd, "dry-run"),
		AllowErrors: flagBool(cmd, "allow-errors"),
	})
	if err != nil {
		return &ExitError{Code: ExitUsageErr, Err: err}
	}
	return e.reportPatch(cmd, format, res)
}

func newAddCmd(e *env, k edit.ElementKind, flags []elementFlag) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <toc-file> --id <id> ...",
		Short: fmt.Sprintf("Add %s", article(k.Name)),
		Example: fmt.Sprintf(`  theoryofchange %s add theory.yaml --id %s-new %s`,
			k.Name, shortPrefix(k.Name), exampleFlags(flags)),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fields, err := collectFields(cmd, flags)
			if err != nil {
				return err
			}
			doc, err := tocio.LoadDocument(args[0])
			if err != nil {
				return &ExitError{Code: ExitUsageErr, Err: err}
			}
			id := flagString(cmd, "id")
			if id == "" && k.Name == "evidence" {
				title, _ := fields["title"].(string)
				if id, err = evidenceID(doc, title); err != nil {
					return usageErr("%v", err)
				}
			}
			fields["id"] = id
			p, err := edit.AddElement(doc, k, fields)
			if err != nil {
				return usageErr("%v", err)
			}
			return e.applyElementPatch(cmd, args[0], p)
		},
	}
	if k.Name == "evidence" {
		cmd.Use = "add <toc-file> --title <title> ..."
		cmd.Example = "  theoryofchange evidence add theory.yaml --title \"Annual report\" --url https://example.org/report --accessed 2026-01-15"
		cmd.Flags().String("id", "", "the new element's ID, a lowercase kebab-case slug (default: derived from the title)")
	} else {
		cmd.Flags().String("id", "", "the new element's ID, a lowercase kebab-case slug (required)")
		if err := cmd.MarkFlagRequired("id"); err != nil {
			panic(fmt.Sprintf("cli: mark --id required: %v", err))
		}
	}
	addFieldFlags(cmd, flags, true)
	addCommonFlags(cmd)
	return cmd
}

func newSetCmd(e *env, k edit.ElementKind, flags []elementFlag) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <toc-file> <id>",
		Short: fmt.Sprintf("Change fields of %s", article(k.Name)),
		Long: fmt.Sprintf(`Change fields of an existing %s. Give typed flags or --set key=value for any
field, and --unset key to remove an optional field. An element's ID cannot
be changed: remove the element and add a new one.`, k.Name),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			fields, err := collectFields(cmd, flags)
			if err != nil {
				return err
			}
			unset, err := cmd.Flags().GetStringArray("unset")
			if err != nil {
				panic(fmt.Sprintf("cli: flag --unset: %v", err))
			}
			doc, err := tocio.LoadDocument(args[0])
			if err != nil {
				return &ExitError{Code: ExitUsageErr, Err: err}
			}
			p, err := edit.SetFields(doc, k, args[1], fields, unset)
			if err != nil {
				return usageErr("%v", err)
			}
			return e.applyElementPatch(cmd, args[0], p)
		},
	}
	addFieldFlags(cmd, flags, false)
	cmd.Flags().StringArray("unset", nil, "remove an optional field by its JSON name (repeatable)")
	addCommonFlags(cmd)
	return cmd
}

func newRmCmd(e *env, k edit.ElementKind) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rm <toc-file> <id>",
		Short: fmt.Sprintf("Remove %s", article(k.Name)),
		Long: fmt.Sprintf(`Remove %s. If anything still references it, the change is rejected and
the references are listed. With --cascade, references are removed too, and
any element that cannot survive losing one is removed as well: a link whose
end is gone, an indicator whose subject is gone, an intervention left with
no targets, a precondition group left with fewer than two members.`, article(k.Name)),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			doc, err := tocio.LoadDocument(args[0])
			if err != nil {
				return &ExitError{Code: ExitUsageErr, Err: err}
			}
			p, err := edit.RemoveElement(doc, k, args[1], flagBool(cmd, "cascade"))
			if err != nil {
				return &ExitError{Code: ExitUsageErr, Err: err}
			}
			return e.applyElementPatch(cmd, args[0], p)
		},
	}
	cmd.Flags().Bool("cascade", false, "also remove references to the element")
	addCommonFlags(cmd)
	return cmd
}

// shortPrefix is the conventional ID prefix for an element kind.
func shortPrefix(kind string) string {
	switch kind {
	case "outcome":
		return "oc"
	case "group":
		return "grp"
	case "link":
		return "ln"
	case "intervention":
		return "int"
	case "assumption":
		return "asm"
	case "indicator":
		return "ind"
	case "actor":
		return "act"
	case "evidence":
		return "ev"
	case "factor":
		return "ef"
	case "risk":
		return "rsk"
	case "pathway":
		return "pw"
	case "metric":
		return "met"
	case "term":
		return "term"
	case "consultation":
		return "con"
	default:
		return "dec"
	}
}

// exampleFlags renders the required flags for a usage example.
func exampleFlags(flags []elementFlag) string {
	var parts []string
	for _, f := range flags {
		if !f.required {
			continue
		}
		switch f.typ {
		case typeStrings:
			parts = append(parts, fmt.Sprintf("--%s a,b", f.name))
		default:
			parts = append(parts, fmt.Sprintf("--%s <%s>", f.name, f.name))
		}
	}
	return strings.Join(parts, " ")
}

// article prefixes a kind name with "a" or "an".
func article(kind string) string {
	if strings.ContainsRune("aeiou", rune(kind[0])) {
		return "an " + kind
	}
	return "a " + kind
}

// evidenceID derives an unused evidence ID from a title: "ev-" and a
// kebab-case slug, with a numeric suffix if the slug is taken. The same
// document and title always yield the same ID.
func evidenceID(doc *toc.Document, title string) (string, error) {
	var slug strings.Builder
	dash := true
	for _, r := range strings.ToLower(title) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			slug.WriteRune(r)
			dash = false
		case !dash:
			slug.WriteByte('-')
			dash = true
		}
	}
	base := strings.Trim(slug.String(), "-")
	if base == "" {
		return "", fmt.Errorf("cannot derive an ID from the title %q; give --id", title)
	}
	if len(base) > 40 {
		base = strings.Trim(base[:40], "-")
	}
	taken := make(map[string]bool, len(doc.Evidence))
	for _, ev := range doc.Evidence {
		taken[ev.ID] = true
	}
	id := "ev-" + base
	for n := 2; taken[id]; n++ {
		id = fmt.Sprintf("ev-%s-%d", base, n)
	}
	return id, nil
}
