package report

import (
	"strings"
	"testing"

	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
)

func TestDefinitionReportsGolden(t *testing.T) {
	doc := fullDoc(t)
	whole := build(t, doc, nil, Options{})
	pathways, err := PathwayViews(doc, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	golden(t, "logic-models.md", LogicModels(whole, pathways))
	golden(t, "metrics.md", MetricDefinitions(whole))
	golden(t, "glossary.md", Glossary(whole))
}

func TestLogicModelProjection(t *testing.T) {
	doc := fullDoc(t)
	doc.Interventions[0].Resources = "Two engineers and an event budget"
	pathways, err := PathwayViews(doc, nil, nil)
	if err != nil || len(pathways) != 1 {
		t.Fatalf("PathwayViews: %v (%d)", err, len(pathways))
	}
	out := LogicModel(pathways[0])
	for _, want := range []string{
		"## Logic model: Supply side: providers and platforms",
		"| Inputs | Two engineers and an event budget |",
		"| Activities | Run interoperability events; Publish reference implementations |",
		// Outputs combine listed outputs and output-level outcomes.
		"Interop test suite; Event reports; Open reference implementations are available",
		"| Short-term outcomes | A stable, interoperable specification exists |",
		"| Medium-term outcomes | Major identity providers support the standard _(inferred)_; Agent platforms support the standard |",
		"| Long-term outcomes | Enterprises require the standard in procurement _(hypothesis)_ |",
		"| Impact | Agent access across the market is scoped and auditable |",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("logic model is missing %q:\n%s", want, out)
		}
	}
}

func TestLogicModelsFallBackToTheWholeTheory(t *testing.T) {
	doc := fullDoc(t)
	doc.Pathways = nil
	for i := range doc.Indicators {
		doc.Indicators[i].PathwayIDs = nil
	}
	whole := build(t, doc, nil, Options{})
	out := LogicModels(whole, nil)
	if !strings.Contains(out, "## Logic model: Open agent authorization standard adoption") {
		t.Errorf("An impact model without pathways should render one whole-model view:\n%s", out)
	}
	if strings.Count(out, "## Logic model") != 1 {
		t.Errorf("want exactly one logic model:\n%s", out)
	}
}

func TestEmptyStagesShowADash(t *testing.T) {
	var d impact.Model
	fixture(t, "minimal.json", &d)
	out := LogicModel(build(t, &d, nil, Options{}))
	if !strings.Contains(out, "| Inputs | — |") || !strings.Contains(out, "| Activities | — |") {
		t.Errorf("empty stages should show a dash:\n%s", out)
	}
}

func TestMetricDefinitions(t *testing.T) {
	out := MetricDefinitions(build(t, fullDoc(t), nil, Options{}))
	for _, want := range []string{
		"| Providers with GA support _(confirmed)_",
		"Mean weighted by Providers with GA support",
		"Providers with GA support (total)",
	} {
		// "confirmed" is not a marked basis, so no marker is expected on it.
		if want == "| Providers with GA support _(confirmed)_" {
			if strings.Contains(out, "_(confirmed)_") {
				t.Error("only inferred and hypothesis elements are marked")
			}
			continue
		}
		if !strings.Contains(out, want) {
			t.Errorf("metric definitions are missing %q:\n%s", want, out)
		}
	}

	doc := fullDoc(t)
	doc.Metrics[0].Aggregation = impact.AggregationNone
	doc.Indicators[5].RollupFrom = nil // the roll-up needs an aggregating metric
	out = MetricDefinitions(build(t, doc, nil, Options{}))
	if !strings.Contains(out, "Cannot be added up: observed directly at each level") {
		t.Errorf("a non-aggregating metric should say so in plain words:\n%s", out)
	}

	doc.Metrics = nil
	for i := range doc.Indicators {
		doc.Indicators[i].MetricID = ""
	}
	if got := MetricDefinitions(build(t, doc, nil, Options{})); !strings.Contains(got, "defines no shared metrics") {
		t.Errorf("empty library = %q", got)
	}
}

func TestAggregationText(t *testing.T) {
	for agg, want := range map[impact.Aggregation]string{
		impact.AggregationSum: "Sum", impact.AggregationMean: "Mean", impact.AggregationMin: "Minimum",
		impact.AggregationMax: "Maximum", impact.AggregationLatest: "Latest value", "": "—",
	} {
		if got := AggregationText(MetricView{Aggregation: string(agg)}); got != want {
			t.Errorf("AggregationText(%q) = %q, want %q", agg, got, want)
		}
	}
	if got := AggregationText(MetricView{Aggregation: "weighted_mean"}); got != "Weighted mean" {
		t.Errorf("weighted mean without a weight = %q", got)
	}
}

func TestGlossary(t *testing.T) {
	out := Glossary(build(t, fullDoc(t), nil, Options{}))
	if !strings.Contains(out, "| General availability | A release available to all customers") {
		t.Errorf("glossary:\n%s", out)
	}
	doc := fullDoc(t)
	doc.Glossary = nil
	if got := Glossary(build(t, doc, nil, Options{})); !strings.Contains(got, "defines no glossary terms") {
		t.Errorf("empty glossary = %q", got)
	}
}

func TestMarkdownCellsAreEscaped(t *testing.T) {
	doc := fullDoc(t)
	doc.Glossary[0].Definition = "A | pipe and a\nnewline"
	out := Glossary(build(t, doc, nil, Options{}))
	if !strings.Contains(out, `A \| pipe and a newline`) {
		t.Errorf("cell not escaped:\n%s", out)
	}
	if mdText("a  b\n c") != "a b c" || orDash("") != "—" {
		t.Error("helper mismatch")
	}
}

func TestDefinitionReportsAreDeterministic(t *testing.T) {
	doc := fullDoc(t)
	a, b := build(t, doc, nil, Options{}), build(t, doc, nil, Options{})
	if MetricDefinitions(a) != MetricDefinitions(b) || Glossary(a) != Glossary(b) || LogicModel(a) != LogicModel(b) {
		t.Fatal("reports differ between runs")
	}
}
