package share

import (
	"encoding/csv"
	"strings"
	"testing"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/report"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

func fullView(t *testing.T, mutate func(*toc.Document)) *report.View {
	t.Helper()
	doc := fullDoc(t)
	if mutate != nil {
		mutate(doc)
	}
	v, err := report.BuildView(doc, nil, nil, report.Options{})
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestReviewPacketGolden(t *testing.T) {
	golden(t, "review-packet.md", ReviewPacket(fullView(t, nil)))
}

func TestReviewSheetGolden(t *testing.T) {
	got, err := ReviewSheet(fullView(t, nil))
	if err != nil {
		t.Fatal(err)
	}
	golden(t, "review-sheet.csv", got)
}

func TestEveryActiveElementHasARow(t *testing.T) {
	v := fullView(t, nil)
	want := len(v.Outcomes) + len(v.Links) + len(v.Groups) + len(v.Assumptions) +
		len(v.Interventions) + len(v.Indicators) + len(v.Metrics) + len(v.Actors)
	rows := ReviewRows(v)
	if len(rows) != want {
		t.Fatalf("got %d rows, want %d", len(rows), want)
	}
	seen := map[string]bool{}
	for _, r := range rows {
		if r.ID == "" || r.Statement == "" {
			t.Errorf("row %+v lacks an ID or statement", r)
		}
		if seen[r.ID] {
			t.Errorf("duplicate row for %s", r.ID)
		}
		seen[r.ID] = true
	}
}

func TestSheetAndPacketListTheSameIDs(t *testing.T) {
	v := fullView(t, nil)
	sheet, err := ReviewSheet(v)
	if err != nil {
		t.Fatal(err)
	}
	records, err := csv.NewReader(strings.NewReader(sheet)).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	packet := ReviewPacket(v)
	if got := strings.Join(records[0], ","); got != strings.Join(ReviewSheetHeader, ",") {
		t.Errorf("header = %s", got)
	}
	for _, rec := range records[1:] {
		if !strings.Contains(packet, "| "+rec[1]+" |") {
			t.Errorf("the packet has no row for %s", rec[1])
		}
		if rec[5] != "" || rec[6] != "" || rec[7] != "" {
			t.Errorf("response columns of %s should be empty: %v", rec[1], rec[5:])
		}
	}
}

func TestRejectedElementsAreNotPutToReviewers(t *testing.T) {
	var rejected string
	v := fullView(t, func(d *toc.Document) {
		rejected = d.Interventions[0].ID
		d.Interventions[0].Provenance = &toc.Provenance{Basis: toc.BasisRejected}
	})
	if rejected == "" {
		t.Fatal("no element to reject")
	}
	for _, r := range ReviewRows(v) {
		if r.ID == rejected {
			t.Errorf("rejected element %s was exported", r.ID)
		}
	}
}

func TestRestrictedEvidenceNeverLeaksIntoAPacket(t *testing.T) {
	v := fullView(t, func(d *toc.Document) {
		d.Evidence = append(d.Evidence, toc.Evidence{
			ID: "ev-private", Title: "Interview notes", URL: "https://example.org/secret-drive",
			Excerpt: "SECRET-EXCERPT", Summary: "SECRET-SUMMARY", Sensitivity: toc.SensitivityRestricted,
		})
		d.Outcomes[1].Provenance = &toc.Provenance{Basis: toc.BasisStated, EvidenceIDs: []string{"ev-private"}}
	})
	sheet, err := ReviewSheet(v)
	if err != nil {
		t.Fatal(err)
	}
	out := ReviewPacket(v) + sheet
	for _, secret := range []string{"SECRET-EXCERPT", "SECRET-SUMMARY", "secret-drive"} {
		if strings.Contains(out, secret) {
			t.Errorf("restricted evidence leaked %q", secret)
		}
	}
	if !strings.Contains(out, "ev-private: Interview notes") {
		t.Error("the source should still be named by ID and title")
	}
}

func TestPacketEscapesPipes(t *testing.T) {
	v := fullView(t, func(d *toc.Document) { d.Interventions[0].Title = "Run A | B events" })
	if !strings.Contains(ReviewPacket(v), `Run A \| B events`) {
		t.Error("a pipe in a statement would break the table")
	}
}
