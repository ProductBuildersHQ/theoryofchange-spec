# V2MOM

If you already write V2MOMs, this page shows where each part of a V2MOM comes
from in an impact model, and what the impact model adds that a V2MOM written by
hand usually leaves out.

```bash
impact-model export v2mom impact-model.yaml -O observations.yaml -o v2mom.json
```

The output is a V2MOM document in the JSON format of
[prism-roadmap](https://github.com/grokify/prism-roadmap)'s `goals/v2mom`
package, which validates it and can render it as slides.

## The mapping at a glance

| V2MOM | Comes from | In impact model terms |
|-------|-----------|-----------------------|
| **Vision** | `context.vision` | The future the organization works toward |
| **Values** | `context.values`, in order | The principles that guide how it works, most important first |
| **Methods** | `pathways`, one method each | A pathway is a program, product, or strategy: a group of outcomes and the interventions that bring them about |
| A method's **projects** | the pathway's `interventionIds` | An intervention is something the organization does to change an outcome, such as a coaching program or a feature launch |
| A method's **measures** | the pathway's `focus`: North Star, one metric that matters, guardrails | The few indicators the program steers by |
| **Obstacles** | `risks`, and high or critical `assumptions` not yet shown to hold | What could stop the plan |
| A method's **owner** | the pathway's `ownerActorId` | The actor who answers for the pathway |

The sections below explain each row. Two rules hold throughout:

- **Nothing is invented.** Every field in the V2MOM is copied or translated from
  something in the impact model. A section the model cannot fill is left empty,
  and a V2MOM that would fail validation is refused instead of written.
- **Order is kept.** Values, methods, measures, and projects appear in the order
  the impact model lists them, so the same model always gives the same V2MOM.

## One pathway, end to end

This pathway from the fictional [intermediary example](examples.md) is a program
called Learning Exchange:

```yaml
pathways:
  - id: pw-learning
    title: Learning Exchange
    kind: program
    description: Peer learning and leadership coaching for member leaders.
    ownerActorId: act-network
    outcomeIds: [oc-out-cohorts, oc-st-leader-skills, oc-st-peer-ties, oc-im-leaders-practice, oc-im-peer-network]
    interventionIds: [int-peer-cohorts, int-leader-coaching, int-training-series]
    focus:
      northStarId: ind-practice-maturity
      guardrailIds: [ind-retention-learning]
```

It becomes this method, shown trimmed, with observations as of 2026-09-30:

```json
{
  "id": "pw-learning",
  "name": "Learning Exchange",
  "description": "Peer learning and leadership coaching for member leaders.",
  "owner": "Riverbend Nonprofit Network",
  "measures": [
    {
      "id": "ind-practice-maturity",
      "name": "Leadership practice maturity",
      "description": "North Star: Median level of participating members on the four-step leadership practice rubric, rated by a coach and the member together.",
      "baseline": "Emerging",
      "target": "Established",
      "current": "Emerging",
      "timeline": "2028-12-31",
      "status": "On Track"
    }
  ],
  "obstacles": [
    { "id": "rsk-time-burden", "name": "Leaders drop out of cohorts because they cannot spare the time." },
    { "id": "asm-skills-applied", "name": "Leaders put coaching-based skills into practice back at their organizations." }
  ],
  "projects": ["int-peer-cohorts", "int-leader-coaching", "int-training-series"]
}
```

- The pathway's title, description, and owner become the method's.
- The North Star leads the measures, and the guardrail follows it (trimmed here).
- The three interventions become the method's projects, each also listed in the
  V2MOM's project list with `"methodId": "pw-learning"`.
- A risk that applies to the cohorts, and a critical assumption about leaders
  applying their skills, become the method's obstacles.
- The five outcomes and the links between them are not in the V2MOM. They are
  the reasoning behind the method, and `report pathway` draws them.

## Vision

The V2MOM's vision is `context.vision`, word for word.

```yaml
context:
  vision: A county where every resident can count on a strong local nonprofit when they need one.
```

The impact model also has `context.mission`, why the organization exists. A
V2MOM has no mission section, so the mission is not exported.

## Values

The V2MOM's values are `context.values`. The first value has priority 1, the
next priority 2, and so on.

```yaml
context:
  values:
    - name: Members first
      description: Every program answers to what member nonprofits need, not to what is easy to offer.
    - name: Stronger together
      description: Shared services and a shared voice do more than each organization can alone.
```

Values frame the model, like the mission and vision. They are not outcomes:
nothing in the model causes them, and they are not measured.

## Methods

Each **pathway** becomes one method, in the order the model lists them. A
pathway is the impact model's unit of "how we will get there": a program,
product, or strategy, with its own outcomes, its own interventions, and its own
owner. A V2MOM usually has three to seven methods, and a model usually has a
similar number of pathways.

| Method field | From the pathway |
|--------------|------------------|
| ID | `id` |
| Name | `title` |
| Description | `description` |
| Owner | the name of the actor in `ownerActorId` |

A pathway marked rejected, such as a program the organization decided against,
is not exported.

### Projects

A method's projects are the pathway's **interventions**, the things the
organization does. They are listed once in the V2MOM's project list, each with
the ID of the method it belongs to.

| Project field | From the intervention |
|---------------|-----------------------|
| ID, Name, Description | `id`, `title`, `description` |
| Method | the pathway it is listed in |
| Status | `planned` → Approved, `in_progress` → In Progress, `completed` → Completed, `cancelled` → Cancelled; `paused` is left blank |
| External links | `initiativeRefs`, keyed by framework, such as a Jira or VisionStudio item |

"Intervention" is the impact model's word because the model is causal: each
intervention says which outcomes it should change. An *initiative* is the
organizational effort that delivers one, and `initiativeRefs` links the two.

### Measures

A method's measures are the indicators the pathway steers by, from its `focus`:

1. the **North Star**, the one number that best shows the value the program
   delivers
2. the **one metric that matters**, if the pathway names one for now
3. the **guardrails**, which watch what pursuing the North Star could harm

Each measure's description starts with its role, such as "North Star:".

| Measure field | From the indicator |
|---------------|--------------------|
| ID, Name, Unit | `id`, `name`, `unit` |
| Description | the role, then `definition` |
| Baseline | `baseline` |
| Target, Timeline | the latest-dated target and its date |
| Current, Progress, Status | the indicator's status, only when observations are given with `-O` |

Status is On Track, At Risk, Behind, or Achieved, from the indicator's state on
the status date. A pathway without a focus exports a method with no measures;
`impact-model lint` points out programs and products that have no North Star.

The V2MOM has no top-level measures. Every measure belongs to a method, which
the document records as the `nested` structure.

## Obstacles

Obstacles come from two places:

- **Risks**: things that could go wrong. Impact becomes severity, probability
  becomes likelihood, and the mitigation and status carry over.
- **Assumptions** the plan depends on that could still turn out false: those
  with high or critical criticality whose status is untested, at risk, or
  invalidated. With observations (`-O`), an assumption's assessed status is used,
  so one assessed as holding drops out and one assessed as at risk comes in.
  Without them, the status written in the model is used, and an assumption with
  no status counts as untested. The obstacle's severity is the assumption's
  criticality, and its description says how the assumption will be tested.

An obstacle that applies to one of a pathway's outcomes, interventions, or the
links between them is listed under that method. One that applies across
pathways, or to nothing in particular, is listed for the whole V2MOM.

## What a V2MOM leaves out

A V2MOM is a short plan, so the export keeps only what fits it. These parts of
the impact model are not in the V2MOM, and other projections show them:

| Left out | Where to see it |
|----------|-----------------|
| The outcomes and the causal links between them | `report pathway`, the theory of change |
| Every assumption, with its test, evidence, and status | `report assumptions` |
| Indicators other than a pathway's focus | `report metric-tree` |
| What each audience is told | `report headline` |
| Where each element came from: stated, inferred, or confirmed | `report assessment` |
| The mission | `context.mission` in the model |

## What the impact model adds

A V2MOM written by hand often states its methods and measures without saying why
the methods should lead to the vision. Exported from an impact model:

- each method is backed by a pathway of outcomes, each link with its reason
- each North Star is one that `impact-model lint` has checked sits at or below
  the organization's ceiling of accountability, so the V2MOM does not commit
  the organization to change it cannot answer for
- the obstacles include the assumptions the plan rests on, not only external
  risks
- with `-O`, the measures carry the latest observed values and status

## When the export is refused

A V2MOM needs a vision, at least one value, and at least one method. If the
impact model has no `context.vision`, no `context.values`, or no pathway,
`export v2mom` writes nothing and exits with status 2, naming what to add:

```
impact-model: impact-model.yaml: the impact model cannot be drawn as a V2MOM: add context.values
```

Add values with a patch:

```bash
impact-model patch impact-model.yaml - <<'EOF'
[{"op":"add","path":"/context/values","value":[
  {"name":"Members first","description":"Every program answers to what members need."}]}]
EOF
```

## Going the other way

There is no import from a V2MOM. A V2MOM has no outcomes, causal links, or
assumptions to read, so an impact model built from one would have to invent the
reasoning that makes it an impact model. Start the impact model from the V2MOM's
vision, values, and methods instead: each method suggests a pathway, and each
measure an indicator.
