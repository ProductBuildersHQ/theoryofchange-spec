# Authoring a Theory

A theory is one document of connected elements. This guide describes the
elements, how they refer to each other, and the ways to change them.

## The elements

| Element | What it is |
|---------|------------|
| **Outcome** | A change that should happen, at one of five levels: `output`, `short_term`, `medium_term`, `long_term`, or `impact` |
| **Link** | A causal arrow from one outcome to another, with a rationale and a polarity. A feedback link marks a loop |
| **Group** | Preconditions for an outcome: all of them, any of them, or at least *k* of *n*. Groups can nest |
| **Intervention** | Something the program does to bring about outcomes |
| **Assumption** | Something that must hold for a link or outcome to work, with a criticality and a status |
| **Indicator** | How an outcome or assumption is measured, with a baseline, targets, and a frequency |
| **Actor** | A person or organization who changes, delivers, or co-owns an outcome |
| **Evidence** | A source that supports or contradicts an element, with a sensitivity |
| **Pathway** | A named route through the theory, such as one program's contribution |
| **Cause**, **target group** | The problem tree and the groups affected, for equity and root-cause analysis |
| **Audience**, **theme** | Who needs to see what reported, and the codebook for coding stories |
| **Metric**, **term** | A shared metric library, and a glossary of the words the theory uses |
| **Risk**, **decision**, **factor**, **consultation** | What could go wrong, what is awaiting a decision, external factors, and the review trail |

An outcome also records its **responsibility**: `own`, `shared`, `partner`, or
`unaddressed`, so a report can separate what the organization answers for from
what it only monitors. The document names a **ceiling of accountability**,
the level above which outcomes are tracked but not claimed.

Every ID is a lowercase kebab-case slug, unique across the document, and
references are by ID.

## Provenance: stated, inferred, or confirmed

Each element carries a **basis** that says why it is in the theory:

| Basis | Meaning |
|-------|---------|
| `stated` | A source says it. Cite the evidence |
| `inferred` | Our reading of the sources |
| `hypothesis` | A guess to test; the first thing to challenge in review |
| `confirmed` | The people it concerns agreed |
| `rejected` | Considered and set aside; kept on record, left out of diagrams and analysis |

Record who decided a confirmed or rejected basis with `--decided-by` as a
**group or role**, not a personal name, and when with `--decided-at`.

## Change a theory with commands

Every element kind has `add`, `set`, and `rm`:

```bash
theoryofchange outcome add theory.yaml --id oc-skills --title "Staff write strong applications" --level short_term
theoryofchange link add theory.yaml --id ln-skills-adoption --from oc-skills --to oc-adoption --rationale "Skilled staff apply more often"
theoryofchange outcome set theory.yaml oc-skills --basis confirmed --decided-by "Program board" --decided-at 2026-04-01
theoryofchange outcome rm theory.yaml oc-skills --cascade
```

Run `theoryofchange <element> add --help` to see every field as a flag.

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

## Change a theory with a patch

For many changes at once, or from a script or an agent, apply an
[RFC 6902](https://www.rfc-editor.org/rfc/rfc6902) JSON Patch:

```bash
echo '[{"op":"replace","path":"/outcomes/@oc-skills/title","value":"Staff write strong applications"}]' \
  | theoryofchange patch theory.yaml -
```

Pointers can address an array element by ID: `/outcomes/@oc-skills/title`
selects the outcome with ID `oc-skills`, and `/outcomes/-` appends. The patch is
applied to a copy, and all of the guarantees above apply.

## Keep a revision log

When a theory changes after review, record why. `theoryofchange diff` compares
two versions by element ID, so a retitled outcome is one changed element, and
the same theory in YAML and JSON has no differences:

```bash
theoryofchange diff theory-v1.yaml theory-v2.yaml
theoryofchange diff theory-v1.yaml theory-v2.yaml --revision --version 0.2 --date 2026-03-01 \
  > revision.patch.json
theoryofchange patch theory-v2.yaml revision.patch.json
```

`--revision` prints a patch that appends a draft revision listing every changed
element ID. Review it, then apply it.

## Theories that belong together

- **Parent and child.** A program's theory can contribute to an organization's.
  `validate program.yaml --parent organization.yaml` checks that the
  `contributesTo` references resolve in the parent.
- **Competing futures.** Theories describing different futures of one market
  share one actor registry. `validate a.yaml b.yaml c.yaml --registry actors.yaml --futures`
  checks that they name their futures and perspectives, do not both speak for the
  same actor in the same future, and refer to each other's elements correctly.
  See the `competing-futures` [example](examples.md).

## Schemas for editors and agents

The JSON Schemas are generated from the Go types and embedded in the binary:

```bash
theoryofchange schema                                   # list them
theoryofchange schema theory-of-change > theory-of-change.schema.json
```

Seven schemas are available: `theory-of-change`, `observation-set`,
`theory-of-change-status`, `status-diff`, `actor-registry`, `review-response`,
and `csv-mapping`.

## Next

[Measuring progress](measurement.md) · [Quality and judging](quality-and-judging.md)
