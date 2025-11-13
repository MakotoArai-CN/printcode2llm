package compress

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// 代码语言映射
var codeLanguages = map[string]bool{
	"javascript": true,
	"typescript": true,
	"python":     true,
	"java":       true,
	"c":          true,
	"cpp":        true,
	"go":         true,
	"rust":       true,
	"php":        true,
	"ruby":       true,
	"swift":      true,
	"kotlin":     true,
	"scala":      true,
	"dart":       true,
	"csharp":     true,
	"objc":       true,
}

// C风格注释的语言
var cStyleCommentLanguages = map[string]bool{
	"javascript": true,
	"typescript": true,
	"java":       true,
	"c":          true,
	"cpp":        true,
	"go":         true,
	"rust":       true,
	"php":        true,
	"swift":      true,
	"kotlin":     true,
	"scala":      true,
	"dart":       true,
	"csharp":     true,
	"objc":       true,
}

// Python风格注释的语言
var pythonStyleCommentLanguages = map[string]bool{
	"python": true,
	"ruby":   true,
	"shell":  true,
	"bash":   true,
}

// Compress 压缩代码
func Compress(content, language string, ultraMode bool) string {
	if content == "" {
		return content
	}

	// 标准化语言名称
	language = strings.ToLower(language)

	// 非代码文件，只做基础处理
	if !codeLanguages[language] {
		return basicCompress(content)
	}

	// 创建压缩器
	compressor := &Compressor{
		language:  language,
		ultraMode: ultraMode,
		tokens:    make([]string, 0),
	}

	return compressor.Compress(content)
}

// Compressor 代码压缩器
type Compressor struct {
	language  string
	ultraMode bool
	tokens    []string
}

// Compress 执行压缩
func (c *Compressor) Compress(content string) string {
	// 1. 保护字符串和正则表达式
	content = c.protectStrings(content)

	// 2. 移除注释
	content = c.removeComments(content)

	// 3. 清理空白字符
	content = c.cleanWhitespace(content)

	// 4. 应用压缩模式
	if c.ultraMode {
		content = c.ultraCompress(content)
	} else {
		content = c.standardCompress(content)
	}

	// 5. 恢复字符串和正则表达式
	content = c.restoreTokens(content)

	// 6. 最终清理
	content = c.finalCleanup(content)

	return strings.TrimSpace(content)
}

// basicCompress 基础压缩（非代码文件）
func basicCompress(content string) string {
	lines := strings.Split(content, "\n")
	result := make([]string, 0, len(lines))

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			// 移除行尾空格，但保留行首缩进的相对关系
			result = append(result, strings.TrimRight(line, " \t"))
		}
	}

	return strings.Join(result, "\n")
}

// protectStrings 保护字符串和正则表达式
func (c *Compressor) protectStrings(content string) string {
	// 保护双引号字符串
	content = c.protectPattern(content, `"(?:[^"\\]|\\.)*"`)

	// 保护单引号字符串
	content = c.protectPattern(content, `'(?:[^'\\]|\\.)*'`)

	// 保护反引号字符串（模板字符串）
	if c.language == "javascript" || c.language == "typescript" {
		content = c.protectPattern(content, "`(?:[^`\\\\]|\\\\.)*`")
	}

	// 保护正则表达式（JavaScript/TypeScript）
	if c.language == "javascript" || c.language == "typescript" {
		content = c.protectRegex(content)
	}

	// 保护 Python 的三引号字符串
	if c.language == "python" {
		content = c.protectPattern(content, `"""(?:[^\\]|\\.)*?"""`)
		content = c.protectPattern(content, `'''(?:[^\\]|\\.)*?'''`)
	}

	// 保护 Go 的原始字符串
	if c.language == "go" {
		content = c.protectPattern(content, "`[^`]*`")
	}

	return content
}

// protectPattern 保护匹配的模式
func (c *Compressor) protectPattern(content, pattern string) string {
	re := regexp.MustCompile(pattern)
	content = re.ReplaceAllStringFunc(content, func(match string) string {
		token := c.createToken()
		c.tokens = append(c.tokens, match)
		return token
	})
	return content
}

