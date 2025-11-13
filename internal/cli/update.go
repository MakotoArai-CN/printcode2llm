package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"printcode2llm/internal/ui"
	"printcode2llm/internal/version"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "检查并更新到最新版本",
	RunE:  runUpdate,
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

type GithubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func runUpdate(cmd *cobra.Command, args []string) error {
	ui.PrintHeader("检查更新")
	ui.PrintInfo("当前版本: %s", version.Version)
	ui.PrintStep("检查最新版本...")

	apiURL := strings.Replace(version.Repo, "github.com", "api.github.com/repos", 1) + "/releases/latest"

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return fmt.Errorf("网络请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("获取版本信息失败: HTTP %d", resp.StatusCode)
	}

	var release GithubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("解析版本信息失败: %w", err)
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	ui.PrintInfo("最新版本: %s", latestVersion)

	if latestVersion == version.Version {
		fmt.Println()
		ui.PrintSuccess("当前已是最新版本")
		return nil
	}

	fmt.Println()
	ui.PrintInfo("发现新版本，开始下载...")

	assetName := getAssetName()
	var downloadURL string

	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("未找到适合当前系统的安装包: %s", assetName)
	}

	tmpFile := filepath.Join(os.TempDir(), assetName)
	if err := downloadFile(downloadURL, tmpFile); err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取可执行文件路径失败: %w", err)
	}

	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return fmt.Errorf("解析路径失败: %w", err)
	}

	backupPath := exePath + ".old"
	if err := os.Rename(exePath, backupPath); err != nil {
		return fmt.Errorf("备份当前版本失败: %w", err)
	}

	input, err := os.ReadFile(tmpFile)
	if err != nil {
		os.Rename(backupPath, exePath)
		return fmt.Errorf("读取下载文件失败: %w", err)
	}

	if err := os.WriteFile(exePath, input, 0755); err != nil {
		os.Rename(backupPath, exePath)
		return fmt.Errorf("替换文件失败: %w", err)
	}

	os.Remove(tmpFile)
	os.Remove(backupPath)

	fmt.Println()
	ui.PrintSuccess("更新完成！")
	ui.PrintInfo("新版本: %s", latestVersion)
	ui.PrintInfo("重新运行 ptlm 以使用新版本")

	return nil
}

func getAssetName() string {
	switch runtime.GOOS {
	case "windows":
		return "ptlm.exe"
	case "darwin":
		return "ptlm-darwin"
	case "linux":
		return "ptlm-linux"
	default:
		return "ptlm"
	}
}

func downloadFile(url, filepath string) error {
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}