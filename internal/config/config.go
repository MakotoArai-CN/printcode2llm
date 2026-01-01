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
var targetDirs []string

func SetConfigPath(path string) {
	userConfigPath = path
}

func SetTargetDirs(dirs []string) {
	targetDirs = dirs
}

func Load() (*Config, error) {
	if userConfigPath != "" {
		return LoadFrom(userConfigPath)
	}

	if _, err := os.Stat(".ptlm.yaml"); err == nil {
		cfg, err := LoadFrom(".ptlm.yaml")
		if err == nil {
			return cfg, nil
		}
	}

	for _, dir := range targetDirs {
		absDir, err := filepath.Abs(dir)
		if err != nil {
			continue
		}
		cfgPath := filepath.Join(absDir, ".ptlm.yaml")
		if _, err := os.Stat(cfgPath); err == nil {
			cfg, err := LoadFrom(cfgPath)
			if err == nil {
				return cfg, nil
			}
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
			SplitMode:     "char",
			IncludeTree:   true,
			OutputPrefix:  "LLM_CODE",
		},
		Prompts: Prompts{
			SectionInfo:         "项目概况",
			SectionTree:         "目录结构",
			SectionCode:         "源码清单",
			SectionStats:        "统计信息",
			CompressNotice:      "代码已压缩，建议格式化后阅读。",
			UltraCompressNotice: "代码深度压缩，必须格式化后阅读。",
			ContinueNotice:      "内容未完，请查看后续部分。",
			CompleteNotice:      "全部内容已展示完毕。",
			FileInfoFormat:      "**类型**: %s | **行数**: %d | **大小**: %s",
			NonCodeFileNotice:   "配置/文档类型",
			BinaryFileSkip:      "跳过二进制文件",
			StatsTableHeader:    "| 指标 | 数值 |\n|------|------|\n",
		},
	}

	cfg.LanguageMap = map[string]string{
		".js": "javascript", ".jsx": "javascript",
		".ts": "typescript", ".tsx": "typescript",
		".py": "python", ".pyw": "python",
		".go": "go",
		".java": "java",
		".c": "c", ".h": "c",
		".cpp": "cpp", ".hpp": "cpp", ".cc": "cpp", ".cxx": "cpp",
		".rs": "rust",
		".rb": "ruby",
		".php": "php",
		".swift": "swift",
		".kt": "kotlin", ".kts": "kotlin",
		".scala": "scala",
		".dart": "dart",
		".cs": "csharp",
		".vue": "vue",
		".svelte": "svelte",
		".html": "html", ".htm": "html",
		".css": "css",
		".scss": "scss", ".sass": "sass",
		".less": "less",
		".json": "json",
		".xml": "xml",
		".yaml": "yaml", ".yml": "yaml",
		".toml": "toml",
		".ini": "ini",
		".md": "markdown",
		".txt": "text",
		".sh": "shell", ".bash": "shell", ".zsh": "shell",
		".ps1": "powershell",
		".sql": "sql",
		".r": "r", ".R": "r",
		".lua": "lua",
		".pl": "perl",
		".ex": "elixir", ".exs": "elixir",
		".erl": "erlang",
		".hs": "haskell",
		".ml": "ocaml",
		".clj": "clojure",
		".groovy": "groovy",
		".gradle": "gradle",
	}

	cfg.DefaultIgnore = []string{
		"node_modules", "vendor", "venv", ".venv", "env", ".env",
		"__pycache__", ".pytest_cache", ".mypy_cache",
		".git", ".svn", ".hg", ".bzr",
		".vscode", ".idea", ".vs", ".eclipse", "*.code-workspace",
		"dist", "build", "out", "bin", "target", "output",
		"coverage", ".nyc_output", "htmlcov",
		".gradle", ".maven",
		".DS_Store", "Thumbs.db", "desktop.ini",
		"*.log", "logs", "*.tmp", "*.temp", "*.bak", "*.swp", "*.swo",
		"package-lock.json", "yarn.lock", "pnpm-lock.yaml",
		"Gemfile.lock", "Cargo.lock", "go.sum", "composer.lock",
		"LLM_CODE*.md",
		"*.min.js", "*.min.css", "*.map",
		"assets", "static", "public/assets",
	}

	cfg.BinaryExtensions = []string{
		".jpg", ".jpeg", ".png", ".gif", ".bmp", ".svg", ".ico", ".webp",
		".mp4", ".avi", ".mov", ".wmv", ".flv", ".mkv", ".webm",
		".mp3", ".wav", ".ogg", ".flac", ".aac", ".m4a",
		".pdf", ".doc", ".docx", ".ppt", ".pptx", ".xls", ".xlsx",
		".zip", ".tar", ".gz", ".rar", ".7z", ".iso", ".dmg",
		".exe", ".dll", ".so", ".dylib", ".lib", ".a",
		".jar", ".war", ".apk", ".ipa",
		".ttf", ".otf", ".woff", ".woff2", ".eot",
		".db", ".sqlite", ".sqlite3",
		".pyc", ".pyo", ".class", ".o", ".obj", ".wasm",
		".pth", ".pt", ".onnx", ".pb", ".h5", ".pkl", ".pickle",
		".npy", ".npz", ".parquet", ".feather",
	}

	cfg.NonCodeExtensions = []string{
		".md", ".txt", ".log", ".csv",
		".json", ".xml", ".yml", ".yaml", ".toml", ".ini",
		".conf", ".config", ".env", ".env.example",
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

	if len(override.DefaultIgnore) > 0 {
		base.DefaultIgnore = append(base.DefaultIgnore, override.DefaultIgnore...)
	}

	if len(override.BinaryExtensions) > 0 {
		base.BinaryExtensions = append(base.BinaryExtensions, override.BinaryExtensions...)
	}

	if len(override.NonCodeExtensions) > 0 {
		base.NonCodeExtensions = append(base.NonCodeExtensions, override.NonCodeExtensions...)
	}

	if len(override.CustomIgnore.Patterns) > 0 {
		base.CustomIgnore.Patterns = append(base.CustomIgnore.Patterns, override.CustomIgnore.Patterns...)
	}

	if len(override.CustomIgnore.Regex) > 0 {
		base.CustomIgnore.Regex = append(base.CustomIgnore.Regex, override.CustomIgnore.Regex...)
	}

	if override.Output.MaxChars > 0 {
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
		base.Prompts.HeaderPrompt = override.Prompts.HeaderPrompt
	}
	if override.Prompts.SectionInfo != "" {
		base.Prompts.SectionInfo = override.Prompts.SectionInfo
	}
	if override.Prompts.SectionTree != "" {
		base.Prompts.SectionTree = override.Prompts.SectionTree
	}
	if override.Prompts.SectionCode != "" {
		base.Prompts.SectionCode = override.Prompts.SectionCode
	}
	if override.Prompts.SectionStats != "" {
		base.Prompts.SectionStats = override.Prompts.SectionStats
	}
}