---
name: toc-status-refresh
description: Bring a Theory of Change up to date with a new period's numbers from a spreadsheet, and report where it stands and what changed. Use when asked to load monthly or quarterly data, refresh status, or produce an executive summary from a spreadsheet.
---

# Refresh status from a spreadsheet

## Purpose

Turn a spreadsheet of the period's numbers into a status report without
anyone editing JSON, and without a single row disappearing unseen.

## Inputs

- The theory, `theory.yaml`.
- The period's spreadsheet, `period.csv`.
- A column mapping, `mapping.yaml`, which says which columns hold the
  indicator, the date, and the value (see `examples/intermediary/csv/`).
- The previous status snapshot, `status-prev.json`, if there is one.

## Rules

- Judgment is limited to reading the spreadsheet's headings and helping to
  write the mapping. Every number comes from a `theoryofchange` command.
- Never type a value into an observation file by hand.
- Every unmatched row must be shown to the user with its reason. Do not
  guess an indicator for a row that does not match; ask.
- Do not compute progress, status, or totals yourself.

## Command sequence

1. Import the sheet. Exit code 1 means some rows did not match:

   ```bash
   theoryofchange import observations --csv period.csv --map mapping.yaml \
     --toc theory.yaml --out observations.yaml --unmatched unmatched.csv
   ```

   Show the user `unmatched.csv`. Fix the sheet or the mapping, and import again.

2. Check the observations against the theory:

   ```bash
   theoryofchange validate observations.yaml --toc theory.yaml
   ```

3. Compute the status, and compare it with the last period:

   ```bash
   theoryofchange status theory.yaml -O observations.yaml --prev status-prev.json --out status.json
   ```

4. Render the executive report:

   ```bash
   theoryofchange report exec theory.yaml -O observations.yaml --prev status-prev.json --out exec.md
   ```

## Checks to report

- Rows imported and rows unmatched, with the reasons.
- Indicators that have new data this period.
- What changed since the previous snapshot.

## Outputs

`observations.yaml`, `status.json` (keep it as next period's `--prev`), and
`exec.md`.
