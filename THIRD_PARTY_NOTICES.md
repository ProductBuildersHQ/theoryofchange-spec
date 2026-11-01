# Third-party notices

The `impact-model` command and its MCP server are built from this repository
together with the Go modules below. This repository is licensed under the
[MIT License](LICENSE). Each module keeps its own license, and the full text of
every license and notice file is in [`third_party/licenses/`](third_party/licenses/),
copied unchanged from the module.

All of them are permissive licenses (MIT, BSD-3-Clause, and Apache-2.0). None
requires that this project be licensed the same way. They do require that the
copyright notice and license text travel with any copy of the software that you
redistribute, and this file and `third_party/licenses/` are how this project
meets that requirement for a binary release.

This repository is distributed as source and installed with `go install`, which
fetches each module with its own license files. If you redistribute a built
binary, include this file and `third_party/licenses/` with it.

Regenerate this file with `scripts/third-party-notices.sh`.

## Compiled into the binary

| Module | Version | License | Copyright |
|--------|---------|---------|-----------|
| [`github.com/google/jsonschema-go`](third_party/licenses/github.com/google/jsonschema-go) | v0.4.3 | MIT | Copyright (c) 2025 JSON Schema Go Project Authors |
| [`github.com/grokify/d2vision`](third_party/licenses/github.com/grokify/d2vision) | v0.5.0 | MIT | Copyright (c) 2026 John Wang |
| [`github.com/grokify/oscompat`](third_party/licenses/github.com/grokify/oscompat) | v0.12.0 | MIT | Copyright (c) 2025 John Wang |
| [`github.com/grokify/prism-core`](third_party/licenses/github.com/grokify/prism-core) | v0.5.0 | MIT | Copyright (c) 2026 John Wang |
| [`github.com/grokify/prism-roadmap`](third_party/licenses/github.com/grokify/prism-roadmap) | v0.21.0 | MIT | Copyright (c) 2026 John Wang |
| [`github.com/grokify/structured-goals`](third_party/licenses/github.com/grokify/structured-goals) | v0.2.0 | MIT | Copyright (c) 2026 John Wang |
| [`github.com/modelcontextprotocol/go-sdk`](third_party/licenses/github.com/modelcontextprotocol/go-sdk) | v1.8.0 | MIT AND Apache-2.0 | Copyright (c) 2024-2025 Model Context Protocol a Series of LF Projects, LLC. |
| [`github.com/plexusone/structured-evaluation`](third_party/licenses/github.com/plexusone/structured-evaluation) | v0.15.0 | MIT | Copyright (c) 2026 John Wang |
| [`github.com/santhosh-tekuri/jsonschema/v6`](third_party/licenses/github.com/santhosh-tekuri/jsonschema/v6) | v6.0.3 | Apache-2.0 | see the license file |
| [`github.com/segmentio/asm`](third_party/licenses/github.com/segmentio/asm) | v1.2.1 | MIT | Copyright 2023 Segment |
| [`github.com/segmentio/encoding`](third_party/licenses/github.com/segmentio/encoding) | v0.5.4 | MIT | Copyright (c) 2019 Segment.io, Inc. |
| [`github.com/spf13/cobra`](third_party/licenses/github.com/spf13/cobra) | v1.10.2 | Apache-2.0 | see the license file |
| [`github.com/spf13/pflag`](third_party/licenses/github.com/spf13/pflag) | v1.0.10 | BSD-3-Clause | Copyright (c) 2012 Alex Ogier. All rights reserved. and others |
| [`github.com/yosida95/uritemplate/v3`](third_party/licenses/github.com/yosida95/uritemplate/v3) | v3.0.2 | BSD-3-Clause | Copyright (C) 2016, Kohei YOSHIDA <https://yosida95.com/>. All rights reserved. |
| [`golang.org/x/oauth2`](third_party/licenses/golang.org/x/oauth2) | v0.37.0 | BSD-3-Clause | Copyright 2009 The Go Authors. |
| [`golang.org/x/sync`](third_party/licenses/golang.org/x/sync) | v0.24.0 | BSD-3-Clause | Copyright 2009 The Go Authors. |
| [`golang.org/x/sys`](third_party/licenses/golang.org/x/sys) | v0.49.0 | BSD-3-Clause | Copyright 2009 The Go Authors. |
| [`golang.org/x/text`](third_party/licenses/golang.org/x/text) | v0.43.0 | BSD-3-Clause | Copyright 2009 The Go Authors. |
| [`golang.org/x/time`](third_party/licenses/golang.org/x/time) | v0.16.0 | BSD-3-Clause | Copyright 2009 The Go Authors. |
| [`gopkg.in/yaml.v3`](third_party/licenses/gopkg.in/yaml.v3) | v3.0.1 | MIT AND Apache-2.0 | Copyright (c) 2006-2010 Kirill Simonov and others |

## Notices that must be carried forward

`gopkg.in/yaml.v3` is covered by two licenses: MIT for the files ported from
libyaml, and Apache-2.0 for the rest. Its `NOTICE` file reads:

    Copyright 2011-2016 Canonical Ltd.

    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at

        http://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.

`github.com/modelcontextprotocol/go-sdk` is mid-way through a move from the MIT
License to Apache-2.0. New code is Apache-2.0, and contributions whose authors
have not agreed to relicense remain MIT. Both licenses are permissive, and its
`LICENSE` file, copied in full, explains the split.

## Used to build, not distributed

These modules are used by the schema generator and the tests. They are not
compiled into the `impact-model` binary, so no notice travels with it, and
they are listed here for completeness.

| Module | Version | License |
|--------|---------|---------|
| `github.com/bahlo/generic-list-go` | v0.2.0 | BSD-3-Clause |
| `github.com/buger/jsonparser` | v1.6.1 | MIT |
| `github.com/invopop/jsonschema` | v0.14.0 | MIT |
| `github.com/pb33f/go-yaml` | v0.1.1 | Apache-2.0 |
| `github.com/pb33f/ordered-map/v2` | v2.3.2 | Apache-2.0 |

## Programs this project can call but does not include

[Marp](https://marp.app), the [Mermaid CLI](https://github.com/mermaid-js/mermaid-cli),
and [Pandoc](https://pandoc.org) are optional programs that some commands run
when they are installed. They are not part of this repository or its binary, and
each is used under its own license.

Optional documentation tooling ([MkDocs](https://www.mkdocs.org) with the
[Material theme](https://squidfunk.github.io/mkdocs-material/)) builds the
documentation site and is likewise not included.
