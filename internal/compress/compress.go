package compress

import (
	"regexp"
	"strings"
	"unicode"
)

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

var pythonStyleCommentLanguages = map[string]bool{
	"python": true,
	"ruby":   true,
	"shell":  true,
	"bash":   true,
}

func Compress(content, language string, ultraMode bool) string {
	if content == "" {
		return content
	}

	language = strings.ToLower(language)

	if !codeLanguages[language] {
		return basicCompress(content)
	}

	compressor := &Compressor{
		language:  language,
		ultraMode: ultraMode,
		tokens:    make([]string, 0),
	}

	return compressor.Compress(content)
}

type Compressor struct {
	language  string
	ultraMode bool
	tokens    []string
}

func (c *Compressor) Compress(content string) string {
	content = c.protectStrings(content)
	content = c.removeComments(content)
	content = c.cleanWhitespace(content)

	if c.ultraMode {
		content = c.ultraCompress(content)
	} else {
		content = c.standardCompress(content)
	}

	content = c.restoreTokens(content)
	content = c.finalCleanup(content)

	return strings.TrimSpace(content)
}

func basicCompress(content string) string {
	lines := strings.Split(content, "\n")
	result := make([]string, 0, len(lines))

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			result = append(result, strings.TrimRight(line, " \t"))
		}
	}

	return strings.Join(result, "\n")
}

func (c *Compressor) protectStrings(content string) string {
	content = c.protectQuotedStrings(content, '"')
	content = c.protectQuotedStrings(content, '\'')

	if c.language == "javascript" || c.language == "typescript" {
		content = c.protectTemplateStrings(content)
	}

	if c.language == "python" {
		content = c.protectTripleQuotes(content, `"""`)
		content = c.protectTripleQuotes(content, `'''`)
	}

	if c.language == "go" {
		content = c.protectRawStrings(content)
	}

	return content
}

func (c *Compressor) protectQuotedStrings(content string, quote rune) string {
	var result strings.Builder
	inString := false
	escape := false

	for i, ch := range content {
		if escape {
			result.WriteRune(ch)
			escape = false
			continue
		}

		if ch == '\\' && inString {
			result.WriteRune(ch)
			escape = true
			continue
		}

		if ch == quote {
			if inString {
				result.WriteRune(ch)
				inString = false
			} else {
				inString = true
				result.WriteRune(ch)
			}
			continue
		}

		if inString {
			result.WriteRune(ch)
		} else {
			if ch == '\n' || ch == '\r' {
				result.WriteRune(ch)
			} else {
				_ = i
				result.WriteRune(ch)
			}
		}
	}

	return result.String()
}

func (c *Compressor) protectTemplateStrings(content string) string {
	var result strings.Builder
	inTemplate := false
	braceDepth := 0

	runes := []rune(content)
	for i := 0; i < len(runes); i++ {
		ch := runes[i]

		if ch == '`' && !inTemplate {
			inTemplate = true
			result.WriteRune(ch)
			continue
		}

		if inTemplate {
			result.WriteRune(ch)

			if ch == '$' && i+1 < len(runes) && runes[i+1] == '{' {
				braceDepth++
			} else if ch == '{' && braceDepth > 0 {
				braceDepth++
			} else if ch == '}' && braceDepth > 0 {
				braceDepth--
			} else if ch == '`' && braceDepth == 0 {
				inTemplate = false
			}
			continue
		}

		result.WriteRune(ch)
	}

	return result.String()
}

func (c *Compressor) protectTripleQuotes(content, quote string) string {
	var result strings.Builder
	i := 0

	for i < len(content) {
		if i+3 <= len(content) && content[i:i+3] == quote {
			start := i
			i += 3

			for i+3 <= len(content) {
				if content[i:i+3] == quote {
					i += 3
					break
				}
				i++
			}

			token := c.createToken()
			c.tokens = append(c.tokens, content[start:i])
			result.WriteString(token)
		} else {
			result.WriteByte(content[i])
			i++
		}
	}

	return result.String()
}

func (c *Compressor) protectRawStrings(content string) string {
	var result strings.Builder
	i := 0

	for i < len(content) {
		if content[i] == '`' {
			start := i
			i++

			for i < len(content) && content[i] != '`' {
				i++
			}

			if i < len(content) {
				i++
			}

			token := c.createToken()
			c.tokens = append(c.tokens, content[start:i])
			result.WriteString(token)
		} else {
			result.WriteByte(content[i])
			i++
		}
	}

	return result.String()
}

func (c *Compressor) createToken() string {
	return "___PTLM_TOKEN_" + strings.Repeat("X", len(c.tokens)) + "___"
}

func (c *Compressor) restoreTokens(content string) string {
	for i := len(c.tokens) - 1; i >= 0; i-- {
		token := "___PTLM_TOKEN_" + strings.Repeat("X", i) + "___"
		content = strings.ReplaceAll(content, token, c.tokens[i])
	}
	return content
}

func (c *Compressor) removeComments(content string) string {
	if cStyleCommentLanguages[c.language] {
		return c.removeCStyleComments(content)
	}

	if pythonStyleCommentLanguages[c.language] {
		return c.removePythonStyleComments(content)
	}

	return content
}

