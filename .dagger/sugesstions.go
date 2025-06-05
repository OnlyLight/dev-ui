package main

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
)

type CodeSuggestion struct {
	File       string
	Line       int
	Suggestion []string
	DiffHunk   string // Add diff hunk for GitHub API
}

func parseDiff(diffText string) ([]CodeSuggestion, error) {
	var suggestions []CodeSuggestion
	var currentFile string
	var currentLine int
	var newCode []string
	var currentHunk []string
	var inHunk bool

	// Regular expressions
	fileRegex := regexp.MustCompile(`^\+\+\+ b/(.+)`)
	lineRegex := regexp.MustCompile(`^@@ -(\d+),\d+ \+(\d+),\d+ @@`)

	scanner := bufio.NewScanner(strings.NewReader(diffText))
	for scanner.Scan() {
		line := scanner.Text()

		// Detect file name
		if matches := fileRegex.FindStringSubmatch(line); matches != nil {
			if inHunk && len(newCode) > 0 {
				suggestions = append(suggestions, CodeSuggestion{
					File:       currentFile,
					Line:       currentLine,
					Suggestion: newCode,
					DiffHunk:   strings.Join(currentHunk, "\n"),
				})
				newCode = []string{}
				currentHunk = []string{}
			}
			currentFile = matches[1]
			inHunk = false
			continue
		}

		// Detect hunk
		if matches := lineRegex.FindStringSubmatch(line); matches != nil {
			if inHunk && len(newCode) > 0 {
				suggestions = append(suggestions, CodeSuggestion{
					File:       currentFile,
					Line:       currentLine,
					Suggestion: newCode,
					DiffHunk:   strings.Join(currentHunk, "\n"),
				})
				newCode = []string{}
				currentHunk = []string{}
			}
			currentLine = atoi(matches[2]) // 1-based line number
			inHunk = true
			currentHunk = append(currentHunk, line)
			continue
		}

		if inHunk {
			currentHunk = append(currentHunk, line)
			if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
				newCode = append(newCode, strings.TrimPrefix(line, "+"))
				currentLine++
			} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
				// Do not increment for removed lines
			} else if strings.HasPrefix(line, " ") {
				currentLine++ // Context lines
			}
		}
	}

	// Add final suggestion
	if inHunk && len(newCode) > 0 && currentFile != "" {
		suggestions = append(suggestions, CodeSuggestion{
			File:       currentFile,
			Line:       currentLine,
			Suggestion: newCode,
			DiffHunk:   strings.Join(currentHunk, "\n"),
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan diff: %w", err)
	}

	return suggestions, nil
}

func atoi(s string) int {
	var i int
	fmt.Sscanf(s, "%d", &i)
	return i
}
