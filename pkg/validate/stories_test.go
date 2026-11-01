package validate

import (
	"testing"

	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
)

func TestAudienceAndThemeIssues(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(d *impact.Model)
		code   string
		path   string
	}{
		{"audience without name", func(d *impact.Model) { d.Audiences[0].Name = " " }, "VAL-AUD-001", "/audiences/0/name"},
		{"theme without label", func(d *impact.Model) { d.Themes[0].Label = "" }, "VAL-THM-001", "/themes/0/label"},
		{"theme without definition", func(d *impact.Model) { d.Themes[0].Definition = "" }, "VAL-THM-001", "/themes/0/definition"},
		{"theme outcome unknown", func(d *impact.Model) { d.Themes[0].OutcomeIDs = []string{"oc-ghost"} }, "VAL-REF-001", "/themes/0/outcomeIds/0"},
		{"outcome audience unknown", func(d *impact.Model) { d.Outcomes[0].AudienceIDs = []string{"aud-ghost"} }, "VAL-REF-001", "/outcomes/0/audienceIds/0"},
		{"indicator audience is a theme", func(d *impact.Model) { d.Indicators[0].AudienceIDs = []string{"thm-first-adoption"} }, "VAL-REF-002", "/indicators/0/audienceIds/0"},
		{"indicator theme unknown", func(d *impact.Model) { d.Indicators[6].ThemeIDs = []string{"thm-ghost"} }, "VAL-REF-001", "/indicators/6/themeIds/0"},
		{"theme count without themes", func(d *impact.Model) { d.Indicators[6].ThemeIDs = nil }, "VAL-THM-003", "/indicators/6/themeIds"},
		{"themes on another kind of indicator", func(d *impact.Model) { d.Indicators[0].ThemeIDs = []string{"thm-first-adoption"} }, "VAL-THM-004", "/indicators/0/themeIds"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := fullDoc(t)
			if doc.Indicators[6].ID != "ind-adoption-stories" {
				t.Fatalf("test setup: Indicators[6] is %q", doc.Indicators[6].ID)
			}
			tt.mutate(doc)
			for _, i := range Document(doc) {
				if i.Code == tt.code && i.Path == tt.path {
					return
				}
			}
			t.Fatalf("expected %s at %s; got:%s", tt.code, tt.path, describe(Document(doc)))
		})
	}
}

func TestStoryIssues(t *testing.T) {
	tests := []struct {
		name     string
		mutate   func(s *impact.Story)
		code     string
		path     string
		severity Severity
	}{
		{"unknown theme", func(s *impact.Story) { s.ThemeIDs = []string{"thm-ghost"} }, "VAL-REF-001", "/stories/0/themeIds/0", SeverityError},
		{"theme is an outcome", func(s *impact.Story) { s.ThemeIDs = []string{"oc-idp-support"} }, "VAL-REF-002", "/stories/0/themeIds/0", SeverityError},
		{"unknown outcome", func(s *impact.Story) { s.OutcomeIDs = []string{"oc-ghost"} }, "VAL-REF-001", "/stories/0/outcomeIds/0", SeverityError},
		{"unknown audience", func(s *impact.Story) { s.AudienceIDs = []string{"aud-ghost"} }, "VAL-REF-001", "/stories/0/audienceIds/0", SeverityError},
		{"missing consent", func(s *impact.Story) { s.Consent = "" }, "VAL-STORY-002", "/stories/0/consent", SeverityError},
		{"invalid consent", func(s *impact.Story) { s.Consent = "maybe" }, "VAL-ENUM-001", "/stories/0/consent", SeverityError},
		{"quote without consent", func(s *impact.Story) { s.Consent = impact.ConsentUnknown }, "VAL-STORY-003", "/stories/0/quote", SeverityWarning},
		{"missing date", func(s *impact.Story) { s.Date = "" }, "VAL-DATE-003", "/stories/0/date", SeverityError},
		{"unknown actor", func(s *impact.Story) { s.ActorID = "act-ghost" }, "VAL-REF-001", "/stories/0/actorId", SeverityError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := obsSet(t)
			tt.mutate(&set.Stories[0])
			for _, i := range Observations(fullDoc(t), set) {
				if i.Code == tt.code && i.Path == tt.path && i.Severity == tt.severity {
					return
				}
			}
			t.Fatalf("expected %s %s at %s; got:%s", tt.severity, tt.code, tt.path, describe(Observations(fullDoc(t), set)))
		})
	}
}

func TestDuplicateStoryIDAndReviewPeriod(t *testing.T) {
	set := obsSet(t)
	set.Stories = append(set.Stories, set.Stories[0])
	set.Reviews[0].Period = ""
	set.Reviews[0].RelatedIDs = []string{"ghost"}
	want := map[string]bool{"VAL-STORY-001": false, "VAL-REV-001": false, "VAL-REF-003": false}
	for _, i := range Observations(fullDoc(t), set) {
		if _, ok := want[i.Code]; ok {
			want[i.Code] = true
		}
	}
	for code, found := range want {
		if !found {
			t.Errorf("missing %s", code)
		}
	}
}
