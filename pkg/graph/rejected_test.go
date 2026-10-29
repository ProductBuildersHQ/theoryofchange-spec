package graph

import (
	"encoding/json"
	"os"
	"slices"
	"testing"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

func loadFixture(t *testing.T) *toc.Document {
	t.Helper()
	data, err := os.ReadFile("../toc/testdata/full.json")
	if err != nil {
		t.Fatal(err)
	}
	var d toc.Document
	if err := json.Unmarshal(data, &d); err != nil {
		t.Fatal(err)
	}
	return &d
}

func rejected() *toc.Provenance { return &toc.Provenance{Basis: toc.BasisRejected} }

func nodeIDs(g *Graph) []string {
	var ids []string
	for _, n := range g.Nodes() {
		ids = append(ids, n.ID)
	}
	return ids
}

func TestRejectedOutcomeLeavesTheGraphWithItsEdges(t *testing.T) {
	doc := loadFixture(t)
	before, err := Build(doc)
	if err != nil {
		t.Fatal(err)
	}

	doc.Outcomes[3].Provenance = rejected() // oc-platform-support
	if doc.Outcomes[3].ID != "oc-platform-support" {
		t.Fatalf("test setup: outcome 3 is %q", doc.Outcomes[3].ID)
	}
	g, err := Build(doc)
	if err != nil {
		t.Fatalf("a document with a rejected outcome must still build: %v", err)
	}
	if slices.Contains(nodeIDs(g), "oc-platform-support") || len(g.Nodes()) != len(before.Nodes())-1 {
		t.Fatalf("rejected outcome still a node: %v", nodeIDs(g))
	}
	for _, e := range append(g.Edges(), g.FeedbackEdges()...) {
		if e.From == "oc-platform-support" || e.To == "oc-platform-support" {
			t.Fatalf("an edge touches the rejected outcome: %+v", e)
		}
	}
	// The output beneath it lost its route to impact, and so did its intervention.
	if dead := g.DeadEnds(); !slices.Equal(dead, []string{"oc-reference-impl", "int-reference"}) {
		t.Fatalf("DeadEnds = %v", dead)
	}
	// Its group keeps only the surviving member.
	groups := g.GroupsFor("oc-enterprise-require")
	if len(groups) != 1 || !slices.Equal(groups[0].Members, []string{"oc-idp-support"}) {
		t.Fatalf("group members = %+v", groups)
	}
	// The document itself is untouched.
	if len(doc.Groups[0].Members) != 2 || len(doc.Links) != 6 {
		t.Fatal("Build changed the document")
	}
}

func TestRejectedLinkIsNotAnEdge(t *testing.T) {
	doc := loadFixture(t)
	for i := range doc.Links {
		if doc.Links[i].ID == "ln-spec-idp" {
			doc.Links[i].Provenance = rejected()
		}
	}
	g, err := Build(doc)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range g.Edges() {
		if e.LinkID == "ln-spec-idp" {
			t.Fatal("rejected link is still an edge")
		}
	}
	if !slices.Contains(g.DeadEnds(), "oc-spec-stable") {
		t.Fatalf("oc-spec-stable should have lost its route to impact: %v", g.DeadEnds())
	}
	// The intervention still reaches the same outcome by its implicit edge.
	if !slices.Contains(g.Upstream("oc-idp-support"), "int-interop") {
		t.Fatalf("the intervention's implicit edge should remain: %v", g.Upstream("oc-idp-support"))
	}
}

func TestRejectedInterventionAndGroupAreExcluded(t *testing.T) {
	doc := loadFixture(t)
	doc.Interventions[0].Provenance = rejected() // int-interop
	doc.Groups[0].Provenance = rejected()
	g, err := Build(doc)
	if err != nil {
		t.Fatal(err)
	}
	if slices.Contains(nodeIDs(g), "int-interop") {
		t.Fatal("rejected intervention is still a node")
	}
	if len(g.GroupsFor("oc-enterprise-require")) != 0 {
		t.Fatal("rejected group is still reported")
	}
	if !slices.Contains(g.Orphans(), "oc-spec-stable") {
		t.Fatalf("oc-spec-stable lost its only intervention and should be an orphan: %v", g.Orphans())
	}
}

func TestRejectedOutcomeLeavesTheCeiling(t *testing.T) {
	doc := loadFixture(t)
	doc.Ceiling = &toc.Ceiling{OutcomeIDs: []string{"oc-enterprise-require", "oc-platform-support"}}
	doc.Outcomes[3].Provenance = rejected()
	g, err := Build(doc)
	if err != nil {
		t.Fatal(err)
	}
	if g.BelowCeiling("oc-platform-support") {
		t.Fatal("a rejected outcome cannot be owned")
	}
	if !g.BelowCeiling("oc-enterprise-require") {
		t.Fatal("the surviving ceiling outcome should still be owned")
	}
}
