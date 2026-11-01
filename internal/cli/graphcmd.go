package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/impact-model/pkg/graph"
	"github.com/ProductBuildersHQ/impact-model/pkg/impactio"
	"github.com/ProductBuildersHQ/impact-model/pkg/service"
)

// selectors maps each query flag to its query.
var selectors = []struct {
	flag  string
	query service.GraphQuery
	help  string
}{
	{"upstream", service.QueryUpstream, "list everything that must hold for --from (transitive preconditions)"},
	{"downstream", service.QueryDownstream, "list everything that --from contributes to"},
	{"paths", service.QueryPaths, "list the paths from --from to the impact outcomes"},
	{"cycles", service.QueryCycles, "list cycles that are not marked as feedback"},
	{"levels", service.QueryLevels, "list each node's longest distance to an impact outcome"},
	{"orphans", service.QueryOrphans, "list outcomes with no preconditions and no intervention"},
	{"dead-ends", service.QueryDeadEnds, "list nodes with no path to an impact outcome"},
}

func newGraphCmd(e *env) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "graph <model-file>",
		Short: "Analyze the causal graph of an impact model",
		Long: `Summarize or query the causal graph: outcomes and interventions joined by
causal links. Feedback links are excluded from every structural query.

With no query flag, prints a summary. With --format dot, prints the whole
graph in Graphviz DOT.`,
		Example: `  impact-model graph impact-model.yaml
  impact-model graph impact-model.yaml --upstream --from oc-enterprise-require
  impact-model graph impact-model.yaml --paths --from oc-spec-stable
  impact-model graph impact-model.yaml --cycles --format json
  impact-model graph impact-model.yaml --format dot | dot -Tsvg > graph.svg`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			format := flagString(cmd, "format")
			if format != "text" && format != "json" && format != "dot" {
				return usageErr("--format must be text, json, or dot, not %q", format)
			}
			query, err := chosenQuery(cmd)
			if err != nil {
				return err
			}
			if format == "dot" && query != service.QuerySummary {
				return usageErr("--format dot prints the whole graph; it cannot be combined with --%s", query)
			}
			id := flagString(cmd, "from")
			if query.NeedsID() && id == "" {
				return usageErr("--%s needs --from <id>", query)
			}

			doc, err := impactio.LoadDocument(args[0])
			if err != nil {
				return &ExitError{Code: ExitUsageErr, Err: err}
			}
			g, err := service.BuildGraph(doc)
			if err != nil {
				if errors.Is(err, graph.ErrInvalidDocument) {
					return &ExitError{Code: ExitInvalid, Err: fmt.Errorf("%s: %w", args[0], err)}
				}
				return &ExitError{Code: ExitUsageErr, Err: err}
			}
			if format == "dot" {
				fmt.Fprint(e.out, service.DOT(g))
				return nil
			}
			result, err := service.RunGraphQuery(g, query, id, flagInt(cmd, "max-paths"))
			if err != nil {
				return usageErr("%v", err)
			}
			if format == "json" {
				if err := writeJSON(e.out, result); err != nil {
					return &ExitError{Code: ExitUsageErr, Err: err}
				}
				return nil
			}
			renderGraphText(e, result)
			return nil
		},
	}
	// The graph command's own --format adds dot to the persistent text/json.
	cmd.Flags().String("format", "text", "output format: text, json, or dot")
	cmd.Flags().String("from", "", "outcome or intervention ID for --upstream, --downstream, and --paths")
	cmd.Flags().Int("max-paths", graph.DefaultMaxPaths, "maximum paths to list with --paths")
	for _, s := range selectors {
		cmd.Flags().Bool(s.flag, false, s.help)
	}
	return cmd
}

// chosenQuery returns the single selected query, or the summary.
func chosenQuery(cmd *cobra.Command) (service.GraphQuery, error) {
	var chosen []service.GraphQuery
	for _, s := range selectors {
		if flagBool(cmd, s.flag) {
			chosen = append(chosen, s.query)
		}
	}
	switch len(chosen) {
	case 0:
		return service.QuerySummary, nil
	case 1:
		return chosen[0], nil
	default:
		names := make([]string, len(chosen))
		for i, q := range chosen {
			names[i] = "--" + string(q)
		}
		return "", usageErr("choose one query, not %s", strings.Join(names, " and "))
	}
}

func renderGraphText(e *env, result any) {
	switch r := result.(type) {
	case service.Summary:
		fmt.Fprintf(e.out, "nodes:           %d\n", r.Nodes)
		fmt.Fprintf(e.out, "edges:           %d (%d feedback)\n", r.Edges, r.FeedbackEdges)
		fmt.Fprintf(e.out, "impact outcomes: %s\n", orNone(r.ImpactOutcomes))
		fmt.Fprintf(e.out, "cycles:          %s\n", orNone(joinCycles(r.Cycles)))
		fmt.Fprintf(e.out, "orphans:         %s\n", orNone(r.Orphans))
		fmt.Fprintf(e.out, "dead ends:       %s\n", orNone(r.DeadEnds))
	case []string:
		for _, id := range r {
			fmt.Fprintln(e.out, id)
		}
	case [][]string:
		for _, p := range r {
			fmt.Fprintln(e.out, strings.Join(p, " -> "))
		}
	case []service.LevelEntry:
		for _, l := range r {
			fmt.Fprintf(e.out, "%d\t%s\n", l.Level, l.ID)
		}
	case service.Export:
		for _, edge := range r.Edges {
			fmt.Fprintf(e.out, "%s -> %s\n", edge.From, edge.To)
		}
	}
}

func orNone(items []string) string {
	if len(items) == 0 {
		return "none"
	}
	return strings.Join(items, ", ")
}

func joinCycles(cycles [][]string) []string {
	out := make([]string, len(cycles))
	for i, c := range cycles {
		out[i] = "(" + strings.Join(c, ", ") + ")"
	}
	return out
}
