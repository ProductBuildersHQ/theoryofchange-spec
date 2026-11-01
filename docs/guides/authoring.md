# Authoring an Impact Model

An impact model is one document of connected elements. This guide describes the
elements, how they refer to each other, and the ways to change them.

## The elements

| Element | What it is |
|---------|------------|
| **Outcome** | A change that should happen, at one of five levels: `output`, `short_term`, `medium_term`, `long_term`, or `impact` |
| **Link** | A causal arrow from one outcome to another, with a rationale and a polarity. A feedback link marks a loop |
| **Group** | Preconditions for an outcome: all of them, any of them, or at least *k* of *n*. Groups can nest |
| **Intervention** | Something the program does to bring about outcomes |
| **Assumption** | Something that must hold for a link or outcome to work, with a criticality and a status |
| **Indicator** | How an outcome, activity, or assumption is measured, with a baseline, targets, and a frequency. It can also be worked out from other indicators, such as a cost per outcome |
| **Actor** | A person or organization who changes, delivers, or co-owns an outcome |
| **Evidence** | A source that supports or contradicts an element, with a sensitivity |
| **Pathway** | A named route through the impact model, such as one program's or product's contribution, with the North Star it steers by |
| **Cause**, **target group** | The problem tree and the groups affected, for equity and root-cause analysis |
| **Audience**, **theme** | Who needs to see what reported, the short list of numbers each is told, and the codebook for coding stories |
| **Metric**, **term** | A shared metric library, and a glossary of the words the impact model uses |
| **Risk**, **decision**, **factor**, **consultation** | What could go wrong, what is awaiting a decision, external factors, and the review trail |

An outcome also records its **responsibility**: `own`, `shared`, `partner`, or
`unaddressed`, so a report can separate what the organization answers for from
what it only monitors. The document names a **ceiling of accountability**,
the level above which outcomes are tracked but not claimed.

Every ID is a lowercase kebab-case slug, unique across the document, and
references are by ID.

## What a program steers by, and what each audience is told

Three choices sit on top of the chain from activities to impact. They are
different things, and an impact model can make all three.

- **A North Star** is the one number a program or product steers by. It is set on
  the pathway, with an optional *one metric that matters* and the guardrails that
  watch what chasing it could harm. It should measure an outcome at or below the
  ceiling of accountability, never an output or an activity.
- **A headline list** is what an audience is told: a short, ordered list on the
  audience, each entry an indicator with an optional label and note worded for
  that reader. The board, funders, and members can each lead with different
  numbers drawn from the same indicators, and every audience sees the same figure
  for an indicator.
- **Purpose** is `context.mission`, `context.vision`, and `context.values`:
  why the organization exists, what it works toward, and the principles that
  guide how it works, most important first. They are not part of the causal
  chain. A V2MOM needs the vision and at least one value.

```bash
# A pathway's North Star, and what watches it
impact-model pathway set impact-model.yaml pw-learning \
  --north-star ind-practice-maturity --guardrails ind-retention-learning

# What the board is told, in order
impact-model audience set impact-model.yaml aud-board --set 'headline=[
  {"indicatorId":"ind-practice-maturity","label":"Leadership practice across members"},
  {"indicatorId":"ind-reach-network"}]'

# Why the organization exists
impact-model patch impact-model.yaml - <<'EOF'
[{"op":"add","path":"/context/mission","value":"Strengthen local nonprofits."}]
EOF
```

Some of the numbers people ask for are arithmetic on others, such as the cost
per applied change. Give that indicator a `composedOf`, and its value is worked
out from the latest value of each part when it has none of its own:

```bash
impact-model indicator set impact-model.yaml ind-cost-per-change \
  --composed-op ratio --composed-of ind-cost,ind-changes-applied
```

A ratio and a difference take exactly two indicators, in order. A composition is
arithmetic and not a causal link. See [Measuring progress](measurement.md) for
how the value is worked out and what is withheld, and
[Quality and judging](quality-and-judging.md) for the rules that check all of it.

