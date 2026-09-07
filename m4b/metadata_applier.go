package m4b

import (
	"fmt"
	"io"
	"os"

	"github.com/achwo/narr/utils"
)

// MetadataApplyExtensions lists the audio file extensions that narr can retag
// in place.
var MetadataApplyExtensions = []string{".m4b", ".m4a", ".mp3", ".flac"}

// metadataProcessor reads and writes the metadata of a single audio file.
type metadataProcessor interface {
	ReadMetadata(file string) (string, error)
	WriteMetadata(file string, metadata string, verbose bool) error
}

// mp4TagPatcher writes single tags of an MP4 file without rewriting the audio
// stream. Whether that works for a file depends on its container and on the
// tools at hand, so the patcher decides that itself and reports false when the
// file has to go through WriteMetadata instead.
type mp4TagPatcher interface {
	PatchMP4Tags(file string, tags []utils.TagWithValue, verbose bool) bool
}

// ApplyOptions configures a metadata apply run.
type ApplyOptions struct {
	DryRun  bool
	Verbose bool
}

// ApplyStatus describes what happened to a file during a metadata apply run.
type ApplyStatus string

const (
	// StatusUnchanged means the rules did not modify any tag of the file.
	StatusUnchanged ApplyStatus = "unchanged"
	// StatusUpdated means the new metadata was written to the file.
	StatusUpdated ApplyStatus = "updated"
	// StatusDryRun means changes were found but not written.
	StatusDryRun ApplyStatus = "dry-run"
	// StatusSkippedHardLink means the file was left untouched because it has
	// more than one hard link.
	StatusSkippedHardLink ApplyStatus = "skipped-hardlink"
)

// ChangeKind describes how a single tag was modified by the rules.
type ChangeKind string

const (
	// ChangeAdded means the tag did not exist before.
	ChangeAdded ChangeKind = "added"
	// ChangeUpdated means the value of an existing tag changed.
	ChangeUpdated ChangeKind = "updated"
	// ChangeDeleted means the tag was removed.
	ChangeDeleted ChangeKind = "deleted"
)

// TagChange describes the modification of a single metadata tag.
type TagChange struct {
	Tag    string
	Before string
	After  string
	Kind   ChangeKind
}

// String renders the change as "tag : before -> after".
func (c TagChange) String() string {
	before := c.Before
	after := c.After

	switch c.Kind {
	case ChangeAdded:
		before = "<none>"
	case ChangeDeleted:
		after = "<deleted>"
	}

	return fmt.Sprintf("%s : %s -> %s", c.Tag, before, after)
}

// FileResult reports the outcome of applying metadata rules to one file.
type FileResult struct {
	File    string
	Status  ApplyStatus
	Changes []TagChange
}

// MetadataApplier applies the metadata rules of a project config directly to
// existing audio files, without converting them.
type MetadataApplier struct {
	AudioProcessor metadataProcessor
	Out            io.Writer
}

// NewMetadataApplier creates a MetadataApplier that uses ffmpeg and writes its
// output to stdout.
func NewMetadataApplier() *MetadataApplier {
	return &MetadataApplier{
		AudioProcessor: NewFFmpegAudioProcessor(),
		Out:            os.Stdout,
	}
}

// ApplyToProject applies the metadata rules of the config to every supported
// audio file below the project path. It returns one result per file.
func (a *MetadataApplier) ApplyToProject(config ProjectConfig, opts ApplyOptions) ([]FileResult, error) {
	path, err := config.FullAudioFilePath()
	if err != nil {
		return nil, fmt.Errorf("could not resolve project path: %w", err)
	}

	files, err := utils.GetFilesByExtensions(path, MetadataApplyExtensions)
	if err != nil {
		return nil, fmt.Errorf("could not get audio files of %s: %w", path, err)
	}

	results := make([]FileResult, 0, len(files))

	for _, file := range files {
		result, err := a.ApplyToFile(file, config.MetadataRules, opts)
		if err != nil {
			return results, err
		}
		results = append(results, result)
	}

	return results, nil
}

