# Reports and Decks

Every report, deck, and table is built from the same view of the theory, so a
number or a label never differs between artifacts. The output is deterministic:
the same inputs give byte-identical files.

## Definition reports

These describe the theory and need no observations, so a workshop can review
them before any data exists.

```bash
theoryofchange report logic-model theory.yaml --pathway pw-supply
theoryofchange report definitions theory.yaml --out definitions.md
```

| Kind | What it shows |
|------|---------------|
| `pathway` | The causal pathways, as diagrams |
| `logic-model` | Inputs, activities, outputs, and outcomes for each pathway. `--diagram` draws it left to right (also `logic-model-diagram`) |
| `logic-model-layered` | The logic model as columns, from inputs to impact, with a box for each item and an arrow wherever the theory says one leads to another. Written as Mermaid |
| `logic-model-d2` | The same layered diagram as [D2](https://d2lang.com), where a grid keeps every stage in its own column. Render it with `d2 diagram.d2 diagram.svg` |
| `outcome-map` | The outcomes framework, marking what is monitored but not owned |
| `logframe` | A results-framework matrix with indicators, means of verification, and assumptions |
| `metrics` | Metric definitions and how they combine across pathways |
| `metric-selection` | Candidate indicators scored against your criteria, best first |
| `collection-plan` | What is collected, how, how often, and the burden against the budget |
| `glossary` | The words the theory uses |
| `assessment` | How much of the theory is stated, inferred, or still a hypothesis |
| `definitions` | All of the above as one document |

## Status reports

These need observations (`-O`) and describe how far the theory has come.

```bash
theoryofchange report exec theory.yaml -O obs-q3.yaml --prev status-q2.json
theoryofchange report story theory.yaml -O obs-q3.yaml --audience aud-funders
theoryofchange report learning-review theory.yaml -O obs-q3.yaml --period 2026-Q3
```

| Kind | What it shows |
|------|---------------|
| `exec` | The executive summary: where things stand and what changed |
| `results` | Indicator and outcome results, separating what the organization answers for from what it only monitors |
| `assumptions` | The assumption register and what moved |
| `interventions` | Each intervention and what it is delivering |
| `evidence-gaps` | Where claims rest on weak or missing evidence |
| `full` | Every status report in one document |
| `story` | An update for one audience (`--audience`), using the numbers and consented stories |
| `headline` | The numbers each audience is told, from its headline list, with the latest figure, status, and caveats (`--audience` narrows it to one) |
| `metric-tree` | Per pathway: its North Star and guardrails, then its indicators from activity to impact, as a table and a diagram (`--pathway` narrows it) |
| `learning-review` | A period's review (`--period`): what changed, surprises, decisions |
| `portfolio` | Pathways side by side, described without recommending |
| `narrative` | The theory as prose, checking the solution against the problem |
| `actors` | Who does what, including partners |
| `custom` | Your own Go `text/template` over the report view (`--template`) |

`narrative`, `actors`, and `custom` also work without observations.

### Consent and sensitivity

Stories and quotes are gated. A story appears only if consent was granted, and a
quote only if its sensitivity is `public` or `internal`. Evidence marked
restricted is never rendered: its excerpt, summary, and URL are withheld from
every report, deck, and judge bundle, since a URL can reveal where a confidential
document lives.

## Slide decks

```bash
theoryofchange deck workshop theory.yaml -o workshop.md
theoryofchange deck exec theory.yaml -O obs-q3.yaml --export exec.pdf
```

| Deck | Contents |
|------|----------|
| `workshop` | The problem, the goal and ceiling, a diagram, a logic model table, and the logic model drawn as columns with arrows per pathway, key assumptions, metric definitions, the glossary, and what is still our reading. Needs no observations |
| `exec` | Where we stand, and what changed since the last snapshot (`--prev`) |
| `learning-review` | Behind plan, what we decided, and questions for discussion (`--period`) |
| `story` | One audience's deck: what you told us you need, the numbers, what we ask of you (`--audience`) |

A deck is written as [Marp](https://marp.app) Markdown. With `--export file`
(`.pdf`, `.pptx`, or `.html`) it is also rendered by `marp`, and by `mmdc` when it
has diagrams. Each deck carries speaker notes and shows the export stamp (theory, version, and
hash) on its title slide.

## Word and PowerPoint

```bash
theoryofchange export docx definitions theory.yaml --out theory.docx --reference-doc template.docx
theoryofchange export pptx workshop theory.yaml --out workshop.pptx
```

These convert the Markdown to a Pandoc-ready profile (diagrams as PNG, basis
badges as text, page breaks) and run `mmdc` and `pandoc`. `--reference-doc`
styles the output with your own template. If a tool is missing, nothing is
written and the exit code is 3.

## Spreadsheets

```bash
theoryofchange export csv indicators theory.yaml --out indicators.csv
```

Tables: `indicators`, `metrics`, `glossary`, `metric-selection`, and
`collection-plan`. The files follow RFC 4180. Multi-valued cells are joined with
`; `.

## The export stamp

Every export starts with a stamp, `toc:<id> version:<v> hash:<h>`, where the hash
is of the theory's canonical form. The same hash appears in every format, so you
can tell that a PDF, a Word file, and a spreadsheet came from the same version.
`--no-stamp` leaves it off, and `--as-of` records the date the artifact describes.

## PRISM exports

```bash
theoryofchange export okr theory.yaml -O obs-q3.yaml -o okr.json
theoryofchange export journey theory.yaml -o journeys.json
```

`export okr` turns each owned outcome that has measured indicators into an
objective, with each indicator as a key result carrying baseline, target,
current value, score, status, and confidence. `export journey` writes outcome
journeys. Both are JSON for the PRISM tools.

## Next

[Review round trip](review-round-trip.md)

## The layered logic model diagram

`logic-model --diagram` draws each stage as a group and joins the groups, which
says only that one stage follows another. The layered diagram draws the real
relationships:

- an **input** leads to the activities that use it, and one input can feed several;
- an **activity** leads to the outputs it lists, which lead on to the outcomes it targets;
- an **outcome** leads to the outcomes its causal links point to, so one box can lead to
  several, and several can lead to one.

The order within a column is chosen to keep arrows from crossing. A link that
runs back to an earlier stage is feedback. It is left out of the diagram, where
it would cut across the boxes between, and listed under it.

Both diagrams are available, so you can compare them. The layered one is made
by `pkg/logicmodel`, which builds a layered graph from the theory and writes it
through [d2vision](https://github.com/grokify/d2vision)'s diagram spec, the same
spec that writes the D2 and the Mermaid.
