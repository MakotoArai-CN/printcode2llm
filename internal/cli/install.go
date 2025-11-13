package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"printcode2llm/internal/ui"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "安装 ptlm 到系统环境变量",
	Long:  "将 ptlm 可执行文件复制到系统路径，并添加到环境变量",
	RunE:  runInstall,
}

func init() {
	rootCmd.AddCommand(installCmd)
}

func runInstall(cmd *cobra.Command, args []string) error {
	ui.PrintHeader("安装 ptlm")

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取可执行文件路径失败: %w", err)
	}

	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return fmt.Errorf("解析符号链接失败: %w", err)
	}

	ui.PrintInfo("当前路径: %s", exePath)

	var targetDir string
	var targetPath string

	switch runtime.GOOS {
	case "windows":
		targetDir = filepath.Join(os.Getenv("USERPROFILE"), "bin")
		targetPath = filepath.Join(targetDir, "ptlm.exe")

		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("创建目录失败: %w", err)
		}

		input, err := os.ReadFile(exePath)
		if err != nil {
			return fmt.Errorf("读取源文件失败: %w", err)
		}

		if err := os.WriteFile(targetPath, input, 0755); err != nil {
			return fmt.Errorf("复制文件失败: %w", err)
		}

		ui.PrintSuccess("文件已复制到: %s", targetPath)

		if err := addToSystemPath(targetDir); err != nil {
			ui.PrintWarning("自动添加环境变量失败: %v", err)
			ui.PrintInfo("请手动添加 %s 到系统 PATH", targetDir)
		} else {
			ui.PrintSuccess("已添加到用户环境变量 PATH")
			ui.PrintInfo("请重新打开命令行窗口生效")
		}

	case "darwin", "linux":
		if os.Geteuid() == 0 {
			targetDir = "/usr/local/bin"
		} else {
			homeDir, _ := os.UserHomeDir()
			targetDir = filepath.Join(homeDir, "bin")
		}

		targetPath = filepath.Join(targetDir, "ptlm")

		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("创建目录失败: %w", err)
		}

		input, err := os.ReadFile(exePath)
		if err != nil {
			return fmt.Errorf("读取源文件失败: %w", err)
		}

		if err := os.WriteFile(targetPath, input, 0755); err != nil {
			return fmt.Errorf("写入文件失败: %w", err)
		}

		ui.PrintSuccess("文件已复制到: %s", targetPath)

		if targetDir != "/usr/local/bin" {
			if err := addToSystemPath(targetDir); err != nil {
				ui.PrintWarning("自动添加 PATH 失败: %v", err)
				ui.PrintInfo("请手动添加以下行到 ~/.bashrc 或 ~/.zshrc:")
				ui.PrintInfo(`export PATH="$HOME/bin:$PATH"`)
			} else {
				ui.PrintSuccess("已添加到 PATH")
				ui.PrintInfo("请运行: source ~/.bashrc 或重新打开终端")
			}
		}

	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}

	fmt.Println()
	ui.PrintSuccess("安装完成！")
	ui.PrintInfo("运行 'ptlm --help' 查看使用说明")

	return nil
}