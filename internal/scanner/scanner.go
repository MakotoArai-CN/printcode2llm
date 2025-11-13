package scanner

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"printcode2llm/internal/config"
)

type FileInfo struct {
	Path       string
	RelPath    string
	Language   string
	Content    string
	IsCode     bool
	IsBinary   bool
	HasNewline bool
	LineCount  int
	Size       int64
	Encoding   string
}

// ScanDirectory 扫描目录
func ScanDirectory(dir string, cfg *config.Config) ([]*FileInfo, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("获取绝对路径失败: %w", err)
	}

	var files []*FileInfo
	ignoreChecker := NewIgnoreChecker(cfg)

	err = filepath.Walk(absDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// 记录错误但继续处理
			return nil
		}

		relPath, err := filepath.Rel(absDir, path)
		if err != nil {
			return nil
		}

		if relPath == "." {
			return nil
		}

		// 检查是否应该忽略
		if ignoreChecker.ShouldIgnore(path, info.IsDir()) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// 只处理文件
		if info.IsDir() {
			return nil
		}

		// 检查文件大小（跳过过大的文件）
		if info.Size() > 10*1024*1024 { // 10MB
			return nil
		}

		// 检查是否是二进制文件（通过扩展名）
		if isBinaryExtension(path, cfg) {
			return nil
		}

		// 读取文件内容
		content, err := os.ReadFile(path)
		if err != nil {
			// 无法读取的文件跳过
			return nil
		}

		// 检测是否是二进制文件（通过内容）
		if isBinaryContent(content) {
			return nil
		}

		// 转换为字符串
		contentStr := string(content)

		// 检测编码
		encoding := detectEncoding(content)

		// 检测换行符
		hasNewline := detectNewline(contentStr)

		// 获取语言类型
		ext := strings.ToLower(filepath.Ext(path))
		language := cfg.LanguageMap[ext]
		if language == "" {
			language = "text"
		}

		// 判断是否是代码文件
		isCode := !isNonCodeFile(ext, cfg)

		// 统计行数
		lineCount := countLines(contentStr)

		files = append(files, &FileInfo{
			Path:       path,
			RelPath:    filepath.ToSlash(relPath),
			Language:   language,
			Content:    contentStr,
			IsCode:     isCode,
			IsBinary:   false,
			HasNewline: hasNewline,
			LineCount:  lineCount,
			Size:       info.Size(),
			Encoding:   encoding,
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("扫描目录失败: %w", err)
	}

	// 排序：目录优先，然后按名称
	sort.Slice(files, func(i, j int) bool {
		return files[i].RelPath < files[j].RelPath
	})

	return files, nil
}

// isBinaryExtension 检查是否是二进制文件扩展名
func isBinaryExtension(path string, cfg *config.Config) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, binExt := range cfg.BinaryExtensions {
		if ext == binExt {
			return true
		}
	}
	return false
}

// isBinaryContent 检测文件内容是否是二进制
func isBinaryContent(content []byte) bool {
	// 空文件不是二进制
	if len(content) == 0 {
		return false
	}

	// 检查前8KB内容
	sampleSize := 8192
	if len(content) < sampleSize {
		sampleSize = len(content)
	}

	sample := content[:sampleSize]

	// 检查是否包含 NULL 字节
	if bytes.IndexByte(sample, 0) != -1 {
		return true
	}

	// 检查是否是有效的 UTF-8
	if !utf8.Valid(sample) {
		return true
	}

	// 计算非打印字符的比例
	nonPrintable := 0
	for _, b := range sample {
		// ASCII 控制字符（除了常见的空白字符）
		if b < 32 && b != '\t' && b != '\n' && b != '\r' {
			nonPrintable++
		}
		// DEL 字符
		if b == 127 {
			nonPrintable++
		}
	}

	// 如果超过 30% 是非打印字符，认为是二进制
	if float64(nonPrintable)/float64(len(sample)) > 0.3 {
		return true
	}

	return false
}

// detectEncoding 检测文件编码
func detectEncoding(content []byte) string {
	// 检查 BOM
	if len(content) >= 3 {
		// UTF-8 BOM
		if content[0] == 0xEF && content[1] == 0xBB && content[2] == 0xBF {
			return "UTF-8 with BOM"
		}
	}

	if len(content) >= 2 {
		// UTF-16 LE BOM
		if content[0] == 0xFF && content[1] == 0xFE {
			return "UTF-16 LE"
		}
		// UTF-16 BE BOM
		if content[0] == 0xFE && content[1] == 0xFF {
			return "UTF-16 BE"
		}
	}

	// 检查是否是有效的 UTF-8
	if utf8.Valid(content) {
		return "UTF-8"
	}

	return "Unknown"
}

