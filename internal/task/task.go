package task

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

func Create(slug string) {
	var buf strings.Builder
	CreateTo(slug, &buf)
	fmt.Print(buf.String())
}

func CreateTo(slug string, w io.Writer) {
	slug = sanitizeSlug(slug)
	tasksDir := getTasksDir()
	filename := fmt.Sprintf("%s.md", slug)
	fp := filepath.Join(tasksDir, filename)

	if _, err := os.Stat(fp); err == nil {
		fmt.Fprintf(w, "Error: Task file already exists: %s\n", filename)
		return
	}

	content := fmt.Sprintf(`# %s

## Description

<TODO: Add task description>

## Acceptance criteria

- <TODO: Add criterion 1>
- <TODO: Add criterion 2>

## Constraints

- <TODO: Add constraints or links to RULES.md>

---

## Status

PENDING

---
`, slug)

	if err := os.WriteFile(fp, []byte(content), 0644); err != nil {
		fmt.Fprintf(w, "Error creating task file: %v\n", err)
		return
	}

	fmt.Fprintf(w, "Created task file: %s\n", filename)
	fmt.Fprintln(w, "Edit the file to add description, acceptance criteria, and constraints")
	fmt.Fprintln(w, "Use 'bots task update <slug> IN_PROGRESS' to start working")
}

func Read(slug string) {
	var buf strings.Builder
	ReadTo(slug, &buf)
	fmt.Print(buf.String())
}

func ReadTo(slug string, w io.Writer) {
	tasksDir := getTasksDir()
	filename := findTaskFile(slug)

	if filename == "" {
		fmt.Fprintf(w, "No task file found for slug: %s\n", slug)
		fmt.Fprintln(w, "Use 'bots task list' to see available tasks")
		return
	}

	fp := filepath.Join(tasksDir, filename)
	content, err := os.ReadFile(fp)
	if err != nil {
		fmt.Fprintf(w, "Error reading task file: %v\n", err)
		return
	}

	fmt.Fprintln(w, string(content))
}

func Status(slug string) {
	tasksDir := getTasksDir()
	filename := findTaskFile(slug)

	if filename == "" {
		fmt.Fprintf(os.Stderr, "No task file found for slug: %s\n", slug)
		os.Exit(1)
	}

	fp := filepath.Join(tasksDir, filename)
	content, err := os.ReadFile(fp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading task file: %v\n", err)
		os.Exit(1)
	}

	lines := strings.Split(string(content), "\n")
	inStatusSection := false
	statusFound := false

	for _, line := range lines {
		if strings.TrimSpace(line) == "## Status" {
			inStatusSection = true
			continue
		}

		if inStatusSection {
			if strings.HasPrefix(line, "## ") {
				break
			}
			trimmed := strings.TrimSpace(line)
			if trimmed != "" && trimmed != "---" {
				fmt.Printf("Task %s: %s\n", slug, trimmed)
				statusFound = true
				break
			}
		}
	}

	if !statusFound {
		fmt.Printf("Task %s: Status not found\n", slug)
	}
}

func UpdateStatus(slug string, newStatus string) {
	var buf strings.Builder
	UpdateStatusTo(slug, newStatus, &buf)
	fmt.Print(buf.String())
}

func UpdateStatusTo(slug string, newStatus string, w io.Writer) {
	tasksDir := getTasksDir()
	filename := findTaskFile(slug)

	if filename == "" {
		fmt.Fprintf(w, "No task file found for slug: %s\n", slug)
		return
	}

	validStatuses := []string{"PENDING", "IN_PROGRESS", "READY_FOR_REVIEW", "CHANGES_REQUESTED", "DONE"}
	valid := false
	for _, validStatus := range validStatuses {
		if strings.ToUpper(newStatus) == validStatus {
			valid = true
			newStatus = validStatus
			break
		}
	}

	if !valid {
		fmt.Fprintf(w, "Invalid status: %s\n", newStatus)
		fmt.Fprintf(w, "Valid statuses: %s\n", strings.Join(validStatuses, ", "))
		return
	}

	fp := filepath.Join(tasksDir, filename)
	content, err := os.ReadFile(fp)
	if err != nil {
		fmt.Fprintf(w, "Error reading task file: %v\n", err)
		return
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string
	inStatusSection := false
	statusReplaced := false

	for i, line := range lines {
		if strings.TrimSpace(line) == "## Status" {
			inStatusSection = true
			newLines = append(newLines, line)
			continue
		}

		if inStatusSection {
			if strings.HasPrefix(line, "## ") || strings.TrimSpace(line) == "---" {
				if !statusReplaced {
					newLines = append(newLines, newStatus)
					newLines = append(newLines, "")
					newLines = append(newLines, "---")
					statusReplaced = true
				}
				inStatusSection = false
				if strings.TrimSpace(line) != "---" {
					newLines = append(newLines, line)
				}
				continue
			}

			if strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "---") {
				continue
			}
		}

		if i == len(lines)-1 && inStatusSection && !statusReplaced {
			newLines = append(newLines, newStatus)
			newLines = append(newLines, "")
			newLines = append(newLines, "---")
			statusReplaced = true
			break
		}

		if !inStatusSection || strings.HasPrefix(line, "## ") {
			newLines = append(newLines, line)
		}
	}

	if !statusReplaced {
		for i, line := range newLines {
			if strings.TrimSpace(line) == "## Status" && i+1 < len(newLines) {
				newLines = append(newLines[:i+1], append([]string{newStatus, "", "---"}, newLines[i+1:]...)...)
				statusReplaced = true
				break
			}
		}
	}

	if !statusReplaced {
		fmt.Fprintln(w, "Could not find status section in task file")
		return
	}

	if err := os.WriteFile(fp, []byte(strings.Join(newLines, "\n")), 0644); err != nil {
		fmt.Fprintf(w, "Error writing task file: %v\n", err)
		return
	}

	fmt.Fprintf(w, "Updated task %s status to: %s\n", slug, newStatus)

	switch newStatus {
	case "IN_PROGRESS":
		fmt.Fprintln(w, "Task started. Remember to log decisions with 'bots log append'")
	case "READY_FOR_REVIEW":
		fmt.Fprintln(w, "Task ready for review. Master model will review and update status.")
	case "DONE":
		fmt.Fprintln(w, "Task completed and approved!")
	case "CHANGES_REQUESTED":
		fmt.Fprintln(w, "Changes requested. Address feedback and mark READY_FOR_REVIEW when done.")
	}
}

