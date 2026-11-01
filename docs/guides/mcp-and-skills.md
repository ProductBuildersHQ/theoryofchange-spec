# MCP Server and Agent Skills

The same validated operations the command line offers are available to an AI
agent over the [Model Context Protocol](https://modelcontextprotocol.io), and
five agent skills describe the common workflows.

## The MCP server

```bash
impact-model mcp --root ./impact-models
```

The server speaks MCP over standard input and output. Configure an MCP client to
run that command; for example, in a client that takes a JSON configuration:

```json
{
  "mcpServers": {
    "impact-model": {
      "command": "impact-model",
      "args": ["mcp", "--root", "./impact-models"]
    }
  }
}
```

It makes no network calls and no model calls.

### Tools

| Tool | What it does |
|------|--------------|
| `impact_schema` | List the embedded JSON Schemas or return one |
| `impact_validate` | Validate an impact model |
| `impact_lint` | Lint an impact model with a profile |
| `impact_graph` | Query the causal graph |
| `impact_status` | Compute a status snapshot from observations |
| `impact_report` | Render a definition or status report |
| `impact_deck` | Generate a deck as Marp Markdown |
| `impact_export` | Export for the PRISM tools as JSON: `okr`, `journey`, `v2mom`, or `dmaic`. With observations, the export carries the latest values and status. A model that lacks what the format needs is refused, naming what to add |
| `impact_diff` | Compare two versions of an impact model |
| `impact_judge_bundle` | Build the judge bundle |
| `impact_patch` | Apply a JSON Patch, atomically |

A tool takes an impact model **inline** (the `document` argument) or **by path**.
`impact_patch` rejects any patch that adds validation errors, exactly as the
`patch` command does, so an agent cannot leave an impact model worse than it found it.

### Where an agent can read and write

A path must lie inside `--root`, which defaults to the current directory. These
are refused with the same answer on every platform: a path with a `..` element,
a rooted or drive path outside the root, and a symbolic link that leads outside
the root. An inline document is limited in size.

## Agent skills

A skill is a Markdown file that tells an agent how to carry out one task with the
commands above. They live in the repository's `skills/` directory.

| Skill | Use it to |
|-------|-----------|
| `impact-draft-from-sources` | Draft an impact model for an organization from its public pages and reports. Every source is recorded as evidence, and every element is labelled stated, inferred, or hypothesis |
| `impact-author` | Add, change, or remove outcomes, links, assumptions, indicators, actors, and other elements with validated edits |
| `impact-status-refresh` | Load a new period's numbers from a spreadsheet, then report where the impact model stands and what changed |
| `impact-learning-review` | Prepare and record a learning review for a period |
| `impact-audience-story` | Draft a short update for one audience from the impact model's numbers and consented stories |

Each skill is tested: the commands and flags it names must exist, and its
sequence of steps is run against the examples to check it works and repeats.

## Next

[Examples](examples.md) · [CLI reference](cli.md)
