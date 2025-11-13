package main

import (
	"os"

	"printcode2llm/internal/cli"
	"printcode2llm/internal/ui"
)

func main() {
	// 显示 Miku 横幅
	ui.PrintBanner()

	// 执行命令
	if err := cli.Execute(); err != nil {
		ui.PrintError("执行失败: %v", err)
		os.Exit(1)
	}
}
