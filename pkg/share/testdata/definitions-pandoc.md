---
title: "Open agent authorization standard adoption"
version: "0.1"
date: "2026-09-30"
model-id: "open-agent-auth"
content-hash: "6c01012150da0627ca84b9b55e045a952149af0a1ea2f8ee02727bfeed5866c1"
---

## Causal pathways

### The theory of change on one page

![Diagram](diagrams/d1.png)

### Supply side: providers and platforms

![Diagram](diagrams/d2.png)

\newpage

## Logic model: Supply side: providers and platforms

Getting identity providers and agent platforms to ship the standard.

| Stage | Elements |
|-------|----------|
| Inputs | Two engineers and an event budget |
| Activities | Run interoperability events; Publish reference implementations |
| Outputs | Interop test suite; Event reports; Open reference implementations are available |
| Short-term outcomes | A stable, interoperable specification exists |
| Medium-term outcomes | Major identity providers support the standard [inferred]; Agent platforms support the standard |
| Long-term outcomes | Enterprises require the standard in procurement [hypothesis] |
| Impact | Agent access across the market is scoped and auditable |

\newpage

## Outcome map: Supply side: providers and platforms

_This is a view of the causal model. It flattens the links between outcomes; the pathway diagram shows them._

### Impact

| Outcome | Sphere | Category | Change | Responsibility | Delivered by | Addresses | Measured by |
|---------|--------|----------|--------|----------------|--------------|-----------|-------------|
| Agent access across the market is scoped and auditable (`oc-safer-agents`) | Interest | respond | — | Own · above the ceiling, monitored only | — | — | ind-adoption-stories |

### Long-term outcome

| Outcome | Sphere | Category | Change | Responsibility | Delivered by | Addresses | Measured by |
|---------|--------|----------|--------|----------------|--------------|-----------|-------------|
| Enterprises require the standard in procurement (`oc-enterprise-require`) [hypothesis] | Influence | expand | Policy | Own | — | — | — |

### Medium-term outcome

| Outcome | Sphere | Category | Change | Responsibility | Delivered by | Addresses | Measured by |
|---------|--------|----------|--------|----------------|--------------|-----------|-------------|
| Major identity providers support the standard (`oc-idp-support`) [inferred] | Influence | adopt | Adoption | Shared | Standards working group, Identity providers | cause-no-demand-signal | ind-idp-count, ind-idp-count-na, ind-idp-count-eu, ind-idp-total |
| Agent platforms support the standard (`oc-platform-support`) | Influence | adapt | Adoption | Partner · monitored only | Identity providers | — | — |

### Short-term outcome

| Outcome | Sphere | Category | Change | Responsibility | Delivered by | Addresses | Measured by |
|---------|--------|----------|--------|----------------|--------------|-----------|-------------|
| A stable, interoperable specification exists (`oc-spec-stable`) | Control | — | Condition | Own | — | — | — |

### Output

| Outcome | Sphere | Category | Change | Responsibility | Delivered by | Addresses | Measured by |
|---------|--------|----------|--------|----------------|--------------|-----------|-------------|
| Open reference implementations are available (`oc-reference-impl`) | Control | — | — | Own | — | — | — |

\newpage

## Logframe: Supply side: providers and platforms

_This is a view of the causal model. It flattens the links between outcomes; the pathway diagram shows them._

| Level | Narrative | Indicators and targets | Means of verification | Assumptions |
|-------|-----------|------------------------|-----------------------|-------------|
| Goal | Agent access across the market is scoped and auditable (`oc-safer-agents`) | Stories of provider adoption (`ind-adoption-stories`) | not stated | — |
| Purpose | Enterprises require the standard in procurement (`oc-enterprise-require`) [hypothesis] | — | — | — |
| Purpose | Major identity providers support the standard (`oc-idp-support`) [inferred] | Identity providers with generally available support (`ind-idp-count`): 2 providers (by 2026-12-31) — First movers; 5 providers (by 2027-12-31) (Covers half the market by seats); Providers with GA support (North America) (`ind-idp-count-na`); Providers with GA support (Europe) (`ind-idp-count-eu`); Providers with GA support (total) (`ind-idp-total`) | Provider release notes; Desk review; P3M; not stated; not stated; not stated | Identity providers see commercial advantage in supporting agent delegation. (`asm-idp-incentive`) |
| Purpose | Agent platforms support the standard (`oc-platform-support`) | — | — | — |
| Outputs | A stable, interoperable specification exists (`oc-spec-stable`) | — | — | — |
| Outputs | Open reference implementations are available (`oc-reference-impl`) | — | — | — |
| Activities | Run interoperability events (`int-interop`) | — | — | — |
| Activities | Publish reference implementations (`int-reference`) | Reference implementation downloads (`ind-reference-downloads`) | not stated | — |

\newpage

## Metric definitions

| Metric | Definition | Formula | Unit | Direction | Combines across pathways | Layer | Used by |
|--------|------------|---------|------|-----------|--------------------------|-------|---------|
| Providers with GA support | Identity providers shipping the standard in a general-availability release. | Count of distinct providers with a GA release implementing the standard. | providers | Increase | Sum | Common | Identity providers with generally available support; Providers with GA support (North America); Providers with GA support (Europe); Providers with GA support (total) |
| Share of provider seats covered | Share of provider seats that can use the standard. | Seats at providers with GA support divided by seats at all tracked providers, weighted by seats. | percent | Increase | Mean weighted by Providers with GA support | Pathway | — |

\newpage

## Glossary

| Term | Definition | Examples |
|------|------------|----------|
| General availability | A release available to all customers without special arrangement, as opposed to a preview. | A vendor's public release notes list the feature without a preview label. |


*Impact model open-agent-auth · version 0.1 · hash 6c01012150da · as of 2026-09-30*
