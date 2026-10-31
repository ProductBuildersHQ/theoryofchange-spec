package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/diff"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/edit"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/judge"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/lint"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/report"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/service"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/slides"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/tocio"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/validate"
	"github.com/ProductBuildersHQ/theoryofchange-spec/schema"
)

// Version is the server's reported version.
const Version = "0.1.0"

// Server holds the confinement and builds the MCP server.
type Server struct {
	root *Root
}

// New returns a server whose file access is confined to root.
func New(root string) (*Server, error) {
	r, err := NewRoot(root)
	if err != nil {
		return nil, err
	}
	return &Server{root: r}, nil
}

// MCP returns the MCP server with every tool registered.
func (s *Server) MCP() *sdk.Server {
	srv := sdk.NewServer(&sdk.Implementation{Name: "theoryofchange", Version: Version}, nil)
	sdk.AddTool(srv, &sdk.Tool{Name: "toc_schema", Description: "List the embedded JSON Schemas, or return one by name."}, s.schemaTool)
	sdk.AddTool(srv, &sdk.Tool{Name: "toc_validate", Description: "Validate a theory against its schema and structurally. Returns whether it is valid and every issue found."}, s.validateTool)
	sdk.AddTool(srv, &sdk.Tool{Name: "toc_lint", Description: "Check the quality of a theory with the lint rules. Returns a structured-evaluation summary report."}, s.lintTool)
	sdk.AddTool(srv, &sdk.Tool{Name: "toc_graph", Description: "Query the causal graph: summary, upstream, downstream, paths, cycles, levels, orphans, dead-ends, or a Mermaid diagram."}, s.graphTool)
	sdk.AddTool(srv, &sdk.Tool{Name: "toc_status", Description: "Compute a status snapshot of a theory from observation sets as of a date."}, s.statusTool)
	sdk.AddTool(srv, &sdk.Tool{Name: "toc_report", Description: "Render a report as Markdown: a definition report, or a status report such as exec or story."}, s.reportTool)
	sdk.AddTool(srv, &sdk.Tool{Name: "toc_patch", Description: "Apply an RFC 6902 JSON Patch to a theory atomically. Elements may be addressed by id: /outcomes/@oc-skills/title. A patch that adds validation errors is rejected."}, s.patchTool)
	sdk.AddTool(srv, &sdk.Tool{Name: "toc_deck", Description: "Generate a slide deck as Marp Markdown: workshop, exec, learning-review, or story. Nothing is rendered to PDF."}, s.deckTool)
	sdk.AddTool(srv, &sdk.Tool{Name: "toc_judge_bundle", Description: "Build the bundle a judge reads to review a theory against the rubric. Makes no model calls."}, s.judgeTool)
	sdk.AddTool(srv, &sdk.Tool{Name: "toc_diff", Description: "Compare two versions of a theory by element id."}, s.diffTool)
	return srv
}

// Run serves over stdio until the client disconnects.
func (s *Server) Run(ctx context.Context) error {
	return s.MCP().Run(ctx, &sdk.StdioTransport{})
}

// ---- toc_schema -------------------------------------------------------------

type schemaIn struct {
	Name string `json:"name,omitempty" jsonschema:"the schema to return; omit to list the names"`
}

func (s *Server) schemaTool(_ context.Context, _ *sdk.CallToolRequest, in schemaIn) (*sdk.CallToolResult, any, error) {
	if in.Name == "" {
		return nil, map[string]any{"names": schema.Names()}, nil
	}
	data, err := schema.Get(in.Name)
	if err != nil {
		return nil, nil, err
	}
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, nil, fmt.Errorf("schema %s: %w", in.Name, err)
	}
	return nil, v, nil
}

// ---- toc_validate -----------------------------------------------------------

type validateIn struct {
	Source
	Kind string `json:"kind,omitempty" jsonschema:"ignored; the kind is read from the document"`
}

type validateOut struct {
	Valid  bool             `json:"valid"`
	Kind   string           `json:"kind,omitempty"`
	Issues []validate.Issue `json:"issues"`
}