// ApplyToFile applies the metadata rules to a single audio file. It prints the
// diff before writing and skips files with more than one hard link.
func (a *MetadataApplier) ApplyToFile(file string, rules []MetadataRule, opts ApplyOptions) (FileResult, error) {
	rawMetadata, err := a.AudioProcessor.ReadMetadata(file)
	if err != nil {
		return FileResult{File: file}, fmt.Errorf("could not read metadata of %s: %w", file, err)
	}

	before := utils.ParseFFMetadata(rawMetadata)
	after := before.Clone()

	if err := applyRulesToMetadata(after, rules); err != nil {
		return FileResult{File: file}, fmt.Errorf("could not apply metadata rules to %s: %w", file, err)
	}

	changes := tagChanges(before, after)

	if len(changes) == 0 {
		if opts.Verbose {
			a.printf("# %s\nNothing to do.\n", file)
		}
		return FileResult{File: file, Status: StatusUnchanged}, nil
	}

	a.printf("# %s\n", file)
	for _, change := range changes {
		a.printf("%s\n", change.String())
	}

	if opts.Verbose {
		a.printf("Metadata after update:\n%s\n", after.String())
	}

	if opts.DryRun {
		a.printf("\n")
		return FileResult{File: file, Status: StatusDryRun, Changes: changes}, nil
	}

	links, err := utils.HardLinkCount(file)
	if err != nil {
		return FileResult{File: file, Changes: changes}, fmt.Errorf("could not check hard links of %s: %w", file, err)
	}

	if links > 1 {
		a.printf("WARNING: skipping %s, it has %d hard links, writing would change the other copies too\n\n", file, links)
		return FileResult{File: file, Status: StatusSkippedHardLink, Changes: changes}, nil
	}

	if err := a.write(file, after, changes, opts.Verbose); err != nil {
		return FileResult{File: file, Changes: changes}, fmt.Errorf("could not write metadata of %s: %w", file, err)
	}

	a.printf("Metadata successfully changed for file %s\n\n", file)

	return FileResult{File: file, Status: StatusUpdated, Changes: changes}, nil
}

// write puts the new metadata into the file, patching only the changed tags
// when the processor can do that, and rewriting the whole file otherwise.
func (a *MetadataApplier) write(
	file string,
	after *utils.FFMetadata,
	changes []TagChange,
	verbose bool,
) error {
	if patcher, ok := a.AudioProcessor.(mp4TagPatcher); ok {
		if patcher.PatchMP4Tags(file, changedTags(after, changes), verbose) {
			return nil
		}
	}

	return a.AudioProcessor.WriteMetadata(file, after.String(), verbose)
}

// changedTags returns the changed tags under the spelling they have in the
// file, with an empty value for the deleted ones.
func changedTags(after *utils.FFMetadata, changes []TagChange) []utils.TagWithValue {
	tags := make([]utils.TagWithValue, 0, len(changes))
	for _, change := range changes {
		tags = append(tags, utils.TagWithValue{Tag: after.OriginalName(change.Tag), Value: change.After})
	}
	return tags
}

func (a *MetadataApplier) printf(format string, args ...any) {
	if a.Out == nil {
		return
	}
	fmt.Fprintf(a.Out, format, args...)
}

func applyRulesToMetadata(metadata *utils.FFMetadata, rules []MetadataRule) error {
	for _, rule := range rules {
		if err := rule.Apply(metadata.Tags); err != nil {
			return fmt.Errorf("rule for tag %s failed: %w", rule.Tag, err)
		}
		metadata.SyncTagOrder()
	}
	return nil
}

func tagChanges(before *utils.FFMetadata, after *utils.FFMetadata) []TagChange {
	var changes []TagChange

	for _, tag := range before.TagOrder {
		oldValue := before.Tags[tag]
		newValue, exists := after.Tags[tag]

		if !exists {
			changes = append(changes, TagChange{Tag: tag, Before: oldValue, Kind: ChangeDeleted})
			continue
		}

		if oldValue != newValue {
			changes = append(changes, TagChange{Tag: tag, Before: oldValue, After: newValue, Kind: ChangeUpdated})
		}
	}

	for _, tag := range after.TagOrder {
		if _, existed := before.Tags[tag]; existed {
			continue
		}
		changes = append(changes, TagChange{Tag: tag, After: after.Tags[tag], Kind: ChangeAdded})
	}

	return changes
}
