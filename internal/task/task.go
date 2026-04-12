package task

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Create creates a new task file with the given slug
func Create(slug string) {
	tasksDir := getTasksDir()
	filename := fmt.Sprintf("%s.md", slug)
	filepath := filepath.Join(tasksDir, filename)

	// Check if file already exists
	if _, err := os.Stat(filepath); err == nil {
		fmt.Fprintf(os.Stderr, "Task file already exists: %s\n", filename)
		os.Exit(1)
	}

	// Create the file with template
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

	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating task file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created task file: %s\n", filename)
	fmt.Println("Edit the file to add description, acceptance criteria, and constraints")
	fmt.Println("Use 'bots task update <slug> IN_PROGRESS' to start working")
}

// Read reads and displays the task file content
func Read(slug string) {
	tasksDir := getTasksDir()
	filename := findTaskFile(slug)

	if filename == "" {
		fmt.Fprintf(os.Stderr, "No task file found for slug: %s\n", slug)
		fmt.Fprintln(os.Stderr, "Use 'bots task list' to see available tasks")
		os.Exit(1)
	}

	filepath := filepath.Join(tasksDir, filename)
	content, err := os.ReadFile(filepath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading task file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(content))
}

// Status reads and displays only the status section
func Status(slug string) {
	tasksDir := getTasksDir()
	filename := findTaskFile(slug)

	if filename == "" {
		fmt.Fprintf(os.Stderr, "No task file found for slug: %s\n", slug)
		os.Exit(1)
	}

	filepath := filepath.Join(tasksDir, filename)
	content, err := os.ReadFile(filepath)
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

// UpdateStatus updates the status of a task file
func UpdateStatus(slug string, newStatus string) {
	tasksDir := getTasksDir()
	filename := findTaskFile(slug)

	if filename == "" {
		fmt.Fprintf(os.Stderr, "No task file found for slug: %s\n", slug)
		os.Exit(1)
	}

	// Validate status
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
		fmt.Fprintf(os.Stderr, "Invalid status: %s\n", newStatus)
		fmt.Fprintf(os.Stderr, "Valid statuses: %s\n", strings.Join(validStatuses, ", "))
		os.Exit(1)
	}

	filepath := filepath.Join(tasksDir, filename)
	content, err := os.ReadFile(filepath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading task file: %v\n", err)
		os.Exit(1)
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
			// Check if we've reached the next section or end of status block
			if strings.HasPrefix(line, "## ") || strings.TrimSpace(line) == "---" {
				if !statusReplaced {
					// Insert status before we leave the section
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

			// Skip the old status line
			if strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "---") {
				continue
			}
		}

		// Handle case where status is at end of file
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

	// Handle case where status section was empty
	if !statusReplaced {
		// Find and insert in the status section
		for i, line := range newLines {
			if strings.TrimSpace(line) == "## Status" && i+1 < len(newLines) {
				// Insert after the header
				newLines = append(newLines[:i+1], append([]string{newStatus, "", "---"}, newLines[i+1:]...)...)
				statusReplaced = true
				break
			}
		}
	}

	if !statusReplaced {
		fmt.Fprintf(os.Stderr, "Could not find status section in task file\n")
		os.Exit(1)
	}

	if err := os.WriteFile(filepath, []byte(strings.Join(newLines, "\n")), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing task file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Updated task %s status to: %s\n", slug, newStatus)

	// Add context-specific message
	switch newStatus {
	case "IN_PROGRESS":
		fmt.Println("Task started. Remember to log decisions with 'bots log append'")
	case "READY_FOR_REVIEW":
		fmt.Println("Task ready for review. Master model will review and update status.")
	case "DONE":
		fmt.Println("Task completed and approved!")
	case "CHANGES_REQUESTED":
		fmt.Println("Changes requested. Address feedback and mark READY_FOR_REVIEW when done.")
	}
}

// List lists all task files
func List() {
	tasksDir := getTasksDir()
	files, err := os.ReadDir(tasksDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading tasks directory: %v\n", err)
		os.Exit(1)
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

			// Read status
			status := readTaskStatus(filepath.Join(tasksDir, file.Name()))
			tasks = append(tasks, TaskInfo{
				Slug:     slug,
				Status:   status,
				Modified: info.ModTime(),
			})
		}
	}

	if len(tasks) == 0 {
		fmt.Println("No task files found")
		fmt.Println("Use 'bots task create <slug>' to create one")
		return
	}

	// Sort by modification time (newest first)
	for i := 0; i < len(tasks)-1; i++ {
		for j := i + 1; j < len(tasks); j++ {
			if tasks[i].Modified.Before(tasks[j].Modified) {
				tasks[i], tasks[j] = tasks[j], tasks[i]
			}
		}
	}

	fmt.Println("Task files:")
	for _, task := range tasks {
		statusIcon := getStatusIcon(task.Status)
		fmt.Printf("  %s %s (%s)\n", statusIcon, task.Slug, task.Modified.Format("2006-01-02"))
	}
}

// Helper functions

func getTasksDir() string {
	// Look for .bots/tasks in current directory or parents
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
			// Reached root, create default structure
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
			if baseName == slug || strings.Contains(baseName, slug) {
				return file.Name()
			}
		}
	}

	return ""
}

func readTaskStatus(filepath string) string {
	content, err := os.ReadFile(filepath)
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
