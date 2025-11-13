package scanner

import (
	"path/filepath"
	"regexp"
	"strings"

	"printcode2llm/internal/config"
)

type IgnoreChecker struct {
	patterns  []string
	regexList []*regexp.Regexp
	cfg       *config.Config
}

func NewIgnoreChecker(cfg *config.Config) *IgnoreChecker {
	checker := &IgnoreChecker{
		patterns:  make([]string, 0),
		regexList: make([]*regexp.Regexp, 0),
		cfg:       cfg,
	}

	checker.patterns = append(checker.patterns, cfg.DefaultIgnore...)
	checker.patterns = append(checker.patterns, cfg.CustomIgnore.Patterns...)

	for _, regexStr := range cfg.CustomIgnore.Regex {
		if re, err := regexp.Compile(regexStr); err == nil {
			checker.regexList = append(checker.regexList, re)
		}
	}

	return checker
}

func (ic *IgnoreChecker) ShouldIgnore(path string, isDir bool) bool {
	name := filepath.Base(path)
	cleanPath := filepath.ToSlash(path)

	for _, pattern := range ic.patterns {
		if strings.Contains(pattern, "*") || strings.Contains(pattern, "?") {
			matched, _ := filepath.Match(pattern, name)
			if matched {
				return true
			}
			
			matched, _ = filepath.Match(pattern, cleanPath)
			if matched {
				return true
			}
		} else {
			if name == pattern {
				return true
			}
			
			if strings.Contains(cleanPath, pattern) {
				return true
			}
		}
	}

	for _, re := range ic.regexList {
		if re.MatchString(cleanPath) || re.MatchString(name) {
			return true
		}
	}

	return false
}