func (c *Compressor) removeCStyleComments(content string) string {
	re := regexp.MustCompile(`/\*[\s\S]*?\*/`)
	content = re.ReplaceAllString(content, "")

	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if idx := strings.Index(line, "//"); idx != -1 {
			beforeComment := line[:idx]
			if !c.isInString(beforeComment) {
				lines[i] = beforeComment
			}
		}
	}

	return strings.Join(lines, "\n")
}

func (c *Compressor) isInString(s string) bool {
	inDouble := false
	inSingle := false
	escape := false

	for _, ch := range s {
		if escape {
			escape = false
			continue
		}

		if ch == '\\' {
			escape = true
			continue
		}

		if ch == '"' && !inSingle {
			inDouble = !inDouble
		} else if ch == '\'' && !inDouble {
			inSingle = !inSingle
		}
	}

	return inDouble || inSingle
}

func (c *Compressor) removePythonStyleComments(content string) string {
	lines := strings.Split(content, "\n")
	result := make([]string, 0, len(lines))

	for _, line := range lines {
		if idx := strings.Index(line, "#"); idx != -1 {
			beforeHash := line[:idx]
			if !c.isInString(beforeHash) {
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

func (c *Compressor) cleanWhitespace(content string) string {
	lines := strings.Split(content, "\n")
	result := make([]string, 0, len(lines))

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		trimmed = c.compressSpaces(trimmed)
		result = append(result, trimmed)
	}

	return strings.Join(result, "\n")
}

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

func (c *Compressor) standardCompress(content string) string {
	replacements := []struct {
		pattern     string
		replacement string
	}{
		{`\n\s*\{`, " {"},
		{`\}\s*\n\s*else`, "} else"},
		{`\}\s*\n\s*catch`, "} catch"},
		{`\}\s*\n\s*finally`, "} finally"},
		{`\}\s*\n\s*elif`, "} elif"},
		{`\}\s*\n\s*except`, "} except"},
	}

	for _, r := range replacements {
		re := regexp.MustCompile(r.pattern)
		content = re.ReplaceAllString(content, r.replacement)
	}

	content = regexp.MustCompile(`\n{3,}`).ReplaceAllString(content, "\n\n")

	return content
}

func (c *Compressor) ultraCompress(content string) string {
	content = c.standardCompress(content)

	replacements := map[string]string{
		"\n{": "{",
		"}\n": "}",
		";\n": ";",
		",\n": ",",
		"{\n": "{",
		"\n}": "}",
		"\n;": ";",
		"\n,": ",",

		" {":  "{",
		"{ ":  "{",
		" }":  "}",
		"} ":  "}",
		" (":  "(",
		"( ":  "(",
		" )":  ")",
		") ":  ")",
		" [":  "[",
		"[ ":  "[",
		" ]":  "]",
		"] ":  "]",
		" ;":  ";",
		" ,":  ",",
		", ":  ",",

		" = ":   "=",
		" == ":  "==",
		" != ":  "!=",
		" === ": "===",
		" !== ": "!==",
		" += ":  "+=",
		" -= ":  "-=",
		" *= ":  "*=",
		" /= ":  "/=",

		" < ":  "<",
		" > ":  ">",
		" <= ": "<=",
		" >= ": ">=",

		" && ": "&&",
		" || ": "||",

		" => ": "=>",
		" := ": ":=",
	}

	for i := 0; i < 3; i++ {
		changed := false
		for old, new := range replacements {
			if strings.Contains(content, old) {
				content = strings.ReplaceAll(content, old, new)
				changed = true
			}
		}
		if !changed {
			break
		}
	}

	switch c.language {
	case "javascript", "typescript":
		content = c.ultraCompressJS(content)
	case "go":
		content = c.ultraCompressGo(content)
	}

	content = regexp.MustCompile(`\n+`).ReplaceAllString(content, "\n")

	return content
}

func (c *Compressor) ultraCompressJS(content string) string {
	lines := strings.Split(content, "\n")
	var result strings.Builder

	keywords := []string{
		"import", "export", "class", "function", "const", "let", "var",
		"if", "else", "for", "while", "switch", "case", "return",
		"break", "continue", "try", "catch", "finally", "async", "await",
	}

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		result.WriteString(line)

		if strings.HasSuffix(strings.TrimSpace(line), ";") && i < len(lines)-1 {
			nextLine := strings.TrimSpace(lines[i+1])
			if !c.startsWithKeyword(nextLine, keywords) {
				continue
			}
		}

		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}

func (c *Compressor) ultraCompressGo(content string) string {
	content = regexp.MustCompile(`(package\s+\w+)\n+`).ReplaceAllString(content, "$1\n")
	content = regexp.MustCompile(`\)\s*\{\s*return`).ReplaceAllString(content, "){return")
	return content
}

func (c *Compressor) startsWithKeyword(s string, keywords []string) bool {
	for _, keyword := range keywords {
		if strings.HasPrefix(s, keyword) {
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

func (c *Compressor) finalCleanup(content string) string {
	content = strings.TrimSpace(content)

	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	content = regexp.MustCompile(`\n{3,}`).ReplaceAllString(content, "\n\n")

	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	content = strings.Join(lines, "\n")

	return content
}

func EstimateTokens(content string) int {
	words := strings.Fields(content)
	charCount := len(content)
	tokenEstimate := (len(words) + charCount/4) / 2
	return tokenEstimate
}