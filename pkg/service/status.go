package service

import (
	"fmt"
	"time"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/measure"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/report"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

// StatusInputFor evaluates doc against the observation sets as of asOf and
// returns what the status reports and decks are made from. prev, if given,
// is an earlier snapshot of the same theory; the input then carries what
// changed since.
func StatusInputFor(doc *toc.Document, sets []*toc.ObservationSet, asOf time.Time, opts toc.StatusOptions, prev *toc.Status) (report.StatusInput, error) {
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
		if prev.TocID != st.TocID {
			return report.StatusInput{}, fmt.Errorf("the earlier snapshot is of %q, not %q", prev.TocID, st.TocID)
		}
		in.Diff = measure.Diff(prev, st)
	}
	return in, nil
}
