package validate

import (
	"testing"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

func TestChangeAnalysisIssues(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(d *toc.Document)
		code   string
		path   string
	}{
		{"cause without statement", func(d *toc.Document) { d.Causes[0].Statement = " " }, "VAL-CAU-001", "/causes/0/statement"},
		{"cause without level", func(d *toc.Document) { d.Causes[0].Level = "" }, "VAL-CAU-002", "/causes/0/level"},
		{"cause level invalid", func(d *toc.Document) { d.Causes[0].Level = "deep" }, "VAL-ENUM-001", "/causes/0/level"},
		{"cause parent unknown", func(d *toc.Document) { d.Causes[1].Parent = "cause-ghost" }, "VAL-REF-001", "/causes/1/parent"},
		{"cause parent is an outcome", func(d *toc.Document) { d.Causes[1].Parent = "oc-idp-support" }, "VAL-REF-002", "/causes/1/parent"},
		{"cause is its own parent", func(d *toc.Document) { d.Causes[0].Parent = d.Causes[0].ID }, "VAL-CAU-003", "/causes/0/parent"},
		{"causes form a loop", func(d *toc.Document) { d.Causes[0].Parent = "cause-no-reference" }, "VAL-CAU-004", "/causes/0/parent"},
		{"cause affected group unknown", func(d *toc.Document) { d.Causes[1].AffectedGroupIDs = []string{"tg-ghost"} }, "VAL-REF-001", "/causes/1/affectedGroupIds/0"},
		{"cause evidence unknown", func(d *toc.Document) { d.Causes[0].EvidenceIDs = []string{"ev-ghost"} }, "VAL-REF-001", "/causes/0/evidenceIds/0"},
		{"cause set aside without reason", func(d *toc.Document) { d.Causes[0].OutOfScope.Reason = "" }, "VAL-CAU-005", "/causes/0/outOfScope/reason"},
		{"target group without name", func(d *toc.Document) { d.TargetGroups[0].Name = "" }, "VAL-TGT-001", "/targetGroups/0/name"},
		{"target group actor unknown", func(d *toc.Document) { d.TargetGroups[0].ActorIDs = []string{"act-ghost"} }, "VAL-REF-001", "/targetGroups/0/actorIds/0"},
		{"responsibility invalid", func(d *toc.Document) { d.Outcomes[2].Responsibility = "mine" }, "VAL-ENUM-001", "/outcomes/2/responsibility"},
		{"delivering actor unknown", func(d *toc.Document) { d.Outcomes[2].DeliveredBy = []string{"act-ghost"} }, "VAL-REF-001", "/outcomes/2/deliveredBy/0"},
		{"addressed cause unknown", func(d *toc.Document) { d.Outcomes[2].AddressesCauses = []string{"cause-ghost"} }, "VAL-REF-001", "/outcomes/2/addressesCauses/0"},
		{"addressed cause is an actor", func(d *toc.Document) { d.Outcomes[2].AddressesCauses = []string{"act-idp"} }, "VAL-REF-002", "/outcomes/2/addressesCauses/0"},
		{"outcome target group unknown", func(d *toc.Document) { d.Outcomes[2].TargetGroupIDs = []string{"tg-ghost"} }, "VAL-REF-001", "/outcomes/2/targetGroupIds/0"},
		{"indicator target group unknown", func(d *toc.Document) { d.Indicators[0].TargetGroupIDs = []string{"tg-ghost"} }, "VAL-REF-001", "/indicators/0/targetGroupIds/0"},
		{"outcome score out of range", func(d *toc.Document) { d.Outcomes[2].Priority.Scores[0].Score = 9 }, "VAL-SEL-002", "/outcomes/2/priority/scores/0/score"},
		{"outcome scored with an intervention criterion", func(d *toc.Document) { d.Outcomes[2].Priority.Scores[0].CriterionID = "crit-feasibility" }, "VAL-SEL-004", "/outcomes/2/priority/scores/0/criterionId"},
		{"outcome criterion unknown", func(d *toc.Document) { d.Outcomes[2].Priority.Scores[0].CriterionID = "crit-ghost" }, "VAL-REF-001", "/outcomes/2/priority/scores/0/criterionId"},
		{"intervention scored with an outcome criterion", func(d *toc.Document) { d.Interventions[0].Priority.Scores[0].CriterionID = "crit-equity" }, "VAL-SEL-004", "/interventions/0/priority/scores/0/criterionId"},
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
	for r, want := range map[toc.Responsibility]bool{
		"": true, toc.ResponsibilityOwn: true, toc.ResponsibilityShared: true,
		toc.ResponsibilityPartner: false, toc.ResponsibilityUnaddressed: false,
	} {
		if got := (toc.Outcome{Responsibility: r}).Accountable(); got != want {
			t.Errorf("Accountable(%q) = %v, want %v", r, got, want)
		}
	}
}
