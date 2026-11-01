# CLI Reference

`theoryofchange <command> --help` prints the full help for any command. This page
is the map.

## Conventions

- **Output files.** `-o` / `--out` names the file a command writes. Without it,
  most commands write to standard output.
- **Observation files.** `-O` / `--observations` names observation sets, and is
  repeatable.
- **Dates.** `--as-of YYYY-MM-DD` sets the date a status or report describes.
- **Format.** `--format text|json` is global. `graph` also takes `dot`, and
  `export review-packet` takes `md` or `docx`.
- **Quiet.** `-q` / `--quiet` prints only problems.
- **Dry run.** Commands that change a file take `--dry-run`, which prints the
  result and writes nothing.
- **Both formats.** A theory can be YAML or JSON, chosen by extension. Edits keep
  the file's own format.

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | The input is invalid, a change was rejected, or a check did not pass |
| 2 | A usage or I/O error |
| 3 | An optional external tool (`marp`, `mmdc`, `pandoc`) is missing |

## Commands

### Start and format

| Command | What it does |
|---------|--------------|
| `init` | Write a starter theory that already passes validation |
| `fmt` | Rewrite documents in canonical form (`--check`, `--sort`) |
| `schema` | List the seven embedded JSON Schemas or print one |

### Edit

| Command | What it does |
|---------|--------------|
| `<element> add`, `set`, `rm` | Add, change, or remove one element through the validated patch engine |
| `patch` | Apply an RFC 6902 JSON Patch atomically, with `@id` pointers |
| `diff` | Compare two versions by element ID, or print a draft revision patch |

The twenty element kinds are `actor`, `assumption`, `audience`, `cause`,
`consultation`, `criterion`, `decision`, `evidence`, `factor`, `group`,
`indicator`, `intervention`, `link`, `metric`, `outcome`, `pathway`, `risk`,
`target-group`, `term`, and `theme`. See [Authoring a theory](authoring.md).

### Check

| Command | What it does |
|---------|--------------|
| `validate` | Schema and structure (`--toc`, `--parent`, `--registry`, `--futures`) |
| `lint` | 66 deterministic rules in three profiles (`--profile`, `--disable`, `--list`) |
| `graph` | Summarize or query the causal graph (`--upstream`, `--downstream`, `--paths`, `--cycles`, `--orphans`, `--dead-ends`, `--levels`) |
| `judge show`, `record`, `summary` | Prepare a theory for a judge and combine the verdicts |

### Measure

| Command | What it does |
|---------|--------------|
| `import observations` | Load a spreadsheet through a column mapping, reporting every row it cannot place |
| `status` | Compute indicator, outcome, pathway, link, and assumption status as of a date (`--prev`, `--rollup`, `--tolerance`) |

### Report and export

| Command | What it does |
|---------|--------------|
| `report <kind>` | Markdown reports: definition kinds (no data needed) and status kinds |
| `deck workshop`, `exec`, `learning-review`, `story` | Marp slide decks (`--export` renders with `marp`) |
| `export csv <table>` | `indicators`, `metrics`, `glossary`, `metric-selection`, `collection-plan` |
| `export docx`, `export pptx` | Word and PowerPoint through `mmdc` and `pandoc` |
| `export okr`, `export journey` | JSON for the PRISM tools |

### Review

| Command | What it does |
|---------|--------------|
| `export review-packet`, `export review-sheet` | Send a theory out for review |
| `import review` | Turn reviewers' answers into a patch, flagged rows, and a consultation record |
| `approve` | Record approval, with approvers, a version, and a date |

### Serve

| Command | What it does |
|---------|--------------|
| `mcp` | Serve the tools to an AI agent over MCP on standard input and output (`--root`) |

## Shell completion

`theoryofchange completion bash|zsh|fish|powershell` prints a completion script.
