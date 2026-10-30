# Examples

Three complete examples are in the repository's
[`examples/`](https://github.com/ProductBuildersHQ/theoryofchange-spec/tree/main/examples)
directory. **Everything in them is fictional**: the organizations, people,
numbers, and URLs (all under `example.org`) are invented for illustration.

Each is valid, is in canonical form, and is checked by the test suite, and the two
reference theories lint with no errors, so you can copy one as a starting point.

## Membership network

`examples/intermediary`: *Riverbend Nonprofit Network*, a network of about 150
nonprofits and eight staff, with three program areas modelled as pathways:
Learning Exchange, Shared Services, and Civic Voice.

It shows two impact layers (member change and resident change) with the ceiling
of accountability between them; 23 outcomes across all five levels with
`all`, `any`, and `at_least` precondition logic; a metric library; measurement
design with weighted selection criteria and a collection budget; a problem tree
and target groups; audiences and a story codebook; and a spreadsheet import with
its column mapping. The folder holds the theory, a quarter of observations with
stories and a learning review, and a monthly CSV with its `mapping.yaml`.

```bash
theoryofchange status examples/intermediary/theory.yaml \
  -O examples/intermediary/observations-2026-q3.yaml
theoryofchange import observations --csv examples/intermediary/csv/monthly-q4.csv \
  --map examples/intermediary/csv/mapping.yaml --toc examples/intermediary/theory.yaml
```

## Market adoption

`examples/market-adoption`: a fictional open standard driven to adoption in a
market with two competing approaches, a proprietary single-vendor scheme and an
extension of an older authorization framework. It shows a theory about a market
rather than a program: adoption outcomes, the competition modelled as outcomes in
the `interest` sphere, and a shared actor registry. It uses the
Adopt-Adapt-Expand-Respond market-systems framing.

## Competing futures

`examples/competing-futures`: three theories that describe different futures of
one market and share one registry of actors. Validate them together:

```bash
theoryofchange validate examples/competing-futures/future-*.yaml \
  --registry examples/competing-futures/actors.yaml --futures
```

## Next

[CLI reference](cli.md)
