package task

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"bots/internal/document"
	"bots/internal/workspace"
)

type TaskStatus string

const (
	StatusPending          TaskStatus = "PENDING"
	StatusInProgress       TaskStatus = "IN_PROGRESS"
	StatusReadyForReview   TaskStatus = "READY_FOR_REVIEW"
	StatusChangesRequested TaskStatus = "CHANGES_REQUESTED"
	StatusDone             TaskStatus = "DONE"
	StatusUnknown          TaskStatus = "UNKNOWN"
)

var (
	ErrTaskNotFound  = errors.New("task handoff not found")
	ErrTaskExists    = errors.New("task handoff already exists")
	ErrInvalidStatus = errors.New("invalid task status")
	ErrStatusMissing = errors.New("task status not found")
)

// Store manages task handoff files in a project workspace.
type Store struct {
	workspace workspace.Workspace
}

type CreateResult struct {
	Slug string
	Path string
}

type ReadResult struct {
	Slug    string
	Path    string
	Content string
}

type StatusResult struct {
	Slug   string
	Path   string
	Status TaskStatus
}

type UpdateStatusResult struct {
	Slug   string
	Path   string
	Status TaskStatus
}

type TaskInfo struct {
	Slug     string
	Path     string
	Status   TaskStatus
	Modified time.Time
}

type ListResult struct {
	Tasks []TaskInfo
}

func NewStore(ws workspace.Workspace) Store {
	return Store{workspace: ws}
}

func NewDefaultStore() (Store, error) {
	ws, err := workspace.FromCurrent(true)
	if err != nil {
		return Store{}, err
	}
	return NewStore(ws), nil
}

func (s Store) Create(slug string) (CreateResult, error) {
	slug = sanitizeSlug(slug)
	if err := s.workspace.EnsureTasksDir(); err != nil {
		return CreateResult{}, err
	}

	path := filepath.Join(s.workspace.TasksDir(), slug+".md")
	if _, err := os.Stat(path); err == nil {
		return CreateResult{Slug: slug, Path: path}, ErrTaskExists
	} else if !os.IsNotExist(err) {
		return CreateResult{}, fmt.Errorf("stat task handoff: %w", err)
	}

	if err := os.WriteFile(path, []byte(createTaskContent(slug)), 0644); err != nil {
		return CreateResult{}, fmt.Errorf("write task handoff: %w", err)
	}

	return CreateResult{Slug: slug, Path: path}, nil
}

func (s Store) Read(slug string) (ReadResult, error) {
	filename, err := s.findTaskFile(slug)
	if err != nil {
		return ReadResult{}, err
	}

	path := filepath.Join(s.workspace.TasksDir(), filename)
	content, err := os.ReadFile(path)
	if err != nil {
		return ReadResult{}, fmt.Errorf("read task handoff: %w", err)
	}

	return ReadResult{Slug: strings.TrimSuffix(filename, ".md"), Path: path, Content: string(content)}, nil
}

func (s Store) Status(slug string) (StatusResult, error) {
	read, err := s.Read(slug)
	if err != nil {
		return StatusResult{}, err
	}

	status, err := readTaskStatus(read.Content)
	if err != nil {
		return StatusResult{}, err
	}

	return StatusResult{Slug: read.Slug, Path: read.Path, Status: status}, nil
}

func (s Store) UpdateStatus(slug string, newStatus string) (UpdateStatusResult, error) {
	status, err := ParseStatus(newStatus)
	if err != nil {
		return UpdateStatusResult{}, err
	}

	read, err := s.Read(slug)
	if err != nil {
		return UpdateStatusResult{}, err
	}

	newContent := document.UpsertSection(read.Content, "## Status", string(status)+"\n\n---")
	if err := os.WriteFile(read.Path, []byte(newContent), 0644); err != nil {
		return UpdateStatusResult{}, fmt.Errorf("write task handoff: %w", err)
	}

	return UpdateStatusResult{Slug: read.Slug, Path: read.Path, Status: status}, nil
}

