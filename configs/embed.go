package configs

import (
	"embed"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

//go:embed default.yaml prompts.yaml
var embeddedFS embed.FS

type Config struct {
	LanguageMap       map[string]string `yaml:"language_map"`
	DefaultIgnore     []string          `yaml:"default_ignore"`
	BinaryExtensions  []string          `yaml:"binary_extensions"`
	NonCodeExtensions []string          `yaml:"non_code_extensions"`
	CustomIgnore      CustomIgnore      `yaml:"custom_ignore"`
	Output            Output            `yaml:"output"`
	Prompts           Prompts           `yaml:"prompts"`
}

type CustomIgnore struct {
	Patterns []string `yaml:"patterns"`
	Regex    []string `yaml:"regex"`
}

type Output struct {
	MaxChars      int    `yaml:"max_chars"`
	Compress      bool   `yaml:"compress"`
	UltraCompress bool   `yaml:"ultra_compress"`
	SplitMode     string `yaml:"split_mode"`
	IncludeTree   bool   `yaml:"include_tree"`
	OutputPrefix  string `yaml:"output_prefix"`
}

type Prompts struct {
	SectionInfo         string `yaml:"section_info"`
	SectionTree         string `yaml:"section_tree"`
	SectionCode         string `yaml:"section_code"`
	SectionStats        string `yaml:"section_stats"`
	HeaderPrompt        string `yaml:"header_prompt"`
	CompressNotice      string `yaml:"compress_notice"`
	UltraCompressNotice string `yaml:"ultra_compress_notice"`
	ContinueNotice      string `yaml:"continue_notice"`
	CompleteNotice      string `yaml:"complete_notice"`
	ProjectSeparator    string `yaml:"project_separator"`
	FileInfoFormat      string `yaml:"file_info_format"`
	NonCodeFileNotice   string `yaml:"non_code_file_notice"`
	BinaryFileSkip      string `yaml:"binary_file_skip"`
	StatsTableHeader    string `yaml:"stats_table_header"`
	UsageInstructions   string `yaml:"usage_instructions"`
}

func LoadEmbedded() (*Config, error) {
	cfg := &Config{
		LanguageMap:       make(map[string]string),
		DefaultIgnore:     []string{},
		BinaryExtensions:  []string{},
		NonCodeExtensions: []string{},
		CustomIgnore: CustomIgnore{
			Patterns: []string{},
			Regex:    []string{},
		},
		Output: Output{
			MaxChars:      50000,
			Compress:      true,
			UltraCompress: false,
			SplitMode:     "char",
			IncludeTree:   true,
			OutputPrefix:  "LLM_CODE",
		},
		Prompts: Prompts{},
	}

	defaultData, err := embeddedFS.ReadFile("default.yaml")
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(defaultData, cfg); err != nil {
		return nil, err
	}

	promptsData, err := embeddedFS.ReadFile("prompts.yaml")
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(promptsData, &cfg.Prompts); err != nil {
		return nil, err
	}

	return cfg, nil
}

func ExportEmbedded(targetPath string) error {
	cfg, err := LoadEmbedded()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	dir := filepath.Dir(targetPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return os.WriteFile(targetPath, data, 0644)
}

func HasEmbedded() bool {
	_, err1 := embeddedFS.ReadFile("default.yaml")
	_, err2 := embeddedFS.ReadFile("prompts.yaml")
	return err1 == nil && err2 == nil
}

func GetEmbeddedRaw(filename string) ([]byte, error) {
	return embeddedFS.ReadFile(filename)
}