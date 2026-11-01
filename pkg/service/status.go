package service

import (
	"fmt"
	"time"

	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
	"github.com/ProductBuildersHQ/impact-model/pkg/measure"
	"github.com/ProductBuildersHQ/impact-model/pkg/report"
)

// StatusInputFor evaluates doc against the observation sets as of asOf and
// returns what the status reports and decks are made from. prev, if given,
// is an earlier snapshot of the same impact model; the input then carries what
// changed since.
func StatusInputFor(doc *impact.Model, sets []*impact.ObservationSet, asOf time.Time, opts impact.StatusOptions, prev *impact.Status) (report.StatusInput, error) {
	g, err := BuildGraph(doc)
	if err != nil {
		return report.StatusInput{}, err
	}
	st, err := measure.Evaluate(doc, g, sets, asOf, opts)
	if err != nil {
		return report.StatusInput{}, err
	}
	v, err := report.BuildView(doc, g, st, report.Options{})
	if err != nil {
		return report.StatusInput{}, err
	}
	in := report.StatusInput{View: v, Doc: doc, Status: st}
	if prev != nil {
		if prev.ModelID != st.ModelID {
			return report.StatusInput{}, fmt.Errorf("the earlier snapshot is of %q, not %q", prev.ModelID, st.ModelID)
		}
		in.Diff = measure.Diff(prev, st)
	}
	return in, nil
}
