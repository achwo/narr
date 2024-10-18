package config

type ProjectConfig struct {
	AudioFilePath string
	CoverPath     string
	HasChapters   bool
	MetadataRules []RegexRule
	ChapterRules  []RegexRule
	OutputRules   []RegexRule
}

type RegexRule struct {
}
