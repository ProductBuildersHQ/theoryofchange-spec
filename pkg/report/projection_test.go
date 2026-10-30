package report

import (
	"regexp"
	"strings"
	"testing"
)

func TestProjectionGoldens(t *testing.T) {
	doc := fullDoc(t)
	whole := build(t, doc, nil, Options{})
	pathways, err := PathwayViews(doc, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	golden(t, "outcome-map.md", OutcomeMaps(whole, pathways))
	golden(t, "logframe.md", Logframes(whole, pathways))
	golden(t, "logic-model-diagram.md", LogicModelDiagrams(whole, pathways))
}

// A theory with no pathways projects the whole theory once.
func TestProjectionsOfATheoryWithNoPathways(t *testing.T) {
	doc := fullDoc(t)
	doc.Pathways = nil
	for i := range doc.Indicators {
		doc.Indicators[i].PathwayIDs = nil
	}
	for i := range doc.Metrics {
		doc.Metrics[i].Layer = ""
	}
	whole := build(t, doc, nil, Options{})
	for name, out := range map[string]string{
		"outcome map": OutcomeMaps(whole, nil), "logframe": Logframes(whole, nil), "diagram": LogicModelDiagrams(whole, nil),
	} {
		if n := strings.Count(out, "\n## "); n != 0 || !strings.HasPrefix(out, "## ") {
			t.Errorf("%s: want exactly one section, got %d extra", name, n)
		}
	}
}

func TestMultiplePathwaysProjectOneSectionEach(t *testing.T) {
	in := exampleInput(t)
	whole := in.View
	pathways, err := PathwayViews(in.Doc, nil, in.Status)
	if err != nil {
		t.Fatal(err)
	}
	if len(pathways) < 2 {
		t.Fatal("the example should have several pathways")
	}
	for name, out := range map[string]string{"outcome map": OutcomeMaps(whole, pathways), "logframe": Logframes(whole, pathways)} {
		if n := strings.Count(out, "\n## ") + 1; n != len(pathways) {
			t.Errorf("%s: %d sections for %d pathways", name, n, len(pathways))
		}
	}
}

var idInBackticks = regexp.MustCompile("`[a-z][a-z0-9-]+`")

func TestEveryProjectedRowNamesItsElements(t *testing.T) {
	in := exampleInput(t)
	pathways, _ := PathwayViews(in.Doc, nil, in.Status)
	for name, out := range map[string]string{"outcome map": OutcomeMaps(in.View, pathways), "logframe": Logframes(in.View, pathways)} {
		checked := 0
		for _, line := range strings.Split(out, "\n") {
			if !strings.HasPrefix(line, "| ") || strings.HasPrefix(line, "| Outcome") || strings.HasPrefix(line, "| Level") || strings.HasPrefix(line, "|--") {
				continue
			}
			cells := strings.Split(line, " | ")
			first := cells[0]
			if name == "logframe" {
				first = cells[1] // the narrative follows the level
			}
			if !idInBackticks.MatchString(first) {
				t.Errorf("%s: a row does not name its element: %.120s", name, line)
			}
			checked++
		}
		if checked == 0 {
			t.Errorf("%s: no rows checked", name)
		}
	}
}

func TestOutcomeMapMarksWhatIsMonitoredOnly(t *testing.T) {
	in := exampleInput(t)
	out := OutcomeMap(in.View)
	if !strings.Contains(out, "Partner · monitored only") {
		t.Errorf("a partner-delivered outcome is not marked:\n%.800s", out)
	}
	if !strings.Contains(out, "above the ceiling, monitored only") {
		t.Errorf("an outcome above the ceiling is not marked")
	}
}

func TestLogframeCarriesAssumptionsAndVerification(t *testing.T) {
	in := exampleInput(t)
	out := Logframe(in.View)
	for _, want := range []string{"| Goal |", "| Purpose |", "| Outputs |", "| Activities |", "Attendance sheets", "`asm-"} {
		if !strings.Contains(out, want) {
			t.Errorf("logframe is missing %q", want)
		}
	}
	if !strings.Contains(out, "a view of the causal model") {
		t.Error("the projection note is missing")
	}
}

func TestLogicModelDiagramJoinsOnlyTheStagesThatHaveContent(t *testing.T) {
	doc := fullDoc(t)
	v := build(t, doc, nil, Options{})
	d := LogicModelMermaid(v)
	if !strings.HasPrefix(d, "flowchart LR\n") || strings.Count(d, "subgraph ") != strings.Count(d, "\n  end\n") {
		t.Errorf("malformed diagram:\n%s", d)
	}
	stages := strings.Count(d, "subgraph ")
	if arrows := strings.Count(d, " --> "); arrows != stages-1 {
		t.Errorf("%d stages need %d arrows, got %d", stages, stages-1, arrows)
	}
	if d != LogicModelMermaid(v) {
		t.Error("the diagram is not deterministic")
	}
}
