//go:build !windows
// +build !windows

package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"printcode2llm/internal/ui"
)

// addToSystemPath Unix/Linux 系统添加到 PATH（通过修改 shell 配置文件）
func addToSystemPath(dir string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	shells := []string{".bashrc", ".zshrc", ".profile"}
	exportLine := `export PATH="$HOME/bin:$PATH"`

	for _, shell := range shells {
		shellPath := filepath.Join(homeDir, shell)
		if _, err := os.Stat(shellPath); os.IsNotExist(err) {
			continue
		}

		content, err := os.ReadFile(shellPath)
		if err != nil {
			continue
		}

		if strings.Contains(string(content), exportLine) {
			ui.PrintInfo("PATH 已在 %s 中配置", shell)
			return nil
		}

		f, err := os.OpenFile(shellPath, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			continue
		}
		defer f.Close()

		_, err = f.WriteString("\n# Added by ptlm\n" + exportLine + "\n")
		if err != nil {
			continue
		}

		ui.PrintSuccess("已添加到 %s", shell)
		return nil
	}

	return fmt.Errorf("未找到 shell 配置文件")
}

// removeFromSystemPath Unix/Linux 系统从 PATH 移除（从 shell 配置文件删除）
func removeFromSystemPath(dir string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	shells := []string{".bashrc", ".zshrc", ".profile"}
	removed := false

	for _, shell := range shells {
		shellPath := filepath.Join(homeDir, shell)
		if _, err := os.Stat(shellPath); os.IsNotExist(err) {
			continue
		}

		content, err := os.ReadFile(shellPath)
		if err != nil {
			continue
		}

		lines := strings.Split(string(content), "\n")
		newLines := make([]string, 0)
		skip := false

		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "# Added by ptlm" {
				skip = true
				continue
			}
			if skip && strings.Contains(line, "export PATH") && strings.Contains(line, "$HOME/bin") {
				skip = false
				removed = true
				continue
			}
			if strings.Contains(line, "export PATH") &&
				(strings.Contains(line, filepath.Join(homeDir, "bin")) || strings.Contains(line, "$HOME/bin")) &&
				(strings.Contains(line, "ptlm") || strings.Contains(line, "Added by ptlm")) {
				removed = true
				continue
			}

			newLines = append(newLines, line)
			skip = false
		}

		if removed {
			newContent := strings.Join(newLines, "\n")
			if err := os.WriteFile(shellPath, []byte(newContent), 0644); err != nil {
				return err
			}
			ui.PrintSuccess("已从 %s 中移除配置", shell)
		}
	}

	if !removed {
		return fmt.Errorf("未找到 PATH 配置")
	}

	return nil
}