package cli

import (
	"fmt"
	"os"

	"printcode2llm/configs"
	"printcode2llm/internal/config"
	"printcode2llm/internal/ui"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "配置文件管理",
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "生成默认配置文件",
	RunE:  runConfigInit,
}

var configExportCmd = &cobra.Command{
	Use:   "export [路径]",
	Short: "导出内嵌配置到指定路径",
	Long:  "导出编译时嵌入的配置文件，默认导出到 .ptlm.yaml",
	RunE:  runConfigExport,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "显示当前使用的配置信息",
	RunE:  runConfigShow,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configExportCmd)
	configCmd.AddCommand(configShowCmd)
}

func runConfigInit(cmd *cobra.Command, args []string) error {
	configPath := ".ptlm.yaml"

	if _, err := os.Stat(configPath); err == nil {
		ui.PrintWarning("配置文件已存在: %s", configPath)
		fmt.Print("是否覆盖？(y/N): ")
		var answer string
		fmt.Scanln(&answer)
		if answer != "y" && answer != "Y" {
			ui.PrintInfo("已取消")
			return nil
		}
	}

	if configs.HasEmbedded() {
		if err := configs.ExportEmbedded(configPath); err != nil {
			return fmt.Errorf("导出内嵌配置失败: %w", err)
		}
		ui.PrintSuccess("✓ 从内嵌配置生成")
	} else {
		cfg := config.Default()
		if err := config.Save(cfg, configPath); err != nil {
			return fmt.Errorf("保存配置失败: %w", err)
		}
		ui.PrintWarning("✓ 从默认配置生成（无内嵌配置）")
	}

	ui.PrintSuccess("配置文件: %s", configPath)
	ui.PrintInfo("可编辑此文件自定义配置")

	return nil
}

func runConfigExport(cmd *cobra.Command, args []string) error {
	configPath := ".ptlm.yaml"
	if len(args) > 0 {
		configPath = args[0]
	}

	if _, err := os.Stat(configPath); err == nil {
		ui.PrintWarning("文件已存在: %s", configPath)
		fmt.Print("是否覆盖？(y/N): ")
		var answer string
		fmt.Scanln(&answer)
		if answer != "y" && answer != "Y" {
			ui.PrintInfo("已取消")
			return nil
		}
	}

	if configs.HasEmbedded() {
		if err := configs.ExportEmbedded(configPath); err != nil {
			return fmt.Errorf("导出失败: %w", err)
		}
		ui.PrintSuccess("✓ 已导出内嵌配置")
	} else {
		cfg := config.Default()
		if err := config.Save(cfg, configPath); err != nil {
			return fmt.Errorf("保存失败: %w", err)
		}
		ui.PrintWarning("✓ 已导出默认配置（无内嵌配置）")
	}

	ui.PrintSuccess("文件: %s", configPath)
	return nil
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	ui.PrintHeader("当前配置")

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	fmt.Println()
	if configPath != "" {
		ui.PrintSuccess("📄 来源: 命令行指定 (%s)", configPath)
	} else if _, err := os.Stat(".ptlm.yaml"); err == nil {
		ui.PrintSuccess("📄 来源: 当前目录 .ptlm.yaml")
	} else if configs.HasEmbedded() {
		ui.PrintSuccess("📄 来源: 内嵌配置（configs/ 文件夹）")
	} else {
		ui.PrintWarning("📄 来源: 默认值")
	}

	fmt.Println()
	ui.PrintInfo("输出设置:")
	ui.PrintStep("字符限制: %s", ui.FormatNumber(cfg.Output.MaxChars))
	ui.PrintStep("压缩模式: %v (超级: %v)", cfg.Output.Compress, cfg.Output.UltraCompress)
	ui.PrintStep("分割模式: %s", cfg.Output.SplitMode)
	ui.PrintStep("包含目录树: %v", cfg.Output.IncludeTree)
	ui.PrintStep("输出前缀: %s", cfg.Output.OutputPrefix)

	fmt.Println()
	ui.PrintInfo("规则统计:")
	ui.PrintStep("语言映射: %d 种", len(cfg.LanguageMap))
	ui.PrintStep("默认忽略: %d 项", len(cfg.DefaultIgnore))
	ui.PrintStep("二进制扩展: %d 个", len(cfg.BinaryExtensions))
	ui.PrintStep("非代码扩展: %d 个", len(cfg.NonCodeExtensions))
	if len(cfg.CustomIgnore.Patterns) > 0 || len(cfg.CustomIgnore.Regex) > 0 {
		ui.PrintStep("自定义模式: %d | 正则: %d",
			len(cfg.CustomIgnore.Patterns), len(cfg.CustomIgnore.Regex))
	}

	fmt.Println()
	ui.PrintInfo("提示词配置:")
	ui.PrintStep("章节标题: %s / %s / %s / %s",
		cfg.Prompts.SectionInfo,
		cfg.Prompts.SectionTree,
		cfg.Prompts.SectionCode,
		cfg.Prompts.SectionStats)

	return nil
}