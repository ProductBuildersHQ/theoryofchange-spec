# MCP Server and Agent Skills

The same validated operations the command line offers are available to an AI
agent over the [Model Context Protocol](https://modelcontextprotocol.io), and
five agent skills describe the common workflows.

## The MCP server

```bash
theoryofchange mcp --root ./theories
```

The server speaks MCP over standard input and output. Configure an MCP client to
run that command; for example, in a client that takes a JSON configuration:

```json
{
  "mcpServers": {
    "theoryofchange": {
      "command": "theoryofchange",
      "args": ["mcp", "--root", "./theories"]
    }
  }
}
```

It makes no network calls and no model calls.

### Tools

| Tool | What it does |
|------|--------------|
| `toc_schema` | List the embedded JSON Schemas or return one |
| `toc_validate` | Validate a theory |
| `toc_lint` | Lint a theory with a profile |
| `toc_graph` | Query the causal graph |
| `toc_status` | Compute a status snapshot from observations |
| `toc_report` | Render a definition or status report |
| `toc_deck` | Generate a deck as Marp Markdown |
| `toc_diff` | Compare two versions of a theory |
| `toc_judge_bundle` | Build the judge bundle |
| `toc_patch` | Apply a JSON Patch, atomically |

A tool takes a theory **inline** (the `document` argument) or **by path**.
`toc_patch` rejects any patch that adds validation errors, exactly as the
`patch` command does, so an agent cannot leave a theory worse than it found it.

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
| `toc-draft-from-sources` | Draft a theory for an organization from its public pages and reports. Every source is recorded as evidence, and every element is labelled stated, inferred, or hypothesis |
| `toc-author` | Add, change, or remove outcomes, links, assumptions, indicators, actors, and other elements with validated edits |
| `toc-status-refresh` | Load a new period's numbers from a spreadsheet, then report where the theory stands and what changed |
| `toc-learning-review` | Prepare and record a learning review for a period |
| `toc-audience-story` | Draft a short update for one audience from the theory's numbers and consented stories |

Each skill is tested: the commands and flags it names must exist, and its
sequence of steps is run against the examples to check it works and repeats.

## Next

[Examples](examples.md) · [CLI reference](cli.md)
