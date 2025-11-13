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

将项目代码打包成 Markdown 格式，方便发送给大模型进行分析。

特性：
  • 智能代码压缩
  • 多项目支持
  • 灵活的过滤规则
  • 详细的统计信息

基本用法：
  ptlm ./myproject           整理指定目录
  ptlm ./proj1 ./proj2       整理多个项目
  ptlm -c 30000 ./src        限制每段字符数
  ptlm -u ./project          使用超级压缩模式

管理命令：
  ptlm install               安装到系统
  ptlm uninstall             从系统卸载
  ptlm update                检查更新
  ptlm version               查看版本`,
	RunE: runMain,
	// 当没有参数时显示帮助
	Args: func(cmd *cobra.Command, args []string) error {
		// 如果没有参数且没有设置任何 flag
		if len(args) == 0 && !cmd.Flags().Changed("dir") {
			return fmt.Errorf("请指定项目目录")
		}
		return nil
	},
	// 自定义错误处理
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	rootCmd.Flags().StringSliceVarP(&projectDirs, "dir", "d", []string{}, "项目目录（可多个）")
	rootCmd.Flags().StringVarP(&outputPrefix, "output", "o", "", "输出文件前缀")
	rootCmd.Flags().IntVarP(&maxChars, "chars", "c", 0, "每段最大字符数")
	rootCmd.Flags().BoolVar(&compress, "compress", true, "压缩代码")
	rootCmd.Flags().BoolVarP(&ultraCompress, "ultra-compress", "u", false, "超级压缩模式")
	rootCmd.Flags().StringVarP(&splitMode, "split-mode", "s", "", "分割模式: file/line")
	rootCmd.Flags().BoolVar(&includeTree, "tree", true, "包含目录树")
	rootCmd.Flags().StringVar(&excludePatterns, "exclude", "", "排除模式（逗号分隔）")
	rootCmd.Flags().StringVar(&regexPatterns, "regex", "", "正则排除模式（逗号分隔）")
	rootCmd.Flags().StringVarP(&configPath, "config", "f", "", "指定配置文件路径")

	// 隐藏不常用的 flag
	rootCmd.Flags().MarkHidden("multi")
}

func Execute() error {
	// 如果没有任何参数和子命令，显示帮助
	if len(os.Args) == 1 {
		return rootCmd.Help()
	}

	err := rootCmd.Execute()
	if err != nil {
		// 检查是否是参数错误
		if strings.Contains(err.Error(), "请指定项目目录") {
			ui.PrintWarning("未指定项目目录")
			fmt.Println()
			ui.PrintInfo("使用示例:")
			ui.PrintInfo("  ptlm ./myproject          # 整理指定目录")
			ui.PrintInfo("  ptlm .                    # 整理当前目录")
			ui.PrintInfo("  ptlm -h                   # 查看完整帮助")
			fmt.Println()
			return nil
		}
		return err
	}
	return nil
}

func runMain(cmd *cobra.Command, args []string) error {
	var err error

	if configPath != "" {
		config.SetConfigPath(configPath)
	}

	cfg, err = config.Load()
	if err != nil {
		ui.PrintWarning("配置加载失败，使用默认配置: %v", err)
		cfg = config.Default()
	}

	// 处理项目目录参数
	if len(args) > 0 {
		projectDirs = args
	} else if cmd.Flags().Changed("dir") {
		// 使用 -d 指定的目录
	} else {
		// 不应该到达这里（Args 验证会拦截）
		return fmt.Errorf("请指定项目目录")
	}

	// 应用命令行参数覆盖配置
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

	// 处理排除模式
	if excludePatterns != "" {
		patterns := strings.Split(excludePatterns, ",")
		cfg.CustomIgnore.Patterns = append(cfg.CustomIgnore.Patterns, patterns...)
	}
	if regexPatterns != "" {
		patterns := strings.Split(regexPatterns, ",")
		cfg.CustomIgnore.Regex = append(cfg.CustomIgnore.Regex, patterns...)
	}

	// 开始处理
	ui.PrintHeader("项目代码整理工具")
	ui.PrintInfo("项目数量: %d", len(projectDirs))
	ui.PrintInfo("字符限制: %s", ui.FormatNumber(cfg.Output.MaxChars))
	ui.PrintInfo("压缩模式: %s", getCompressMode(cfg))
	ui.PrintInfo("分割模式: %s", cfg.Output.SplitMode)
	fmt.Println()

	// 清理旧文件
	if err := output.CleanOldFiles(cfg.Output.OutputPrefix); err != nil {
		ui.PrintWarning("清理旧文件失败: %v", err)
	}

	// 处理所有项目
	allResults := make([]*generator.Result, 0)
	for _, projectDir := range projectDirs {
		ui.PrintSection("处理项目: %s", projectDir)

		// 检查目录是否存在
		if _, err := os.Stat(projectDir); os.IsNotExist(err) {
			ui.PrintError("目录不存在: %s", projectDir)
			continue
		}

		// 扫描文件
		ui.PrintStep("扫描文件...")
		files, err := scanner.ScanDirectory(projectDir, cfg)
		if err != nil {
			ui.PrintError("扫描失败: %v", err)
			continue
		}
		ui.PrintSuccess("找到 %d 个文件", len(files))

		// 生成内容
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

	// 检查是否有成功的结果
	if len(allResults) == 0 {
		ui.PrintWarning("没有成功处理任何项目")
		return nil
	}

	// 写入文件
	ui.PrintSection("写入文件")
	totalSize, err := output.WriteResults(allResults, cfg)
	if err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	// 显示完成信息
	fmt.Println()
	ui.PrintSuccess("全部完成！")
	ui.PrintInfo("总输出大小: %s", ui.FormatBytes(totalSize))

	totalSegments := 0
	for _, r := range allResults {
		totalSegments += len(r.Segments)
	}

	if totalSegments > 1 {
		fmt.Println()
		ui.PrintSection("使用说明")
		ui.PrintInfo("共生成 %d 个文件", totalSegments)
		ui.PrintInfo("请按顺序发送给大模型")
		ui.PrintInfo("建议在对话开始时说明这是分段发送的代码")
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