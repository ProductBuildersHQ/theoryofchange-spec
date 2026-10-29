// Package examples tests the reference examples shipped in this directory.
// Every example is fictional and must stay valid, canonical, and rich enough
// to demonstrate the features it is there to show.
package examples

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/graph"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/tocio"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/validate"
)

func exampleFiles(t *testing.T) []string {
	t.Helper()
	var files []string
	for _, pattern := range []string{"*/*.yaml", "*/*.yml", "*/*.json"} {
		m, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, m...)
	}
	slices.Sort(files)
	if len(files) == 0 {
		t.Fatal("no example files found")
	}
	return files
}

func read(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func loadTheory(t *testing.T, dir string) *toc.Document {
	t.Helper()
	doc, err := tocio.LoadDocument(filepath.Join(dir, "theory.yaml"))
	if err != nil {
		t.Fatalf("%s/theory.yaml: %v", dir, err)
	}
	return doc
}

func TestEveryExampleFileIsCanonical(t *testing.T) {
	for _, path := range exampleFiles(t) {
		t.Run(path, func(t *testing.T) {
			data := read(t, path)
			f, err := tocio.FormatFromPath(path)
			if err != nil {
				t.Fatal(err)
			}
			out, err := tocio.Reformat(data, f)
			if err != nil {
				t.Fatalf("Reformat: %v", err)
			}
			if string(out) != string(data) {
				t.Fatalf("%s is not canonical; run: theoryofchange fmt %s", path, path)
			}
		})
	}
}

var urlPattern = regexp.MustCompile(`https?://[^\s"'<>)]+`)

func TestExamplesAreFictional(t *testing.T) {
	for _, path := range exampleFiles(t) {
		for _, u := range urlPattern.FindAllString(string(read(t, path)), -1) {
			host := strings.TrimPrefix(strings.TrimPrefix(u, "https://"), "http://")
			host = strings.SplitN(host, "/", 2)[0]
			if host != "example.org" && !strings.HasSuffix(host, ".example.org") {
				t.Errorf("%s: URL %q is not under example.org", path, u)
			}
		}
	}
	for _, dir := range exampleDirs(t) {
		if !strings.Contains(strings.ToLower(string(read(t, filepath.Join(dir, "README.md")))), "fictional") {
			t.Errorf("%s/README.md must say the example is fictional", dir)
		}
	}
}

func exampleDirs(t *testing.T) []string {
	t.Helper()
	dirs, err := filepath.Glob("*/theory.yaml")
	if err != nil {
		t.Fatal(err)
	}
	for i := range dirs {
		dirs[i] = filepath.Dir(dirs[i])
	}
	if len(dirs) < 2 {
		t.Fatalf("expected at least two examples, found %v", dirs)
	}
	return dirs
}

func TestEveryTheoryIsValidWithAGoodGraph(t *testing.T) {
	for _, dir := range exampleDirs(t) {
		t.Run(dir, func(t *testing.T) {
			doc := loadTheory(t, dir)
			if issues := validate.Document(doc); len(issues) != 0 {
				for _, i := range issues {
					t.Errorf("%s", i)
				}
			}
			g, err := graph.Build(doc)
			if err != nil {
				t.Fatalf("graph: %v", err)
			}
			if c := g.Cycles(); len(c) != 0 {
				t.Errorf("cycles: %v", c)
			}
			if o := g.Orphans(); len(o) != 0 {
				t.Errorf("orphans: %v", o)
			}
			if d := g.DeadEnds(); len(d) != 0 {
				t.Errorf("dead ends: %v", d)
			}
			if len(g.FeedbackEdges()) == 0 {
				t.Error("an example should demonstrate a feedback loop")
			}
		})
	}
}

func TestObservationsMatchTheirTheory(t *testing.T) {
	for _, dir := range exampleDirs(t) {
		doc := loadTheory(t, dir)
		paths, err := filepath.Glob(filepath.Join(dir, "observations-*.yaml"))
		if err != nil {
			t.Fatal(err)
		}
		if len(paths) == 0 {
			t.Errorf("%s has no observation set", dir)
		}
		for _, p := range paths {
			sets, err := tocio.LoadObservations(p)
			if err != nil {
				t.Fatalf("%s: %v", p, err)
			}
			if issues := validate.Observations(doc, sets...); len(issues) != 0 {
				for _, i := range issues {
					t.Errorf("%s: %s", p, i)
				}
			}
			if len(sets[0].Observations) < 12 {
				t.Errorf("%s: only %d observations", p, len(sets[0].Observations))
			}
		}
	}
}

// basesUsed returns every provenance basis that appears in the document.
func basesUsed(doc *toc.Document) map[toc.Basis]int {
	count := map[toc.Basis]int{}
	add := func(p *toc.Provenance) {
		if p != nil {
			count[p.Basis]++
		}
	}
	for _, v := range doc.Outcomes {
		add(v.Provenance)
	}
	for _, v := range doc.Groups {
		add(v.Provenance)
	}
	for _, v := range doc.Links {
		add(v.Provenance)
	}
	for _, v := range doc.Interventions {
		add(v.Provenance)
	}
	for _, v := range doc.Assumptions {
		add(v.Provenance)
	}
	for _, v := range doc.Indicators {
		add(v.Provenance)
	}
	for _, v := range doc.Actors {
		add(v.Provenance)
	}
	for _, v := range doc.ExternalFactors {
		add(v.Provenance)
	}
	for _, v := range doc.Risks {
		add(v.Provenance)
	}
	for _, v := range doc.Metrics {
		add(v.Provenance)
	}
	for _, v := range doc.Pathways {
		add(v.Provenance)
	}
	return count
}

func TestEveryBasisAndEveryStatedElementIsCited(t *testing.T) {
	for _, dir := range exampleDirs(t) {
		t.Run(dir, func(t *testing.T) {
			doc := loadTheory(t, dir)
			used := basesUsed(doc)
			for _, b := range toc.AllBases() {
				if used[b] == 0 {
					t.Errorf("no element uses basis %q", b)
				}
			}
			if used[toc.BasisRejected] < 2 {
				t.Errorf("want at least two rejected elements, have %d", used[toc.BasisRejected])
			}
			check := func(id string, p *toc.Provenance) {
				if p == nil {
					return
				}
				if p.Basis == toc.BasisStated && len(p.EvidenceIDs) == 0 {
					t.Errorf("%s is stated but cites no evidence", id)
				}
				if p.Basis == toc.BasisConfirmed && (p.DecidedAt == "" || p.DecidedBy == "") {
					t.Errorf("%s is confirmed but not dated and attributed", id)
				}
			}
			for _, v := range doc.Outcomes {
				check(v.ID, v.Provenance)
			}
			for _, v := range doc.Links {
				check(v.ID, v.Provenance)
			}
			for _, v := range doc.Interventions {
				check(v.ID, v.Provenance)
			}
			for _, v := range doc.Assumptions {
				check(v.ID, v.Provenance)
			}
			for _, v := range doc.Indicators {
				check(v.ID, v.Provenance)
			}
			for _, v := range doc.Actors {
				check(v.ID, v.Provenance)
			}
			for _, v := range doc.Metrics {
				check(v.ID, v.Provenance)
			}
			for _, v := range doc.Pathways {
				check(v.ID, v.Provenance)
			}
		})
	}
}

func TestIndicatorsCoverEveryDataAvailability(t *testing.T) {
	for _, dir := range exampleDirs(t) {
		doc := loadTheory(t, dir)
		seen := map[toc.DataAvailability]bool{}
		for _, ind := range doc.Indicators {
			seen[ind.DataAvailability] = true
		}
		for _, d := range toc.AllDataAvailabilities() {
			if !seen[d] {
				t.Errorf("%s: no indicator has data availability %q", dir, d)
			}
		}
	}
}

func TestIntermediaryExampleDemonstratesItsFeatures(t *testing.T) {
	doc := loadTheory(t, "intermediary")
	if doc.ID != "riverbend-network" {
		t.Fatalf("id = %q", doc.ID)
	}
	live := 0
	for _, p := range doc.Pathways {
		if !p.Rejected() {
			live++
		}
	}
	if len(doc.Outcomes) < 18 || live != 3 {
		t.Errorf("outcomes=%d live pathways=%d (a rejected pathway is allowed in addition)", len(doc.Outcomes), live)
	}
	if doc.Ceiling == nil || doc.Ceiling.Level != toc.LevelLongTerm {
		t.Error("the resident-level impact should sit above a long_term ceiling")
	}
	var any, rollup, none, weighted bool
	for _, g := range doc.Groups {
		if g.Logic.Type == toc.LogicAny || g.Logic.Type == toc.LogicAtLeast {
			any = true
		}
	}
	claims := map[toc.Claim]bool{}
	for _, ind := range doc.Indicators {
		claims[ind.EffectiveClaim()] = true
		if len(ind.RollupFrom) > 0 {
			rollup = true
		}
	}
	for _, m := range doc.Metrics {
		none = none || m.Aggregation == toc.AggregationNone
		weighted = weighted || (m.Aggregation == toc.AggregationWeightedMean && m.WeightMetricID != "")
	}
	if !any || !rollup || !none || !weighted {
		t.Errorf("any/at_least group=%v rollup=%v none-aggregation=%v weighted mean=%v", any, rollup, none, weighted)
	}
	if !claims[toc.ClaimAttribution] || !claims[toc.ClaimContribution] {
		t.Errorf("both claims should appear: %v", claims)
	}
	if len(doc.Glossary) < 6 || len(doc.Consultations) < 3 {
		t.Errorf("glossary=%d consultations=%d", len(doc.Glossary), len(doc.Consultations))
	}
	qualitative := 0
	for _, ind := range doc.Indicators {
		if len(ind.Levels) > 0 {
			qualitative++
		}
	}
	if qualitative < 2 {
		t.Errorf("want at least two qualitative indicators, have %d", qualitative)
	}
}

func TestMarketAdoptionExampleDemonstratesItsFeatures(t *testing.T) {
	doc := loadTheory(t, "market-adoption")
	if doc.ID != "delegation-standard-adoption" {
		t.Fatalf("id = %q", doc.ID)
	}
	categories := map[string]bool{}
	atLeast := false
	for _, o := range doc.Outcomes {
		categories[o.Category] = true
		if o.Logic != nil && o.Logic.Type == toc.LogicAtLeast {
			atLeast = true
		}
	}
	for _, want := range []string{"adopt", "adapt", "expand", "respond"} {
		if !categories[want] {
			t.Errorf("no outcome has AAER category %q", want)
		}
	}
	if !atLeast {
		t.Error("want an outcome with at_least logic")
	}
	if doc.Ceiling == nil {
		t.Error("the market-wide impact should sit above a ceiling")
	}

	groupLogic := map[toc.LogicType]bool{}
	nested := false
	groups := map[string]bool{}
	for _, g := range doc.Groups {
		groups[g.ID] = true
	}
	for _, g := range doc.Groups {
		groupLogic[g.Logic.Type] = true
		for _, m := range g.Members {
			nested = nested || groups[m]
		}
	}
	if !groupLogic[toc.LogicAll] || !groupLogic[toc.LogicAny] || !nested {
		t.Errorf("group logic=%v nested=%v", groupLogic, nested)
	}

	loops, negative := map[toc.LoopType]bool{}, false
	for _, l := range doc.Links {
		if l.Feedback {
			loops[l.LoopType] = true
		}
		negative = negative || l.Polarity == toc.PolarityNegative
	}
	if !loops[toc.LoopReinforcing] || !loops[toc.LoopBalancing] || !negative {
		t.Errorf("loops=%v negative polarity=%v", loops, negative)
	}

	statuses := map[toc.AssumptionStatus]bool{}
	for _, a := range doc.Assumptions {
		statuses[a.Status] = true
	}
	if !statuses[toc.AssumptionAtRisk] || !statuses[toc.AssumptionInvalidated] {
		t.Errorf("assumption statuses = %v", statuses)
	}
	contradicting := false
	for _, e := range doc.Evidence {
		contradicting = contradicting || e.Direction == toc.EvidenceContradicts
	}
	if !contradicting {
		t.Error("want contradicting evidence")
	}
	if doc.ActorRegistry != "delegation-market-actors" {
		t.Errorf("actorRegistry = %q", doc.ActorRegistry)
	}
	reg, err := tocio.ParseActorRegistry(read(t, "market-adoption/actors.yaml"), tocio.FormatYAML)
	if err != nil || reg.ID != doc.ActorRegistry || len(reg.Actors) < 4 {
		t.Errorf("actor registry: %v %+v", err, reg)
	}
	if issues := validate.Document(doc, validate.WithActorRegistry(reg)); len(issues) != 0 {
		t.Errorf("with the registry supplied: %v", issues)
	}
}
