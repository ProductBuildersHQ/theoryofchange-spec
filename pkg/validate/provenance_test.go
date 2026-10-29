package validate

import (
	"testing"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

func TestProvenanceAndConsultationIssues(t *testing.T) {
	tests := []struct {
		name     string
		mutate   func(d *toc.Document)
		code     string
		path     string
		severity Severity
	}{
		{"provenance without basis", func(d *toc.Document) { d.Outcomes[0].Provenance = &toc.Provenance{} }, "VAL-PRV-001", "/outcomes/0/provenance/basis", SeverityError},
		{"invalid basis", func(d *toc.Document) { d.Outcomes[0].Provenance = &toc.Provenance{Basis: "rumor"} }, "VAL-ENUM-001", "/outcomes/0/provenance/basis", SeverityError},
		{"provenance cites unknown evidence", func(d *toc.Document) {
			d.Links[0].Provenance = &toc.Provenance{Basis: toc.BasisStated, EvidenceIDs: []string{"ev-ghost"}}
		}, "VAL-REF-001", "/links/0/provenance/evidenceIds/0", SeverityError},
		{"provenance cites a non-evidence element", func(d *toc.Document) {
			d.Actors[0].Provenance = &toc.Provenance{Basis: toc.BasisStated, EvidenceIDs: []string{"oc-idp-support"}}
		}, "VAL-REF-002", "/actors/0/provenance/evidenceIds/0", SeverityError},
		{"decided-at not a date", func(d *toc.Document) {
			d.Metrics[0].Provenance = &toc.Provenance{Basis: toc.BasisConfirmed, DecidedAt: "last Tuesday"}
		}, "VAL-DATE-001", "/metrics/0/provenance/decidedAt", SeverityError},
		{"pathway provenance checked", func(d *toc.Document) { d.Pathways[0].Provenance = &toc.Provenance{} }, "VAL-PRV-001", "/pathways/0/provenance/basis", SeverityError},
		{"invalid data availability", func(d *toc.Document) { d.Indicators[0].DataAvailability = "somewhere" }, "VAL-ENUM-001", "/indicators/0/dataAvailability", SeverityError},
		{"consultation without date", func(d *toc.Document) { d.Consultations[0].Date = "" }, "VAL-DATE-003", "/consultations/0/date", SeverityError},
		{"consultation bad date", func(d *toc.Document) { d.Consultations[0].Date = "February" }, "VAL-DATE-001", "/consultations/0/date", SeverityError},
		{"consultation without step", func(d *toc.Document) { d.Consultations[0].Step = "" }, "VAL-CON-001", "/consultations/0/step", SeverityError},
		{"consultation invalid step", func(d *toc.Document) { d.Consultations[0].Step = "everything" }, "VAL-ENUM-001", "/consultations/0/step", SeverityError},
		{"consultation invalid method", func(d *toc.Document) { d.Consultations[0].Method = "seance" }, "VAL-ENUM-001", "/consultations/0/method", SeverityError},
		{"consultation negative count", func(d *toc.Document) { d.Consultations[0].Count = -1 }, "VAL-CON-002", "/consultations/0/count", SeverityError},
		{"consultation changed element removed", func(d *toc.Document) { d.Consultations[0].ChangedIDs = []string{"oc-gone"} }, "VAL-REF-003", "/consultations/0/changedIds/0", SeverityWarning},
		{"consultation id collides", func(d *toc.Document) { d.Consultations[0].ID = "oc-idp-support" }, "VAL-ID-002", "/consultations/0/id", SeverityError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := fullDoc(t)
			tt.mutate(doc)
			for _, i := range Document(doc) {
				if i.Code == tt.code && i.Path == tt.path {
					if i.Severity != tt.severity {
						t.Fatalf("%s severity = %s, want %s", tt.code, i.Severity, tt.severity)
					}
					return
				}
			}
			t.Fatalf("expected %s at %s; got:%s", tt.code, tt.path, describe(Document(doc)))
		})
	}
}

func TestEveryProvenanceBasisIsAccepted(t *testing.T) {
	for _, b := range toc.AllBases() {
		doc := fullDoc(t)
		doc.Outcomes[0].Provenance = &toc.Provenance{Basis: b}
		if issues := Document(doc); len(issues) != 0 {
			t.Errorf("basis %q:%s", b, describe(issues))
		}
	}
}

func TestRejectedElementsStillResolveAsReferences(t *testing.T) {
	// Rejected elements stay in the document, so references to them remain
	// valid; analysis, not validation, excludes them.
	doc := fullDoc(t)
	doc.Outcomes[2].Provenance = &toc.Provenance{Basis: toc.BasisRejected}
	if issues := Document(doc); len(issues) != 0 {
		t.Fatalf("rejecting an outcome should not break references:%s", describe(issues))
	}
}
