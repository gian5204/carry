# Carry

> Move local config files between Git clones without putting them in Git.

Files such as `.env`, `.env.local`, local JSON configuration, and development secrets are intentionally ignored by Git. That keeps machine-specific values and credentials out of version control, but it becomes inconvenient when the same repository has several local clones.

Carry tracks the ignored local files you want to manage and copies them between local clones of the same repository. The manifest stores paths, not file contents. Carry v0.1.0 is intentionally local: it does not upload files, synchronize through a server, or transfer files between devices.

## Quick start

From the clone that already has your local configuration:

```bash
# Discover ignored local files
carry discover

# See what's managed
carry list

# Copy them to another clone of the same repository
carry copy ../my-project-copy
```

`discover` shows ignored, untracked candidates and asks whether to add all of them. You can also manage paths directly:

```bash
carry add .env .env.local config/local.json
carry remove .env.local
```

Carry stores managed repository-relative paths in `.carry.json`. This file does **not** contain the contents of `.env` files, secrets, or any other managed file. You can commit the manifest if you want clones to share the same selection.

## Installation

### From source

If you have Go installed:

```bash
go install github.com/gian5204/carry@latest
```

Carry also requires Git.

Pre-built binaries will be available with the v0.1.0 release.

## Usage

Run `carry`, `carry help`, `carry --help`, or `carry -h` for the top-level command reference.

### `discover`

```bash
carry discover
```

Carry asks Git for files that are both ignored and untracked. It then applies its built-in exclusions, applies `.carryignore`, removes paths already present in `.carry.json`, sorts the result, and displays the remaining candidates.

If candidates are found, Carry prompts:

```text
Add all discovered files to Carry? [y/N]
```

Only `y` or `Y` adds the discovered files. Any other answer leaves the manifest unchanged.

### `add` and `remove`

Add one or more ignored local files:

```bash
carry add .env .env.local config/local.json
```

Before changing the manifest, `add` validates the complete batch. Every path must exist, be untracked by Git, and be ignored by Git. A validation error aborts the command without saving partial manifest changes. Paths already managed by Carry produce a warning and are otherwise harmless.

Remove one or more paths from the manifest:

```bash
carry remove .env .env.local
```

Paths that are not managed produce a warning and do not prevent other paths from being removed. Both commands ignore duplicate arguments. Carry currently manages individual file paths; recursive directory copying is not implemented.

### `list`

```bash
carry list
```

Lists the repository-relative paths currently managed through `.carry.json`.

### `copy`

```bash
carry copy ../my-project-copy
```

The destination must be another local clone of the same Git repository. Carry verifies this using repository identity derived from the normalized Git `origin` URL and refuses to copy when the identities differ.

Copy uses the source clone's `.carry.json`, and only its managed paths are considered. Existing destination files are never silently overwritten: each conflict requires an individual decision. Enter or `n` skips the file, while `y` approves replacement. Carry collects all overwrite decisions before copying any files. Skipped files remain untouched, and missing parent directories are created when needed.

### Version

```bash
carry version
carry --version
carry -v
```

All three forms print the build version. Normal unversioned development builds report:

```text
Carry dev
```

## How it works

Carry keeps selection separate from content:

```text
repo-a/
├── .carry.json       paths selected for copying
├── .env              ignored local file
└── config/local.json ignored local file
        │
        │ carry copy ../repo-b
        ▼
repo-b/
├── .env
└── config/local.json
```

A manifest is ordinary JSON at the repository root:

```json
{
  "version": 1,
  "files": [
    ".env",
    "config/local.json"
  ]
}
```

The listed files remain normal files in the working tree and remain subject to the repository's Git ignore rules. Carry reads or copies their contents only for an explicit copy operation. It does not maintain a hidden backup of those contents.

## `.carryignore`

`.gitignore` controls what Git tracks. `.carryignore` controls what `carry discover` suggests.

Place `.carryignore` at the repository root:

```text
# Local test data
local_testing/

# Local databases
*.sqlite

screenshots/
```

The current matcher supports three forms:

- `local_testing/` excludes files under that directory, including nested occurrences.
- `*.sqlite` excludes files with that suffix at any depth.
- `config/local.json` excludes that exact repository-relative path.

Blank lines and lines beginning with `#` are ignored, surrounding whitespace is trimmed, and `/` or `\` separators are normalized. Matching is case-sensitive. Negation, general globbing, and full `.gitignore` semantics are not supported. `.carryignore` itself is never suggested by discovery.

## Discovery exclusions

Carry conservatively filters a few obvious generated and runtime paths before applying `.carryignore`. Excluded directory names are `node_modules`, `dist`, `build`, and `coverage`, wherever they occur in a path. Excluded file extensions are `.exe`, `.dll`, `.so`, `.dylib`, `.log`, and `.tmp`. Built-in matching is case-insensitive.

## Safety

- `add` accepts only existing paths that Git reports as ignored and untracked.
- `.carry.json` stores paths only, never managed file contents.
- `copy` verifies that the source and destination are clones of the same repository before writing files.
- Existing destination files require explicit per-file confirmation.
- Carry performs no uploads or network transfer in v0.1.0.

## Current limitations

v0.1.0 supports local clone-to-clone copying only. There is no cross-device transfer, remote synchronization, encryption, or device pairing yet. Repository identity depends on both clones having an `origin` remote whose normalized URLs match. `.carryignore` has deliberately limited matching rules, and copy is not a filesystem transaction: a later I/O error can stop execution after earlier planned files were copied.

The CLI and manifest format are early and may evolve as the project gains real-world use.

## Roadmap

Possible future directions include:

- encrypted peer-to-peer transfer between devices
- device pairing
- richer discovery controls
- package-manager distribution
- `carry init`

These are ideas, not commitments.

## Contributing

Issues and pull requests are welcome. The usual development checks are:

```bash
go test ./...
go vet ./...
go build ./...
```

## License

Carry is available under the [MIT License](LICENSE).
