package m4b

import (
	"bytes"
	"fmt"
	"slices"
	"strings"

	"github.com/achwo/narr/utils"
)

// atomicParsleySwitches maps the ffmetadata tag names that ffmpeg reads from a
// standard MP4 atom to the AtomicParsley option writing that same atom.
var atomicParsleySwitches = map[string]string{
	"album":        "--album",
	"album_artist": "--albumArtist",
	"artist":       "--artist",
	"comment":      "--comment",
	"composer":     "--composer",
	"copyright":    "--copyright",
	"date":         "--year",
	"description":  "--description",
	"disc":         "--disk",
	"encoder":      "--encodingTool",
	"genre":        "--genre",
	"grouping":     "--grouping",
	"lyrics":       "--lyrics",
	"synopsis":     "--longdesc",
	"title":        "--title",
	"track":        "--tracknum",
}

// standardMP4TagsWithoutSwitch are the remaining tag names ffmpeg reads from a
// standard MP4 atom. Writing one of them as a freeform atom would leave the
// standard atom in place with its old value, so they cannot be patched.
var standardMP4TagsWithoutSwitch = []string{
	"compatible_brands",
	"compilation",
	"episode_id",
	"episode_sort",
	"gapless_playback",
	"hd_video",
	"keywords",
	"major_brand",
	"media_type",
	"minor_version",
	"network",
	"podcast",
	"purchase_date",
	"rating",
	"season_number",
	"show",
	"sort_album",
	"sort_album_artist",
	"sort_artist",
	"sort_composer",
	"sort_name",
	"sort_show",
}

// PatchMP4Tags writes the given tags into an MP4 file with AtomicParsley, which
// touches only the atoms it is told about. Chapters, cover art and atoms narr
// knows nothing about stay as they are, and a large file takes a second instead
// of a minute, because the audio stream is not rewritten. An empty value
// deletes the tag.
//
// It reports whether the file now carries the tags. AtomicParsley is optional
// and refuses files with more atoms than its fixed limit, so a false result is
// the expected signal to fall back to a remux, not an error.
func (p *FFmpegAudioProcessor) PatchMP4Tags(file string, tags []utils.TagWithValue, verbose bool) bool {
	if !isMP4(file) || len(tags) == 0 {
		return false
	}

	args, patchable := atomicParsleyArgs(file, tags)
	if !patchable {
		return false
	}

	if err := p.Command.LookPath(AtomicParsleyCommand); err != nil {
		if verbose {
			fmt.Printf("%s is not on the PATH, writing %s with ffmpeg instead\n", AtomicParsleyCommand, file)
		}
		return false
	}

	var outBuf bytes.Buffer
	err := p.Command.Create(AtomicParsleyCommand, args...).Run(&outBuf, &outBuf)

	if verbose {
		fmt.Printf("Command output:\n%s\n", outBuf.String())
	}

	if err != nil {
		fmt.Printf(
			"%s could not write the tags of %s, falling back to ffmpeg: %v\n%s\n",
			AtomicParsleyCommand, file, err, outBuf.String(),
		)
		return false
	}

	return p.mp4TagsWritten(file, tags)
}

// mp4TagsWritten reports whether the file carries every wanted tag. Reading the
// file back is necessary because AtomicParsley exits successfully when it
// truncates a value at 255 characters or skips an option it does not know.
func (p *FFmpegAudioProcessor) mp4TagsWritten(file string, tags []utils.TagWithValue) bool {
	metadata, err := p.ReadMetadata(file)
	if err != nil {
		return false
	}

	written := utils.ParseFFMetadata(metadata)

	for _, tag := range tags {
		if written.Tags[strings.ToLower(tag.Tag)] != tag.Value {
			return false
		}
	}

	return true
}

func atomicParsleyArgs(file string, tags []utils.TagWithValue) ([]string, bool) {
	args := make([]string, 0, len(tags)*4+2)
	args = append(args, file)

	for _, tag := range tags {
		name := strings.ToLower(tag.Tag)

		if slices.Contains(standardMP4TagsWithoutSwitch, name) {
			return nil, false
		}

		if option, standard := atomicParsleySwitches[name]; standard {
			args = append(args, option, tag.Value)
			continue
		}

		args = append(args, "--rDNSatom", tag.Value, "name="+tag.Tag, "domain="+freeformDomain)
	}

	return append(args, "--overWrite"), true
}