func (s *Server) validateTool(_ context.Context, _ *sdk.CallToolRequest, in validateIn) (*sdk.CallToolResult, validateOut, error) {
	var (
		data []byte
		f    tocio.Format
		err  error
	)
	switch {
	case in.Document != nil:
		data, f, err = inlineBytes(in.Document)
	case in.Path != "":
		var path string
		if path, err = s.root.Resolve(in.Path); err == nil {
			var rep service.Report
			if rep, err = service.ValidateFile(path, nil); err == nil {
				return nil, validateOut{Valid: rep.Valid(), Kind: string(rep.Kind), Issues: rep.Issues}, nil
			}
		}
	default:
		err = fmt.Errorf("give a document or a path")
	}
	if err != nil {
		return nil, validateOut{}, err
	}
	rep := service.ValidateBytes("inline", data, f, nil)
	return nil, validateOut{Valid: rep.Valid(), Kind: string(rep.Kind), Issues: rep.Issues}, nil
}

// ---- toc_lint ---------------------------------------------------------------

type lintIn struct {
	Source
	Profile string   `json:"profile,omitempty" jsonschema:"draft, standard (default), or full"`
	Disable []string `json:"disable,omitempty" jsonschema:"rule ids to skip"`
}

func (s *Server) lintTool(_ context.Context, _ *sdk.CallToolRequest, in lintIn) (*sdk.CallToolResult, any, error) {
	doc, _, err := s.load(in.Source)
	if err != nil {
		return nil, nil, err
	}
	g, err := service.BuildGraph(doc)
	if err != nil {
		return nil, nil, err
	}
	res, err := lint.Run(doc, g, lint.Options{Profile: lint.Profile(in.Profile), Disable: in.Disable})
	if err != nil {
		return nil, nil, err
	}
	return nil, res.SummaryReport(doc.ID, doc.Metadata.Version, time.Time{}), nil
}

// ---- toc_graph --------------------------------------------------------------

type graphIn struct {
	Source
	Query string `json:"query" jsonschema:"summary, upstream, downstream, paths, cycles, levels, orphans, dead-ends, export, or mermaid"`
	ID    string `json:"id,omitempty" jsonschema:"the outcome or intervention for upstream, downstream, and paths"`
}

func (s *Server) graphTool(_ context.Context, _ *sdk.CallToolRequest, in graphIn) (*sdk.CallToolResult, any, error) {
	doc, _, err := s.load(in.Source)
	if err != nil {
		return nil, nil, err
	}
	g, err := service.BuildGraph(doc)
	if err != nil {
		return nil, nil, err
	}
	if in.Query == "mermaid" {
		v, err := report.BuildView(doc, g, nil, report.Options{})
		if err != nil {
			return nil, nil, err
		}
		return nil, map[string]any{"format": "mermaid", "content": report.Mermaid(v, report.MermaidOptions{})}, nil
	}
	q := service.GraphQuery(in.Query)
	out, err := service.RunGraphQuery(g, q, in.ID, 0)
	if err != nil {
		return nil, nil, err
	}
	return nil, map[string]any{"query": in.Query, "result": out}, nil
}

// ---- toc_status -------------------------------------------------------------

type statusIn struct {
	Source
	Observations
	AsOf string `json:"asOf,omitempty" jsonschema:"the date to describe, YYYY-MM-DD; default: the latest date in the observations"`
}

func (s *Server) statusTool(_ context.Context, _ *sdk.CallToolRequest, in statusIn) (*sdk.CallToolResult, any, error) {
	doc, _, err := s.load(in.Source)
	if err != nil {
		return nil, nil, err
	}
	sets, err := s.loadObservations(in.Observations)
	if err != nil {
		return nil, nil, err
	}
	asOf, err := asOfDate(in.AsOf, sets)
	if err != nil {
		return nil, nil, err
	}
	if issues := validate.Observations(doc, sets...); validate.HasErrors(issues) {
		return nil, nil, fmt.Errorf("the observations are invalid: %s: %s", issues[0].Code, issues[0].Message)
	}
	input, err := service.StatusInputFor(doc, sets, asOf, toc.DefaultStatusOptions(), nil)
	if err != nil {
		return nil, nil, err
	}
	return nil, input.Status, nil
}

// ---- toc_report -------------------------------------------------------------

type reportIn struct {
	Source
	Observations
	Kind     string `json:"kind" jsonschema:"a definition report (pathway, logic-model, outcome-map, logframe, metrics, glossary, assessment, definitions, ...) or a status report (exec, results, assumptions, interventions, evidence-gaps, full, story, headline, learning-review, portfolio, narrative, actors, custom)"`
	AsOf     string `json:"asOf,omitempty" jsonschema:"the status date, YYYY-MM-DD"`
	Pathway  string `json:"pathway,omitempty" jsonschema:"restrict a definition report to one pathway"`
	Audience string `json:"audience,omitempty" jsonschema:"the audience id, for the story report; narrows the headline report"`
	Period   string `json:"period,omitempty" jsonschema:"the period label, for the learning-review report"`
	Template string `json:"template,omitempty" jsonschema:"the text/template source, for the custom report"`
}