// detectNewline 检测换行符类型
func detectNewline(content string) bool {
	return strings.Contains(content, "\n") ||
		strings.Contains(content, "\r\n") ||
		strings.Contains(content, "\r")
}

// getNewlineType 获取换行符类型
func getNewlineType(content string) string {
	hasCRLF := strings.Contains(content, "\r\n")
	hasLF := strings.Contains(content, "\n")
	hasCR := strings.Contains(content, "\r")

	if hasCRLF {
		return "CRLF (Windows)"
	} else if hasLF {
		return "LF (Unix)"
	} else if hasCR {
		return "CR (Old Mac)"
	}

	return "None"
}

// isNonCodeFile 检查是否是非代码文件
func isNonCodeFile(ext string, cfg *config.Config) bool {
	for _, nonCodeExt := range cfg.NonCodeExtensions {
		if ext == nonCodeExt {
			return true
		}
	}
	return false
}

// countLines 统计行数
func countLines(content string) int {
	if content == "" {
		return 0
	}

	// 处理不同的换行符
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")

	count := 1
	for _, ch := range content {
		if ch == '\n' {
			count++
		}
	}

	// 如果文件以换行符结尾，不计入额外的一行
	if strings.HasSuffix(content, "\n") {
		count--
	}

	// 至少返回 1
	if count < 1 {
		count = 1
	}

	return count
}

// ValidateContent 验证文件内容
func ValidateContent(file *FileInfo) []string {
	var warnings []string

	// 检查空文件
	if len(file.Content) == 0 {
		warnings = append(warnings, "文件为空")
		return warnings
	}

	// 检查换行符
	if !file.HasNewline && file.Size > 0 {
		warnings = append(warnings, "文件不包含换行符")
	}

	// 检查混合换行符
	hasCRLF := strings.Contains(file.Content, "\r\n")
	hasLF := strings.Contains(file.Content, "\n")
	hasCR := strings.Contains(file.Content, "\r")

	mixedNewlines := 0
	if hasCRLF {
		mixedNewlines++
	}
	if hasLF && !hasCRLF {
		mixedNewlines++
	}
	if hasCR && !hasCRLF {
		mixedNewlines++
	}

	if mixedNewlines > 1 {
		warnings = append(warnings, "文件包含混合的换行符类型")
	}

	// 检查行尾空格
	lines := strings.Split(file.Content, "\n")
	trailingSpaceLines := 0
	for _, line := range lines {
		if len(line) > 0 && (line[len(line)-1] == ' ' || line[len(line)-1] == '\t') {
			trailingSpaceLines++
		}
	}

	if trailingSpaceLines > 0 {
		warnings = append(warnings,
			fmt.Sprintf("有 %d 行包含行尾空格", trailingSpaceLines))
	}

	// 检查文件是否以换行符结尾
	if !strings.HasSuffix(file.Content, "\n") {
		warnings = append(warnings, "文件不以换行符结尾")
	}

	// 检查超长行
	maxLineLength := 0
	for _, line := range lines {
		if len(line) > maxLineLength {
			maxLineLength = len(line)
		}
	}

	if maxLineLength > 500 {
		warnings = append(warnings,
			fmt.Sprintf("包含超长行（最长 %d 字符）", maxLineLength))
	}

	// 检查 Tab 字符
	if strings.Contains(file.Content, "\t") {
		tabCount := strings.Count(file.Content, "\t")
		warnings = append(warnings,
			fmt.Sprintf("包含 %d 个 Tab 字符", tabCount))
	}

	return warnings
}

// GetFileStats 获取文件统计信息
func GetFileStats(files []*FileInfo) map[string]interface{} {
	stats := make(map[string]interface{})

	totalFiles := len(files)
	totalLines := 0
	totalSize := int64(0)
	codeFiles := 0
	configFiles := 0

	languageCount := make(map[string]int)
	encodingCount := make(map[string]int)

	for _, file := range files {
		totalLines += file.LineCount
		totalSize += file.Size

		if file.IsCode {
			codeFiles++
		} else {
			configFiles++
		}

		languageCount[file.Language]++
		encodingCount[file.Encoding]++
	}

	stats["total_files"] = totalFiles
	stats["total_lines"] = totalLines
	stats["total_size"] = totalSize
	stats["code_files"] = codeFiles
	stats["config_files"] = configFiles
	stats["language_count"] = languageCount
	stats["encoding_count"] = encodingCount

	return stats
}
