package cli

import (
	"os"
	"strings"
	"testing"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/report"
)

func TestReportKinds(t *testing.T) {
	path := fixtureCopy(t, "full.json", "theory.json")
	for _, kind := range report.DefinitionKinds {
		r := run(t, "", "report", kind, path)
		if r.code != ExitOK || r.stdout == "" {
			t.Errorf("report %s: code %d, stderr %s", kind, r.code, r.err)
		}
	}
}

func TestReportDefinitionsHasEveryPart(t *testing.T) {
	path := fixtureCopy(t, "full.json", "theory.json")
	r := run(t, "", "report", "definitions", path)
	for _, want := range []string{"## Causal pathways", "## Logic model:", "## Metric definitions", "## Glossary"} {
		if !strings.Contains(r.stdout, want) {
			t.Errorf("missing %q", want)
		}
	}
}

func TestReportPathwayRestrictsLogicModel(t *testing.T) {
	path := fixtureCopy(t, "full.json", "theory.json")
	r := run(t, "", "report", "logic-model", path, "--pathway", "pw-supply")
	if r.code != ExitOK || strings.Count(r.stdout, "## Logic model:") != 1 {
		t.Errorf("code %d, output:\n%s%s", r.code, r.stdout, r.err)
	}
}

func TestReportUnknownKindAndPathway(t *testing.T) {
	path := fixtureCopy(t, "full.json", "theory.json")
	if r := run(t, "", "report", "nope", path); r.code != ExitUsageErr || !strings.Contains(r.err, "glossary") {
		t.Errorf("unknown kind: code %d, stderr %s", r.code, r.err)
	}
	if r := run(t, "", "report", "glossary", path, "--pathway", "pw-missing"); r.code != ExitUsageErr {
		t.Errorf("unknown pathway: code %d, stderr %s", r.code, r.err)
	}
}

func TestExportCSVToFile(t *testing.T) {
	path := fixtureCopy(t, "full.json", "theory.json")
	out := path + ".csv"
	r := run(t, "", "export", "csv", "metrics", path, "--out", out)
	if r.code != ExitOK || r.stdout != "" {
		t.Fatalf("code %d, stdout %q, stderr %s", r.code, r.stdout, r.err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(data), "id,name,definition,formula") {
		t.Errorf("unexpected CSV:\n%s", data)
	}
	if r := run(t, "", "export", "csv", "nope", path); r.code != ExitUsageErr {
		t.Errorf("unknown table: code %d", r.code)
	}
}
