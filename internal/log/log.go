package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

func StartLog(topic string) {
	var buf strings.Builder
	StartLogTo(topic, &buf)
	fmt.Print(buf.String())
}

func StartLogTo(topic string, w io.Writer) {
	logsDir := getLogsDir()
	slug := createSlug(topic)
	filename := fmt.Sprintf("%s-%s.md", time.Now().Format("2006-01-02"), slug)
	fp := filepath.Join(logsDir, filename)

	if _, err := os.Stat(fp); err == nil {
		fmt.Fprintf(w, "Log file already exists: %s\n", filename)
		fmt.Fprintln(w, "Use a different topic or append to existing log")
		return
	}

	content := fmt.Sprintf(`# Session Log: %s

## %s

- Started session on topic: %s

`, topic, time.Now().Format("2006-01-02"), topic)

	if err := os.WriteFile(fp, []byte(content), 0644); err != nil {
		fmt.Fprintf(w, "Error creating log file: %v\n", err)
		return
	}

	fmt.Fprintf(w, "Created session log: %s\n", filename)
	fmt.Fprintf(w, "Use 'bots log append %s \"<message>\"' to add entries\n", slug)
}

func AppendEntry(slug string, message string) {
	var buf strings.Builder
	AppendEntryTo(slug, message, &buf)
	fmt.Print(buf.String())
}

func AppendEntryTo(slug string, message string, w io.Writer) {
	logsDir := getLogsDir()
	filename := findLogFile(slug)

	if filename == "" {
		fmt.Fprintf(w, "No log file found for slug: %s\n", slug)
		fmt.Fprintln(w, "Use 'bots log list' to see available logs")
		return
	}

	fp := filepath.Join(logsDir, filename)

	fileContent, err := os.ReadFile(fp)
	if err != nil {
		fmt.Fprintf(w, "Error reading log file: %v\n", err)
		return
	}

	now := time.Now()
	entry := fmt.Sprintf("\n## %s\n\n- %s\n\n", now.Format("2006-01-02"), message)
	newContent := string(fileContent) + entry

	if err := os.WriteFile(fp, []byte(newContent), 0644); err != nil {
		fmt.Fprintf(w, "Error writing to log file: %v\n", err)
		return
	}

	fmt.Fprintf(w, "Appended to %s\n", filename)
}

func SearchLogs(query string) {
	var buf strings.Builder
	SearchLogsTo(query, &buf)
	fmt.Print(buf.String())
}

func SearchLogsTo(query string, w io.Writer) {
	logsDir := getLogsDir()
	files, err := os.ReadDir(logsDir)
	if err != nil {
		fmt.Fprintf(w, "Error reading logs directory: %v\n", err)
		return
	}

	type Match struct {
		File    string
		Line    int
		Content string
	}

	var matches []Match
	queryLower := strings.ToLower(query)

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
			continue
		}

		fileContent, err := os.ReadFile(filepath.Join(logsDir, file.Name()))
		if err != nil {
			continue
		}

		lines := strings.Split(string(fileContent), "\n")
		for i, line := range lines {
			if strings.Contains(strings.ToLower(line), queryLower) {
				matches = append(matches, Match{
					File:    file.Name(),
					Line:    i + 1,
					Content: strings.TrimSpace(line),
				})
			}
		}
	}

	if len(matches) == 0 {
		fmt.Fprintln(w, "No matches found")
		return
	}

	fmt.Fprintf(w, "Found %d matches for '%s':\n\n", len(matches), query)
	for _, match := range matches {
		fmt.Fprintf(w, "%s:%d: %s\n", match.File, match.Line, match.Content)
	}
}

func SummarizeLog(slug string) {
	var buf strings.Builder
	SummarizeLogTo(slug, &buf)
	fmt.Print(buf.String())
}

