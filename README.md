# narr

A Go command-line tool for building M4B audiobooks from ripped audio files, and for
inspecting and rewriting audio file metadata.

The idea is to keep the original files as ripped from cd in a lossless codec while
creating a lossily compressed m4b file for usage on phone etc. When there is an
error in the metadata or chapters, you fix it in the `narr.yaml`, rerun narr and
get a new file with the corrected data.

## Installation

```
go install
```

`make install` does nothing: the Makefile's `install` target has no recipe, the
`go install` line belongs to the `push` target.

## Prerequisites

- Go 1.23.2 or higher (see `go.mod`)
- `ffmpeg` and `ffprobe` on the PATH (metadata, cover, concat, duration)
- `mp4chaps` (mp4v2), only needed for projects with `hasChapters: true`
- With `shouldConvert: true`, narr encodes via `ffmpeg -c:a aac_at`, so the ffmpeg
  build must provide the AudioToolbox AAC encoder (macOS)

## Commands

Every command takes an optional path argument and falls back to the current
directory when it is omitted.

### m4b

- `narr m4b generate [dir]` — writes an empty `narr.yaml` into `dir`.
- `narr m4b check [dir]` — prints input tracks, chapters (if `hasChapters`),
  metadata and the resulting output filename, without writing anything.
  `-r, --recursive` checks all projects below the path. The subcommands
  `check chapters`, `check metadata`, `check filename` and `check files` print
  only that one part for a single project.
- `narr m4b run [dir]` — runs the conversion and writes the result to
  `~/narr/<artist>/<album>/<album>.m4b`. `-r, --recursive` converts all projects
  below the path. A project is skipped when the output file already exists and its
  duration is within 5% of the summed input duration.

### metadata

- `narr metadata show [path]` — prints the ffmpeg metadata of every `.m4b` file
  below `path`. `-t, --tag` limits the output to the given tags (repeatable).
- `narr metadata edit [path]` — applies one regex/format replacement to the given
  tags of every `.m4b` file below `path`. Requires `--regex`, `--format` and
  `-t, --tag`; `--dryRun` only prints the diff, `-v, --verbose` prints the full
  metadata.
- `narr metadata apply [path]` — applies the `metadataRules` of a project config to
  the project's own audio files (`.m4b`, `.m4a`, `.mp3`, `.flac`) in place, without
  converting them. `-r, --recursive` processes all projects below the path.
  Use `--dryRun` first: it prints the tag diff per file and writes nothing.
  When writing, files with more than one hard link are skipped, because writing them
  would change the other copies as well. Each run ends with a summary of updated,
  unchanged, dry-run and skipped files.

### file

- `narr file list [path]` — lists all `.m4b` files below `path`. `--noPath` prints
  only the file names.
- `narr file rename [path]` — renames all `.m4b` files below `path` by applying
  `--regex` and `--format` (both required) to the file name. `--dryRun` only prints
  the planned renames.

## m4b workflow

narr uses a docker-compose like project file named `narr.yaml`. It sits in the root
of the directory containing the audio files of the audiobook.

1. Go to the base directory of your project.
2. Run `narr m4b generate` to create a `narr.yaml`.
3. Fill the `narr.yaml` according to your use case.
4. Run `narr m4b check` to see tracks, chapters, metadata and output filename
   without executing anything.
5. When you're satisfied with the output, run `narr m4b run`.
6. When the conversion is done, find your output file(s) in `~/narr/`.

Input files are picked up recursively below the project directory (`.m4a`, `.mp3`,
`.flac`) and sorted by disc number, then track number, then file name.

### Project configuration

```yaml
# Cover image for the m4b, absolute or relative to the project directory.
# If it does not exist, the cover is extracted from the first audio file.
coverPath: ""

# Write chapter markers into the m4b (needs mp4chaps).
hasChapters: false

# Re-encode the input files to aac before concatenating them.
shouldConvert: true

# Treat every subdirectory next to this file as a separate project with this config.
multi: false

# Rewrite metadata tags before they are used for m4b metadata, chapters and filename.
metadataRules: []

# Rewrite track titles before they are grouped into chapters.
chapterRules: []
```

`ProjectConfig` also has a `projectPath` field, but narr sets it itself (the
directory of the `narr.yaml`, or the respective subdirectory for `multi: true`);
setting it in the file has no effect.

### metadataRules

Rules are applied in order to the tags of a track. Tag names are matched
case-insensitively. There are three types:

```yaml
metadataRules:
  # regex: needs tag, regex and format. The tag must exist.
  # One %s per capture group; the counts must match, otherwise the rule errors.
  # Input that does not match the regex is left unchanged.
  - type: regex
    tag: album
    regex: "Folge (\\d+): (.*)"
    format: "%s. %s"

  # set: needs tag and value, no regex or format. Creates the tag if it is missing.
  - type: set
    tag: genre
    value: "Hörspiel"

  # delete: needs only tag, no value, regex or format.
  - type: delete
    tag: comment
```

Regexes use Go regex syntax, so backslashes have to be escaped in YAML.

### chapterRules

Chapter titles are taken from the `title` tag of each track and then piped through
all chapter rules. Tracks that end up with the same title form one chapter.
A chapter rule needs both `regex` and `format`, with the same `%s` semantics as a
regex metadata rule:

```yaml
chapterRules:
  - regex: "^(Kapitel \\d+).*"
    format: "%s"
```
