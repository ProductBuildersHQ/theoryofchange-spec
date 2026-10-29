package review

import (
	"strings"
	"testing"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

const sheet = "# toc:riverbend-network version:0.2.0 hash:sha256:current\r\n" +
	"kind,id,statement,basis,sources,decision,suggested_change,comment\r\n" +
	"outcome,oc-st-leader-skills,Leaders gain skills,stated,,Agree,,\r\n" +
	"outcome,oc-resident-access,Residents get services,inferred,,change,\"title: Residents reach services\ncategory: access\",\r\n" +
	"outcome,oc-resident-voice,Residents are heard,inferred,,,,\r\n" +
	"assumption,asm-positions-hold,Positions hold,,,,,Doubt this one\r\n"

func TestFromSheet(t *testing.T) {
	resp, err := FromSheet([]byte(sheet), SheetMeta{Reviewers: []string{"program staff"}, Date: "2026-03-01"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.TocID != "riverbend-network" || resp.Version != "0.2.0" || resp.ContentHash != "sha256:current" || resp.Method != toc.ReviewSheet {
		t.Fatalf("response = %+v", resp)
	}
	if len(resp.Rows) != 3 {
		t.Fatalf("rows = %+v", resp.Rows)
	}
	if resp.Rows[0].Decision != toc.DecisionAgree || resp.Rows[0].Origin != toc.OriginSheet {
		t.Errorf("row 0 = %+v", resp.Rows[0])
	}
	if !strings.Contains(resp.Rows[1].SuggestedChange, "\ncategory: access") {
		t.Errorf("row 1 = %+v", resp.Rows[1])
	}
	if resp.Rows[2].Comment != "Doubt this one" || resp.Rows[2].Decision != "" {
		t.Errorf("row 2 = %+v", resp.Rows[2])
	}

	res, err := Import(load(t), resp, opts())
	if err != nil {
		t.Fatal(err)
	}
	if res.Stale || res.Mapped != 2 || len(res.Flags) != 1 || res.Flags[0].Row != 3 {
		t.Fatalf("result = %+v", res)
	}
}

func TestFromSheetWithoutStampIsStale(t *testing.T) {
	body := sheet[strings.Index(sheet, "kind,"):]
	resp, err := FromSheet([]byte(body), SheetMeta{Reviewers: []string{"x"}, Date: "2026-03-01", TocID: "riverbend-network"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.ContentHash != "" {
		t.Fatalf("hash = %q", resp.ContentHash)
	}
	if res, _ := Import(load(t), resp, opts()); !res.Stale {
		t.Error("a sheet with no hash was not stale")
	}
	over, _ := FromSheet([]byte(sheet), SheetMeta{ContentHash: "sha256:other"})
	if over.ContentHash != "sha256:other" {
		t.Errorf("override ignored: %q", over.ContentHash)
	}
}

func TestFromSheetErrors(t *testing.T) {
	for name, in := range map[string]string{
		"empty":          "",
		"missing column": "kind,id,statement\r\n",
	} {
		if _, err := FromSheet([]byte(in), SheetMeta{}); err == nil {
			t.Errorf("%s: no error", name)
		}
	}
}
