package validate

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
)

// Observations checks observation sets against the impact model they measure.
// Every indicator and assumption reference must resolve (nothing is dropped
// silently), values must fit the indicator's measurement type, and dates
// must be valid. Issues carry Source "observations[i]" when more than one
// set is given.
func Observations(doc *impact.Model, sets ...*impact.ObservationSet) []Issue {
	c := newChecker(doc, nil)
	for _, e := range doc.Elements() {
		if _, dup := c.kinds[e.ID]; !dup && e.ID != "" {
			c.kinds[e.ID] = e.Kind
		}
	}
	for i, set := range sets {
		c.source = ""
		if len(sets) > 1 {
			c.source = fmt.Sprintf("observations[%d]", i)
		}
		c.observationSet(set)
	}
	return c.issues
}

func (c *checker) observationSet(set *impact.ObservationSet) {
	if set.Kind != impact.KindObservationSet {
		c.errorf("VAL-OBS-001", "/kind", "", "kind is %q; expected %q", string(set.Kind), string(impact.KindObservationSet))
	}
	if set.ModelID != c.doc.ID {
		c.errorf("VAL-OBS-002", "/modelId", "", "observation set measures %q, not %q", set.ModelID, c.doc.ID)
	}
	for i, o := range set.Observations {
		p := "/observations/" + strconv.Itoa(i)
		c.observation(p, o)
	}
	for i, a := range set.Assessments {
		p := "/assumptionAssessments/" + strconv.Itoa(i)
		c.ref(p+"/assumptionId", a.AssumptionID, a.AssumptionID, impact.ElementAssumption)
		if a.AssumptionID == "" {
			c.errorf("VAL-OBS-003", p+"/assumptionId", "", "assessment has no assumptionId")
		}
		c.date(p+"/date", a.AssumptionID, a.Date)
		if a.Date == "" {
			c.errorf("VAL-DATE-003", p+"/date", a.AssumptionID, "assessment has no date")
		}
		if a.Status == "" {
			c.errorf("VAL-OBS-004", p+"/status", a.AssumptionID, "assessment has no status")
		}
		enum(c, p+"/status", a.AssumptionID, a.Status)
		c.refs(p+"/evidenceIds", a.AssumptionID, a.EvidenceIDs, impact.ElementEvidence)
	}
	c.stories(set)
	for i, r := range set.Reviews {
		p := "/learningReviews/" + strconv.Itoa(i)
		if r.Period == "" {
			c.errorf("VAL-REV-001", p+"/period", "", "learning review has no period")
		}
		for j, id := range r.RelatedIDs {
			if _, ok := c.kinds[id]; !ok && id != "" {
				c.add("VAL-REF-003", SeverityWarning, p+"/relatedIds/"+strconv.Itoa(j), "", "learning review refers to unknown element %q", id)
			}
		}
	}
	for i, n := range set.Notes {
		p := "/notes/" + strconv.Itoa(i)
		c.date(p+"/date", "", n.Date)
		for j, id := range n.RelatedIDs {
			if _, ok := c.kinds[id]; !ok && id != "" {
				c.add("VAL-REF-003", SeverityWarning, p+"/relatedIds/"+strconv.Itoa(j), "", "note refers to unknown element %q", id)
			}
		}
	}
}

func (c *checker) observation(p string, o impact.Observation) {
	if o.IndicatorID == "" {
		c.errorf("VAL-OBS-005", p+"/indicatorId", "", "observation has no indicatorId")
		return
	}
	if o.Date == "" {
		c.errorf("VAL-DATE-003", p+"/date", o.IndicatorID, "observation has no date")
	}
	c.date(p+"/date", o.IndicatorID, o.Date)
	enum(c, p+"/quality", o.IndicatorID, o.Quality)
	c.refs(p+"/evidenceIds", o.IndicatorID, o.EvidenceIDs, impact.ElementEvidence)

	kind, known := c.kinds[o.IndicatorID]
	if !known {
		c.errorf("VAL-REF-001", p+"/indicatorId", "", "observation refers to unknown indicator %q", o.IndicatorID)
		return
	}
	if kind != impact.ElementIndicator {
		c.errorf("VAL-REF-002", p+"/indicatorId", "", "references %q, which is a %s; expected indicator", o.IndicatorID, kind)
		return
	}
	ind, _ := c.doc.Indicator(o.IndicatorID)
	hasValue, hasLevel := o.Value != nil, o.Level != ""
	if !hasValue && !hasLevel {
		c.errorf("VAL-OBS-006", p, o.IndicatorID, "observation has neither value nor level")
	}
	switch ind.MeasurementType {
	case impact.MeasurementQualitative:
		if hasValue {
			c.add("VAL-OBS-007", SeverityWarning, p+"/value", o.IndicatorID, "qualitative indicator observed with a numeric value; use level")
		}
		if !hasLevel && hasValue {
			c.errorf("VAL-OBS-006", p+"/level", o.IndicatorID, "qualitative indicator needs a level")
		}
	case impact.MeasurementQuantitative, "":
		if hasLevel && !hasValue {
			c.errorf("VAL-OBS-006", p+"/value", o.IndicatorID, "quantitative indicator needs a numeric value")
		}
	}
	if hasLevel && len(ind.Levels) > 0 && !slices.Contains(ind.Levels, o.Level) {
		c.errorf("VAL-OBS-008", p+"/level", o.IndicatorID, "level %q is not one of the indicator's levels %v", o.Level, ind.Levels)
	}
}

// stories checks stories against the impact model's codebook and audiences, and
// holds them to the consent rules: every story states its consent, and a
// quote without granted consent is reported so it is never shown by accident.
func (c *checker) stories(set *impact.ObservationSet) {
	seen := map[string]bool{}
	for i, s := range set.Stories {
		p := "/stories/" + strconv.Itoa(i)
		if s.ID == "" {
			c.errorf("VAL-STORY-001", p+"/id", "", "story has no id")
		} else if seen[s.ID] {
			c.errorf("VAL-STORY-001", p+"/id", s.ID, "story id %q is used twice", s.ID)
		}
		seen[s.ID] = true
		if s.Date == "" {
			c.errorf("VAL-DATE-003", p+"/date", s.ID, "story has no date")
		}
		c.date(p+"/date", s.ID, s.Date)
		if s.Consent == "" {
			c.errorf("VAL-STORY-002", p+"/consent", s.ID, "story does not say whether consent was given (granted, not_granted, or unknown)")
		}
		enum(c, p+"/consent", s.ID, s.Consent)
		enum(c, p+"/sensitivity", s.ID, s.Sensitivity)
		if s.Quote != "" && s.Consent != impact.ConsentGranted {
			c.add("VAL-STORY-003", SeverityWarning, p+"/quote", s.ID, "the quote will not be shown: consent is not granted")
		}
		c.refs(p+"/themeIds", s.ID, s.ThemeIDs, impact.ElementTheme)
		c.refs(p+"/outcomeIds", s.ID, s.OutcomeIDs, impact.ElementOutcome)
		c.refs(p+"/audienceIds", s.ID, s.AudienceIDs, impact.ElementAudience)
		c.actor(p+"/actorId", s.ID, s.ActorID)
	}
}
