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
	Short: "生成配置文件",
	RunE:  runConfigInit,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "显示当前配置",
	RunE:  runConfigShow,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd)
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
			return fmt.Errorf("导出配置失败: %w", err)
		}
	} else {
		cfg := config.Default()
		if err := config.Save(cfg, configPath); err != nil {
			return fmt.Errorf("保存配置失败: %w", err)
		}
	}

	ui.PrintSuccess("配置文件已生成: %s", configPath)
	ui.PrintInfo("可以编辑此文件来自定义配置")

	return nil
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	ui.PrintHeader("当前配置")

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	fmt.Println()
	ui.PrintInfo("输出设置:")
	ui.PrintStep("字符限制: %s", ui.FormatNumber(cfg.Output.MaxChars))
	ui.PrintStep("压缩: %v (超级: %v)", cfg.Output.Compress, cfg.Output.UltraCompress)
	ui.PrintStep("分割模式: %s", cfg.Output.SplitMode)
	ui.PrintStep("输出前缀: %s", cfg.Output.OutputPrefix)

	fmt.Println()
	ui.PrintInfo("规则统计:")
	ui.PrintStep("语言映射: %d 种", len(cfg.LanguageMap))
	ui.PrintStep("默认忽略: %d 项", len(cfg.DefaultIgnore))
	ui.PrintStep("二进制扩展: %d 个", len(cfg.BinaryExtensions))

	if len(cfg.CustomIgnore.Patterns) > 0 {
		ui.PrintStep("自定义模式: %d 个", len(cfg.CustomIgnore.Patterns))
	}

	return nil
}