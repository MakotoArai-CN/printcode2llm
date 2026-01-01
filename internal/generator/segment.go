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
	PartNum   int
	TotalPart int
	CharCount int
	FileRange string
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

type fileBlock struct {
	fileNum   int
	file      *scanner.FileInfo
	content   string
	lines     []string
	startLine int
	endLine   int
}

func Generate(projectDir string, files []*scanner.FileInfo, cfg *config.Config) (*Result, error) {
	projectName := filepath.Base(projectDir)

	result := &Result{
		ProjectName: projectName,
		ProjectPath: projectDir,
		Segments:    make([]*Segment, 0),
		FileCount:   len(files),
	}

	for _, file := range files {
		result.TotalLines += file.LineCount
		result.TotalChars += len(file.Content)
		if file.IsCode {
			result.CodeFiles++
		} else {
			result.ConfigFiles++
		}
	}

	var allBlocks []fileBlock
	for i, file := range files {
		content := file.Content
		if cfg.Output.Compress && file.IsCode {
			content = compress.Compress(content, file.Language, cfg.Output.UltraCompress)
		}

		lines := strings.Split(content, "\n")
		allBlocks = append(allBlocks, fileBlock{
			fileNum:   i + 1,
			file:      file,
			content:   content,
			lines:     lines,
			startLine: 1,
			endLine:   len(lines),
		})
	}

	headerTemplate := generateHeaderTemplate(projectName, result, cfg)
	treeSection := ""
	if cfg.Output.IncludeTree {
		tree, err := GenerateTree(projectDir, cfg)
		if err == nil {
			treeSection = generateTreeSection(projectName, tree, cfg)
		}
	}

	firstPartHeader := headerTemplate + treeSection + fmt.Sprintf("## %s\n\n", cfg.Prompts.SectionCode)
	contHeader := func(partNum int) string {
		return generateContinuationHeader(projectName, partNum, cfg)
	}

	maxChars := cfg.Output.MaxChars
	segments := splitBlocksIntoSegments(allBlocks, maxChars, firstPartHeader, contHeader, cfg)

	totalParts := len(segments)
	for i, seg := range segments {
		seg.PartNum = i + 1
		seg.TotalPart = totalParts

		if i == totalParts-1 {
			seg.Content += generateFooter(result, totalParts, cfg)
		}
	}

	result.Segments = segments
	return result, nil
}

func splitBlocksIntoSegments(blocks []fileBlock, maxChars int, firstHeader string, contHeader func(int) string, cfg *config.Config) []*Segment {
	var segments []*Segment
	var currentBuilder strings.Builder
	currentChars := 0
	partNum := 1

	header := firstHeader
	currentBuilder.WriteString(header)
	currentChars = len(header)

	blockIdx := 0
	lineIdx := 0

	for blockIdx < len(blocks) {
		block := blocks[blockIdx]
		file := block.file
		fileNum := block.fileNum
		lines := block.lines

		for lineIdx < len(lines) {
			remainingLines := lines[lineIdx:]
			startLine := lineIdx + 1

			blockContent := buildFileBlockContent(fileNum, file, remainingLines, startLine, startLine+len(remainingLines)-1, lineIdx == 0, cfg)
			blockLen := len(blockContent)

			if currentChars+blockLen <= maxChars {
				currentBuilder.WriteString(blockContent)
				currentChars += blockLen
				lineIdx = len(lines)
			} else {
				availableChars := maxChars - currentChars
				if availableChars < 500 {
					segments = append(segments, &Segment{
						Content:   currentBuilder.String() + generateContinueNotice(cfg),
						CharCount: currentChars,
					})

					partNum++
					currentBuilder.Reset()
					header = contHeader(partNum)
					currentBuilder.WriteString(header)
					currentChars = len(header)
					continue
				}

				linesForThisPart, charsUsed := fitLinesIntoChars(fileNum, file, remainingLines, startLine, availableChars, lineIdx == 0, cfg)

				if linesForThisPart == 0 {
					segments = append(segments, &Segment{
						Content:   currentBuilder.String() + generateContinueNotice(cfg),
						CharCount: currentChars,
					})

					partNum++
					currentBuilder.Reset()
					header = contHeader(partNum)
					currentBuilder.WriteString(header)
					currentChars = len(header)
					continue
				}

				partialContent := buildFileBlockContent(fileNum, file, remainingLines[:linesForThisPart], startLine, startLine+linesForThisPart-1, lineIdx == 0, cfg)
				currentBuilder.WriteString(partialContent)
				currentChars += charsUsed

				segments = append(segments, &Segment{
					Content:   currentBuilder.String() + generateContinueNotice(cfg),
					CharCount: currentChars,
				})

				partNum++
				currentBuilder.Reset()
				header = contHeader(partNum)
				currentBuilder.WriteString(header)
				currentChars = len(header)

				lineIdx += linesForThisPart
			}
		}

		blockIdx++
		lineIdx = 0
	}

	if currentChars > len(contHeader(partNum)) {
		segments = append(segments, &Segment{
			Content:   currentBuilder.String(),
			CharCount: currentChars,
		})
	}

	return segments
}

