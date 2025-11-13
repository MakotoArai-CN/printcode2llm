package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

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

var userConfigPath string

func SetConfigPath(path string) {
	userConfigPath = path
}

func Load() (*Config, error) {
	cfg := Default()

	if userConfigPath != "" {
		return LoadFrom(userConfigPath)
	}

	userCfgPath := ".ptlm.yaml"
	if _, err := os.Stat(userCfgPath); err == nil {
		userCfg, err := LoadFrom(userCfgPath)
		if err != nil {
			return nil, err
		}
		merge(cfg, userCfg)
	}

	return cfg, nil
}

func LoadFrom(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	baseCfg := Default()
	merge(baseCfg, &cfg)

	return baseCfg, nil
}

func Default() *Config {
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
			SplitMode:     "file",
			IncludeTree:   true,
			OutputPrefix:  "LLM_CODE",
		},
		Prompts: Prompts{
			SectionInfo:         "项目概况",
			SectionTree:         "目录结构",
			SectionCode:         "源码清单",
			SectionStats:        "数据统计",
			CompressNotice:      "代码已经过压缩处理，建议使用前先格式化。",
			UltraCompressNotice: "代码已经过深度压缩，使用前必须格式化，否则可读性较差。",
			ContinueNotice:      "当前内容未完，请等待后续部分。",
			CompleteNotice:      "全部内容已展示完毕。",
			FileInfoFormat:      "**类型**: %s | **行数**: %d | **大小**: %s",
			NonCodeFileNotice:   "此文件为配置/文档类型，非可执行代码",
			BinaryFileSkip:      "跳过二进制文件",
		},
	}

	defaultPath := "configs/default.yaml"
	if data, err := os.ReadFile(defaultPath); err == nil {
		yaml.Unmarshal(data, cfg)
	}

	promptsPath := "configs/prompts.yaml"
	if data, err := os.ReadFile(promptsPath); err == nil {
		yaml.Unmarshal(data, &cfg.Prompts)
	}

	return cfg
}

func Save(cfg *Config, path string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return os.WriteFile(path, data, 0644)
}

func merge(base, override *Config) {
	if len(override.LanguageMap) > 0 {
		if base.LanguageMap == nil {
			base.LanguageMap = make(map[string]string)
		}
		for k, v := range override.LanguageMap {
			base.LanguageMap[k] = v
		}
	}

	if len(override.CustomIgnore.Patterns) > 0 {
		base.CustomIgnore.Patterns = override.CustomIgnore.Patterns
	}
	if len(override.CustomIgnore.Regex) > 0 {
		base.CustomIgnore.Regex = override.CustomIgnore.Regex
	}

	if override.Output.MaxChars != 0 {
		base.Output.MaxChars = override.Output.MaxChars
	}
	if override.Output.SplitMode != "" {
		base.Output.SplitMode = override.Output.SplitMode
	}
	if override.Output.OutputPrefix != "" {
		base.Output.OutputPrefix = override.Output.OutputPrefix
	}

	base.Output.Compress = override.Output.Compress
	base.Output.UltraCompress = override.Output.UltraCompress
	base.Output.IncludeTree = override.Output.IncludeTree

	if override.Prompts.HeaderPrompt != "" {
		base.Prompts = override.Prompts
	}
}