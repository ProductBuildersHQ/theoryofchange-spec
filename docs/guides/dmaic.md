# DMAIC

If you already run DMAIC (Define, Measure, Analyze, Improve, Control) efforts,
this page shows where each part of a DMAIC metrics document comes from in an
impact model, how to mark the programs that suit DMAIC, and what the impact
model checks before it lets you start.

```bash
impact-model export dmaic impact-model.yaml -O observations.yaml -o dmaic.json
```

The output is a DMAIC metrics document in the JSON format of
[prism-roadmap](https://github.com/grokify/prism-roadmap)'s `goals/dmaic`
package, which validates it and can render it as slides or as Markdown for PDF.

## Which programs suit DMAIC

DMAIC improves a process with statistics: control charts need many measurements
taken on a steady schedule. A program whose numbers arrive weekly, such as an
online course catalog with completion rates and a post-course survey, suits it.
A program measured once a year does not: twenty points would take twenty years.

So DMAIC is opt-in, per pathway. Mark a pathway, the impact model's program or
product, with `projections`:

```yaml
pathways:
  - id: pw-catalog
    title: Course catalog
    projections:
      - dmaic
```

Only marked pathways are checked by the DMAIC lint rules and exported.

## What lint checks first

`impact-model lint` checks every indicator of a marked pathway:

| Rule | Checks | Why |
|------|--------|-----|
| `IMP-DMA-001` | It has a `frequency` and a data source (`dataSource` or `collection.system`) | Without them nothing is measured on a schedule |
| `IMP-DMA-002` | It is measured at least 20 times a year | Control limits need 20 to 25 points; fewer a year means waiting years |
| `IMP-DMA-003` | A survey indicator gets at least 30 responses per measurement | Below 30 a score swings with chance; an NPS needs about 100 for a margin near ±10 points |

Responses per measurement are `collection.respondentsPerYear`, times
`collection.sampleRate` when set, divided by measurements per year. A weekly
survey of 6,500 learners a year gives about 125 per week. The same survey of 780
learners gives 15, and lint suggests measuring every month, pooling across
courses, or asking more people.

The rules are warnings, not errors: an indicator that fails them can still be
exported, but its chart will not be ready for a long time.

## The mapping at a glance

| DMAIC | Comes from | In impact model terms |
|-------|-----------|-----------------------|
| **Category** | each pathway marked `dmaic` | A program or product: a group of outcomes and the interventions that bring them about |
| **Metric** | each indicator of the pathway with a numeric baseline and target | An indicator measures one outcome or intervention |
| Metric **baseline**, **target** | `baseline.value`, and the value of the latest-dated target | |
| Metric **current** | the latest observation, or the baseline before any | |
| Metric **data points** | the indicator's observations, given with `-O` | Observations are dated values recorded against an indicator |
| Metric **control limits** | computed from the data points, only from 20 or more | |
| Metric **trend direction** | `direction`: increase, decrease, or maintain | |
| Metric **status** | the indicator's health on the status date: Green, Yellow, Red | |
| Metric **phase** | derived; see [Phases](#phases) | |
| Metric **root causes** | the causes the measured outcome addresses | The problem tree: `causes`, each immediate, underlying, or structural |
| **Initiatives** | the pathway's interventions aimed at the measured outcome | An intervention is something the organization does to change an outcome |

## One metric, end to end

The [course catalog example](examples.md) is a fictional nonprofit's free
on-demand certificate courses. Its North Star is the weekly completion rate:

```yaml
indicators:
  - id: ind-completion-rate
    measures: oc-mt-certified
    name: Course completion rate
    unit: percent
    direction: increase
    baseline: { value: 38, date: "2025-12-29" }
    targets:
      - { value: 45, date: "2026-12-31" }
      - { value: 50, date: "2027-12-31" }
    dataSource: Learning platform
    frequency: P1W
```

With 24 weeks of observations it becomes this metric, shown trimmed:

```json
{
  "id": "ind-completion-rate",
  "name": "Course completion rate",
  "baseline": 38,
  "current": 40.9,
  "target": 50,
  "unit": "percent",
  "phase": "Improve",
  "trendDirection": "higher_better",
  "controlLimits": { "ucl": 42.855, "lcl": 35.987, "centerLine": 39.421, "sigma": 1.145 },
  "frequency": "Weekly",
  "dataSource": "Learning platform",
  "status": "Green",
  "dataPoints": [
    { "timestamp": "2026-01-05T00:00:00Z", "value": 38 },
    { "timestamp": "2026-01-12T00:00:00Z", "value": 39 }
  ],
  "rootCauses": [
    { "id": "cause-no-time", "description": "Staff cannot set aside an hour at a time for a course.", "category": "Immediate", "validated": true }
  ],
  "initiativeIds": ["int-course-redesign"]
}
```

- The baseline is the indicator's baseline and the target its 2027 target, the
  latest one.
- The data points are the 24 weekly observations, and the current value is the
  last of them.
- With 24 points, the control limits are computed.
- The outcome it measures, learners completing courses, addresses the cause
  "Staff cannot set aside an hour at a time". The cause cites evidence, so it is
  marked validated.
- The intervention aimed at that outcome, shortening modules, is in progress,
  so the metric is in the Improve phase and lists it as its initiative.

## Data points and the current value

Data points are the indicator's observations from the files given with `-O`, in
date order. An observation broken down by a group, with a `disaggregation` such
as `member size: small`, describes only that group, so it is not a point on the
total's chart and is left out.

The current value is the latest data point. Before there are any, it is the
baseline.

## Control limits

Control limits are those of an individuals chart (I-MR), the usual chart for one
measurement per period:

- the center line is the mean of the data points
- sigma is estimated as the average moving range, the mean difference between
  consecutive points, divided by 1.128
- the upper and lower limits are the center line plus and minus three sigma

They are computed only from 20 or more data points. With fewer the metric has
no limits, because limits from a handful of points mislead more than they help.

A point outside the limits is a signal that the process has changed. A metric
that is improving on purpose will cross a limit; that is the improvement
showing, and the limits should be recomputed from the new level.

Process capability (Cp, Cpk, sigma level) is never filled. It compares the
process with specification limits, the bounds a result must stay within, and an
impact model records targets, not specification limits.

## Phases

DMAIC tracks which phase each metric is in. The impact model does not record a
phase, so the export derives one, as a convention:

| Phase | When |
|-------|------|
| **Control** | The indicator has achieved its target on the status date. Hold the gain |
| **Measure** | There are fewer than 20 data points, too few for control limits |
| **Improve** | An intervention aimed at the measured outcome is in progress |
| **Analyze** | Otherwise: enough data, the target not yet met, and nothing in progress to change it |

Define is never assigned: a metric with a baseline and a target has already been
defined.

## Root causes and initiatives

A metric's root causes are the causes its outcome addresses, from the model's
problem tree. Each cause's level becomes the category, such as Immediate or
Structural, and a cause that cites evidence is marked validated.

The document's initiatives are the interventions of the marked pathways that aim
at a metric's outcome, or that the metric measures directly. Each lists the
metrics it should move, and its expected impact names the outcomes it targets.

| Initiative status | From the intervention |
|-------------------|-----------------------|
| Planned | `planned` |
| In Progress | `in_progress` |
| Completed | `completed` |
| Cancelled | not exported: a cancelled intervention is not an initiative |

A paused intervention is exported with no status. External links come from
`initiativeRefs`, such as a VisionStudio or Jira item.

## What is left out

| Left out | Why |
|----------|-----|
| An indicator with no numeric baseline or target, such as a qualitative rating | A DMAIC metric is a number moving from a baseline to a target. Each is named on standard error |
| Pathways not marked `dmaic` | They are not DMAIC efforts |
| The causal links, assumptions, and evidence | `report pathway` and `report assumptions` show them |
| Thresholds and process capability | The model records targets, not warning thresholds or specification limits |

## When the export is refused

`export dmaic` writes nothing and exits with status 2 when no pathway is marked
`dmaic`, or when no indicator of a marked pathway has a numeric baseline and
target:

```
impact-model: impact-model.yaml: the impact model cannot be drawn as DMAIC: mark a pathway with projections: [dmaic]
```
