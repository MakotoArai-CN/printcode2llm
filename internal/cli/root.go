package cli

import (
	"fmt"
	"os"
	"strings"

	"printcode2llm/internal/config"
	"printcode2llm/internal/generator"
	"printcode2llm/internal/output"
	"printcode2llm/internal/scanner"
	"printcode2llm/internal/ui"

	"github.com/spf13/cobra"
)

var (
	cfg             *config.Config
	projectDirs     []string
	outputPrefix    string
	maxChars        int
	compress        bool
	ultraCompress   bool
	splitMode       string
	includeTree     bool
	excludePatterns string
	regexPatterns   string
	configPath      string
)

var rootCmd = &cobra.Command{
	Use:   "ptlm [项目目录...]",
	Short: "将项目代码整理成适合大模型阅读的格式",
	Long: `PrintCode2LLM (ptlm) - 代码整理工具

将项目代码整理为 Markdown 格式，方便发送给大模型分析。

基本用法:
  ptlm .                     整理当前目录
  ptlm ./project             整理指定目录
  ptlm ./p1 ./p2             整理多个项目
  ptlm -c 80000 .            限制每段字符数
  ptlm -u .                  超级压缩模式

管理命令:
  ptlm config init           生成配置文件
  ptlm install               安装到系统
  ptlm uninstall             卸载
  ptlm version               查看版本`,
	RunE:          runMain,
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	rootCmd.Flags().StringSliceVarP(&projectDirs, "dir", "d", []string{}, "项目目录")
	rootCmd.Flags().StringVarP(&outputPrefix, "output", "o", "", "输出文件前缀")
	rootCmd.Flags().IntVarP(&maxChars, "chars", "c", 0, "每段最大字符数")
	rootCmd.Flags().BoolVar(&compress, "compress", true, "压缩代码")
	rootCmd.Flags().BoolVarP(&ultraCompress, "ultra-compress", "u", false, "超级压缩")
	rootCmd.Flags().StringVarP(&splitMode, "split-mode", "s", "", "分割模式: char/file")
	rootCmd.Flags().BoolVar(&includeTree, "tree", true, "包含目录树")
	rootCmd.Flags().StringVar(&excludePatterns, "exclude", "", "排除模式(逗号分隔)")
	rootCmd.Flags().StringVar(&regexPatterns, "regex", "", "正则排除(逗号分隔)")
	rootCmd.Flags().StringVarP(&configPath, "config", "f", "", "配置文件路径")
}

func Execute() error {
	if len(os.Args) == 1 {
		return rootCmd.Help()
	}
	return rootCmd.Execute()
}

func runMain(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		projectDirs = append(projectDirs, args...)
	}

	if len(projectDirs) == 0 {
		ui.PrintWarning("未指定项目目录")
		ui.PrintInfo("使用示例:")
		ui.PrintInfo("  ptlm .              整理当前目录")
		ui.PrintInfo("  ptlm ./project      整理指定目录")
		return nil
	}

	if configPath != "" {
		config.SetConfigPath(configPath)
	}
	config.SetTargetDirs(projectDirs)

	var err error
	cfg, err = config.Load()
	if err != nil {
		ui.PrintWarning("配置加载失败: %v", err)
		cfg = config.Default()
	}

	if outputPrefix != "" {
		cfg.Output.OutputPrefix = outputPrefix
	}
	if maxChars > 0 {
		cfg.Output.MaxChars = maxChars
	}
	if cmd.Flags().Changed("compress") {
		cfg.Output.Compress = compress
	}
	if ultraCompress {
		cfg.Output.UltraCompress = true
		cfg.Output.Compress = true
	}
	if splitMode != "" {
		cfg.Output.SplitMode = splitMode
	}
	if cmd.Flags().Changed("tree") {
		cfg.Output.IncludeTree = includeTree
	}

	if excludePatterns != "" {
		patterns := strings.Split(excludePatterns, ",")
		for _, p := range patterns {
			p = strings.TrimSpace(p)
			if p != "" {
				cfg.CustomIgnore.Patterns = append(cfg.CustomIgnore.Patterns, p)
			}
		}
	}

	if regexPatterns != "" {
		patterns := strings.Split(regexPatterns, ",")
		for _, p := range patterns {
			p = strings.TrimSpace(p)
			if p != "" {
				cfg.CustomIgnore.Regex = append(cfg.CustomIgnore.Regex, p)
			}
		}
	}

	ui.PrintHeader("PrintCode2LLM")
	ui.PrintInfo("项目数量: %d", len(projectDirs))
	ui.PrintInfo("字符限制: %s", ui.FormatNumber(cfg.Output.MaxChars))
	ui.PrintInfo("压缩模式: %s", getCompressMode(cfg))
	fmt.Println()

	if err := output.CleanOldFiles(cfg.Output.OutputPrefix); err != nil {
		ui.PrintWarning("清理旧文件失败: %v", err)
	}

	allResults := make([]*generator.Result, 0)

	for _, projectDir := range projectDirs {
		ui.PrintSection("处理: %s", projectDir)

		if _, err := os.Stat(projectDir); os.IsNotExist(err) {
			ui.PrintError("目录不存在: %s", projectDir)
			continue
		}

		ui.PrintStep("扫描文件...")
		files, err := scanner.ScanDirectory(projectDir, cfg)
		if err != nil {
			ui.PrintError("扫描失败: %v", err)
			continue
		}
		ui.PrintSuccess("找到 %d 个文件", len(files))

		ui.PrintStep("生成内容...")
		result, err := generator.Generate(projectDir, files, cfg)
		if err != nil {
			ui.PrintError("生成失败: %v", err)
			continue
		}

		allResults = append(allResults, result)
		ui.PrintSuccess("生成 %d 个分段", len(result.Segments))
		fmt.Println()
	}

	if len(allResults) == 0 {
		ui.PrintWarning("没有成功处理任何项目")
		return nil
	}

	ui.PrintSection("写入文件")
	totalSize, err := output.WriteResults(allResults, cfg)
	if err != nil {
		return fmt.Errorf("写入失败: %w", err)
	}

	fmt.Println()
	ui.PrintSuccess("完成!")
	ui.PrintInfo("总大小: %s", ui.FormatBytes(totalSize))

	totalSegments := 0
	for _, r := range allResults {
		totalSegments += len(r.Segments)
	}
	if totalSegments > 1 {
		fmt.Println()
		ui.PrintInfo("共 %d 个文件，请按顺序发送给大模型", totalSegments)
	}

	return nil
}

func getCompressMode(cfg *config.Config) string {
	if !cfg.Output.Compress {
		return "不压缩"
	}
	if cfg.Output.UltraCompress {
		return "超级压缩"
	}
	return "标准压缩"
}