#!/usr/bin/env bash
# Regenerate THIRD_PARTY_NOTICES.md and third_party/licenses/ from the modules
# that are compiled into the impact-model binary.
#
# Each module's LICENSE, COPYING, and NOTICE files are copied verbatim. The
# license is named from the text of those files, and the script stops with an
# error for any license text it does not recognize, so a new dependency with
# unusual terms gets a human review instead of a silent label.
#
# Run it after any change to go.mod, and commit the result.
set -euo pipefail
cd "$(dirname "$0")/.."

out=third_party/licenses
notices=THIRD_PARTY_NOTICES.md
cli=./cmd/impact-model

rm -rf "$out"
mkdir -p "$out"

# module path, version, directory, for everything compiled into the binary
modules() {
  go list -deps -f '{{with .Module}}{{if not .Main}}{{.Path}} {{.Version}} {{.Dir}}{{end}}{{end}}' "$@" | sort -u
}

# classify prints the SPDX names found in the given files, joined with AND.
classify() {
  local found=() text
  text=$(cat "$@")
  grep -qi 'Permission is hereby granted, free of charge' <<<"$text" && found+=("MIT")
  if grep -qi 'Redistribution and use in source and binary forms' <<<"$text"; then
    if grep -qi 'Neither the name' <<<"$text"; then found+=("BSD-3-Clause"); else found+=("BSD-2-Clause"); fi
  fi
  grep -qi 'Apache License' <<<"$text" && found+=("Apache-2.0")
  if [ ${#found[@]} -eq 0 ]; then
    echo "UNRECOGNIZED"
    return 0
  fi
  local IFS=,
  local joined="${found[*]}"
  echo "${joined//,/ AND }"
}

# copyright prints the first specific copyright line of a license file.
copyright() {
  local lines first n
  lines=$(grep -hiE '^[[:space:]]*(copyright[[:space:]]+(\(c\)|©)?[[:space:]]*[0-9]{4}|copyright[[:space:]]+\(c\))' "$@" 2>/dev/null |
    grep -v '\[yyyy\]' | sed -E 's/^[[:space:]]+//; s/[[:space:]]+$//; s/\|/\\|/g' | sort -u || true)
  [ -z "$lines" ] && return 0
  first=$(head -1 <<<"$lines")
  n=$(wc -l <<<"$lines" | tr -d ' ')
  if [ "$n" -gt 1 ]; then echo "$first and others"; else echo "$first"; fi
}

{
  cat <<'HEADER'
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
HEADER

  unrecognized=0
  while read -r path ver dir; do
    dest="$out/$path"
    mkdir -p "$dest"
    files=()
    for f in "$dir"/LICENSE* "$dir"/LICENCE* "$dir"/COPYING* "$dir"/NOTICE*; do
      [ -f "$f" ] && files+=("$f")
    done
    if [ ${#files[@]} -eq 0 ]; then
      echo "no license file in $path $ver" >&2
      unrecognized=1
      continue
    fi
    for f in "${files[@]}"; do
      install -m 0644 "$f" "$dest/$(basename "$f")"
    done
    lic=$(classify "${files[@]}")
    if [ "$lic" = "UNRECOGNIZED" ]; then
      echo "license text of $path $ver is not one this script recognizes" >&2
      unrecognized=1
    fi
    cp=$(copyright "${files[@]}")
    printf '| [`%s`](third_party/licenses/%s) | %s | %s | %s |\n' "$path" "$path" "$ver" "$lic" "${cp:-see the license file}"
  done < <(modules "$cli")

  cat <<'DEVHEAD'

## Notices that must be carried forward

`gopkg.in/yaml.v3` is covered by two licenses: MIT for the files ported from
libyaml, and Apache-2.0 for the rest. Its `NOTICE` file reads:

DEVHEAD
  yamldir=$(modules "$cli" | awk '$1 == "gopkg.in/yaml.v3" {print $3}')
  if [ -n "$yamldir" ] && [ -f "$yamldir/NOTICE" ]; then
    sed 's/^\(.\)/    \1/' "$yamldir/NOTICE"
  fi

  cat <<'MCP'

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
MCP

  cli_set=$(modules "$cli" | awk '{print $1" "$2}')
  while read -r path ver dir; do
    grep -qxF "$path $ver" <<<"$cli_set" && continue
    files=()
    for f in "$dir"/LICENSE* "$dir"/LICENCE* "$dir"/COPYING* "$dir"/NOTICE*; do
      [ -f "$f" ] && files+=("$f")
    done
    if [ ${#files[@]} -eq 0 ]; then
      echo "no license file in $path $ver" >&2
      unrecognized=1
      continue
    fi
    lic=$(classify "${files[@]}")
    if [ "$lic" = "UNRECOGNIZED" ]; then
      echo "license text of $path $ver is not one this script recognizes" >&2
      unrecognized=1
    fi
    printf '| `%s` | %s | %s |\n' "$path" "$ver" "$lic"
  done < <(modules ./...)

  cat <<'OPTIONAL'

## Programs this project can call but does not include

[Marp](https://marp.app), the [Mermaid CLI](https://github.com/mermaid-js/mermaid-cli),
and [Pandoc](https://pandoc.org) are optional programs that some commands run
when they are installed. They are not part of this repository or its binary, and
each is used under its own license.

Optional documentation tooling ([MkDocs](https://www.mkdocs.org) with the
[Material theme](https://squidfunk.github.io/mkdocs-material/)) builds the
documentation site and is likewise not included.
OPTIONAL
  echo "UNRECOGNIZED=$unrecognized" >"$out/.status"
} >"$notices"

status=$(sed -n 's/^UNRECOGNIZED=//p' "$out/.status")
rm -f "$out/.status"
if [ "$status" != "0" ]; then
  echo "review the messages above, then update classify() or the dependency" >&2
  exit 1
fi
echo "wrote $notices and $(find "$out" -type f | wc -l | tr -d ' ') license files under $out/"
