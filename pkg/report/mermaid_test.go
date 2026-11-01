package report

import (
	"regexp"
	"strings"
	"testing"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

func TestMermaidGolden(t *testing.T) {
	doc := fullDoc(t)
	golden(t, "pathway.mmd", Mermaid(build(t, doc, nil, Options{}), MermaidOptions{}))
	golden(t, "pathway-status.mmd", Mermaid(build(t, doc, fullStatus(t), Options{}), MermaidOptions{ShowIDs: true}))
}

func TestMermaidStructure(t *testing.T) {
	out := Mermaid(build(t, fullDoc(t), nil, Options{}), MermaidOptions{})
	for _, want := range []string{
		"flowchart BT\n",
		// The ceiling is a boundary: impact above, the rest within.
		`subgraph above["Above the ceiling of accountability: tracked, not owned"]`,
		`subgraph within["Within accountability: Long-term outcome and below"]`,
		`subgraph lv_impact["Impacts"]`,
		`subgraph lv_medium_term["Medium-term outcomes"]`,
		// A group is a logic node wired member -> group -> target.
		`n_grp_supply{"AND"}`,
		"n_oc_idp_support --> n_grp_supply\n",
		"n_oc_platform_support --> n_grp_supply\n",
		"n_grp_supply --> n_oc_enterprise_require\n",
		// A feedback link is dashed and says what kind of loop it is.
		`n_oc_enterprise_require -.->|"reinforcing loop"| n_oc_platform_support`,
		// Interventions are their own subgraph, linked to what they target.
		`subgraph interventions["Interventions"]`,
		`n_int_interop(["Run interoperability events"])`,
		"n_int_interop --> n_oc_spec_stable\n",
		// Inferred elements are visibly marked, in text and in style.
		"«inferred»",
		"class n_oc_idp_support inferred",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("diagram is missing %q", want)
		}
	}
	// Links carried by a group are not drawn a second time.
	for _, dup := range []string{"n_oc_idp_support --> n_oc_enterprise_require", "n_oc_platform_support --> n_oc_enterprise_require"} {
		if strings.Contains(out, dup) {
			t.Errorf("link %q duplicates the group wiring", dup)
		}
	}
	// Impact is above the ceiling, so it is not inside the within subgraph.
	within := out[strings.Index(out, "subgraph within"):]
	if strings.Contains(within[:strings.Index(within, "\n  end\n")], "oc_safer_agents") {
		t.Error("the impact outcome appears inside the accountability boundary")
	}
}

func TestMermaidStatusColoursAndLabels(t *testing.T) {
	out := Mermaid(build(t, fullDoc(t), fullStatus(t), Options{}), MermaidOptions{})
	if !strings.Contains(out, "class n_oc_idp_support green") {
		t.Errorf("an outcome with green health should be styled green:\n%s", out)
	}
	if !strings.Contains(out, "<i>«inferred» · On track</i>") {
		t.Errorf("the label should carry basis and pathway status:\n%s", out)
	}
	plain := Mermaid(build(t, fullDoc(t), nil, Options{}), MermaidOptions{})
	for _, colour := range []string{"class n_oc_idp_support green", "class n_oc_idp_support grey"} {
		if strings.Contains(plain, colour) {
			t.Errorf("without status there must be no health styling: %q", colour)
		}
	}
}

func TestMermaidOptions(t *testing.T) {
	v := build(t, fullDoc(t), nil, Options{})
	noInt := Mermaid(v, MermaidOptions{HideInterventions: true})
	if strings.Contains(noInt, "interventions[") || strings.Contains(noInt, "n_int_interop") {
		t.Errorf("interventions should be hidden:\n%s", noInt)
	}
	ids := Mermaid(v, MermaidOptions{ShowIDs: true})
	if !strings.Contains(ids, "oc-spec-stable") {
		t.Error("ShowIDs should put element IDs in labels")
	}
}

func TestMermaidWithoutACeilingHasNoBoundary(t *testing.T) {
	doc := fullDoc(t)
	doc.Ceiling = nil
	out := Mermaid(build(t, doc, nil, Options{}), MermaidOptions{})
	if strings.Contains(out, "subgraph above") || strings.Contains(out, "subgraph within") {
		t.Errorf("no ceiling means no boundary:\n%s", out)
	}
	if !strings.Contains(out, `subgraph lv_impact["Impacts"]`) {
		t.Error("levels should still be grouped")
	}
}

var (
	nodeDef   = regexp.MustCompile(`(?m)^\s*(n_[a-z0-9_]+)(?:\[|\(\[|\{)`)
	classLine = regexp.MustCompile(`(?m)^\s*class ([a-z0-9_,]+) [a-z]+$`)
	edgeLine  = regexp.MustCompile(`(?m)^\s*(n_[a-z0-9_]+) (?:-->|-\.->)(?:\|"[^"]*"\|)? (n_[a-z0-9_]+)$`)
)

func TestMermaidReferencesOnlyDefinedNodes(t *testing.T) {
	for _, st := range []*toc.Status{nil, fullStatus(t)} {
		out := Mermaid(build(t, fullDoc(t), st, Options{}), MermaidOptions{})
		defined := map[string]int{}
		for _, m := range nodeDef.FindAllStringSubmatch(out, -1) {
			defined[m[1]]++
		}
		for id, n := range defined {
			if n != 1 {
				t.Errorf("node %s is defined %d times", id, n)
			}
		}
		for _, m := range edgeLine.FindAllStringSubmatch(out, -1) {
			for _, id := range m[1:] {
				if defined[id] == 0 {
					t.Errorf("edge references undefined node %s", id)
				}
			}
		}
		for _, m := range classLine.FindAllStringSubmatch(out, -1) {
			for _, id := range strings.Split(m[1], ",") {
				if defined[id] == 0 {
					t.Errorf("class line references undefined node %s", id)
				}
			}
		}
	}
}

func TestMermaidEscapesLabels(t *testing.T) {
	doc := fullDoc(t)
	doc.Outcomes[2].Title = `Say "hello" <b>& goodbye</b>`
	out := Mermaid(build(t, doc, nil, Options{}), MermaidOptions{})
	if !strings.Contains(out, "Say #quot;hello#quot; #lt;b#gt;#amp; goodbye#lt;/b#gt;") {
		t.Errorf("label not escaped:\n%s", out)
	}
	if strings.Contains(out, `"hello"`) {
		t.Error("an unescaped quote would break the diagram")
	}
}

func TestMermaidRendersNegativePolarityAndDeterminism(t *testing.T) {
	doc := fullDoc(t)
	doc.Links[1].Polarity = toc.PolarityNegative
	v := build(t, doc, nil, Options{})
	out := Mermaid(v, MermaidOptions{})
	if !strings.Contains(out, `-->|"reduces"|`) {
		t.Errorf("a negative link should be labelled:\n%s", out)
	}
	if out != Mermaid(v, MermaidOptions{}) {
		t.Error("Mermaid is not deterministic")
	}
}

func TestMermaidPathwayFilter(t *testing.T) {
	doc := fullDoc(t)
	doc.Pathways = append(doc.Pathways, toc.Pathway{ID: "pw-platform", Title: "Platforms", OutcomeIDs: []string{"oc-platform-support"}})
	out := Mermaid(build(t, doc, nil, Options{Pathway: "pw-platform"}), MermaidOptions{})
	if strings.Contains(out, "oc_spec_stable") || !strings.Contains(out, "n_oc_platform_support") {
		t.Errorf("filtered diagram:\n%s", out)
	}
}
