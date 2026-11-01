# Quality and Judging

Three layers check an impact model, from cheapest to most thorough. Each one assumes the
one before it passes.

| Layer | Question | Command | Needs a model? |
|-------|----------|---------|----------------|
| Validate | Is the document well formed? | `validate` | No |
| Lint | Is the impact model sound? | `lint` | No |
| Judge | Does an expert reader think it is good? | `judge` | A judge, human or agent |

## Validate

```bash
impact-model validate impact-model.yaml
```

Checks the schema and the structure: duplicate IDs, references that do not
resolve, invalid values, bad dates, group logic, and that nested groups share a
target. It also checks observation sets (`--model`), a child against its parent
(`--parent`), and competing futures (`--registry --futures`). Exit code 1 means
at least one file has errors.

## Lint

```bash
impact-model lint impact-model.yaml --profile standard
impact-model lint --list
```

Validation says a document is well formed; lint says whether the impact model is
sound. There are 72 deterministic rules, each with an ID such as `IMP-LOG-001`,
grouped by what they look at:

| Area | What the rules look for |
|------|-------------------------|
| Structure (`STR`) | At least one impact outcome; no cycles unless marked as feedback; every outcome leads to impact and has a precondition or an intervention; links run from lower levels to higher ones; precondition groups are complete |
| Logic (`LOG`) | Links into medium-term and higher outcomes say why they hold and name an assumption; impact outcomes have at least two preconditions; several incoming links say whether all, any, or some are needed |
| Assumptions and risks (`ASM`, `RSK`) | High and critical assumptions say how they will be tested, cite evidence, and have a related risk with a mitigation |
| Measurement (`MEA`, `MET`, `CLM`) | Every outcome within the ceiling is measured; every indicator has a baseline, target, source, and frequency; targets fall inside the time horizon; claims above the ceiling are contribution, not attribution; common metrics are measured in every pathway |
| Collection budget (`BUD`) | Selected indicators fit the staff-time and question budgets and say how data is collected |
| Accountability and responsibility (`ACC`, `RSP`, `ACT`, `INT`) | A ceiling is declared; outcomes above it are not owned; shared and partner outcomes name who delivers them; every intervention has an owner |
| The chain from activity to impact (`INT`) | Every intervention says what it produces, leads through the outcomes it targets to an impact, agrees with the links that name it, is measured, and states its resources |
| North Stars and headline lists (`NSM`, `HLN`, `AUD`, `CTX`) | A North Star measures an outcome at or below the ceiling, has a baseline, a target, and a guardrail, and belongs to its own pathway; a headline list is short, reported to its audience, and producible; an audience told about outcomes is also told numbers; the impact model states its mission or vision |
| Provenance and evidence (`PRV`, `EVD`) | Every element says where it comes from; stated elements cite a source; an approved impact model has no hypotheses among what the organization answers for |
| Equity, causes, audiences (`EQU`, `CAU`, `AUD`) | Outcomes for target groups are measured by those groups; each deeper cause is addressed or set aside with a reason, and no cause contributes to a deeper one; every audience is the target of something reported |
| Effects (`EFF`) | Each effect in the problem tree is relieved by an outcome or set aside with a reason, and effects are relieved by long-term or impact outcomes, not only earlier ones |
| Pathways and validation (`PTH`, `VAL`) | Every pathway reaches impact and has a selected indicator; the main steps have been checked with the people the impact model concerns |
| DMAIC readiness (`DMA`) | For a pathway marked `projections: [dmaic]`: every indicator says how often it is measured and where the data comes from, is measured at least 20 times a year, and, for a survey, gets at least 30 responses per measurement. Pathways not marked are not checked |

Profiles choose how much is reported:

| Profile | Reports |
|---------|---------|
| `draft` | Errors and structural rules. A small valid impact model is GO |
| `standard` | Also the warnings (the default) |
| `full` | Also the advice |

The result is GO unless a rule reports an error (exit code 1). `--disable` skips
rules by ID, and `--format json` prints a
[structured-evaluation](https://github.com/plexusone/structured-evaluation)
summary report. Rejected elements are ignored.

## Judge

A rubric is better than a checklist when the question needs judgment: is the
causal logic plausible, is the problem framed fairly, would the people it
concerns recognise themselves. The `judge` commands prepare an impact model for a judge
and combine the verdicts. **They make no model calls**: a human reviewer or an AI
agent reads the bundle and returns a verdict.

```bash
impact-model judge show impact-model.yaml --profile development --out bundle.json
impact-model judge record impact-model.yaml verdict.json --judge reviewer-a
impact-model judge record impact-model.yaml verdict2.json --judge reviewer-b
impact-model judge summary impact-model.yaml --aggregation majority
```

1. `show` prints the **bundle**: the rubric, the document, a narrative, the lint
   report, instructions, and a hash of all of them. Restricted evidence is
   redacted.
2. The judge returns a verdict, a structured-evaluation report that scores each
   rubric category and names its findings.
3. `record` checks the verdict against the bundle and stores it under
   `.toc/judgements/<hash>/<judge>.json`. It is rejected, with exit code 1, if it
   judged a different version of the impact model, is not a valid report, scores a
   category the rubric does not have, or leaves a required category unscored.
4. `summary` combines every verdict for the current version: per-category scores,
   how closely the judges agree, and whether the impact model passes.

Rubric profiles: `core` has 12 categories; `development` adds categories for
programs working on development outcomes. Aggregation is `majority` (a tie goes
to partial), `conservative` (the lowest score), `optimistic` (the highest), or
`unanimous` (anything else is partial). The impact model passes when every required
category passes, no blocking category fails, and the findings stay within the
rubric's limits.

!!! note "How well does the rubric discriminate?"
    The [rubric evaluation](../evals/rubric-discrimination.md) records two judges
    scoring a reference impact model and a deliberately weak one. The weak impact model fails
    most categories and the reference has no high or critical findings.

## Next

[MCP server and agent skills](mcp-and-skills.md)