// protectRegex 保护 JavaScript/TypeScript 的正则表达式
func (c *Compressor) protectRegex(content string) string {
	// 匹配正则表达式字面量：/pattern/flags
	// 需要区分除法运算符和正则表达式
	re := regexp.MustCompile(`(?:^|[=([,;:!&|?{}\n])\s*/(?![*/])(?:[^\\/\n]|\\.)+/[gimsuvy]*`)

	content = re.ReplaceAllStringFunc(content, func(match string) string {
		// 提取前缀
		prefix := ""
		regexPart := match

		// 找到正则表达式的开始位置
		idx := strings.LastIndex(match, "/")
		if idx > 0 {
			for i := idx - 1; i >= 0; i-- {
				if match[i] == '/' {
					prefix = match[:i]
					regexPart = match[i:]
					break
				}
			}
		}

		// 保护正则表达式部分
		token := c.createToken()
		c.tokens = append(c.tokens, regexPart)

		return prefix + token
	})

	return content
}

// createToken 创建占位符
func (c *Compressor) createToken() string {
	return "___TOKEN_" + strings.Repeat("X", len(c.tokens)) + "___"
}

// restoreTokens 恢复所有 token
func (c *Compressor) restoreTokens(content string) string {
	for i := len(c.tokens) - 1; i >= 0; i-- {
		token := "___TOKEN_" + strings.Repeat("X", i) + "___"
		content = strings.ReplaceAll(content, token, c.tokens[i])
	}
	return content
}

// removeComments 移除注释
func (c *Compressor) removeComments(content string) string {
	if cStyleCommentLanguages[c.language] {
		return c.removeCStyleComments(content)
	}

	if pythonStyleCommentLanguages[c.language] {
		return c.removePythonStyleComments(content)
	}

	return content
}

// removeCStyleComments 移除 C 风格注释
func (c *Compressor) removeCStyleComments(content string) string {
	// 移除单行注释 //
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if idx := strings.Index(line, "//"); idx != -1 {
			lines[i] = line[:idx]
		}
	}
	content = strings.Join(lines, "\n")

	// 移除多行注释 /* */
	re := regexp.MustCompile(`/\*[\s\S]*?\*/`)
	content = re.ReplaceAllString(content, "")

	return content
}

