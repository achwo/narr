package m4b

import (
	"bytes"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/achwo/narr/utils"
)

// AtomicParsleyCommand is the executable that writes the freeform atoms the
// ffmpeg mov muxer cannot write.
const AtomicParsleyCommand = "AtomicParsley"

// freeformDomain is the reverse-DNS domain that iTunes, Audible and the common
// audiobook players use for tags outside the standard MP4 key set.
const freeformDomain = "com.apple.iTunes"

var mp4Extensions = []string{".m4b", ".m4a", ".m4p", ".m4v", ".mp4"}

func isMP4(file string) bool {
	return slices.Contains(mp4Extensions, strings.ToLower(filepath.Ext(file)))
}

// preserveFreeformTags writes the tags of wanted that outputFile no longer has
// back into outputFile.
//
// ffmpeg reads freeform MP4 atoms (----:com.apple.iTunes:SERIES and friends) but
// its mov muxer only writes its own fixed key set and drops the rest without a
// word, so every remux of an Audible file would lose SERIES, SUBTITLE and
// AUDIBLE_ASIN. Which keys survive depends on the ffmpeg build, so the tags that
// were lost are determined by reading the file that was actually written instead
// of from a list kept in this repo.
//
// originalFile supplies the spelling the tags had before the write, so that a
// SERIES atom is not silently renamed to series. It may be empty.
func (p *FFmpegAudioProcessor) preserveFreeformTags(originalFile string, outputFile string, wanted string) error {
	if !isMP4(outputFile) {
		return nil
	}

	written, err := p.ReadMetadata(outputFile)
	if err != nil {
		return fmt.Errorf("could not read back metadata of %s: %w", outputFile, err)
	}

	lost := lostTags(wanted, written)
	if len(lost) == 0 {
		return nil
	}

	if originalFile != "" {
		original, err := p.ReadMetadata(originalFile)
		if err != nil {
			return fmt.Errorf("could not read metadata of %s: %w", originalFile, err)
		}
		useOriginalSpelling(lost, utils.ParseFFMetadata(original))
	}

	if err := p.Command.LookPath(AtomicParsleyCommand); err != nil {
		return fmt.Errorf(
			"ffmpeg cannot write the MP4 tags %s and %s is not on the PATH to write them instead; "+
				"install %s and run again, no file was changed",
			tagNames(lost),
			AtomicParsleyCommand,
			AtomicParsleyCommand,
		)
	}

	args := make([]string, 0, len(lost)*4+2)
	args = append(args, outputFile)
	for _, tag := range lost {
		args = append(args, "--rDNSatom", tag.Value, "name="+tag.Tag, "domain="+freeformDomain)
	}
	args = append(args, "--overWrite")

	cmd := p.Command.Create(AtomicParsleyCommand, args...)

	var outBuf bytes.Buffer
	if err := cmd.Run(&outBuf, &outBuf); err != nil {
		return fmt.Errorf(
			"could not write the MP4 tags %s of %s with %s: %w\n%s",
			tagNames(lost),
			outputFile,
			AtomicParsleyCommand,
			err,
			outBuf.String(),
		)
	}

	return nil
}

// mp4TagsOf returns the metadata of an MP4 file, so that a following write can
// be checked against it. Non-MP4 files yield an empty document.
func (p *FFmpegAudioProcessor) mp4TagsOf(file string) (string, error) {
	if !isMP4(file) {
		return "", nil
	}
	return p.ReadMetadata(file)
}

// lostTags returns the tags of wanted that are missing from written.
func lostTags(wanted string, written string) []utils.TagWithValue {
	want := utils.ParseFFMetadata(wanted)
	have := utils.ParseFFMetadata(written)

	var lost []utils.TagWithValue

	for _, tag := range want.TagOrder {
		value, exists := want.Tags[tag]
		if !exists || value == "" {
			continue
		}
		if _, kept := have.Tags[tag]; kept {
			continue
		}
		lost = append(lost, utils.TagWithValue{Tag: want.OriginalName(tag), Value: value})
	}

	return lost
}

func useOriginalSpelling(tags []utils.TagWithValue, original *utils.FFMetadata) {
	for i, tag := range tags {
		lowered := strings.ToLower(tag.Tag)
		if _, known := original.Tags[lowered]; known {
			tags[i].Tag = original.OriginalName(lowered)
		}
	}
}

func tagNames(tags []utils.TagWithValue) string {
	names := make([]string, 0, len(tags))
	for _, tag := range tags {
		names = append(names, tag.Tag)
	}
	return strings.Join(names, ", ")
}
