package toc

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func roundTrip[T any](t *testing.T, name string) T {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var v T
	if err := dec.Decode(&v); err != nil {
		t.Fatalf("decode %s: %v", name, err)
	}
	out, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal %s: %v", name, err)
	}
	if !reflect.DeepEqual(normalize(t, data), normalize(t, out)) {
		t.Fatalf("%s round trip changed the document:\n in: %s\nout: %s", name, data, out)
	}
	return v
}

func TestObservationSetRoundTrip(t *testing.T) {
	set := roundTrip[ObservationSet](t, "observations.json")
	if set.Kind != KindObservationSet || len(set.Observations) != 2 || len(set.Assessments) != 1 || len(set.Notes) != 1 {
		t.Fatalf("unexpected observation set %+v", set)
	}
	if set.Observations[0].Value == nil || *set.Observations[0].Value != 2 || set.Observations[1].Level != "committed" {
		t.Fatalf("observation values not decoded: %+v", set.Observations)
	}
}

func TestStatusRoundTrip(t *testing.T) {
	st := roundTrip[Status](t, "status.json")
	if st.Kind != KindStatus || st.Options != DefaultStatusOptions() {
		t.Fatalf("unexpected status header %+v", st)
	}
	if st.Outcomes[0].PathwayState != StateOnTrack || st.Links[0].State != LinkStateOK || st.Assumptions[0].Source != SourceAssessment {
		t.Fatalf("unexpected status body %+v", st)
	}
}

func TestActorRegistryRoundTrip(t *testing.T) {
	reg := roundTrip[ActorRegistry](t, "registry.json")
	if reg.Kind != KindActorRegistry || len(reg.Actors) != 2 {
		t.Fatalf("unexpected registry %+v", reg)
	}
}

func TestStatusDiffRoundTrip(t *testing.T) {
	diff := roundTrip[StatusDiff](t, "diff.json")
	if len(diff.Changes) != 1 || diff.Changes[0].ElementKind != ElementIndicator {
		t.Fatalf("unexpected diff %+v", diff)
	}
}

func TestNewObservationSet(t *testing.T) {
	s := NewObservationSet("doc", "2026-Q4")
	if s.Kind != KindObservationSet || s.SpecVersion != SpecVersion || s.TocID != "doc" || s.Period != "2026-Q4" {
		t.Fatalf("unexpected observation set %+v", s)
	}
}

func TestStatusEnums(t *testing.T) {
	enums := map[string][]string{
		"DataQuality":      strs(AllDataQualities()),
		"State":            strs(AllStates()),
		"LinkState":        strs(AllLinkStates()),
		"RollupMethod":     strs(AllRollupMethods()),
		"AssumptionSource": strs(AllAssumptionSources()),
	}
	for name, values := range enums {
		seen := map[string]bool{}
		for _, v := range values {
			if !snakeCase.MatchString(v) || seen[v] {
				t.Errorf("%s value %q is not unique snake_case", name, v)
			}
			seen[v] = true
		}
	}
	if !StateBaselineOnly.Valid() || State("").Valid() {
		t.Error("State validity mismatch")
	}
	if !QualityVerified.Valid() || DataQuality("guess").Valid() {
		t.Error("DataQuality validity mismatch")
	}
	if !LinkStateBroken.Valid() || !RollupMean.Valid() || !SourceDocument.Valid() {
		t.Error("status enum validity mismatch")
	}
}