## Provenance: stated, inferred, or confirmed

Each element carries a **basis** that says why it is in the impact model:

| Basis | Meaning |
|-------|---------|
| `stated` | A source says it. Cite the evidence |
| `inferred` | Our reading of the sources |
| `hypothesis` | A guess to test; the first thing to challenge in review |
| `confirmed` | The people it concerns agreed |
| `rejected` | Considered and set aside; kept on record, left out of diagrams and analysis |

Record who decided a confirmed or rejected basis with `--decided-by` as a
**group or role**, not a personal name, and when with `--decided-at`.

## Change an impact model with commands

Every element kind has `add`, `set`, and `rm`:

```bash
impact-model outcome add impact-model.yaml --id oc-skills --title "Staff write strong applications" --level short_term
impact-model link add impact-model.yaml --id ln-skills-adoption --from oc-skills --to oc-adoption --rationale "Skilled staff apply more often"
impact-model outcome set impact-model.yaml oc-skills --basis confirmed --decided-by "Program board" --decided-at 2026-04-01
impact-model outcome rm impact-model.yaml oc-skills --cascade
```

Run `impact-model <element> add --help` to see every field as a flag.

Each command is compiled to a JSON Patch and applied through the same validated
path, which has these guarantees:

- **Atomic.** If any step fails, nothing is written.
- **Never worse.** A change is rejected if the result would have more structural
  errors than before. `--allow-errors` accepts structural errors but never a
  schema violation.
- **Canonical.** The file is rewritten in canonical form in its own format.
- **Previewable.** `--dry-run` prints the resulting document and writes nothing.
- **Honest removal.** Removing an element that others refer to is rejected;
  `rm --cascade` removes the references as well and tells you what it removed.

## Change an impact model with a patch

For many changes at once, or from a script or an agent, apply an
[RFC 6902](https://www.rfc-editor.org/rfc/rfc6902) JSON Patch:

```bash
echo '[{"op":"replace","path":"/outcomes/@oc-skills/title","value":"Staff write strong applications"}]' \
  | impact-model patch impact-model.yaml -
```

Pointers can address an array element by ID: `/outcomes/@oc-skills/title`
selects the outcome with ID `oc-skills`, and `/outcomes/-` appends. The patch is
applied to a copy, and all of the guarantees above apply.

## Keep a revision log

When an impact model changes after review, record why. `impact-model diff` compares
two versions by element ID, so a retitled outcome is one changed element, and
the same impact model in YAML and JSON has no differences:

```bash
impact-model diff impact-model-v1.yaml impact-model-v2.yaml
impact-model diff impact-model-v1.yaml impact-model-v2.yaml --revision --version 0.2 --date 2026-03-01 \
  > revision.patch.json
impact-model patch impact-model-v2.yaml revision.patch.json
```

`--revision` prints a patch that appends a draft revision listing every changed
element ID. Review it, then apply it.

## Impact models that belong together

- **Parent and child.** A program's impact model can contribute to an organization's.
  `validate program.yaml --parent organization.yaml` checks that the
  `contributesTo` references resolve in the parent.
- **Competing futures.** Impact models describing different futures of one market
  share one actor registry. `validate a.yaml b.yaml c.yaml --registry actors.yaml --futures`
  checks that they name their futures and perspectives, do not both speak for the
  same actor in the same future, and refer to each other's elements correctly.
  See the `competing-futures` [example](examples.md).

## Schemas for editors and agents

The JSON Schemas are generated from the Go types and embedded in the binary:

```bash
impact-model schema                                   # list them
impact-model schema impact-model > impact-model.schema.json
```

Seven schemas are available: `theory-of-change`, `observation-set`,
`impact-model-status`, `status-diff`, `actor-registry`, `review-response`,
and `csv-mapping`.

## Next

[Measuring progress](measurement.md) · [Quality and judging](quality-and-judging.md)
