# narr

A Go-based command-line tool for processing and manipulating M4B audiobook files. This tool helps you manage, merge, and process audio files specifically for audiobook formats.

The idea is to keep the original files as ripped from cd in a lossless codec while
creating a lossily compressed m4b file for usage on phone etc. When there is an
error in the metadata or chapters, you simply fix it in the narr.yml, rerun narr and 
get a new file with the corrected data.

## Usage

narr uses a docker-compose like project-file named `narr.yml`. It should be located on the root
of the directory containing the audio files of the audio book.

## Basic workflow

1. Go to the base directory of your project
1. Run `narr m4b generate` to create a `narr.yml`.
1. Fill the narr.yml according to your use case.
4. Run `narr m4b check` to check your changes without executing them.
5. When you're satisfied with the output, run `narr m4b run`.
6. Wenn the conversion is done, find your output file(s) in `~/narr/`

### Project Configuration

The tool uses a YAML configuration file to define project settings. Here's an example configuration:

```yaml
# Path to the cover image for the audiobook. Uses cover from the first audio file if empty.
coverPath: "" 

# Whether to generate chapters from the audio files metadata titles
hasChapters: false

# manipulation of meta tags via regex (using go regex syntax)
metadataRules: 
  - tag: album
    type: regex
    regex: "Folge (\\d): (.*)"
    format: "Folge 00%s: %s"
  - tag: album
    type: regex
    regex: "Folge (\\d\\d): (.*)"
    format: "Folge 0%s: %s"
  - tag: album
    type: regex
    regex: "Folge (\\d+): (.*)"
    format: "%s. %s"

# rules to map title tags into chapters (same continuous title = no new chapter)
chapterRules:
  - pattern: "Chapter \\d+"
    format: "Chapter %d"
chapterRules: []
shouldConvert: true
multi: true

# Whether to convert audio files into aac before concatenating them
shouldConvert: true
```

## Prerequisites

- Go 1.16 or higher
- FFmpeg installed on your system

## Installation

```
make install
```

