package ui

import (
	"fmt"
)

// PrintBanner 打印 Miku 主题横幅
func PrintBanner() {
	banner := `
    ___       _   _             ____   _     _     __  __ 
   / _  _ __(_)_| |_ _ __     |___  | |   | |   |  /  |
  / /_)/ '__| | ' _| '_  _____  __) || |   | |   | |/| |
 / ___/| |  | | | | | | | |_____/ __/ | |___| |___| |  | |
 /    |_|  |_|_| |_| |_|      |_____|_____|_____|_|  |_|
                                                           
            代码整理工具 · Miku Edition
`

	// 使用 Miku 配色打印
	colorCyan.Print(banner)
	colorPink.Println("            ♫ 让代码传递给 AI 更轻松 ♫")
	fmt.Println()
}
