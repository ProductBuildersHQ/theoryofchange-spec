package logicmodel

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
	"github.com/ProductBuildersHQ/impact-model/pkg/report"
)

func TestSpecHasAColumnForEachStage(t *testing.T) {
	g := Build(view())
	spec := g.Spec()
	if spec.GridColumns != len(g.Stages) || len(spec.Containers) != len(g.Stages) {
		t.Errorf("%d columns and %d containers for %d stages", spec.GridColumns, len(spec.Containers), len(g.Stages))
	}
	for i, c := range spec.Containers {
		if c.ID != string(g.Stages[i].Stage) || c.Label != g.Stages[i].Stage.Name() || c.GridColumns != 1 {
			t.Errorf("container %d = %+v", i, c)
		}
	}
	if len(spec.Edges) != len(g.Edges) {
		t.Errorf("%d edges for %d graph edges", len(spec.Edges), len(g.Edges))
	}
}

func TestD2NamesNodesByTheirStage(t *testing.T) {
	d2 := Build(view()).D2()
	for _, want := range []string{
		"grid-columns: 7",
		"horizontal-gap: 90",
		"classes: {",
		"inputs: Inputs {",
		"class: stage",
		`inputs._input1 -> activities."int-a"`,
		`activities."int-a" -> outputs._output1`,
		"Practice applied (inferred)",
	} {
		if !strings.Contains(d2, want) {
			t.Errorf("missing %q:\n%s", want, d2)
		}
	}
	if d2 != Build(view()).D2() {
		t.Error("the same view gave different D2")
	}
}

func TestMermaidDrawsEveryNodeAndEdge(t *testing.T) {
	g := Build(view())
	out, err := g.Mermaid()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out, "flowchart LR\n") {
		t.Errorf("output:\n%s", out)
	}
	if got := strings.Count(out, "-->"); got != len(g.Edges) {
		t.Errorf("%d arrows for %d edges:\n%s", got, len(g.Edges), out)
	}
	for _, s := range g.Stages {
		for _, n := range s.Nodes {
			if !strings.Contains(out, `"`+n.Label+`"`) {
				t.Errorf("no node %q:\n%s", n.Label, out)
			}
		}
	}
}

// A dot in an element's ID must not read as a step into a container.
func TestADotInAnIDIsNotAPath(t *testing.T) {
	v := &report.View{Outcomes: []report.OutcomeView{
		{ID: "a.1", Title: "A", Level: impact.LevelShortTerm},
		{ID: "b.1", Title: "B", Level: impact.LevelMediumTerm},
	}, Links: []report.LinkView{{From: "a.1", To: "b.1"}}}
	d2 := Build(v).D2()
	if !strings.Contains(d2, `short_term.a_1 -> medium_term.b_1`) {
		t.Errorf("D2:\n%s", d2)
	}
}

// With the d2 tool installed, the D2 must compile and every label must reach
// the picture. The test is skipped without it, so a build without d2 still
// passes.
func TestD2CompilesAndKeepsEveryLabel(t *testing.T) {
	tool, err := exec.LookPath("d2")
	if err != nil {
		t.Skip("d2 is not installed")
	}
	g := Build(view())
	dir := t.TempDir()
	in, out := filepath.Join(dir, "lm.d2"), filepath.Join(dir, "lm.svg")
	if err := os.WriteFile(in, []byte(g.D2()), 0o600); err != nil {
		t.Fatal(err)
	}
	if msg, err := exec.Command(tool, "--layout", "elk", in, out).CombinedOutput(); err != nil {
		t.Fatalf("d2 could not compile the diagram: %v\n%s\n%s", err, msg, g.D2())
	}
	svg, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range g.Stages {
		if !strings.Contains(string(svg), s.Stage.Name()) {
			t.Errorf("the heading %q is not in the picture", s.Stage.Name())
		}
		for _, n := range s.Nodes {
			if !strings.Contains(string(svg), n.Label) {
				t.Errorf("the label %q is not in the picture", n.Label)
			}
		}
	}
}
