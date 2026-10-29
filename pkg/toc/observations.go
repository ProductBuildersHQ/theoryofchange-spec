package toc

import "slices"

// ObservationSet holds measurements for one Theory of Change, usually for
// one period. It is a separate document so the slow-changing theory and the
// fast-changing measurements have independent histories.
type ObservationSet struct {
	Schema      string `json:"$schema,omitempty"`
	Kind        Kind   `json:"kind" jsonschema:"required,enum=ObservationSet"`
	SpecVersion string `json:"specVersion" jsonschema:"required"`
	// TocID is the id of the document these observations measure.
	TocID string `json:"tocId" jsonschema:"required"`
	// Period labels the reporting period, e.g. "2026-Q4".
	Period       string                 `json:"period,omitempty"`
	Observations []Observation          `json:"observations,omitempty"`
	Assessments  []AssumptionAssessment `json:"assumptionAssessments,omitempty"`
	Notes        []Note                 `json:"notes,omitempty"`
}

// NewObservationSet returns an observation set with Kind and SpecVersion
// set.
func NewObservationSet(tocID, period string) *ObservationSet {
	return &ObservationSet{Kind: KindObservationSet, SpecVersion: SpecVersion, TocID: tocID, Period: period}
}

// DataQuality is how trustworthy an observed value is.
type DataQuality string

const (
	QualityVerified  DataQuality = "verified"
	QualityReported  DataQuality = "reported"
	QualityEstimated DataQuality = "estimated"
)

// AllDataQualities returns data qualities from most to least trustworthy.
func AllDataQualities() []DataQuality {
	return []DataQuality{QualityVerified, QualityReported, QualityEstimated}
}

// Valid reports whether v is a known DataQuality.
func (v DataQuality) Valid() bool { return slices.Contains(AllDataQualities(), v) }

// Observation is one measured value of an indicator.
type Observation struct {
	IndicatorID string `json:"indicatorId" jsonschema:"required"`
	Date        string `json:"date" jsonschema:"required,format=date"`
	// Value is set for quantitative indicators; Level for qualitative ones.
	Value          *float64          `json:"value,omitempty"`
	Level          string            `json:"level,omitempty"`
	Source         string            `json:"source,omitempty"`
	EvidenceIDs    []string          `json:"evidenceIds,omitempty"`
	Note           string            `json:"note,omitempty"`
	Quality        DataQuality       `json:"quality,omitempty"`
	Disaggregation map[string]string `json:"disaggregation,omitempty"`
}

// AssumptionAssessment records an assumption's status at a point in time.
// The latest assessment overrides the document's static status.
type AssumptionAssessment struct {
	AssumptionID string           `json:"assumptionId" jsonschema:"required"`
	Date         string           `json:"date" jsonschema:"required,format=date"`
	Status       AssumptionStatus `json:"status" jsonschema:"required"`
	Rationale    string           `json:"rationale,omitempty"`
	EvidenceIDs  []string         `json:"evidenceIds,omitempty"`
}

// Note is a dated free-text note about the period.
type Note struct {
	Date       string   `json:"date,omitempty" jsonschema:"format=date"`
	Text       string   `json:"text" jsonschema:"required"`
	RelatedIDs []string `json:"relatedIds,omitempty"`
}

// ActorRegistry is a shared list of actors that several theories reference,
// for example the champions of competing futures.
type ActorRegistry struct {
	Schema      string  `json:"$schema,omitempty"`
	Kind        Kind    `json:"kind" jsonschema:"required,enum=ActorRegistry"`
	SpecVersion string  `json:"specVersion" jsonschema:"required"`
	ID          string  `json:"id" jsonschema:"required"`
	Name        string  `json:"name,omitempty"`
	Actors      []Actor `json:"actors" jsonschema:"required"`
}
