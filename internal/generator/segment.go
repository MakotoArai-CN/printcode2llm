package generator

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"printcode2llm/internal/compress"
	"printcode2llm/internal/config"
	"printcode2llm/internal/scanner"
)

type Segment struct {
	Content   string
	StartFile int
	EndFile   int
	HasMore   bool
}

type Result struct {
	ProjectName string
	ProjectPath string
	Segments    []*Segment
	FileCount   int
	TotalLines  int
	TotalChars  int
	CodeFiles   int
	ConfigFiles int
}

// Generate 生成分段内容
func Generate(projectDir string, files []*scanner.FileInfo, cfg *config.Config) (*Result, error) {
	projectName := filepath.Base(projectDir)

	result := &Result{
		ProjectName: projectName,
		ProjectPath: projectDir,
		Segments:    make([]*Segment, 0),
		FileCount:   len(files),
	}

	// 计算统计信息
	for _, file := range files {
		result.TotalLines += file.LineCount
		result.TotalChars += len(file.Content)
		if file.IsCode {
			result.CodeFiles++
		} else {
			result.ConfigFiles++
		}
	}

	var currentSegment strings.Builder
	currentCharCount := 0
	fileCount := 1
	segmentFileStart := 1

	// 生成头部
	header := generateHeader(projectName, files, result, cfg)
	currentSegment.WriteString(header)
	currentCharCount = len(header)

	// 添加目录树
	if cfg.Output.IncludeTree {
		tree, err := GenerateTree(projectDir, cfg)
		if err == nil {
			treeSection := generateTreeSection(projectName, tree, cfg)
			currentSegment.WriteString(treeSection)
			currentCharCount += len(treeSection)
		}
	}

	// 添加代码区域标题
	codeHeader := generateCodeHeader(cfg)
	currentSegment.WriteString(codeHeader)
	currentCharCount += len(codeHeader)

	// 处理每个文件
	for i, file := range files {
		content := file.Content

		// 压缩处理
		if cfg.Output.Compress && file.IsCode {
			content = compress.Compress(content, file.Language, cfg.Output.UltraCompress)
		}

		// 根据分割模式处理
		if cfg.Output.SplitMode == "line" {
			// 按行分割模式
			if err := processFileByLine(
				file,
				content,
				fileCount,
				&currentSegment,
				&currentCharCount,
				&segmentFileStart,
				result,
				cfg,
			); err != nil {
				return nil, fmt.Errorf("处理文件 %s 失败: %w", file.RelPath, err)
			}
		} else {
			// 按文件分割模式（默认）
			fileBlock := generateFileBlock(fileCount, file, content, cfg)
			fileBlockLen := len(fileBlock)

			// 判断是否需要分割
			if currentCharCount+fileBlockLen > cfg.Output.MaxChars && currentCharCount > 0 {
				// 保存当前分段
				continueNotice := generateContinueNotice(len(files)-i, cfg)
				currentSegment.WriteString(continueNotice)

				result.Segments = append(result.Segments, &Segment{
					Content:   currentSegment.String(),
					StartFile: segmentFileStart,
					EndFile:   fileCount - 1,
					HasMore:   true,
				})

				// 开始新段
				currentSegment.Reset()
				newHeader := generateContinuationHeader(projectName, len(result.Segments)+1, cfg)
				currentSegment.WriteString(newHeader)
				currentCharCount = len(newHeader)
				segmentFileStart = fileCount
			}

			currentSegment.WriteString(fileBlock)
			currentCharCount += fileBlockLen
		}

		fileCount++
	}

	// 添加最后一段
	if currentCharCount > 0 {
		footer := generateFooter(result, cfg)
		currentSegment.WriteString(footer)

		result.Segments = append(result.Segments, &Segment{
			Content:   currentSegment.String(),
			StartFile: segmentFileStart,
			EndFile:   fileCount - 1,
			HasMore:   false,
		})
	}

	return result, nil
}

