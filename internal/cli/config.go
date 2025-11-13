package cli

import (
	"fmt"
	"os"

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

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd)
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

	cfg := config.Default()
	if err := config.Save(cfg, configPath); err != nil {
		return fmt.Errorf("保存配置文件失败: %w", err)
	}

	ui.PrintSuccess("配置文件已生成: %s", configPath)
	ui.PrintInfo("可以编辑此文件来自定义配置")

	return nil
}