---
title: "Open agent authorization standard adoption"
version: "0.1"
date: "2026-09-30"
toc-id: "open-agent-auth"
content-hash: "aa5ec2912e5f9c2c8ff2422d65c3f638f7548b50df4846de3807dcdb380de6bd"
---

## Causal pathways

### The theory on one page

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
| Intermediate outcomes | Major identity providers support the standard [inferred]; Agent platforms support the standard |
| Long-term outcomes | Enterprises require the standard in procurement [hypothesis] |
| Impact | Agent access across the market is scoped and auditable |

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


*Theory of Change open-agent-auth · version 0.1 · hash aa5ec2912e5f · as of 2026-09-30*
