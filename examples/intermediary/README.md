# Intermediary example: Riverbend Nonprofit Network

**Everything here is fictional.** The organization, county, people, numbers,
and URLs (all under `example.org`) are invented for illustration. No real
organization, funder, or program is described. A membership network of about 150 nonprofits and 8 staff, with three program
areas modeled as pathways: Learning Exchange, Shared Services, and Civic Voice.

Files:

- `impact-model.yaml` - the Theory of Change (id `riverbend-network`)
- `observations-2026-q3.yaml` - one period of measurements and assessments

## Features demonstrated

1. Two impact layers: member change (influence, `attribution`) and resident
   change (interest, `contribution`), with the ceiling of accountability at
   `long_term`.
2. Problem context: statement, root causes, population, geography, horizon.
3. 23 outcomes across all five levels; three pathways converging on shared
   long-term outcomes; AND, `any`, and `at_least` (2 of 3) precondition logic.
4. Causal links with rationales and assumptions, plus a reinforcing feedback loop.
5. Interventions with targets, outputs, owners, and statuses.
6. Assumptions of all six types, with tests and monitoring indicators.
7. Actors including the perspective actor, member segments, and partners.
8. A metric library, including a distinct-count metric that cannot be summed
   (`aggregation: none`) and a `weighted_mean` retention rate.
9. Indicators: a roll-up, qualitative levels, caveats, and every data availability.
10. Glossary of Riverbend-specific terms.
11. Evidence and provenance using all five bases, including rejected elements.
12. Risks, external factors, decisions, consultations, revisions, alignments.
13. Measurement design: weighted selection criteria, 11 selected indicators
    (plus candidates and a retired one) with collection cost, and a budget they
    fit inside.
14. Change analysis: a problem tree across three levels with one cause set
    aside, target groups, shared and partner-delivered outcomes, and priority
    scores.
15. Audiences, a story codebook, a theme-count indicator, and an observation
    set with stories (granted, not granted, and unknown consent) and a learning
    review.

## Check it

```bash
impact-model validate impact-model.yaml
impact-model validate observations-2026-q3.yaml --model impact-model.yaml
impact-model graph impact-model.yaml
impact-model fmt --check impact-model.yaml observations-2026-q3.yaml
```
