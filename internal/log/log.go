package log

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// StartLog creates a new session log file with the given topic
func StartLog(topic string) {
	logsDir := getLogsDir()
	slug := createSlug(topic)
	filename := fmt.Sprintf("%s-%s.md", time.Now().Format("2006-01-02"), slug)
	filepath := filepath.Join(logsDir, filename)

	// Check if file already exists
	if _, err := os.Stat(filepath); err == nil {
		fmt.Fprintf(os.Stderr, "Log file already exists: %s\n", filename)
		fmt.Fprintln(os.Stderr, "Use a different topic or append to existing log")
		os.Exit(1)
	}

	// Create the file with header
	content := fmt.Sprintf(`# Session Log: %s

## %s

- Started session on topic: %s

`, topic, time.Now().Format("2006-01-02"), topic)

	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating log file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created session log: %s\n", filename)
	fmt.Printf("Use 'bots log append %s \"<message>\"' to add entries\n", slug)
}

// AppendEntry appends a dated entry to the specified log file
func AppendEntry(slug string, message string) {
	logsDir := getLogsDir()
	filename := findLogFile(slug)

	if filename == "" {
		fmt.Fprintf(os.Stderr, "No log file found for slug: %s\n", slug)
		fmt.Fprintln(os.Stderr, "Use 'bots log list' to see available logs")
		os.Exit(1)
	}

	filepath := filepath.Join(logsDir, filename)

	// Read existing content
	content, err := os.ReadFile(filepath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading log file: %v\n", err)
		os.Exit(1)
	}

	// Append new entry
	now := time.Now()
	entry := fmt.Sprintf("\n## %s\n\n- %s\n\n", now.Format("2006-01-02"), message)
	newContent := string(content) + entry

	if err := os.WriteFile(filepath, []byte(newContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to log file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Appended to %s\n", filename)
}

// SearchLogs searches across all log files for the given query
func SearchLogs(query string) {
	logsDir := getLogsDir()
	files, err := os.ReadDir(logsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading logs directory: %v\n", err)
		os.Exit(1)
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

		content, err := os.ReadFile(filepath.Join(logsDir, file.Name()))
		if err != nil {
			continue
		}

		lines := strings.Split(string(content), "\n")
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
		fmt.Println("No matches found")
		return
	}

	fmt.Printf("Found %d matches for '%s':\n\n", len(matches), query)
	for _, match := range matches {
		fmt.Printf("%s:%d: %s\n", match.File, match.Line, match.Content)
	}
}

// SummarizeLog generates a summary of decisions from a log file
func SummarizeLog(slug string) {
	logsDir := getLogsDir()
	filename := findLogFile(slug)

	if filename == "" {
		fmt.Fprintf(os.Stderr, "No log file found for slug: %s\n", slug)
		os.Exit(1)
	}

	filepath := filepath.Join(logsDir, filename)
	content, err := os.ReadFile(filepath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading log file: %v\n", err)
		os.Exit(1)
	}

	lines := strings.Split(string(content), "\n")
	var decisions []string
	var currentDate string

	for _, line := range lines {
		// Extract date headers
		if strings.HasPrefix(line, "## ") && len(line) > 3 {
			dateRegex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}`)
			if match := dateRegex.FindString(strings.TrimPrefix(line, "## ")); match != "" {
				currentDate = match
			}
		}

		// Extract decision entries
		if strings.Contains(strings.ToLower(line), "decision:") {
			if currentDate != "" {
				decisions = append(decisions, fmt.Sprintf("[%s] %s", currentDate, strings.TrimSpace(line)))
			} else {
				decisions = append(decisions, strings.TrimSpace(line))
			}
		}
	}

	fmt.Printf("Summary of %s:\n\n", filename)
	if len(decisions) == 0 {
		fmt.Println("No decisions recorded")
		return
	}

	fmt.Printf("Total decisions: %d\n\n", len(decisions))
	for _, decision := range decisions {
		fmt.Printf("- %s\n", decision)
	}
}

// ListLogs lists all session log files
func ListLogs() {
	logsDir := getLogsDir()
	files, err := os.ReadDir(logsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading logs directory: %v\n", err)
		os.Exit(1)
	}

	var logFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
			logFiles = append(logFiles, file.Name())
		}
	}

	if len(logFiles) == 0 {
		fmt.Println("No session logs found")
		fmt.Println("Use 'bots log start <topic>' to create one")
		return
	}

	// Sort by modification time (newest first)
	sort.Slice(logFiles, func(i, j int) bool {
		info1, _ := os.Stat(filepath.Join(logsDir, logFiles[i]))
		info2, _ := os.Stat(filepath.Join(logsDir, logFiles[j]))
		if info1 == nil || info2 == nil {
			return logFiles[i] < logFiles[j]
		}
		return info1.ModTime().After(info2.ModTime())
	})

	fmt.Println("Session logs:")
	for _, file := range logFiles {
		info, err := os.Stat(filepath.Join(logsDir, file))
		if err != nil {
			fmt.Printf("  %s\n", file)
		} else {
			fmt.Printf("  %s (%s)\n", file, info.ModTime().Format("2006-01-02"))
		}
	}
}

// Helper functions

func getLogsDir() string {
	// Look for .bots/logs in current directory or parents
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting working directory: %v\n", err)
		os.Exit(1)
	}

	for {
		botsDir := filepath.Join(dir, ".bots")
		logsDir := filepath.Join(botsDir, "logs")
		if _, err := os.Stat(logsDir); err == nil {
			return logsDir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root, create default structure
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
	// Convert to lowercase
	slug := strings.ToLower(topic)
	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove special characters
	slug = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(slug, "")
	// Remove consecutive hyphens
	slug = regexp.MustCompile(`-+`).ReplaceAllString(slug, "-")
	// Trim hyphens from ends
	slug = strings.Trim(slug, "-")
	return slug
}

func findLogFile(slug string) string {
	logsDir := getLogsDir()
	files, err := os.ReadDir(logsDir)
	if err != nil {
		return ""
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
			if strings.Contains(file.Name(), slug) {
				return file.Name()
			}
		}
	}

	return ""
}
