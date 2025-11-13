package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"printcode2llm/internal/ui"

	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "卸载 ptlm",
	Long:  "从系统中移除 ptlm 可执行文件并清理环境变量",
	RunE:  runUninstall,
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}

func runUninstall(cmd *cobra.Command, args []string) error {
	ui.PrintHeader("卸载 ptlm")

	var targetPaths []string
	var targetDir string

	switch runtime.GOOS {
	case "windows":
		userProfile := os.Getenv("USERPROFILE")
		targetDir = filepath.Join(userProfile, "bin")
		targetPaths = []string{
			filepath.Join(targetDir, "ptlm.exe"),
		}

	case "darwin", "linux":
		homeDir, _ := os.UserHomeDir()
		targetPaths = []string{
			"/usr/local/bin/ptlm",
			filepath.Join(homeDir, "bin", "ptlm"),
		}

	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}

	removed := false
	for _, path := range targetPaths {
		if _, err := os.Stat(path); err == nil {
			if err := os.Remove(path); err != nil {
				ui.PrintWarning("删除失败: %s (%v)", path, err)
			} else {
				ui.PrintSuccess("已删除: %s", path)
				removed = true
				if runtime.GOOS == "windows" {
					targetDir = filepath.Dir(path)
				}
			}
		}
	}

	if !removed {
		ui.PrintWarning("未找到已安装的 ptlm")
		return nil
	}

	switch runtime.GOOS {
	case "windows":
		if err := removeFromSystemPath(targetDir); err != nil {
			ui.PrintWarning("清理环境变量失败: %v", err)
			ui.PrintInfo("请手动从环境变量 PATH 中移除: %s", targetDir)
		} else {
			ui.PrintSuccess("已从环境变量 PATH 中移除")
			ui.PrintInfo("请重新打开命令行窗口生效")
		}

	case "darwin", "linux":
		homeDir, _ := os.UserHomeDir()
		binDir := filepath.Join(homeDir, "bin")
		if err := removeFromSystemPath(binDir); err != nil {
			ui.PrintWarning("清理 PATH 配置失败: %v", err)
			ui.PrintInfo("请手动从 ~/.bashrc 或 ~/.zshrc 中删除相关配置")
		} else {
			ui.PrintSuccess("已清理 PATH 配置")
			ui.PrintInfo("请运行: source ~/.bashrc 或重新打开终端")
		}
	}

	fmt.Println()
	ui.PrintSuccess("卸载完成！")

	return nil
}