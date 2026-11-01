# Integration contract for USM and UFM

This page says what the impact model module promises to the Unified
Strategy Model (USM) and the Unified Futures Model (UFM), and what it asks of
them. It is a contract in the plain sense: each side can build against it
without reading the other's code.

**Status: draft, for review by the owners of USM and UFM.** Nothing here is
final until they have read it.

## What each model uses an impact model for

- **USM** describes a strategy from one market participant's point of view,
  combining a theory of change with a balanced scorecard. It reads a
  Impact model's outcomes, indicators, status, and confidence, and attaches its own
  scorecard measures to elements of the impact model.
- **UFM** describes competing futures of a market, such as one scheme winning
  over another. It groups impact models by the future each one is working toward
  and compares them.

The impact model is the shared part: neither model owns it, and it owns
neither of them.

## The promises

### 1. Every impact model and element has a stable address

An impact model is `model:<doc>`; an element is `model:<doc>#<element>`. Both parts are
slugs: lower case letters, digits, and single hyphens, starting with a letter,
at most 128 characters. An ID is permanent for the life of the element:

- the tools never change an element's ID (`impact-model outcome set` refuses
  to), and `impact-model patch` addresses elements by ID, not position;
- renaming means removing the element and adding a new one, which shows as
  such in `impact-model diff`;
- a removed or rejected element keeps its place in history, and its ID should not be reused for something else.

USM and UFM should store these references, not copies of the content.

### 2. An impact model says whose it is and which future it belongs to

```yaml
perspective:
  actorId: act-open-group          # the market participant (USM's participant)
  role: proponent                  # proponent | funder | implementer | observer
  futureRef: ufm:open-standard-wins   # the future (UFM's future)
```

`futureRef` has the form `ufm:<slug>`. UFM owns what a future *is*; the impact model
only points at it. Two impact models may belong to the same future if they speak for
different actors; two impact models may not speak for the same actor in the same
future (`impact-model validate --futures` reports it).

### 3. Impact models in one market share actors through a registry

An `ActorRegistry` document holds the actors several impact models refer to. A
Impact model names it in `actorRegistry`, and actor IDs not defined in the impact model
resolve there. The tools check that competing futures name one registry
(`validate --registry actors.yaml --futures`).

USM and UFM may read the registry to learn who the market's actors are. They
should not add actors to a registry belonging to an impact model they do not own;
add them to their own and have the impact model reference both.

### 4. Strategy data attaches by alignment, and the impact model never holds it

An impact model may carry `alignments`: an element ID and an external target
(`framework`, `ref`, `relation`). Frameworks include `bsc`, `usm`, `ufm`, `okr`,
`v2mom`, and `market_spec`. This is how a scorecard objective or measure points
at an outcome or indicator.

The impact model never stores scorecard data. USM computes its scorecard from the
Impact model's status and its own measures. The impact model does not list the models that
consume it: references run from USM and UFM into the impact model.

### 5. Status is a document

`impact-model status` produces a `ImpactModelStatus` document (schema
`impact-model-status`) that anyone can read without the Go module. For each
indicator, outcome, link, and assumption it gives a state; for each outcome,
a pathway state and a confidence from 0 to 1; and, per impact outcome, the
headline. The snapshot records the settings it was computed with, so any
number can be traced, and the same inputs always give the same bytes.

USM should read states and confidence from the snapshot instead of recomputing
them. A change in how status is computed is a change to the module's version
and is listed in its changelog.

### 6. Comparing futures

Competing-futures impact models compare without any schema change: see
`examples/competing-futures`. A UFM reading three impact models can line up their
impact outcomes (each is `level: impact`), their status and confidence, and
the assumptions each depends on. An impact model may name another impact model's element in
an alignment or an indicator's external reference; `validate --futures` checks
that such a reference resolves.

## Versioning

- Documents carry `specVersion` (currently `0.1`). A change that makes a valid
  document invalid, or renames a field or an enumeration value, raises it.
- The JSON Schemas are published under `schema/v0/` and are generated from the
  Go types, which are the source of truth.
- Additive changes (a new optional field, a new element kind) do not raise the
  version. Readers should ignore fields they do not know.
- The reference grammar and the `futureRef` form are part of this contract and
  change only with a major version.

## How a model reads an impact model

Any of these, in order of preference for a program:

1. The Go packages: `pkg/impact` (types), `pkg/impactio` (read and write), `pkg/measure`
   (status), `pkg/report` (views).
2. The command line, with `--format json` where offered.
3. The MCP server (`impact-model mcp`), which exposes validate, lint, graph, status,
   report, patch, deck, judge bundle, and diff to an agent.
4. The files and schemas alone.

## What the module does not do

- It does not model a market, a scorecard, or a future. Those belong to USM
  and UFM.
- It does not call them, and it does not depend on them.
- It does not decide which future is more likely; it states, for each, what has
  to be true and how well that is supported.

## Open questions for the owners

1. Does USM want the impact model's `alignments` to be one-way (USM points at the
   Impact model) only, or may an impact model point at a scorecard objective, as `bsc`
   alignments do today?
2. Is `ufm:<slug>` the right form for a future's reference, or should UFM
   publish its own reference grammar for the module to adopt?
3. Should confidence be offered to UFM as a number, as now, or only as
   high, medium, and low?
4. Which of the status snapshot's fields does USM need that are not there?