func fitLinesIntoChars(fileNum int, file *scanner.FileInfo, lines []string, startLine, availableChars int, isStart bool, cfg *config.Config) (int, int) {
	if len(lines) == 0 {
		return 0, 0
	}

	headerSize := estimateBlockHeaderSize(fileNum, file, startLine, startLine, isStart, cfg)
	footerSize := 10

	usableChars := availableChars - headerSize - footerSize
	if usableChars <= 0 {
		return 0, 0
	}

	charCount := 0
	lineCount := 0

	for i, line := range lines {
		lineLen := len(line) + 1
		if charCount+lineLen > usableChars {
			break
		}
		charCount += lineLen
		lineCount = i + 1
	}

	if lineCount == 0 && len(lines) > 0 {
		lineCount = 1
		charCount = len(lines[0]) + 1
	}

	totalChars := headerSize + charCount + footerSize
	return lineCount, totalChars
}

func estimateBlockHeaderSize(fileNum int, file *scanner.FileInfo, startLine, endLine int, isStart bool, cfg *config.Config) int {
	var builder strings.Builder

	if isStart {
		builder.WriteString(fmt.Sprintf("### %d. %s\n\n", fileNum, file.RelPath))
	} else {
		builder.WriteString(fmt.Sprintf("### %d. %s (续: 行 %d-%d)\n\n", fileNum, file.RelPath, startLine, endLine))
	}

	builder.WriteString(fmt.Sprintf("```%s\n", file.Language))

	return builder.Len()
}

func buildFileBlockContent(fileNum int, file *scanner.FileInfo, lines []string, startLine, endLine int, isStart bool, cfg *config.Config) string {
	var builder strings.Builder

	if isStart {
		if startLine == 1 && endLine >= len(strings.Split(file.Content, "\n")) {
			builder.WriteString(fmt.Sprintf("### %d. %s\n\n", fileNum, file.RelPath))
		} else {
			builder.WriteString(fmt.Sprintf("### %d. %s (行 %d-%d)\n\n", fileNum, file.RelPath, startLine, endLine))
		}
	} else {
		builder.WriteString(fmt.Sprintf("### %d. %s (续: 行 %d-%d)\n\n", fileNum, file.RelPath, startLine, endLine))
	}

	builder.WriteString(fmt.Sprintf("```%s\n", file.Language))
	builder.WriteString(strings.Join(lines, "\n"))
	if len(lines) > 0 && !strings.HasSuffix(lines[len(lines)-1], "\n") {
		builder.WriteString("\n")
	}
	builder.WriteString("```\n\n")

	return builder.String()
}

func generateHeaderTemplate(projectName string, result *Result, cfg *config.Config) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("# %s\n\n", projectName))

	if cfg.Prompts.HeaderPrompt != "" {
		builder.WriteString(cfg.Prompts.HeaderPrompt)
		builder.WriteString("\n\n")
	}

	builder.WriteString(fmt.Sprintf("## %s\n\n", cfg.Prompts.SectionInfo))
	builder.WriteString(fmt.Sprintf("- **项目**: %s\n", projectName))
	builder.WriteString(fmt.Sprintf("- **时间**: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	builder.WriteString(fmt.Sprintf("- **文件**: %d (代码: %d, 配置: %d)\n", result.FileCount, result.CodeFiles, result.ConfigFiles))
	builder.WriteString(fmt.Sprintf("- **行数**: %s\n", formatNumber(result.TotalLines)))
	builder.WriteString(fmt.Sprintf("- **字符**: %s\n", formatNumber(result.TotalChars)))

	if cfg.Output.Compress {
		mode := "标准"
		if cfg.Output.UltraCompress {
			mode = "深度"
		}
		builder.WriteString(fmt.Sprintf("- **压缩**: %s\n", mode))
	}

	builder.WriteString("\n")

	return builder.String()
}

func generateTreeSection(projectName, tree string, cfg *config.Config) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("## %s\n\n", cfg.Prompts.SectionTree))
	builder.WriteString("```\n")
	builder.WriteString(fmt.Sprintf("%s/\n%s", projectName, tree))
	builder.WriteString("```\n\n")

	return builder.String()
}

func generateContinuationHeader(projectName string, partNum int, cfg *config.Config) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("# %s (第 %d 部分)\n\n", projectName, partNum))
	builder.WriteString(fmt.Sprintf("> 第 %d 部分，接续上文\n\n", partNum))

	if cfg.Output.Compress && cfg.Output.UltraCompress {
		builder.WriteString("> ⚠️ 代码已深度压缩\n\n")
	}

	builder.WriteString(fmt.Sprintf("## %s (续)\n\n", cfg.Prompts.SectionCode))

	return builder.String()
}

func generateContinueNotice(cfg *config.Config) string {
	return "\n---\n\n> 📋 内容续下一部分\n\n"
}

func generateFooter(result *Result, totalParts int, cfg *config.Config) string {
	var builder strings.Builder

	builder.WriteString("\n---\n\n")
	builder.WriteString(fmt.Sprintf("## %s\n\n", cfg.Prompts.SectionStats))
	builder.WriteString("✅ **全部内容已展示**\n\n")

	builder.WriteString("| 指标 | 数值 |\n")
	builder.WriteString("|------|------|\n")
	builder.WriteString(fmt.Sprintf("| 文件总数 | %d |\n", result.FileCount))
	builder.WriteString(fmt.Sprintf("| 代码文件 | %d |\n", result.CodeFiles))
	builder.WriteString(fmt.Sprintf("| 配置文件 | %d |\n", result.ConfigFiles))
	builder.WriteString(fmt.Sprintf("| 总行数 | %s |\n", formatNumber(result.TotalLines)))
	builder.WriteString(fmt.Sprintf("| 总字符 | %s |\n", formatNumber(result.TotalChars)))

	if totalParts > 1 {
		builder.WriteString(fmt.Sprintf("| 分段数 | %d |\n", totalParts))
	}

	builder.WriteString("\n")

	return builder.String()
}

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