func SummarizeLogTo(slug string, w io.Writer) {
	logsDir := getLogsDir()
	filename := findLogFile(slug)

	if filename == "" {
		fmt.Fprintf(w, "No log file found for slug: %s\n", slug)
		return
	}

	fp := filepath.Join(logsDir, filename)
	content, err := os.ReadFile(fp)
	if err != nil {
		fmt.Fprintf(w, "Error reading log file: %v\n", err)
		return
	}

	lines := strings.Split(string(content), "\n")
	var decisions []string
	var currentDate string

	for _, line := range lines {
		if strings.HasPrefix(line, "## ") && len(line) > 3 {
			trimmed := strings.TrimPrefix(line, "## ")
			if dateRegex.MatchString(trimmed) {
				currentDate = dateRegex.FindString(trimmed)
			}
		}

		if strings.Contains(strings.ToLower(line), "decision:") {
			if currentDate != "" {
				decisions = append(decisions, fmt.Sprintf("[%s] %s", currentDate, strings.TrimSpace(line)))
			} else {
				decisions = append(decisions, strings.TrimSpace(line))
			}
		}
	}

	fmt.Fprintf(w, "Summary of %s:\n\n", filename)
	if len(decisions) == 0 {
		fmt.Fprintln(w, "No decisions recorded")
		return
	}

	fmt.Fprintf(w, "Total decisions: %d\n\n", len(decisions))
	for _, decision := range decisions {
		fmt.Fprintf(w, "- %s\n", decision)
	}
}

func ListLogs() {
	var buf strings.Builder
	ListLogsTo(&buf)
	fmt.Print(buf.String())
}

func ListLogsTo(w io.Writer) {
	logsDir := getLogsDir()
	files, err := os.ReadDir(logsDir)
	if err != nil {
		fmt.Fprintf(w, "Error reading logs directory: %v\n", err)
		return
	}

	var logFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
			logFiles = append(logFiles, file.Name())
		}
	}

	if len(logFiles) == 0 {
		fmt.Fprintln(w, "No session logs found")
		fmt.Fprintln(w, "Use 'bots log start <topic>' to create one")
		return
	}

	sort.Slice(logFiles, func(i, j int) bool {
		info1, _ := os.Stat(filepath.Join(logsDir, logFiles[i]))
		info2, _ := os.Stat(filepath.Join(logsDir, logFiles[j]))
		if info1 == nil || info2 == nil {
			return logFiles[i] < logFiles[j]
		}
		return info1.ModTime().After(info2.ModTime())
	})

	fmt.Fprintln(w, "Session logs:")
	for _, file := range logFiles {
		info, err := os.Stat(filepath.Join(logsDir, file))
		if err != nil {
			fmt.Fprintf(w, "  %s\n", file)
		} else {
			fmt.Fprintf(w, "  %s (%s)\n", file, info.ModTime().Format("2006-01-02"))
		}
	}
}

var dateRegex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}`)

func getLogsDir() string {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting working directory: %v\n", err)
		os.Exit(1)
	}

	for {
		botsDir := filepath.Join(dir, ".bots")
		logsDirPath := filepath.Join(botsDir, "logs")
		if _, err := os.Stat(logsDirPath); err == nil {
			return logsDirPath
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			defaultLogsDir := ".bots/logs"
			if err := os.MkdirAll(defaultLogsDir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating logs directory: %v\n", err)
				os.Exit(1)
			}
			return defaultLogsDir
		}
		dir = parent
	}
}

func createSlug(topic string) string {
	slug := strings.ToLower(topic)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = slugRegex.ReplaceAllString(slug, "")
	slug = multiHyphenRegex.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	return slug
}

var slugRegex = regexp.MustCompile(`[^a-z0-9-]`)
var multiHyphenRegex = regexp.MustCompile(`-+`)

func findLogFile(slug string) string {
	return findFileBySlug(getLogsDir(), slug)
}

func findFileBySlug(dir, slug string) string {
	files, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
			baseName := strings.TrimSuffix(file.Name(), ".md")
			if baseName == slug {
				return file.Name()
			}
		}
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
			baseName := strings.TrimSuffix(file.Name(), ".md")
			if strings.Contains(baseName, slug) {
				return file.Name()
			}
		}
	}

	return ""
}