func (s Store) List() (ListResult, error) {
	if err := s.workspace.EnsureTasksDir(); err != nil {
		return ListResult{}, err
	}

	files, err := os.ReadDir(s.workspace.TasksDir())
	if err != nil {
		return ListResult{}, fmt.Errorf("read tasks directory: %w", err)
	}

	var tasks []TaskInfo
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		path := filepath.Join(s.workspace.TasksDir(), file.Name())
		status := readTaskStatusFromFile(path)
		tasks = append(tasks, TaskInfo{
			Slug:     strings.TrimSuffix(file.Name(), ".md"),
			Path:     path,
			Status:   status,
			Modified: info.ModTime(),
		})
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].Modified.After(tasks[j].Modified)
	})

	return ListResult{Tasks: tasks}, nil
}

func ParseStatus(status string) (TaskStatus, error) {
	candidate := TaskStatus(strings.ToUpper(status))
	for _, valid := range ValidStatuses() {
		if candidate == valid {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("%w: %s", ErrInvalidStatus, status)
}

func ValidStatuses() []TaskStatus {
	return []TaskStatus{StatusPending, StatusInProgress, StatusReadyForReview, StatusChangesRequested, StatusDone}
}

func (s Store) findTaskFile(slug string) (string, error) {
	if strings.TrimSpace(slug) == "" {
		return "", fmt.Errorf("%w: %s", ErrTaskNotFound, slug)
	}

	if err := s.workspace.EnsureTasksDir(); err != nil {
		return "", err
	}

	files, err := os.ReadDir(s.workspace.TasksDir())
	if err != nil {
		return "", fmt.Errorf("read tasks directory: %w", err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
			baseName := strings.TrimSuffix(file.Name(), ".md")
			if baseName == slug {
				return file.Name(), nil
			}
		}
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
			baseName := strings.TrimSuffix(file.Name(), ".md")
			if strings.Contains(baseName, slug) {
				return file.Name(), nil
			}
		}
	}

	return "", fmt.Errorf("%w: %s", ErrTaskNotFound, slug)
}

func createTaskContent(slug string) string {
	return fmt.Sprintf(`# %s

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
}

func readTaskStatusFromFile(path string) TaskStatus {
	content, err := os.ReadFile(path)
	if err != nil {
		return StatusUnknown
	}

	status, err := readTaskStatus(string(content))
	if err != nil {
		return StatusUnknown
	}
	return status
}

func readTaskStatus(content string) (TaskStatus, error) {
	body, ok := document.SectionBody(content, "## Status")
	if !ok {
		return StatusUnknown, ErrStatusMissing
	}

	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || trimmed == "---" {
			continue
		}
		return TaskStatus(trimmed), nil
	}

	return StatusUnknown, ErrStatusMissing
}

func Create(slug string) {
	var buf strings.Builder
	CreateTo(slug, &buf)
	fmt.Print(buf.String())
}

func CreateTo(slug string, w io.Writer) {
	store, err := NewDefaultStore()
	if err != nil {
		fmt.Fprintf(w, "Error creating task file: %v\n", err)
		return
	}

	result, err := store.Create(slug)
	if err != nil {
		if errors.Is(err, ErrTaskExists) {
			fmt.Fprintf(w, "Error: Task file already exists: %s.md\n", result.Slug)
			return
		}
		fmt.Fprintf(w, "Error creating task file: %v\n", err)
		return
	}

	WriteCreateResult(w, result)
}

func WriteCreateResult(w io.Writer, result CreateResult) {
	fmt.Fprintf(w, "Created task file: %s.md\n", result.Slug)
	fmt.Fprintln(w, "Edit the file to add description, acceptance criteria, and constraints")
	fmt.Fprintln(w, "Use 'bots task update <slug> IN_PROGRESS' to start working")
}

func Read(slug string) {
	var buf strings.Builder
	ReadTo(slug, &buf)
	fmt.Print(buf.String())
}

func ReadTo(slug string, w io.Writer) {
	store, err := NewDefaultStore()
	if err != nil {
		fmt.Fprintf(w, "Error reading task file: %v\n", err)
		return
	}

	result, err := store.Read(slug)
	if err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			fmt.Fprintf(w, "No task file found for slug: %s\n", slug)
			fmt.Fprintln(w, "Use 'bots task list' to see available tasks")
			return
		}
		fmt.Fprintf(w, "Error reading task file: %v\n", err)
		return
	}

	WriteReadResult(w, result)
}

func WriteReadResult(w io.Writer, result ReadResult) {
	fmt.Fprint(w, result.Content)
	if !strings.HasSuffix(result.Content, "\n") {
		fmt.Fprintln(w)
	}
}

func Status(slug string) {
	store, err := NewDefaultStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading task file: %v\n", err)
		os.Exit(1)
	}

	result, err := store.Status(slug)
	if err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			fmt.Fprintf(os.Stderr, "No task file found for slug: %s\n", slug)
		} else {
			fmt.Fprintf(os.Stderr, "Error reading task file: %v\n", err)
		}
		os.Exit(1)
	}

	fmt.Printf("Task %s: %s\n", slug, result.Status)
}

func UpdateStatus(slug string, newStatus string) {
	var buf strings.Builder
	UpdateStatusTo(slug, newStatus, &buf)
	fmt.Print(buf.String())
}

func UpdateStatusTo(slug string, newStatus string, w io.Writer) {
	store, err := NewDefaultStore()
	if err != nil {
		fmt.Fprintf(w, "Error reading task file: %v\n", err)
		return
	}

	result, err := store.UpdateStatus(slug, newStatus)
	if err != nil {
		switch {
		case errors.Is(err, ErrTaskNotFound):
			fmt.Fprintf(w, "No task file found for slug: %s\n", slug)
		case errors.Is(err, ErrInvalidStatus):
			fmt.Fprintf(w, "Invalid status: %s\n", newStatus)
			fmt.Fprintf(w, "Valid statuses: %s\n", strings.Join(statusStrings(ValidStatuses()), ", "))
		default:
			fmt.Fprintf(w, "Error writing task file: %v\n", err)
		}
		return
	}

	WriteUpdateStatusResult(w, result)
}

func WriteUpdateStatusResult(w io.Writer, result UpdateStatusResult) {
	fmt.Fprintf(w, "Updated task %s status to: %s\n", result.Slug, result.Status)

	switch result.Status {
	case StatusInProgress:
		fmt.Fprintln(w, "Task started. Remember to log decisions with 'bots log append'")
	case StatusReadyForReview:
		fmt.Fprintln(w, "Task ready for review. Master model will review and update status.")
	case StatusDone:
		fmt.Fprintln(w, "Task completed and approved!")
	case StatusChangesRequested:
		fmt.Fprintln(w, "Changes requested. Address feedback and mark READY_FOR_REVIEW when done.")
	}
}

func List() {
	var buf strings.Builder
	ListTo(&buf)
	fmt.Print(buf.String())
}

func ListTo(w io.Writer) {
	store, err := NewDefaultStore()
	if err != nil {
		fmt.Fprintf(w, "Error reading tasks directory: %v\n", err)
		return
	}

	result, err := store.List()
	if err != nil {
		fmt.Fprintf(w, "Error reading tasks directory: %v\n", err)
		return
	}

	WriteListResult(w, result)
}

func WriteListResult(w io.Writer, result ListResult) {
	if len(result.Tasks) == 0 {
		fmt.Fprintln(w, "No task files found")
		fmt.Fprintln(w, "Use 'bots task create <slug>' to create one")
		return
	}

	fmt.Fprintln(w, "Task files:")
	for _, task := range result.Tasks {
		fmt.Fprintf(w, "  %s %s (%s)\n", getStatusIcon(task.Status), task.Slug, task.Modified.Format("2006-01-02"))
	}
}

func getStatusIcon(status TaskStatus) string {
	switch status {
	case StatusPending:
		return "○"
	case StatusInProgress:
		return "◐"
	case StatusReadyForReview:
		return "●"
	case StatusChangesRequested:
		return "▲"
	case StatusDone:
		return "✓"
	default:
		return "?"
	}
}

func statusStrings(statuses []TaskStatus) []string {
	out := make([]string, len(statuses))
	for i, status := range statuses {
		out[i] = string(status)
	}
	return out
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
