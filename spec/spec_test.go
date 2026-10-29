package spec

import (
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/plexusone/structured-evaluation/rubric"
)

func weightSum(rs *rubric.RubricSet) float64 {
	var sum float64
	for _, c := range rs.Categories {
		sum += c.Weight
	}
	return sum
}

func TestRubricsParseAndValidate(t *testing.T) {
	for _, p := range Profiles() {
		rs, err := Rubric(p)
		if err != nil {
			t.Fatalf("%s: %v", p, err)
		}
		if issues := rs.Validate(); len(issues) != 0 {
			t.Errorf("%s: %v", p, issues)
		}
		if sum := weightSum(rs); math.Abs(sum-1) > 1e-9 {
			t.Errorf("%s: weights sum to %v, want 1", p, sum)
		}
		if len(rs.JudgeInstructions) == 0 || rs.EvaluationType != rubric.EvaluationTypeAnalytic {
			t.Errorf("%s: missing judge instructions or not analytic", p)
		}
		seen := map[string]bool{}
		for _, c := range rs.Categories {
			if seen[c.ID] {
				t.Errorf("%s: duplicate category %q", p, c.ID)
			}
			seen[c.ID] = true
			if len(c.Scale.Options) != 3 {
				t.Errorf("%s/%s: want pass, partial, fail options", p, c.ID)
			}
		}
	}
}

// The UN checklist, item by item (research section 9.6), and the category
// that judges it. Sector-neutral items belong to the core rubric.
var checklist = []struct {
	item, category, profile string
}{
	{"Collaborative, participatory process with multiple perspectives", "process", ProfileCore},
	{"Plausible, clear, logical flow without leaps of faith", "logic_flow", ProfileCore},
	{"Root and underlying causes tackled, not only immediate ones", "root_causes", ProfileCore},
	{"States the assumptions and risks most relevant to whether change will happen", "assumptions_risks", ProfileCore},
	{"Identifies who does what and the key partners", "actors", ProfileCore},
	{"Refers to evidence from credible sources", "evidence", ProfileCore},
	{"Identifies limitations in the evidence base", "evidence", ProfileCore},
	{"Names specific changes for targeted groups rather than assuming they benefit", "targeting", ProfileCore},
	{"A diagram embedded in the narrative", "presentation", ProfileCore},
	{"Benefits the most vulnerable", "leave_no_one_behind", ProfileDevelopment},
	{"Targets women and girls", "gender", ProfileDevelopment},
	{"Resilience to shocks", "resilience", ProfileDevelopment},
	{"Poverty-environment links", "poverty_environment", ProfileDevelopment},
	{"Strengthens institutions; structural causes of violence", "institutions_violence", ProfileDevelopment},
}

func TestEveryChecklistItemMapsToACategory(t *testing.T) {
	for _, c := range checklist {
		rs, err := Rubric(c.profile)
		if err != nil {
			t.Fatal(err)
		}
		if rs.GetCategory(c.category) == nil {
			t.Errorf("%q: no category %q in the %s rubric", c.item, c.category, c.profile)
		}
	}
	// Sector-specific items must not leak into the core rubric.
	core, _ := Rubric(ProfileCore)
	for _, c := range checklist {
		if c.profile == ProfileDevelopment && core.GetCategory(c.category) != nil {
			t.Errorf("development category %q is in the core rubric", c.category)
		}
	}
}

func TestDevelopmentExtendsCore(t *testing.T) {
	core, _ := Rubric(ProfileCore)
	dev, _ := Rubric(ProfileDevelopment)
	if len(dev.Categories) != len(core.Categories)+5 {
		t.Fatalf("development has %d categories, core %d", len(dev.Categories), len(core.Categories))
	}
	for i, c := range core.Categories {
		d := dev.Categories[i]
		if d.ID != c.ID || d.Description != c.Description || d.Weight >= c.Weight {
			t.Errorf("category %d: development %q (weight %v) does not extend core %q (weight %v)", i, d.ID, d.Weight, c.ID, c.Weight)
		}
	}
	if core.ID == dev.ID {
		t.Error("the two rubrics share an ID")
	}
}

func TestRubricCallsDoNotShareState(t *testing.T) {
	a, _ := Rubric(ProfileCore)
	a.Categories[0].Weight = 99
	b, _ := Rubric(ProfileCore)
	if b.Categories[0].Weight == 99 {
		t.Error("a change to one rubric leaked into the next")
	}
}

func TestUnknownProfile(t *testing.T) {
	if _, err := Rubric("military"); !errors.Is(err, ErrUnknownProfile) {
		t.Errorf("error = %v", err)
	}
}

func TestTemplateCoversTheDocument(t *testing.T) {
	for _, want := range []string{"## 1. Problem and context", "## 5. Assumptions and risks", "## 8. Measurement", "## 10. The picture"} {
		if !strings.Contains(Template(), want) {
			t.Errorf("template is missing %q", want)
		}
	}
}