// processFileByLine 按行分割处理文件
func processFileByLine(
	file *scanner.FileInfo,
	content string,
	fileCount int,
	currentSegment *strings.Builder,
	currentCharCount *int,
	segmentFileStart *int,
	result *Result,
	cfg *config.Config,
) error {
	lines := strings.Split(content, "\n")
	var currentFileContent strings.Builder
	lineStart := 0

	for j := 0; j < len(lines); j++ {
		line := lines[j]
		if j < len(lines)-1 {
			line += "\n"
		}

		// 检查是否会超出限制
		testLen := *currentCharCount + currentFileContent.Len() + len(line)
		if testLen > cfg.Output.MaxChars && *currentCharCount > 0 {
			// 保存当前文件的部分内容
			if currentFileContent.Len() > 0 {
				partialBlock := generatePartialFileBlock(
					fileCount,
					file,
					currentFileContent.String(),
					lineStart+1,
					j,
					cfg,
				)
				currentSegment.WriteString(partialBlock)
			}

			// 保存当前分段
			continueNotice := generateContinueNotice(0, cfg)
			currentSegment.WriteString(continueNotice)

			result.Segments = append(result.Segments, &Segment{
				Content:   currentSegment.String(),
				StartFile: *segmentFileStart,
				EndFile:   fileCount,
				HasMore:   true,
			})

			// 开始新段
			currentSegment.Reset()
			newHeader := generateContinuationHeader(result.ProjectName, len(result.Segments)+1, cfg)
			currentSegment.WriteString(newHeader)
			*currentCharCount = len(newHeader)
			currentFileContent.Reset()
			lineStart = j
			*segmentFileStart = fileCount
		}

		currentFileContent.WriteString(line)
	}

	// 添加剩余的文件内容
	if currentFileContent.Len() > 0 {
		var fileBlock string
		if lineStart > 0 {
			// 这是文件的一部分
			fileBlock = generatePartialFileBlock(
				fileCount,
				file,
				currentFileContent.String(),
				lineStart+1,
				len(lines),
				cfg,
			)
		} else {
			// 这是完整的文件
			fileBlock = generateFileBlock(fileCount, file, currentFileContent.String(), cfg)
		}

		currentSegment.WriteString(fileBlock)
		*currentCharCount += len(fileBlock)
	}

	return nil
}

// generateHeader 生成文档头部
func generateHeader(projectName string, files []*scanner.FileInfo, result *Result, cfg *config.Config) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("# %s\n\n", projectName))

	// 添加提示词
	if cfg.Prompts.HeaderPrompt != "" {
		builder.WriteString(cfg.Prompts.HeaderPrompt)
		builder.WriteString("\n\n")
	}

	// 项目信息
	builder.WriteString(fmt.Sprintf("## %s\n\n", cfg.Prompts.SectionInfo))
	builder.WriteString(fmt.Sprintf("- **项目名称**: %s\n", projectName))
	builder.WriteString(fmt.Sprintf("- **生成时间**: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	builder.WriteString(fmt.Sprintf("- **文件总数**: %d 个（代码: %d, 配置: %d）\n",
		len(files), result.CodeFiles, result.ConfigFiles))
	builder.WriteString(fmt.Sprintf("- **总行数**: %s 行\n", formatNumber(result.TotalLines)))
	builder.WriteString(fmt.Sprintf("- **总字符数**: %s 字符\n", formatNumber(result.TotalChars)))

	// 压缩信息
	if cfg.Output.Compress {
		compressMode := "标准压缩"
		notice := cfg.Prompts.CompressNotice
		if cfg.Output.UltraCompress {
			compressMode = "超级压缩"
			notice = cfg.Prompts.UltraCompressNotice
		}
		builder.WriteString(fmt.Sprintf("- **压缩模式**: %s\n", compressMode))
		if notice != "" {
			builder.WriteString(fmt.Sprintf("\n> ⚠️ **注意**: %s\n", notice))
		}
	}

	builder.WriteString("\n")

	return builder.String()
}

// generateTreeSection 生成目录树部分
func generateTreeSection(projectName, tree string, cfg *config.Config) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("## %s\n\n", cfg.Prompts.SectionTree))
	builder.WriteString("```tree\n")
	builder.WriteString(fmt.Sprintf("%s/\n%s", projectName, tree))
	builder.WriteString("```\n\n")

	return builder.String()
}

// generateCodeHeader 生成代码区域头部
func generateCodeHeader(cfg *config.Config) string {
	return fmt.Sprintf("## %s\n\n", cfg.Prompts.SectionCode)
}

// generateContinuationHeader 生成续页头部
func generateContinuationHeader(projectName string, partNum int, cfg *config.Config) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("# %s (续 - 第 %d 部分)\n\n", projectName, partNum))
	builder.WriteString(fmt.Sprintf("> 📌 **说明**: 这是第 %d 部分，继续展示项目代码\n\n", partNum))

	// 添加压缩提示
	if cfg.Output.Compress {
		notice := cfg.Prompts.CompressNotice
		if cfg.Output.UltraCompress {
			notice = cfg.Prompts.UltraCompressNotice
		}
		if notice != "" {
			builder.WriteString(fmt.Sprintf("> ⚠️ **注意**: %s\n\n", notice))
		}
	}

	builder.WriteString(fmt.Sprintf("## %s（续）\n\n", cfg.Prompts.SectionCode))

	return builder.String()
}

// generateContinueNotice 生成继续提示
func generateContinueNotice(remainingFiles int, cfg *config.Config) string {
	var builder strings.Builder

	builder.WriteString("\n---\n\n")

	if remainingFiles > 0 {
		builder.WriteString(fmt.Sprintf("> 📋 **提示**: 还有 %d 个文件未展示，内容将在下一部分继续\n", remainingFiles))
	} else {
		builder.WriteString(fmt.Sprintf("> 📋 **提示**: %s\n", cfg.Prompts.ContinueNotice))
	}

	builder.WriteString("\n")

	return builder.String()
}

