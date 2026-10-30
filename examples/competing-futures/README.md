# Competing futures example (fictional)

Everything here is **fictional**. Three short theories of change describe three
competing futures of one invented market, the delegation of authority to AI
agents. Each is written from the perspective of the actor that wants that
future to come true, and all three share one actor registry, so "enterprise
buyers" means the same thing in each.

- `actors.yaml` - the shared actor registry (`agent-delegation-actors`)
- `future-open.yaml` - an open standard wins (`ufm:open-standard-wins`)
- `future-bundle.yaml` - a single vendor's bundled scheme wins (`ufm:bundled-scheme-wins`)
- `future-extension.yaml` - an extension of an older framework wins (`ufm:extension-wins`)

## What it shows

- Several theories of one market, each with its own `perspective` and
  `futureRef`, and no change to the schema.
- Actors defined once in a registry and referenced by ID from every theory.
- References from one theory to another's elements, as `toc:<doc>#<element>`:
  the bundled-scheme theory watches the open standard's pilot count as a
  competitor signal, and the extension theory aligns with one of its outcomes.

## Check it

```bash
theoryofchange validate future-open.yaml future-bundle.yaml future-extension.yaml \
  --registry actors.yaml --futures
```

Without `--registry`, each theory warns that its actors cannot be resolved.
With `--futures`, the theories are also checked against each other: they must
share one registry, name their futures and perspectives, not both speak for the
same actor in the same future, and refer to one another's elements correctly.
