# Examples

Four complete examples are in the repository's
[`examples/`](https://github.com/ProductBuildersHQ/impact-model/tree/main/examples)
directory. **Everything in them is fictional**: the organizations, people,
numbers, and URLs (all under `example.org`) are invented for illustration.

Each is valid, is in canonical form, and is checked by the test suite, and the two
reference impact models lint with no errors, so you can copy one as a starting point.

## Membership network

`examples/intermediary`: *Riverbend Nonprofit Network*, a network of about 150
nonprofits and eight staff, with three program areas modelled as pathways:
Learning Exchange, Shared Services, and Civic Voice.

It shows two impact layers (member change and resident change) with the ceiling
of accountability between them; 23 outcomes across all five levels with
`all`, `any`, and `at_least` precondition logic; a metric library; measurement
design with weighted selection criteria and a collection budget; a problem tree
and target groups; audiences and a story codebook; a mission and vision, a North
Star with a guardrail for each program area, and headline lists that lead the
board, funders, and members with different numbers; and a spreadsheet import with
its column mapping. The folder holds the impact model, a quarter of observations with
stories and a learning review, and a monthly CSV with its `mapping.yaml`.

```bash
impact-model status examples/intermediary/impact-model.yaml \
  -O examples/intermediary/observations-2026-q3.yaml
impact-model import observations --csv examples/intermediary/csv/monthly-q4.csv \
  --map examples/intermediary/csv/mapping.yaml --model examples/intermediary/impact-model.yaml
```

## Market adoption

`examples/market-adoption`: a fictional open standard driven to adoption in a
market with two competing approaches, a proprietary single-vendor scheme and an
extension of an older authorization framework. It shows an impact model about a market
rather than a program: adoption outcomes, the competition modelled as outcomes in
the `interest` sphere, and a shared actor registry. It uses the
Adopt-Adapt-Expand-Respond market-systems framing.

## Competing futures

`examples/competing-futures`: three impact models that describe different futures of
one market and share one registry of actors. Validate them together:

```bash
impact-model validate examples/competing-futures/future-*.yaml \
  --registry examples/competing-futures/actors.yaml --futures
```

## Course catalog

`examples/course-catalog`: a fictional small nonprofit's free, on-demand
certificate courses for staff of other nonprofits. Its one program is marked
`projections: [dmaic]` and is measured weekly: completion rate, courses started,
hours to a first forum answer, and a post-course Net Promoter Score from a survey
large enough for weekly scores. Twenty-four weeks of observations give each of
those enough points for control limits. A monthly catalog count and a yearly
graduate survey are flagged by the DMAIC lint rules on purpose.

```bash
impact-model lint examples/course-catalog/impact-model.yaml
impact-model export dmaic examples/course-catalog/impact-model.yaml \
  -O examples/course-catalog/observations-2026-h1.yaml
```

See the [DMAIC guide](dmaic.md).

## Next

[CLI reference](cli.md)