// removePythonStyleComments 移除 Python 风格注释
func (c *Compressor) removePythonStyleComments(content string) string {
	lines := strings.Split(content, "\n")
	result := make([]string, 0, len(lines))

	for _, line := range lines {
		// 查找 # 注释
		if idx := strings.Index(line, "#"); idx != -1 {
			// 确保不是字符串中的 #
			beforeHash := line[:idx]
			// 简单检查：如果 # 前面有未闭合的引号，则不是注释
			singleQuotes := strings.Count(beforeHash, "'") - strings.Count(beforeHash, "\\'")
			doubleQuotes := strings.Count(beforeHash, "\"") - strings.Count(beforeHash, "\\\"")

			if singleQuotes%2 == 0 && doubleQuotes%2 == 0 {
				line = beforeHash
			}
		}

		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

// cleanWhitespace 清理空白字符
func (c *Compressor) cleanWhitespace(content string) string {
	lines := strings.Split(content, "\n")
	result := make([]string, 0, len(lines))

	for _, line := range lines {
		// 移除行首空格
		trimmed := strings.TrimSpace(line)

		// 跳过空行
		if trimmed == "" {
			continue
		}

		// 压缩连续空格为单个空格
		trimmed = c.compressSpaces(trimmed)

		result = append(result, trimmed)
	}

	return strings.Join(result, "\n")
}

// compressSpaces 压缩连续空格
func (c *Compressor) compressSpaces(s string) string {
	var result strings.Builder
	var prevSpace bool

	for _, ch := range s {
		if unicode.IsSpace(ch) {
			if !prevSpace {
				result.WriteRune(' ')
				prevSpace = true
			}
		} else {
			result.WriteRune(ch)
			prevSpace = false
		}
	}

	return result.String()
}

// standardCompress 标准压缩
func (c *Compressor) standardCompress(content string) string {
	// 合并某些可以安全合并的行
	replacements := []struct {
		pattern     string
		replacement string
	}{
		{`\n\s*\{`, " {"},                    // 将独立的 { 合并到上一行
		{`\}\s*\n\s*else`, "} else"},         // } else
		{`\}\s*\n\s*catch`, "} catch"},       // } catch
		{`\}\s*\n\s*finally`, "} finally"},   // } finally
		{`\}\s*\n\s*elif`, "} elif"},         // } elif (Python)
		{`\}\s*\n\s*except`, "} except"},     // } except (Python)
	}

	for _, r := range replacements {
		re := regexp.MustCompile(r.pattern)
		content = re.ReplaceAllString(content, r.replacement)
	}

	// 移除块结束后的多余换行
	content = regexp.MustCompile(`\n{3,}`).ReplaceAllString(content, "\n\n")

	return content
}

// ultraCompress 超级压缩
func (c *Compressor) ultraCompress(content string) string {
	// 首先应用标准压缩
	content = c.standardCompress(content)

	// 更激进的压缩规则
	aggressiveReplacements := map[string]string{
		// 移除所有换行符周围的空格
		"\n{":  "{",
		"}\n":  "}",
		";\n":  ";",
		",\n":  ",",
		"{\n":  "{",
		"\n}":  "}",
		"\n;":  ";",
		"\n,":  ",",

		// 移除操作符周围的空格
		" {":   "{",
		" }":   "}",
		" (":   "(",
		" )":   ")",
		"( ":   "(",
		") ":   ")",
		" [":   "[",
		" ]":   "]",
		"[ ":   "[",
		"] ":   "]",
		" ;":   ";",
		" ,":   ",",
		", ":   ",",
		" :":   ":",
		": ":   ":",

		// 赋值运算符
		" = ":   "=",
		" == ":  "==",
		" != ":  "!=",
		" === ": "===",
		" !== ": "!==",
		" += ":  "+=",
		" -= ":  "-=",
		" *= ":  "*=",
		" /= ":  "/=",
		" %= ":  "%=",
		" &= ":  "&=",
		" |= ":  "|=",
		" ^= ":  "^=",

		// 比较运算符
		" < ":  "<",
		" > ":  ">",
		" <= ": "<=",
		" >= ": ">=",

		// 逻辑运算符
		" && ": "&&",
		" || ": "||",

		// 位运算符
		" & ":  "&",
		" | ":  "|",
		" ^ ":  "^",
		" << ": "<<",
		" >> ": ">>",

		// 箭头函数 (JavaScript/TypeScript)
		" => ": "=>",

		// Go 的 :=
		" := ": ":=",
	}

	// 应用替换（多次迭代以确保完全压缩）
	maxIterations := 3
	for iteration := 0; iteration < maxIterations; iteration++ {
		changed := false
		for old, new := range aggressiveReplacements {
			if strings.Contains(content, old) {
				content = strings.ReplaceAll(content, old, new)
				changed = true
			}
		}
		if !changed {
			break
		}
	}

	// 特定语言的优化
	switch c.language {
	case "javascript", "typescript":
		content = c.ultraCompressJavaScript(content)
	case "python":
		content = c.ultraCompressPython(content)
	case "go":
		content = c.ultraCompressGo(content)
	}

	// 移除所有多余的换行
	content = regexp.MustCompile(`\n+`).ReplaceAllString(content, "\n")

	return content
}

// ultraCompressJavaScript JavaScript/TypeScript 特定的超级压缩
func (c *Compressor) ultraCompressJavaScript(content string) string {
	// 移除分号后的换行（除非下一行是关键字）
	lines := strings.Split(content, "\n")
	var result strings.Builder

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		result.WriteString(line)

		// 如果行以分号结尾，尝试合并下一行
		if strings.HasSuffix(strings.TrimSpace(line), ";") && i < len(lines)-1 {
			nextLine := strings.TrimSpace(lines[i+1])
			// 检查下一行是否是关键字开头
			if !c.startsWithKeyword(nextLine, []string{
				"import", "export", "class", "function", "const", "let", "var",
				"if", "else", "for", "while", "switch", "case", "return",
				"break", "continue", "try", "catch", "finally", "async", "await",
			}) {
				continue // 不添加换行
			}
		}

		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}

// ultraCompressPython Python 特定的超级压缩
func (c *Compressor) ultraCompressPython(content string) string {
	// Python 需要保留缩进，所以压缩较为保守
	lines := strings.Split(content, "\n")
	result := make([]string, 0, len(lines))

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			// 保留必要的缩进结构
			result = append(result, trimmed)
		}
	}

	return strings.Join(result, "\n")
}

// ultraCompressGo Go 特定的超级压缩
func (c *Compressor) ultraCompressGo(content string) string {
	// 移除 package 和 import 之间的空行
	content = regexp.MustCompile(`(package\s+\w+)\n+`).ReplaceAllString(content, "$1\n")

	// 合并简短的函数
	content = regexp.MustCompile(`\)\s*\{\s*return`).ReplaceAllString(content, "){return")

	return content
}

