# Model feedback from authoring the reference examples

Two fictional reference impact models (`examples/intermediary` and
`examples/market-adoption`) were written against the model to find what it
could not say, or said awkwardly. This page records each finding and what was
decided, so the reasoning outlives the conversation that produced it.

Decisions are **done** (implemented and tested), **documented** (a convention,
now written down), **planned** (assigned to a roadmap item), or **declined**
(with the reason).

## Done

| Finding | Decision |
|---------|----------|
| A network-wide roll-up or common-layer indicator has no single outcome to measure, so it was attached to an arbitrary one | An indicator's `measures` may name a **pathway**. Validation accepts it, and a pathway-filtered view includes pathway-level indicators only for that pathway |
| Ordinal levels cannot be summed or averaged, so a "median level" lived only in formula text | Added `median` and `mode` aggregations |
| No actor type for a government office; no external-factor kind for an opportunity (risks already had one) | Added the `government` actor type and the `opportunity` factor kind |
| What may a restricted evidence record contain, and what do renderers show? | A restricted record (or one with no sensitivity set) may carry a title, type, strength, and source. Its excerpt, summary, **and URL** are withheld by every renderer, since a URL can reveal where a confidential document lives |
| A group's `target` is one outcome; nothing said what a nested group's target is | A nested group targets the **same outcome** as the group containing it. Validation enforces it (`VAL-GRP-004`) |
| The coverage rule wanted an indicator on every owned outcome, which inflates indicator counts | `IMP-MEA-001` accepts coverage through another indicator's `proxyFor` |

## Documented

| Finding | Decision |
|---------|----------|
| `direction: increase` is vague for a qualitative indicator whose rise is bad news | `direction` states the **desired** direction. If a rising level is bad news, the direction is `decrease`; say why in the indicator's definition |
| Negative-polarity links and precondition logic | A negative link is an **inhibitor**, not a precondition: it does not count toward `all`, `any`, or `at_least` logic or toward `k`. How an inhibitor lowers a pathway's status is decided with the roll-up algorithm (`RMI-THEORYOFCHANGESPEC-018`) |
| A competitor's momentum has no element type | Model it as an outcome in the `interest` sphere with category `respond`, linked to the outcomes it inhibits. A first-class competitor-response element is deferred until a second example needs it |
| Capability relation `dependsOn` is camelCase while every other enum is snake_case | It comes from prism-core's shared `CapabilityRef` and is intentionally left alone so the PRISM modules stay interchangeable |
| In YAML flow mappings (`{a: b, c: d}`), an unquoted comma inside a value silently splits the mapping | Schema validation catches it. The authoring guide recommends block style, which `fmt` always produces |
| `provenance.decidedBy` is free text, so nothing stops a personal name | By convention it names a **group or role**. Detecting personal names is not something code can do reliably; the authoring guide and the review checklist say so |
| Responsibility (own, shared, partner) and a problem tree are not modelled | Both are in `RMI-THEORYOFCHANGESPEC-044`. Until then the ceiling is a level and root causes are text |

## Planned

| Finding | Roadmap item |
|---------|--------------|
| A `weighted_mean` roll-up cannot be computed if the weight metric has no indicator | Lint rule `IMP-MET-003`, in `RMI-THEORYOFCHANGESPEC-014` |
| `pathway.outcomeIds` and indicators' `pathwayIds` are two lists kept consistent by hand | Lint rule `IMP-PTH-003`, in `RMI-THEORYOFCHANGESPEC-014` |
| A group's members should each have a link into the group's target | Already covered by `IMP-STR-007`, in `RMI-THEORYOFCHANGESPEC-013` |
| Assumption `appliesTo` and a link's `assumptionIds` can disagree | A consistency lint rule, in `RMI-THEORYOFCHANGESPEC-014` |

## Declined (for now)

| Finding | Reason |
|---------|--------|
| A `window` field on metrics ("trailing twelve months") | Useful, but it changes how observations and targets are compared, so it deserves its own design. It is recorded for a later minor version; until then the window is stated in the formula |
| One indicator measuring both an outcome and an assumption | Two indicators are clearer and cost little. Revisit if authors keep hitting it |
| A "key outcome" or "monitored" flag on outcomes | `proxyFor` already expresses "this indicator also stands in for that outcome" without a second mechanism |
| A `layer` label on outcomes for two-layer impact | Layer one versus layer two is already visible through `sphere`, the indicator `claim`, and the ceiling of accountability. A fourth signal would add overlap, not information |
| `changeType` has no `standardization` | That value describes what an intervention does, not the kind of change an outcome describes; the outcome's `adoption` or `policy` change type fits |
