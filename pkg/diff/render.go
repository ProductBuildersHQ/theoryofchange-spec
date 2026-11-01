package diff

import (
	"fmt"
	"strings"

	"github.com/ProductBuildersHQ/impact-model/pkg/edit"
	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
)

// Counts returns how many elements were added, changed, and removed.
func (r *Result) Counts() (added, changed, removed int) {
	for _, c := range r.Changes {
		switch c.Op {
		case Added:
			added++
		case Changed:
			changed++
		case Removed:
			removed++
		}
	}
	return added, changed, removed
}

// ChangedIDs lists the IDs of every element that was added, changed, or
// removed, in the order of the result.
func (r *Result) ChangedIDs() []string {
	ids := make([]string, 0, len(r.Changes))
	for _, c := range r.Changes {
		ids = append(ids, c.ElementID)
	}
	return ids
}

// Markdown renders the difference for a reader.
func (r *Result) Markdown() string {
	var b strings.Builder
	title := "Changes"
	if r.OldVersion != "" || r.NewVersion != "" {
		title = fmt.Sprintf("Changes from version %s to %s", orUnversioned(r.OldVersion), orUnversioned(r.NewVersion))
	}
	fmt.Fprintf(&b, "# %s\n\n", title)
	if r.Empty() {
		b.WriteString("Nothing changed.\n")
		return b.String()
	}
	added, changed, removed := r.Counts()
	fmt.Fprintf(&b, "%d element(s) added, %d changed, %d removed.\n\n", added, changed, removed)

	if len(r.Document) > 0 {
		b.WriteString("## The impact model\n\n")
		for _, f := range r.Document {
			fmt.Fprintf(&b, "- `%s`: %s\n", f.Path, arrow(f))
		}
		b.WriteString("\n")
	}
	kind := ""
	for _, c := range r.Changes {
		if c.Kind != kind {
			if kind != "" {
				b.WriteString("\n")
			}
			kind = c.Kind
			fmt.Fprintf(&b, "## %s\n\n", heading(kind))
		}
		note := ""
		if c.Rejected {
			note = " (rejected)"
		}
		fmt.Fprintf(&b, "- **%s** `%s`%s", c.Op, c.ElementID, note)
		if c.Title != "" && c.Title != c.ElementID {
			fmt.Fprintf(&b, ": %s", c.Title)
		}
		b.WriteString("\n")
		for _, f := range c.Fields {
			fmt.Fprintf(&b, "  - `%s`: %s\n", f.Path, arrow(f))
		}
	}
	return b.String()
}

func orUnversioned(v string) string {
	if v == "" {
		return "(unversioned)"
	}
	return v
}

func arrow(f FieldChange) string {
	switch {
	case f.From == "" && f.To != "":
		return "set to " + quote(f.To)
	case f.To == "" && f.From != "":
		return "removed (was " + quote(f.From) + ")"
	default:
		return quote(f.From) + " → " + quote(f.To)
	}
}

func quote(s string) string { return "“" + s + "”" }

func heading(kind string) string {
	switch kind {
	case "group":
		return "Precondition groups"
	case "factor":
		return "External factors"
	case "target-group":
		return "Target groups"
	case "term":
		return "Glossary"
	case "evidence":
		return "Evidence"
	}
	return strings.ToUpper(kind[:1]) + kind[1:] + "s"
}

// DraftRevision returns a revision entry that records this difference: the
// version, date, and trigger given, a summary of the counts, and the IDs
// that changed. Nothing is applied; see RevisionPatch.
func (r *Result) DraftRevision(version, date string, trigger impact.RevisionTrigger) impact.Revision {
	added, changed, removed := r.Counts()
	return impact.Revision{
		Version: version, Date: date, Trigger: trigger,
		Summary:    fmt.Sprintf("%d element(s) added, %d changed, %d removed.", added, changed, removed),
		ChangedIDs: r.ChangedIDs(),
	}
}

// RevisionPatch returns the patch that appends a revision to a document
// that has (hasRevisions) or does not have a revisions list yet, ready for
// the patch engine.
func RevisionPatch(rev impact.Revision, hasRevisions bool) (edit.Patch, error) {
	var op edit.Op
	var err error
	if hasRevisions {
		op, err = edit.Add(edit.Pointer("revisions", "-"), rev)
	} else {
		op, err = edit.Add(edit.Pointer("revisions"), []impact.Revision{rev})
	}
	if err != nil {
		return nil, err
	}
	return edit.Patch{op}, nil
}
