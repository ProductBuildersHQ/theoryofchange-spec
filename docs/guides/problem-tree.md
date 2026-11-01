# Problem Tree

A problem tree puts the problem a program exists to solve in the middle, its
causes below as roots, and its effects above as branches. It comes before the
theory of change: the roots say what the program could work on, and the
branches say what harm it should reduce. Drawn from an impact model, it is one
more projection of the same model, and it checks the solution against the
problem: every cause and every effect should be addressed by an outcome or set
aside with a reason.

```bash
impact-model report problem-tree impact-model.yaml        # Markdown, with a Mermaid diagram
impact-model report problem-tree-d2 impact-model.yaml     # the same, with a D2 diagram
impact-model report problem-tree-json impact-model.yaml   # the diagram document itself, as JSON
```

## The parts and where they live

| Diagram | In the model | Notes |
|---------|--------------|-------|
| **Core problem**, in the middle | `context.problemStatement` | The problem the program responds to |
| **Causes** (roots), below | `causes` | Each with a level: `immediate`, `underlying`, or `structural` |
| A cause leading to another | a cause's `parents` | The causes it contributes to; none means it leads straight to the core problem |
| **Effects** (branches), above | `effects` | What the core problem leads to |
| An effect leading to another | an effect's `parents` | The effects it follows from; none means it follows straight from the core problem |
| What addresses a cause | an outcome's `addressesCauses` | Usually earlier outcomes: what the program works on |
| What relieves an effect | an outcome's `addressesEffects` | Usually long-term and impact outcomes: the harm the program reduces |
| Set aside | `outOfScope.reason` on a cause or effect | Real, but not something this program takes on |

The impact model is JSON first: a model file is JSON or YAML, read against one
JSON Schema generated from the Go types.

```yaml
context:
  problemStatement: Poor eating health
causes:
  - id: cause-busy
    statement: Too busy to cook
    level: immediate
  - id: cause-fast-food
    statement: Fast food is everywhere
    level: underlying
    parents: [cause-busy]
  - id: cause-poverty
    statement: Poverty
    level: structural
    outOfScope:
      reason: Beyond what this program can change.
effects:
  - id: eff-diabetes
    statement: Diabetes and heart disease
  - id: eff-costs
    statement: Higher health costs
    parents: [eff-diabetes]
outcomes:
  - id: oc-time-to-cook
    title: People have time to cook
    level: short_term
    addressesCauses: [cause-busy]
  - id: oc-eat-well
    title: People eat well
    level: impact
    addressesEffects: [eff-diabetes]
```

With element commands:

```bash
impact-model cause add impact-model.yaml --id cause-fast-food \
  --statement "Fast food is everywhere" --level underlying --parents cause-busy
impact-model effect add impact-model.yaml --id eff-costs \
  --statement "Higher health costs" --parents eff-diabetes
impact-model outcome set impact-model.yaml oc-eat-well --relieves eff-diabetes
```

## Causes and effects form a network

A cause can contribute to several others, and an effect can follow from several
others, so `parents` is a list. A tree is the case of at most one parent each.
Validation rejects a cause or effect that is its own parent, and a chain of
parents that leads back to where it started through any branch.

## Which way the links point

A cause's `parents` point from the deeper cause toward the problem: structural
causes feed underlying ones, underlying causes feed immediate ones, and
immediate causes lead to the core problem. Written the other way round, the
tree draws upside down, so `IMP-CAU-003` reports a cause whose parent sits at a
deeper level.

## How the diagram is laid out

The arrows flow upward, from the roots to the core problem and on to the
branches:

- The core problem is in the middle row.
- A cause with no parents is in the first row below it. Any other cause is one
  row below its lowest parent, so every arrow from a cause points up.
- An effect with no parents is in the first row above the core problem. Any
  other effect is one row above its highest parent.
- Within a row, boxes keep the order of the model file.
- A cause or effect set aside is drawn faded and labelled "(set aside)". A
  rejected one is left out.

The diagram is built first as a JSON document, a
[d2vision](https://github.com/grokify/d2vision) diagram spec with a zone for the
effects, the core problem, and the causes, a row for each layer, and the edges
between boxes. `problem-tree-json` prints that document, and D2 and Mermaid are
both rendered from it, so another tool can read or restyle the same tree. D2
keeps every row in place; Mermaid draws bottom to top and places the rows
itself.

## The tables

Below the diagram, `problem-tree` and `problem-tree-d2` list:

- each cause with its level, what it leads to, and the outcomes that address it
- each effect with what it follows from and the outcomes that relieve it

A dash means nothing addresses it yet, and "set aside" gives the reason.

## What lint checks

| Rule | Checks |
|------|--------|
| `IMP-CAU-001` | Each underlying and structural cause is addressed by an outcome or set aside |
| `IMP-CAU-002` | Outcomes address some root cause, not only immediate ones |
| `IMP-CAU-003` | A cause contributes only to causes at its own level or nearer the problem |
| `IMP-EFF-001` | Each effect is relieved by an outcome or set aside |
| `IMP-EFF-002` | Effects are relieved by long-term or impact outcomes, not only earlier ones |

## Known limitations

- Labels are not wrapped, so a long problem statement makes a wide box. Keep the
  core problem to a short phrase and put the detail in the context's narrative.
- Within a row, boxes keep file order, so arrows can cross. Reordering causes or
  effects in the file reorders the row.