// startsWithKeyword 检查字符串是否以关键字开头
func (c *Compressor) startsWithKeyword(s string, keywords []string) bool {
	for _, keyword := range keywords {
		if strings.HasPrefix(s, keyword) {
			// 确保关键字后面是空格或其他非字母字符
			if len(s) == len(keyword) {
				return true
			}
			nextChar := s[len(keyword)]
			if !unicode.IsLetter(rune(nextChar)) && nextChar != '_' {
				return true
			}
		}
	}
	return false
}

// finalCleanup 最终清理
func (c *Compressor) finalCleanup(content string) string {
	// 移除开头和结尾的空白
	content = strings.TrimSpace(content)

	// 确保文件以换行符结尾（符合 POSIX 标准）
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	// 移除连续的空行（最多保留一个）
	content = regexp.MustCompile(`\n{3,}`).ReplaceAllString(content, "\n\n")

	// 移除行尾空格
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	content = strings.Join(lines, "\n")

	return content
}

// GetCompressionStats 获取压缩统计信息
func GetCompressionStats(original, compressed string) map[string]interface{} {
	stats := make(map[string]interface{})

	originalSize := len(original)
	compressedSize := len(compressed)
	savings := originalSize - compressedSize
	ratio := 0.0

	if originalSize > 0 {
		ratio = float64(savings) / float64(originalSize) * 100
	}

	originalLines := strings.Count(original, "\n") + 1
	compressedLines := strings.Count(compressed, "\n") + 1

	stats["original_size"] = originalSize
	stats["compressed_size"] = compressedSize
	stats["savings_bytes"] = savings
	stats["compression_ratio"] = ratio
	stats["original_lines"] = originalLines
	stats["compressed_lines"] = compressedLines
	stats["lines_removed"] = originalLines - compressedLines

	return stats
}

// EstimateTokens 估算 token 数量（简单估算）
func EstimateTokens(content string) int {
	// 简单估算：平均每 4 个字符约等于 1 个 token
	// 这是一个粗略估计，实际取决于 tokenizer
	words := strings.Fields(content)
	charCount := len(content)

	// 使用字数和字符数的组合估算
	tokenEstimate := (len(words) + charCount/4) / 2

	return tokenEstimate
}

// ValidateCompression 验证压缩是否安全
func ValidateCompression(original, compressed, language string) []string {
	var warnings []string

	// 检查大小变化
	if len(compressed) > len(original) {
		warnings = append(warnings, "压缩后大小增加了")
	}

	// 检查是否过度压缩
	ratio := float64(len(compressed)) / float64(len(original))
	if ratio < 0.3 {
		warnings = append(warnings, "压缩率过高，可能影响可读性")
	}

	// 检查是否保留了关键结构
	switch language {
	case "python":
		// Python 需要保留缩进
		if !strings.Contains(compressed, "\n") && strings.Contains(original, "\n") {
			warnings = append(warnings, "Python 代码的换行被过度移除")
		}

	case "go":
		// Go 需要保留 package 声明
		if strings.Contains(original, "package ") && !strings.Contains(compressed, "package ") {
			warnings = append(warnings, "Go 的 package 声明丢失")
		}
	}

	// 检查括号匹配
	brackets := map[rune]rune{
		'(': ')',
		'[': ']',
		'{': '}',
	}

	for open, close := range brackets {
		openCount := strings.Count(compressed, string(open))
		closeCount := strings.Count(compressed, string(close))
		if openCount != closeCount {
			warnings = append(warnings, 
				fmt.Sprintf("括号不匹配：%c=%d, %c=%d", open, openCount, close, closeCount))
		}
	}

	return warnings
}

// DecompressHint 生成解压提示
func DecompressHint(language string, ultraMode bool) string {
	hints := []string{
		"使用代码格式化工具进行格式化：",
	}

	switch language {
	case "javascript", "typescript":
		hints = append(hints, "  - Prettier: npx prettier --write file.js")
		hints = append(hints, "  - ESLint: npx eslint --fix file.js")

	case "python":
		hints = append(hints, "  - Black: black file.py")
		hints = append(hints, "  - autopep8: autopep8 --in-place file.py")

	case "go":
		hints = append(hints, "  - gofmt: gofmt -w file.go")
		hints = append(hints, "  - goimports: goimports -w file.go")

	case "rust":
		hints = append(hints, "  - rustfmt: rustfmt file.rs")

	case "java":
		hints = append(hints, "  - google-java-format")

	case "cpp", "c":
		hints = append(hints, "  - clang-format: clang-format -i file.cpp")
	}

	if ultraMode {
		hints = append(hints, "")
		hints = append(hints, "注意：使用了超级压缩模式，必须先格式化才能正常阅读")
	}

	return strings.Join(hints, "\n")
}