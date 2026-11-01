package review

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
)

// SheetMeta supplies what a returned sheet does not say itself.
type SheetMeta struct {
	// Reviewers name groups or roles.
	Reviewers []string
	// Date is the day the review was done.
	Date string
	// ModelID, Version and ContentHash override the sheet's stamp row when
	// set; they fill in a sheet whose stamp was lost on the way back.
	ModelID, Version, ContentHash string
}

// sheetColumns are the columns the importer needs from a returned sheet.
var sheetColumns = []string{"kind", "id", "decision", "suggested_change", "comment"}

// FromSheet reads a returned review sheet (see share.ReviewSheet) into a
// response. The export stamp row, if still present, supplies the impact model's
// ID, version, and hash. Rows the reviewer left untouched are skipped. A
// sheet with no stamp and no override yields an empty hash, which the import
// reports as stale.
func FromSheet(data []byte, m SheetMeta) (*impact.ReviewResponse, error) {
	resp := &impact.ReviewResponse{
		Kind: impact.KindReviewResponse, SpecVersion: impact.SpecVersion,
		ModelID: m.ModelID, Version: m.Version, ContentHash: m.ContentHash,
		Reviewers: m.Reviewers, Method: impact.ReviewSheet, Date: m.Date, Rows: []impact.ReviewItem{},
	}
	r := csv.NewReader(bytes.NewReader(data))
	r.FieldsPerRecord = -1
	r.LazyQuotes = true
	var col map[string]int
	for {
		rec, err := r.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("review: read sheet: %w", err)
		}
		if col == nil {
			if id, version, hash, ok := parseStamp(rec); ok {
				resp.ModelID = orString(m.ModelID, id)
				resp.Version = orString(m.Version, version)
				resp.ContentHash = orString(m.ContentHash, hash)
				continue
			}
			if col, err = header(rec); err != nil {
				return nil, err
			}
			continue
		}
		item, ok := sheetRow(rec, col)
		if ok {
			resp.Rows = append(resp.Rows, item)
		}
	}
	if col == nil {
		return nil, errors.New("review: the sheet has no header row")
	}
	return resp, nil
}

func orString(override, fallback string) string {
	if override != "" {
		return override
	}
	return fallback
}

// parseStamp reads "# model:<id> version:<v> hash:<h>" from the first field of
// a record.
func parseStamp(rec []string) (id, version, hash string, ok bool) {
	if len(rec) == 0 || !strings.HasPrefix(strings.TrimSpace(rec[0]), "# model:") {
		return "", "", "", false
	}
	for _, f := range strings.Fields(strings.TrimPrefix(strings.TrimSpace(rec[0]), "#")) {
		switch {
		case strings.HasPrefix(f, "model:"):
			id = strings.TrimPrefix(f, "model:")
		case strings.HasPrefix(f, "version:"):
			version = strings.TrimPrefix(f, "version:")
		case strings.HasPrefix(f, "hash:"):
			hash = strings.TrimPrefix(f, "hash:")
		}
	}
	return id, version, hash, true
}

func header(rec []string) (map[string]int, error) {
	col := map[string]int{}
	for i, name := range rec {
		col[strings.ToLower(strings.TrimSpace(name))] = i
	}
	for _, want := range sheetColumns {
		if _, ok := col[want]; !ok {
			return nil, fmt.Errorf("review: the sheet has no %q column", want)
		}
	}
	return col, nil
}

// sheetRow reads one record; it reports false for a row the reviewer did not
// answer.
func sheetRow(rec []string, col map[string]int) (impact.ReviewItem, bool) {
	get := func(name string) string {
		if i := col[name]; i < len(rec) {
			return strings.TrimSpace(rec[i])
		}
		return ""
	}
	item := impact.ReviewItem{
		ElementID:       get("id"),
		Decision:        impact.ReviewDecision(strings.ToLower(get("decision"))),
		SuggestedChange: get("suggested_change"),
		Comment:         get("comment"),
		Origin:          impact.OriginSheet,
	}
	if item.Decision == "" && item.SuggestedChange == "" && item.Comment == "" {
		return impact.ReviewItem{}, false
	}
	return item, true
}