// generateFileBlock 生成完整文件块
func generateFileBlock(fileNum int, file *scanner.FileInfo, content string, cfg *config.Config) string {
	var builder strings.Builder

	// 文件标题
	builder.WriteString(fmt.Sprintf("### %d. %s\n\n", fileNum, file.RelPath))

	// 文件信息
	if !cfg.Output.Compress {
		sizeStr := formatSize(file.Size)
		fileType := file.Language
		if !file.IsCode {
			fileType += " (配置)"
		}

		info := fmt.Sprintf(cfg.Prompts.FileInfoFormat, fileType, file.LineCount, sizeStr)
		builder.WriteString(fmt.Sprintf("> %s\n", info))

		if !file.IsCode && cfg.Prompts.NonCodeFileNotice != "" {
			builder.WriteString(fmt.Sprintf("> %s\n", cfg.Prompts.NonCodeFileNotice))
		}

		builder.WriteString("\n")
	}

	// 代码块
	builder.WriteString(fmt.Sprintf("```%s\n", file.Language))
	builder.WriteString(content)
	if !strings.HasSuffix(content, "\n") {
		builder.WriteString("\n")
	}
	builder.WriteString("```\n\n")

	return builder.String()
}

// generatePartialFileBlock 生成部分文件块（按行分割时使用）
func generatePartialFileBlock(fileNum int, file *scanner.FileInfo, content string, lineStart, lineEnd int, cfg *config.Config) string {
	var builder strings.Builder

	// 文件标题（包含行号范围）
	builder.WriteString(fmt.Sprintf("### %d. %s (行 %d-%d)\n\n", fileNum, file.RelPath, lineStart, lineEnd))

	// 文件信息
	if !cfg.Output.Compress {
		fileType := file.Language
		if !file.IsCode {
			fileType += " (配置)"
		}

		actualLines := strings.Count(content, "\n") + 1
		info := fmt.Sprintf("**类型**: %s | **部分**: 行 %d 至 %d | **行数**: %d",
			fileType, lineStart, lineEnd, actualLines)
		builder.WriteString(fmt.Sprintf("> %s\n", info))

		if !file.IsCode && cfg.Prompts.NonCodeFileNotice != "" {
			builder.WriteString(fmt.Sprintf("> %s\n", cfg.Prompts.NonCodeFileNotice))
		}

		builder.WriteString("\n")
	}

	// 代码块
	builder.WriteString(fmt.Sprintf("```%s\n", file.Language))
	builder.WriteString(content)
	if !strings.HasSuffix(content, "\n") {
		builder.WriteString("\n")
	}
	builder.WriteString("```\n\n")

	return builder.String()
}

// generateFooter 生成文档尾部
func generateFooter(result *Result, cfg *config.Config) string {
	var builder strings.Builder

	builder.WriteString("\n---\n\n")
	builder.WriteString(fmt.Sprintf("## %s\n\n", cfg.Prompts.SectionStats))

	// 完成提示
	if cfg.Prompts.CompleteNotice != "" {
		builder.WriteString(fmt.Sprintf("✅ **%s**\n\n", cfg.Prompts.CompleteNotice))
	}

	// 统计表格
	builder.WriteString("### 项目统计\n\n")
	builder.WriteString(cfg.Prompts.StatsTableHeader)
	builder.WriteString(fmt.Sprintf("| 文件总数 | %d |\n", result.FileCount))
	builder.WriteString(fmt.Sprintf("| 代码文件 | %d |\n", result.CodeFiles))
	builder.WriteString(fmt.Sprintf("| 配置文件 | %d |\n", result.ConfigFiles))
	builder.WriteString(fmt.Sprintf("| 总行数 | %s |\n", formatNumber(result.TotalLines)))
	builder.WriteString(fmt.Sprintf("| 总字符数 | %s |\n", formatNumber(result.TotalChars)))

	// 添加分段信息
	if len(result.Segments) > 0 {
		builder.WriteString(fmt.Sprintf("| 分段数量 | %d |\n", len(result.Segments)))
	}

	builder.WriteString("\n")

	// 添加使用说明（如果有多个分段）
	if len(result.Segments) > 1 && cfg.Prompts.UsageInstructions != "" {
		instructions := fmt.Sprintf(cfg.Prompts.UsageInstructions, len(result.Segments))
		builder.WriteString(instructions)
		builder.WriteString("\n")
	}

	return builder.String()
}

// formatSize 格式化文件大小
func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// formatNumber 格式化数字（添加千位分隔符）
func formatNumber(n int) string {
	str := fmt.Sprintf("%d", n)
	if len(str) <= 3 {
		return str
	}

	var result []byte
	for i, ch := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(ch))
	}

	return string(result)
}
