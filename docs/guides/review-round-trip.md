# Review Round Trip

An impact model is only as good as the people who have read it. The round trip sends it
out for review, brings the answers back as proposed changes you can inspect, and
records approval. Nothing changes the impact model until you apply it.

```text
export packet or sheet  ->  reviewers answer  ->  import review  ->  apply patch  ->  approve
```

## 1. Send it out

Two formats carry the same rows, one per element:

```bash
impact-model export review-packet impact-model.yaml --format docx --out packet.docx   # for reviewers who work in a document
impact-model export review-sheet impact-model.yaml --out sheet.csv                    # for reviewers who work in a spreadsheet
```

Each row has the element's ID, its statement, its basis, and its sources, with
empty **Decision**, **Suggested change**, and **Comment** columns. Rejected
elements are not put to reviewers. Sources are named by ID and title only;
excerpts and URLs are never included.

The decision is one of `agree`, `change`, `remove`, or `unsure`. Both formats
carry the export stamp, so a reply can be matched to the version it answers.

## 2. Collect the answers

Reviewers fill in the packet or the sheet and return it. You can also write the
answers as a `ReviewResponse` file (see `impact-model schema review-response`).
How the files travel (email, a shared drive, a form) is up to you.

## 3. Bring them back

```bash
impact-model import review impact-model.yaml --from-csv returned-sheet.csv \
  --reviewer "Program board" --date 2026-06-15 \
  --out review.patch.json --consultation-out consultation.json
```

The impact model is never changed. The command writes a patch of the changes it could
work out, a list of the rows that need a person to decide, and a consultation
record of the review. A row is handled like this:

| Decision | What happens |
|----------|--------------|
| `agree` | The element's basis becomes `confirmed` |
| `remove` | The element's basis becomes `rejected` |
| `change` | If the suggestion is `field: value` lines, those fields are set. Free text is flagged for a person |
| `unsure`, or a comment with no decision | Flagged |

A row with no element ID is matched to an element by its quoted text, and flagged
if no single element matches. If the answers were written against a different
version of the impact model (the hash differs), every row is flagged.

## 4. Apply what you accept

```bash
impact-model patch impact-model.yaml review.patch.json --dry-run
impact-model patch impact-model.yaml review.patch.json
```

Review the patch first. It goes through the same validated path as every other
edit.

## 5. Approve

```bash
impact-model approve impact-model.yaml --version 1.0.0 --date 2026-07-01 \
  --approver "Board chair:chair" --approver "Executive director:director"
```

Approval sets the status to `approved`, sets the version label, and appends a
revision naming the approvers. It is refused, with nothing written, if the impact model
has validation errors; if elements are still hypotheses (confirm or reject them,
or use `--force` with `--reason`, which is recorded); or if the version label is
missing or already used, the date is missing, or there is no approver. Use a name
or role you are happy to publish.

## Sharing through Google Docs and Sheets

The packet and the sheet are ordinary files, so any shared drive works. To use
Google Docs and Sheets, the [`gogoogle`](https://github.com/grokify/gogoogle)
command line can upload the packet as a Google Doc (`gogoogle drive upload
packet.docx --convert doc`), list its comments (`gogoogle drive comments`), show
reviewers' suggested edits (`gogoogle docs suggestions`), and read a returned
sheet as records (`gogoogle sheets records`). This tool makes no network calls
itself. An agent skill that runs that whole workflow is planned.

## Next

[Quality and judging](quality-and-judging.md)