type textOut struct {
	Format  string `json:"format"`
	Content string `json:"content"`
}

func (s *Server) reportTool(_ context.Context, _ *sdk.CallToolRequest, in reportIn) (*sdk.CallToolResult, textOut, error) {
	doc, _, err := s.load(in.Source)
	if err != nil {
		return nil, textOut{}, err
	}
	if !service.IsStatusReport(in.Kind) {
		text, err := service.DefinitionReport(doc, in.Kind, in.Pathway)
		return nil, textOut{Format: "markdown", Content: text}, err
	}
	input, sets, withStatus, err := s.statusInput(doc, in.Observations, in.AsOf, in.Kind)
	if err != nil {
		return nil, textOut{}, err
	}
	text, err := service.StatusReport(in.Kind, input, sets, service.ReportParams{Audience: in.Audience, Period: in.Period, Template: in.Template}, withStatus)
	return nil, textOut{Format: "markdown", Content: text}, err
}

// statusInput evaluates the theory for a status report or deck. A narrative,
// actors, or custom report with no observations and no date has no status.
func (s *Server) statusInput(doc *toc.Document, obs Observations, asOf, kind string) (report.StatusInput, []*toc.ObservationSet, bool, error) {
	sets, err := s.loadObservations(obs)
	if err != nil {
		return report.StatusInput{}, nil, false, err
	}
	if slices.Contains(service.OptionalStatusKinds, kind) && asOf == "" && len(sets) == 0 {
		in, err := service.ViewWithoutStatus(doc)
		return in, nil, false, err
	}
	at, err := asOfDate(asOf, sets)
	if err != nil {
		return report.StatusInput{}, nil, false, err
	}
	if issues := validate.Observations(doc, sets...); validate.HasErrors(issues) {
		return report.StatusInput{}, nil, false, fmt.Errorf("the observations are invalid: %s: %s", issues[0].Code, issues[0].Message)
	}
	in, err := service.StatusInputFor(doc, sets, at, toc.DefaultStatusOptions(), nil)
	return in, sets, true, err
}

// ---- toc_patch --------------------------------------------------------------

// PatchOp is one RFC 6902 operation.
type PatchOp struct {
	Op    string `json:"op" jsonschema:"add, remove, replace, move, copy, or test"`
	Path  string `json:"path" jsonschema:"a JSON pointer; use /outcomes/@id/field to address an element by id, and - to append"`
	From  string `json:"from,omitempty" jsonschema:"the source pointer, for move and copy"`
	Value any    `json:"value,omitempty" jsonschema:"the value, for add, replace, and test"`
}

type patchIn struct {
	Source
	Patch       []PatchOp `json:"patch" jsonschema:"the operations to apply, in order, as one unit"`
	DryRun      bool      `json:"dryRun,omitempty" jsonschema:"compute the result without writing a file"`
	AllowErrors bool      `json:"allowErrors,omitempty" jsonschema:"accept a patch that adds structural errors (never a schema violation)"`
}

type patchOut struct {
	Applied          bool             `json:"applied"`
	Reason           string           `json:"reason,omitempty"`
	Written          bool             `json:"written,omitempty"`
	Document         any              `json:"document,omitempty"`
	IssuesIntroduced []validate.Issue `json:"issuesIntroduced"`
	IssuesResolved   []validate.Issue `json:"issuesResolved"`
}

