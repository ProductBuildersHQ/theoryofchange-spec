package judge

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/plexusone/structured-evaluation/rubric"
)

// Verdict is a structured-evaluation report with the hash of the bundle it
// judged. The report fields sit at the top level beside bundleHash, so any
// consumer of the report format reads a verdict unchanged.
type Verdict struct {
	BundleHash string `json:"bundleHash"`
	rubric.Rubric
}

// Errors returned for a verdict that cannot be accepted.
var (
	ErrWrongBundle   = errors.New("the verdict judged a different version of the theory")
	ErrInvalidReport = errors.New("invalid verdict")
	ErrJudgeName     = errors.New("invalid judge name")
)

// ParseVerdict reads a verdict. Unknown fields are errors, so a misspelled
// field is not silently dropped.
func ParseVerdict(data []byte) (*Verdict, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var v Verdict
	if err := dec.Decode(&v); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidReport, err)
	}
	return &v, nil
}

// Check reports whether v can be accepted for b: it names b's hash, is a
// valid report, and scores only categories of b's rubric, each at most once,
// leaving none of the required ones unscored.
func (v *Verdict) Check(b *Bundle) error {
	if v.BundleHash != b.BundleHash {
		return fmt.Errorf("%w (verdict hash %q, current %q)", ErrWrongBundle, short(v.BundleHash), short(b.BundleHash))
	}
	if res := rubric.ValidateReport(&v.Rubric); res.HasErrors() {
		return fmt.Errorf("%w: %s", ErrInvalidReport, strings.TrimSpace(res.String()))
	}
	scored := map[string]bool{}
	for _, c := range v.Categories {
		if b.Rubric.GetCategory(c.Category) == nil {
			return fmt.Errorf("%w: category %q is not in the %s rubric", ErrInvalidReport, c.Category, b.Profile)
		}
		if scored[c.Category] {
			return fmt.Errorf("%w: category %q is scored twice", ErrInvalidReport, c.Category)
		}
		scored[c.Category] = true
	}
	for _, c := range b.Rubric.GetRequiredCategories() {
		if !scored[c.ID] {
			return fmt.Errorf("%w: required category %q is not scored", ErrInvalidReport, c.ID)
		}
	}
	return nil
}

func short(h string) string {
	if len(h) > 12 {
		return h[:12]
	}
	return h
}

// JudgeID names the judge that made the verdict: the report's judge ID, then
// its reviewer ID.
func (v *Verdict) JudgeID() string {
	if v.Judge != nil && v.Judge.JudgeID != "" {
		return v.Judge.JudgeID
	}
	return v.Metadata.ReviewerID
}

var judgeName = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,63}$`)

// ValidJudgeName reports whether name is safe to use as a file name: lower
// case letters, digits, dot, dash, and underscore, not starting with a
// symbol, so it can never name a path outside the store.
func ValidJudgeName(name string) bool {
	return judgeName.MatchString(name) && !strings.Contains(name, "..")
}

// Dir returns the directory that holds the verdicts for a bundle, beside the
// theory file.
func Dir(tocPath, bundleHash string) string {
	return filepath.Join(filepath.Dir(tocPath), ".toc", "judgements", bundleHash)
}

// Record stores a verdict for b under the judge's name and returns the path.
// The verdict must pass Check. Recording again for the same judge replaces
// the earlier verdict.
func Record(tocPath string, b *Bundle, v *Verdict, name string) (string, error) {
	if name == "" {
		name = strings.ToLower(v.JudgeID())
	}
	if !ValidJudgeName(name) {
		return "", fmt.Errorf("%w: %q (use lower case letters, digits, dot, dash, or underscore)", ErrJudgeName, name)
	}
	if err := v.Check(b); err != nil {
		return "", err
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", fmt.Errorf("judge: encode verdict: %w", err)
	}
	dir := Dir(tocPath, b.BundleHash)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("judge: create %s: %w", dir, err)
	}
	path := filepath.Join(dir, name+".json")
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil { //nolint:gosec // verdicts are meant to be shared
		return "", fmt.Errorf("judge: write %s: %w", path, err)
	}
	return path, nil
}

// Load reads the stored verdicts for b, sorted by judge name. A verdict that
// no longer passes Check is an error, not skipped.
func Load(tocPath string, b *Bundle) (map[string]*Verdict, []string, error) {
	dir := Dir(tocPath, b.BundleHash)
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]*Verdict{}, nil, nil
	}
	if err != nil {
		return nil, nil, fmt.Errorf("judge: read %s: %w", dir, err)
	}
	out := map[string]*Verdict{}
	var names []string
	for _, e := range entries {
		name, ok := strings.CutSuffix(e.Name(), ".json")
		if !ok || e.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, nil, fmt.Errorf("judge: read verdict %s: %w", e.Name(), err)
		}
		v, err := ParseVerdict(data)
		if err == nil {
			err = v.Check(b)
		}
		if err != nil {
			return nil, nil, fmt.Errorf("judge: verdict %s: %w", e.Name(), err)
		}
		out[name] = v
		names = append(names, name)
	}
	slices.Sort(names)
	return out, names, nil
}
