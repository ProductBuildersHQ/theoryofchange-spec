package validate

import (
	"testing"

	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
)

func TestChangeAnalysisIssues(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(d *impact.Model)
		code   string
		path   string
	}{
		{"cause without statement", func(d *impact.Model) { d.Causes[0].Statement = " " }, "VAL-CAU-001", "/causes/0/statement"},
		{"cause without level", func(d *impact.Model) { d.Causes[0].Level = "" }, "VAL-CAU-002", "/causes/0/level"},
		{"cause level invalid", func(d *impact.Model) { d.Causes[0].Level = "deep" }, "VAL-ENUM-001", "/causes/0/level"},
		{"cause parent unknown", func(d *impact.Model) { d.Causes[1].Parents = []string{"cause-ghost"} }, "VAL-REF-001", "/causes/1/parents/0"},
		{"cause parent is an outcome", func(d *impact.Model) { d.Causes[1].Parents = []string{"oc-idp-support"} }, "VAL-REF-002", "/causes/1/parents/0"},
		{"cause is its own parent", func(d *impact.Model) { d.Causes[0].Parents = []string{d.Causes[0].ID} }, "VAL-CAU-003", "/causes/0/parents"},
		{"causes form a loop", func(d *impact.Model) { d.Causes[0].Parents = []string{"cause-no-reference"} }, "VAL-CAU-004", "/causes/0/parents"},
		// A loop through the second of two parents is still a loop.
		{"causes loop through a second parent", func(d *impact.Model) {
			d.Causes = append(d.Causes, impact.Cause{ID: "cause-extra", Statement: "Another cause.", Level: impact.CauseImmediate})
			d.Causes[0].Parents = []string{"cause-extra", "cause-no-reference"}
		}, "VAL-CAU-004", "/causes/0/parents"},
		{"effect without statement", func(d *impact.Model) { d.Effects[0].Statement = " " }, "VAL-EFF-001", "/effects/0/statement"},
		{"effect parent unknown", func(d *impact.Model) { d.Effects[1].Parents = []string{"eff-ghost"} }, "VAL-REF-001", "/effects/1/parents/0"},
		{"effect parent is a cause", func(d *impact.Model) { d.Effects[1].Parents = []string{"cause-no-reference"} }, "VAL-REF-002", "/effects/1/parents/0"},
		{"effect is its own parent", func(d *impact.Model) { d.Effects[0].Parents = []string{d.Effects[0].ID} }, "VAL-EFF-002", "/effects/0/parents"},
		{"effects form a loop", func(d *impact.Model) { d.Effects[0].Parents = []string{"eff-slow-adoption"} }, "VAL-EFF-003", "/effects/0/parents"},
		{"effect set aside without reason", func(d *impact.Model) { d.Effects[1].OutOfScope.Reason = "" }, "VAL-EFF-004", "/effects/1/outOfScope/reason"},
		{"outcome relieves an unknown effect", func(d *impact.Model) { d.Outcomes[0].AddressesEffects = []string{"eff-ghost"} }, "VAL-REF-001", "/outcomes/0/addressesEffects/0"},
		{"cause affected group unknown", func(d *impact.Model) { d.Causes[1].AffectedGroupIDs = []string{"tg-ghost"} }, "VAL-REF-001", "/causes/1/affectedGroupIds/0"},
		{"cause evidence unknown", func(d *impact.Model) { d.Causes[0].EvidenceIDs = []string{"ev-ghost"} }, "VAL-REF-001", "/causes/0/evidenceIds/0"},
		{"cause set aside without reason", func(d *impact.Model) { d.Causes[0].OutOfScope.Reason = "" }, "VAL-CAU-005", "/causes/0/outOfScope/reason"},
		{"target group without name", func(d *impact.Model) { d.TargetGroups[0].Name = "" }, "VAL-TGT-001", "/targetGroups/0/name"},
		{"target group actor unknown", func(d *impact.Model) { d.TargetGroups[0].ActorIDs = []string{"act-ghost"} }, "VAL-REF-001", "/targetGroups/0/actorIds/0"},
		{"responsibility invalid", func(d *impact.Model) { d.Outcomes[2].Responsibility = "mine" }, "VAL-ENUM-001", "/outcomes/2/responsibility"},
		{"delivering actor unknown", func(d *impact.Model) { d.Outcomes[2].DeliveredBy = []string{"act-ghost"} }, "VAL-REF-001", "/outcomes/2/deliveredBy/0"},
		{"addressed cause unknown", func(d *impact.Model) { d.Outcomes[2].AddressesCauses = []string{"cause-ghost"} }, "VAL-REF-001", "/outcomes/2/addressesCauses/0"},
		{"addressed cause is an actor", func(d *impact.Model) { d.Outcomes[2].AddressesCauses = []string{"act-idp"} }, "VAL-REF-002", "/outcomes/2/addressesCauses/0"},
		{"outcome target group unknown", func(d *impact.Model) { d.Outcomes[2].TargetGroupIDs = []string{"tg-ghost"} }, "VAL-REF-001", "/outcomes/2/targetGroupIds/0"},
		{"indicator target group unknown", func(d *impact.Model) { d.Indicators[0].TargetGroupIDs = []string{"tg-ghost"} }, "VAL-REF-001", "/indicators/0/targetGroupIds/0"},
		{"outcome score out of range", func(d *impact.Model) { d.Outcomes[2].Priority.Scores[0].Score = 9 }, "VAL-SEL-002", "/outcomes/2/priority/scores/0/score"},
		{"outcome scored with an intervention criterion", func(d *impact.Model) { d.Outcomes[2].Priority.Scores[0].CriterionID = "crit-feasibility" }, "VAL-SEL-004", "/outcomes/2/priority/scores/0/criterionId"},
		{"outcome criterion unknown", func(d *impact.Model) { d.Outcomes[2].Priority.Scores[0].CriterionID = "crit-ghost" }, "VAL-REF-001", "/outcomes/2/priority/scores/0/criterionId"},
		{"intervention scored with an outcome criterion", func(d *impact.Model) { d.Interventions[0].Priority.Scores[0].CriterionID = "crit-equity" }, "VAL-SEL-004", "/interventions/0/priority/scores/0/criterionId"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := fullDoc(t)
			if doc.Outcomes[2].ID != "oc-idp-support" || doc.Interventions[0].ID != "int-interop" {
				t.Fatal("test setup: fixture order changed")
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

func TestResponsibilityAccountability(t *testing.T) {
	for r, want := range map[impact.Responsibility]bool{
		"": true, impact.ResponsibilityOwn: true, impact.ResponsibilityShared: true,
		impact.ResponsibilityPartner: false, impact.ResponsibilityUnaddressed: false,
	} {
		if got := (impact.Outcome{Responsibility: r}).Accountable(); got != want {
			t.Errorf("Accountable(%q) = %v, want %v", r, got, want)
		}
	}
}
