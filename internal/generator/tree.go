package generator

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"printcode2llm/internal/config"
	"printcode2llm/internal/scanner"
)

// GenerateTree 生成目录树
func GenerateTree(dir string, cfg *config.Config) (string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	ignoreChecker := scanner.NewIgnoreChecker(cfg)

	err = generateTreeRecursive(absDir, "", &builder, ignoreChecker, true)
	if err != nil {
		return "", err
	}

	return builder.String(), nil
}

func generateTreeRecursive(dir, prefix string, builder *strings.Builder, ignoreChecker *scanner.IgnoreChecker, isRoot bool) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	// 过滤和排序
	var validEntries []os.DirEntry
	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		if !ignoreChecker.ShouldIgnore(path, entry.IsDir()) {
			validEntries = append(validEntries, entry)
		}
	}

	sort.Slice(validEntries, func(i, j int) bool {
		if validEntries[i].IsDir() != validEntries[j].IsDir() {
			return validEntries[i].IsDir()
		}
		return validEntries[i].Name() < validEntries[j].Name()
	})

	for i, entry := range validEntries {
		isLast := i == len(validEntries)-1
		connector := "├── "
		if isLast {
			connector = "└── "
		}

		name := entry.Name()
		if entry.IsDir() {
			name += "/"
		}

		if !isRoot || i > 0 {
			builder.WriteString(prefix)
		}
		builder.WriteString(connector)
		builder.WriteString(name)
		builder.WriteString("\n")

		if entry.IsDir() {
			nextPrefix := prefix
			if isLast {
				nextPrefix += "    "
			} else {
				nextPrefix += "│   "
			}

			subDir := filepath.Join(dir, entry.Name())
			if err := generateTreeRecursive(subDir, nextPrefix, builder, ignoreChecker, false); err != nil {
				return err
			}
		}
	}

	return nil
}
