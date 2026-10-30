# Quality and Judging

Three layers check a theory, from cheapest to most thorough. Each one assumes the
one before it passes.

| Layer | Question | Command | Needs a model? |
|-------|----------|---------|----------------|
| Validate | Is the document well formed? | `validate` | No |
| Lint | Is the theory sound? | `lint` | No |
| Judge | Does an expert reader think it is good? | `judge` | A judge, human or agent |

## Validate

```bash
theoryofchange validate theory.yaml
```

Checks the schema and the structure: duplicate IDs, references that do not
resolve, invalid values, bad dates, group logic, and that nested groups share a
target. It also checks observation sets (`--toc`), a child against its parent
(`--parent`), and competing futures (`--registry --futures`). Exit code 1 means
at least one file has errors.

## Lint

```bash
theoryofchange lint theory.yaml --profile standard
theoryofchange lint --list
```

Validation says a document is well formed; lint says whether the theory is
sound. There are 48 deterministic rules, each with an ID such as `TOC-LOG-001`,
grouped by what they look at:

| Area | What the rules look for |
|------|-------------------------|
| Structure (`STR`) | At least one impact outcome; no cycles unless marked as feedback; every outcome leads to impact and has a precondition or an intervention; links run from lower levels to higher ones; precondition groups are complete |
| Logic (`LOG`) | Links into intermediate and higher outcomes say why they hold and name an assumption; impact outcomes have at least two preconditions; several incoming links say whether all, any, or some are needed |
| Assumptions and risks (`ASM`, `RSK`) | High and critical assumptions say how they will be tested, cite evidence, and have a related risk with a mitigation |
| Measurement (`MEA`, `MET`, `CLM`) | Every outcome within the ceiling is measured; every indicator has a baseline, target, source, and frequency; targets fall inside the time horizon; claims above the ceiling are contribution, not attribution; common metrics are measured in every pathway |
| Collection budget (`BUD`) | Selected indicators fit the staff-time and question budgets and say how data is collected |
| Accountability and responsibility (`ACC`, `RSP`, `ACT`, `INT`) | A ceiling is declared; outcomes above it are not owned; shared and partner outcomes name who delivers them; every intervention has an owner |
| Provenance and evidence (`PRV`, `EVD`) | Every element says where it comes from; stated elements cite a source; an approved theory has no hypotheses among what the organization answers for |
| Equity, causes, audiences (`EQU`, `CAU`, `AUD`) | Outcomes for target groups are measured by those groups; each deeper cause is addressed or set aside with a reason; every audience is the target of something reported |
| Pathways and validation (`PTH`, `VAL`) | Every pathway reaches impact and has a selected indicator; the main steps have been checked with the people the theory concerns |

Profiles choose how much is reported:

| Profile | Reports |
|---------|---------|
| `draft` | Errors and structural rules. A small valid theory is GO |
| `standard` | Also the warnings (the default) |
| `full` | Also the advice |

The result is GO unless a rule reports an error (exit code 1). `--disable` skips
rules by ID, and `--format json` prints a
[structured-evaluation](https://github.com/plexusone/structured-evaluation)
summary report. Rejected elements are ignored.

## Judge

A rubric is better than a checklist when the question needs judgment: is the
causal logic plausible, is the problem framed fairly, would the people it
concerns recognise themselves. The `judge` commands prepare a theory for a judge
and combine the verdicts. **They make no model calls**: a human reviewer or an AI
agent reads the bundle and returns a verdict.

```bash
theoryofchange judge show theory.yaml --profile development --out bundle.json
theoryofchange judge record theory.yaml verdict.json --judge reviewer-a
theoryofchange judge record theory.yaml verdict2.json --judge reviewer-b
theoryofchange judge summary theory.yaml --aggregation majority
```

1. `show` prints the **bundle**: the rubric, the document, a narrative, the lint
   report, instructions, and a hash of all of them. Restricted evidence is
   redacted.
2. The judge returns a verdict, a structured-evaluation report that scores each
   rubric category and names its findings.
3. `record` checks the verdict against the bundle and stores it under
   `.toc/judgements/<hash>/<judge>.json`. It is rejected, with exit code 1, if it
   judged a different version of the theory, is not a valid report, scores a
   category the rubric does not have, or leaves a required category unscored.
4. `summary` combines every verdict for the current version: per-category scores,
   how closely the judges agree, and whether the theory passes.

Rubric profiles: `core` has 12 categories; `development` adds categories for
programs working on development outcomes. Aggregation is `majority` (a tie goes
to partial), `conservative` (the lowest score), `optimistic` (the highest), or
`unanimous` (anything else is partial). The theory passes when every required
category passes, no blocking category fails, and the findings stay within the
rubric's limits.

!!! note "How well does the rubric discriminate?"
    The [rubric evaluation](../evals/rubric-discrimination.md) records two judges
    scoring a reference theory and a deliberately weak one. The weak theory fails
    most categories and the reference has no high or critical findings.

## Next

[MCP server and agent skills](mcp-and-skills.md)
