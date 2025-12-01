package m4b_test

import (
	"testing"

	"github.com/achwo/narr/m4b"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetadataRule_Validate_Set(t *testing.T) {
	tests := []struct {
		name    string
		rule    m4b.MetadataRule
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid set rule",
			rule: m4b.MetadataRule{
				Type:  "set",
				Tag:   "album",
				Value: "My Album",
			},
			wantErr: false,
		},
		{
			name: "set rule without value",
			rule: m4b.MetadataRule{
				Type: "set",
				Tag:  "album",
			},
			wantErr: true,
			errMsg:  "set rule requires a value",
		},
		{
			name: "set rule with regex",
			rule: m4b.MetadataRule{
				Type:  "set",
				Tag:   "album",
				Value: "My Album",
				Regex: ".*",
			},
			wantErr: true,
			errMsg:  "set rule cannot have regex or format",
		},
		{
			name: "set rule with format",
			rule: m4b.MetadataRule{
				Type:   "set",
				Tag:    "album",
				Value:  "My Album",
				Format: "%s",
			},
			wantErr: true,
			errMsg:  "set rule cannot have regex or format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rule.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMetadataRule_Apply_Set(t *testing.T) {
	tests := []struct {
		name          string
		rule          m4b.MetadataRule
		initialTags   map[string]string
		expectedTags  map[string]string
		expectedError bool
	}{
		{
			name: "set rule sets value when tag exists",
			rule: m4b.MetadataRule{
				Type:  "set",
				Tag:   "album",
				Value: "New Album Name",
			},
			initialTags: map[string]string{
				"album":  "Old Album Name",
				"artist": "Artist Name",
			},
			expectedTags: map[string]string{
				"album":  "New Album Name",
				"artist": "Artist Name",
			},
			expectedError: false,
		},
		{
			name: "set rule creates tag when it doesn't exist",
			rule: m4b.MetadataRule{
				Type:  "set",
				Tag:   "album",
				Value: "New Album Name",
			},
			initialTags: map[string]string{
				"artist": "Artist Name",
			},
			expectedTags: map[string]string{
				"album":  "New Album Name",
				"artist": "Artist Name",
			},
			expectedError: false,
		},
		{
			name: "set rule overwrites existing value",
			rule: m4b.MetadataRule{
				Type:  "set",
				Tag:   "album",
				Value: "Arkham Horror - Litanei der Träume",
			},
			initialTags: map[string]string{
				"album":  "Some Random Album",
				"title":  "Track 1",
				"artist": "H.P. Lovecraft",
			},
			expectedTags: map[string]string{
				"album":  "Arkham Horror - Litanei der Träume",
				"title":  "Track 1",
				"artist": "H.P. Lovecraft",
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rule.Apply(tt.initialTags)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedTags, tt.initialTags)
			}
		})
	}
}

func TestMetadataRule_Apply_Delete(t *testing.T) {
	rule := m4b.MetadataRule{
		Type: "delete",
		Tag:  "album",
	}

	tags := map[string]string{
		"album":  "Album Name",
		"artist": "Artist Name",
	}

	err := rule.Apply(tags)
	require.NoError(t, err)

	expectedTags := map[string]string{
		"artist": "Artist Name",
	}
	assert.Equal(t, expectedTags, tags)
}

func TestMetadataRule_Integration_SetRule(t *testing.T) {
	// Integration test with Project to verify set rule works end-to-end
	data := make(map[string]m4b.FileData)

	data["file1.m4a"] = m4b.FileData{
		Title:    "Chapter 1",
		Duration: 5000,
		Metadata: `;FFMETADATA1
title=Chapter 01
artist=Hans Wurst
date=2002-09-16`,
	}

	fakeAudioProcessor := &m4b.NullAudioProcessor{Data: data}
	trackFactory := &m4b.FFmpegTrackFactory{AudioProcessor: fakeAudioProcessor}

	config := m4b.ProjectConfig{
		MetadataRules: []m4b.MetadataRule{
			{
				Type:  "set",
				Tag:   "album",
				Value: "Arkham Horror - Litanei der Träume",
			},
		},
	}

	deps := m4b.ProjectDependencies{
		AudioFileProvider: &FakeAudioFileProvider{
			Files: []string{"file1.m4a"},
		},
		AudioProcessor: fakeAudioProcessor,
		TrackFactory:   trackFactory,
	}

	project, err := m4b.NewProjectWithDeps(config, deps)
	require.NoError(t, err)

	metadata, err := project.Metadata()
	require.NoError(t, err)

	// Verify that the album tag was added with the set value
	assert.Contains(t, metadata, "album=Arkham Horror - Litanei der Träume")

	// Verify full metadata structure
	expectedMetadata := `;FFMETADATA1
title=Chapter 01
artist=Hans Wurst
date=2002-09-16
album=Arkham Horror - Litanei der Träume`

	assert.Equal(t, expectedMetadata, metadata)
}
