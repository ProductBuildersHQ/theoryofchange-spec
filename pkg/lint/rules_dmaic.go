package lint

import (
	"fmt"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
	"github.com/ProductBuildersHQ/impact-model/pkg/validate"
)

const (
	// dmaicMinPointsPerYear is the fewest measurements a year that let a DMAIC
	// effort set control limits within about a year: statistical process
	// control asks for 20 to 25 points before the limits mean anything.
	dmaicMinPointsPerYear = 20
	// dmaicMinResponsesPerPoint is the fewest survey responses behind one
	// measurement before its value stops swinging with chance. An NPS needs
	// about 100 for a margin near ±10 points; 30 is the floor.
	dmaicMinResponsesPerPoint = 30
)

// dmaicIndicator is an indicator of a pathway marked for DMAIC.
type dmaicIndicator struct {
	indexedIndicator
	pathwayID string
}

// dmaicIndicators lists, once each and in file order, the active indicators
// that serve an active pathway marked for DMAIC.
func dmaicIndicators(c *Context) []dmaicIndicator {
	var out []dmaicIndicator
	for _, ind := range activeIndicators(c.Doc) {
		for _, p := range activePathways(c.Doc) {
			if slices.Contains(p.v.Projections, impact.ProjectionDMAIC) && c.Doc.IndicatorServes(ind.v, p.v.ID) {
				out = append(out, dmaicIndicator{ind, p.v.ID})
				break
			}
		}
	}
	return out
}

// The rules below check that a pathway marked for DMAIC is measured often
// enough, and from enough people, for control charts and process capability
// to mean something. They say nothing about a pathway that is not marked.
func init() {
	register("IMP-DMA-001", validate.SeverityWarning, "An indicator of a DMAIC pathway says how often it is measured and where the data comes from.", func(c *Context) []Finding {
		var out []Finding
		for _, d := range dmaicIndicators(c) {
			var missing []string
			if d.v.Frequency == "" {
				missing = append(missing, "a frequency")
			}
			if d.v.DataSource == "" && (d.v.Collection == nil || d.v.Collection.System == "") {
				missing = append(missing, "a data source")
			}
			if len(missing) > 0 {
				out = append(out, Finding{Path: at("indicators", d.i), Element: d.v.ID,
					Message: fmt.Sprintf("pathway %q is marked for DMAIC, and this indicator has no %s, so it cannot be measured on a schedule", d.pathwayID, strings.Join(missing, " or "))})
			}
		}
		return out
	})

	register("IMP-DMA-002", validate.SeverityWarning, "An indicator of a DMAIC pathway is measured at least 20 times a year.", func(c *Context) []Finding {
		var out []Finding
		for _, d := range dmaicIndicators(c) {
			n, ok := perYear(d.v.Frequency)
			if !ok || n >= dmaicMinPointsPerYear {
				continue
			}
			out = append(out, Finding{Path: at("indicators", d.i) + "/frequency", Element: d.v.ID,
				Message: fmt.Sprintf("it is measured about %s times a year (%s); control limits need at least %d measurements, so measure it every two weeks or more often, or leave pathway %q out of DMAIC",
					round(n), d.v.Frequency, dmaicMinPointsPerYear, d.pathwayID)})
		}
		return out
	})

	register("IMP-DMA-003", validate.SeverityWarning, "A survey indicator of a DMAIC pathway gets at least 30 responses per measurement.", func(c *Context) []Finding {
		var out []Finding
		for _, d := range dmaicIndicators(c) {
			col := d.v.Collection
			if col == nil || col.Mode != impact.ModeSurvey {
				continue
			}
			path := at("indicators", d.i) + "/collection"
			if col.RespondentsPerYear == 0 {
				out = append(out, Finding{Path: path + "/respondentsPerYear", Element: d.v.ID,
					Message: "it is a survey, and how many people are asked each year is not stated, so it cannot be told whether each measurement has enough responses"})
				continue
			}
			n, ok := perYear(d.v.Frequency)
			if !ok {
				continue // IMP-DMA-001 reports the missing frequency
			}
			asked := float64(col.RespondentsPerYear)
			if col.SampleRate > 0 {
				asked *= col.SampleRate
			}
			if per := asked / n; per < dmaicMinResponsesPerPoint {
				out = append(out, Finding{Path: path, Element: d.v.ID,
					Message: fmt.Sprintf("each measurement rests on about %s responses; below %d a survey score swings with chance (an NPS needs about 100 for ±10 points), so measure less often, pool across groups, or ask more people",
						round(per), dmaicMinResponsesPerPoint)})
			}
		}
		return out
	})
}

var isoDuration = regexp.MustCompile(`^P(?:(\d+)Y)?(?:(\d+)M)?(?:(\d+)W)?(?:(\d+)D)?(?:T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+(?:\.\d+)?)S)?)?$`)

// perYear is how many times a year an ISO 8601 duration repeats, using the
// average year and month. It reports false for an empty or unreadable
// duration and for one of zero length.
func perYear(iso string) (float64, bool) {
	m := isoDuration.FindStringSubmatch(iso)
	if m == nil {
		return 0, false
	}
	const daysPerYear = 365.2425
	unit := []float64{daysPerYear, daysPerYear / 12, 7, 1, 1.0 / 24, 1.0 / 1440, 1.0 / 86400}
	days := 0.0
	for i, part := range m[1:] {
		if part == "" {
			continue
		}
		v, err := strconv.ParseFloat(part, 64)
		if err != nil {
			return 0, false // the pattern admits only numbers; unreachable
		}
		days += v * unit[i]
	}
	if days == 0 {
		return 0, false
	}
	return daysPerYear / days, true
}

// round formats a count for a message: a whole number, or one decimal below 10.
func round(v float64) string {
	if v < 10 {
		return strconv.FormatFloat(math.Round(v*10)/10, 'f', -1, 64)
	}
	return strconv.FormatFloat(math.Round(v), 'f', 0, 64)
}