func List() {
	var buf strings.Builder
	ListTo(&buf)
	fmt.Print(buf.String())
}

func ListTo(w io.Writer) {
	tasksDir := getTasksDir()
	files, err := os.ReadDir(tasksDir)
	if err != nil {
		fmt.Fprintf(w, "Error reading tasks directory: %v\n", err)
		return
	}

	type TaskInfo struct {
		Slug     string
		Status   string
		Modified time.Time
	}

	var tasks []TaskInfo

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
			slug := strings.TrimSuffix(file.Name(), ".md")
			info, err := file.Info()
			if err != nil {
				continue
			}

			status := readTaskStatus(filepath.Join(tasksDir, file.Name()))
			tasks = append(tasks, TaskInfo{
				Slug:     slug,
				Status:   status,
				Modified: info.ModTime(),
			})
		}
	}

	if len(tasks) == 0 {
		fmt.Fprintln(w, "No task files found")
		fmt.Fprintln(w, "Use 'bots task create <slug>' to create one")
		return
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].Modified.After(tasks[j].Modified)
	})

	fmt.Fprintln(w, "Task files:")
	for _, task := range tasks {
		statusIcon := getStatusIcon(task.Status)
		fmt.Fprintf(w, "  %s %s (%s)\n", statusIcon, task.Slug, task.Modified.Format("2006-01-02"))
	}
}

func getTasksDir() string {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting working directory: %v\n", err)
		os.Exit(1)
	}

	for {
		botsDir := filepath.Join(dir, ".bots")
		tasksDir := filepath.Join(botsDir, "tasks")
		if _, err := os.Stat(tasksDir); err == nil {
			return tasksDir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			defaultTasksDir := ".bots/tasks"
			if err := os.MkdirAll(defaultTasksDir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating tasks directory: %v\n", err)
				os.Exit(1)
			}
			return defaultTasksDir
		}
		dir = parent
	}
}

func findTaskFile(slug string) string {
	tasksDir := getTasksDir()
	files, err := os.ReadDir(tasksDir)
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

func readTaskStatus(fp string) string {
	content, err := os.ReadFile(fp)
	if err != nil {
		return "UNKNOWN"
	}

	lines := strings.Split(string(content), "\n")
	inStatusSection := false

	for _, line := range lines {
		if strings.TrimSpace(line) == "## Status" {
			inStatusSection = true
			continue
		}

		if inStatusSection {
			if strings.HasPrefix(line, "## ") || strings.TrimSpace(line) == "---" {
				break
			}
			trimmed := strings.TrimSpace(line)
			if trimmed != "" {
				return trimmed
			}
		}
	}

	return "UNKNOWN"
}

func getStatusIcon(status string) string {
	switch status {
	case "PENDING":
		return "○"
	case "IN_PROGRESS":
		return "◐"
	case "READY_FOR_REVIEW":
		return "●"
	case "CHANGES_REQUESTED":
		return "▲"
	case "DONE":
		return "✓"
	default:
		return "?"
	}
}

var slugRegex = regexp.MustCompile(`[^a-z0-9._-]`)
var multiHyphenRegex = regexp.MustCompile(`-+`)

func sanitizeSlug(slug string) string {
	slug = strings.ToLower(slug)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, ".", "-")
	slug = slugRegex.ReplaceAllString(slug, "")
	slug = multiHyphenRegex.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = "task"
	}
	if strings.HasPrefix(slug, "-") {
		slug = strings.TrimPrefix(slug, "-")
	}
	return slug
}
