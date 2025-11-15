package config

import (
	"os"
	"path/filepath"

	"printcode2llm/configs"

	"gopkg.in/yaml.v3"
)

type Config = configs.Config
type CustomIgnore = configs.CustomIgnore
type Output = configs.Output
type Prompts = configs.Prompts

var userConfigPath string

func SetConfigPath(path string) {
	userConfigPath = path
}

func Load() (*Config, error) {
	if userConfigPath != "" {
		return LoadFrom(userConfigPath)
	}

	userCfgPath := ".ptlm.yaml"
	if _, err := os.Stat(userCfgPath); err == nil {
		cfg, err := LoadFrom(userCfgPath)
		if err == nil {
			return cfg, nil
		}
	}

	if configs.HasEmbedded() {
		cfg, err := configs.LoadEmbedded()
		if err == nil {
			return cfg, nil
		}
	}

	return Default(), nil
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

	var baseCfg *Config
	if configs.HasEmbedded() {
		baseCfg, _ = configs.LoadEmbedded()
	}
	if baseCfg == nil {
		baseCfg = Default()
	}

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

	cfg.LanguageMap = map[string]string{
		".js":   "javascript",
		".ts":   "typescript",
		".py":   "python",
		".go":   "go",
		".java": "java",
		".cpp":  "cpp",
		".c":    "c",
		".rs":   "rust",
		".php":  "php",
		".rb":   "ruby",
		".md":   "markdown",
		".json": "json",
		".yaml": "yaml",
		".yml":  "yaml",
	}

	cfg.DefaultIgnore = []string{
		"node_modules",
		"vendor",
		".git",
		".vscode",
		".idea",
		"dist",
		"build",
		"LLM_CODE*.md",
	}

	cfg.BinaryExtensions = []string{
		".exe",
		".dll",
		".so",
		".jpg",
		".png",
		".pdf",
		".zip",
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