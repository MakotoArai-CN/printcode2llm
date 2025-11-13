package cli

import (
	"fmt"
	"printcode2llm/internal/ui"
	"printcode2llm/internal/version"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"v"},
	Short:   "显示版本信息",
	Run:     runVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func runVersion(cmd *cobra.Command, args []string) {
	ui.PrintHeader("版本信息")
	ui.PrintInfo("版本: %s", version.Version)
	ui.PrintInfo("仓库: %s", version.Repo)
	fmt.Println()
}