func (s *Server) patchTool(_ context.Context, _ *sdk.CallToolRequest, in patchIn) (*sdk.CallToolResult, patchOut, error) {
	raw, err := json.Marshal(in.Patch)
	if err != nil {
		return nil, patchOut{}, fmt.Errorf("the patch cannot be encoded: %w", err)
	}
	patch, err := edit.ParsePatch(raw)
	if err != nil {
		return nil, patchOut{}, err
	}
	doc, path, err := s.load(in.Source)
	if err != nil {
		return nil, patchOut{}, err
	}
	// A theory given by path is written back unless this is a dry run.
	if path != "" && !in.DryRun {
		res, err := service.PatchFile(path, patch, service.PatchOptions{AllowErrors: in.AllowErrors})
		if err != nil {
			return nil, patchOut{}, err
		}
		return nil, patchOut{Applied: res.Applied, Reason: res.Reason, Written: res.Written, IssuesIntroduced: res.IssuesIntroduced, IssuesResolved: res.IssuesResolved}, nil
	}
	res, err := edit.Apply(doc, patch, edit.Options{AllowErrors: in.AllowErrors})
	if err != nil {
		return nil, patchOut{}, err
	}
	out := patchOut{Applied: res.Applied, Reason: res.Reason, IssuesIntroduced: nonNil(res.IssuesIntroduced), IssuesResolved: nonNil(res.IssuesResolved)}
	if res.Applied {
		canon, err := tocio.Canonical(res.Document, tocio.FormatJSON)
		if err != nil {
			return nil, patchOut{}, err
		}
		var v any
		if err := json.Unmarshal(canon, &v); err != nil {
			return nil, patchOut{}, err
		}
		out.Document = v
	}
	return nil, out, nil
}

func nonNil(i []validate.Issue) []validate.Issue {
	if i == nil {
		return []validate.Issue{}
	}
	return i
}

// ---- toc_deck ---------------------------------------------------------------

type deckIn struct {
	Source
	Observations
	Kind     string `json:"kind" jsonschema:"workshop, exec, learning-review, or story"`
	AsOf     string `json:"asOf,omitempty" jsonschema:"the status date, YYYY-MM-DD"`
	Audience string `json:"audience,omitempty" jsonschema:"the audience id, for the story deck"`
	Period   string `json:"period,omitempty" jsonschema:"the period label, for the learning-review deck"`
	Theme    string `json:"theme,omitempty" jsonschema:"the Marp theme"`
}

func (s *Server) deckTool(_ context.Context, _ *sdk.CallToolRequest, in deckIn) (*sdk.CallToolResult, textOut, error) {
	doc, _, err := s.load(in.Source)
	if err != nil {
		return nil, textOut{}, err
	}
	if in.Kind == "workshop" {
		md, err := service.WorkshopDeck(doc, in.Theme, "")
		return nil, textOut{Format: "marp", Content: md}, err
	}
	if !slices.Contains([]string{"exec", "learning-review", "story"}, in.Kind) {
		return nil, textOut{}, fmt.Errorf("unknown deck %q (want workshop, exec, learning-review, or story)", in.Kind)
	}
	input, sets, _, err := s.statusInput(doc, in.Observations, in.AsOf, in.Kind)
	if err != nil {
		return nil, textOut{}, err
	}
	opts := slides.Options{Theme: in.Theme}
	var md string
	switch in.Kind {
	case "exec":
		md = slides.Exec(input, opts)
	case "learning-review":
		md = slides.LearningReview(report.LearningInput{StatusInput: input, Sets: sets, Period: in.Period}, opts)
	default:
		if in.Audience == "" {
			return nil, textOut{}, fmt.Errorf("the story deck needs an audience")
		}
		md, err = slides.Story(input, sets, in.Audience, opts)
		if err != nil {
			return nil, textOut{}, err
		}
	}
	return nil, textOut{Format: "marp", Content: md}, nil
}

// ---- toc_judge_bundle -------------------------------------------------------

type judgeIn struct {
	Source
	Profile string `json:"profile,omitempty" jsonschema:"core (default) or development"`
}

func (s *Server) judgeTool(_ context.Context, _ *sdk.CallToolRequest, in judgeIn) (*sdk.CallToolResult, any, error) {
	doc, _, err := s.load(in.Source)
	if err != nil {
		return nil, nil, err
	}
	profile := in.Profile
	if profile == "" {
		profile = "core"
	}
	b, err := judge.Build(doc, profile)
	if err != nil {
		return nil, nil, err
	}
	return nil, b, nil
}

// ---- toc_diff ---------------------------------------------------------------

type diffIn struct {
	Old Source `json:"old" jsonschema:"the earlier version"`
	New Source `json:"new" jsonschema:"the later version"`
}

func (s *Server) diffTool(_ context.Context, _ *sdk.CallToolRequest, in diffIn) (*sdk.CallToolResult, any, error) {
	old, _, err := s.load(in.Old)
	if err != nil {
		return nil, nil, fmt.Errorf("old: %w", err)
	}
	cur, _, err := s.load(in.New)
	if err != nil {
		return nil, nil, fmt.Errorf("new: %w", err)
	}
	res, err := diff.Compare(old, cur)
	if err != nil {
		return nil, nil, err
	}
	return nil, res, nil
}
