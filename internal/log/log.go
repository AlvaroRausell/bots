package log

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

	"bots/internal/workspace"
)

var (
	ErrLogExists   = errors.New("session log already exists")
	ErrLogNotFound = errors.New("session log not found")
)

// Store manages session logs in a project workspace.
type Store struct {
	workspace workspace.Workspace
	now       func() time.Time
}

type StartResult struct {
	Topic    string
	Slug     string
	Filename string
	Path     string
}

type AppendResult struct {
	Slug     string
	Filename string
	Path     string
}

type Match struct {
	File    string
	Line    int
	Content string
}

type SearchResult struct {
	Query   string
	Matches []Match
}

type Decision struct {
	Date    string
	Content string
}

type SummaryResult struct {
	Slug      string
	Filename  string
	Decisions []Decision
}

type LogInfo struct {
	Filename string
	Path     string
	Modified time.Time
}

type ListResult struct {
	Logs []LogInfo
}

func NewStore(ws workspace.Workspace) Store {
	return NewStoreWithClock(ws, time.Now)
}

func NewStoreWithClock(ws workspace.Workspace, now func() time.Time) Store {
	if now == nil {
		now = time.Now
	}
	return Store{workspace: ws, now: now}
}

func NewDefaultStore() (Store, error) {
	ws, err := workspace.FromCurrent(true)
	if err != nil {
		return Store{}, err
	}
	return NewStore(ws), nil
}

func (s Store) Start(topic string) (StartResult, error) {
	if err := s.workspace.EnsureLogsDir(); err != nil {
		return StartResult{}, err
	}

	slug := createSlug(topic)
	filename := fmt.Sprintf("%s-%s.md", s.now().Format("2006-01-02"), slug)
	path := filepath.Join(s.workspace.LogsDir(), filename)
	result := StartResult{Topic: topic, Slug: slug, Filename: filename, Path: path}

	if _, err := os.Stat(path); err == nil {
		return result, ErrLogExists
	} else if !os.IsNotExist(err) {
		return StartResult{}, fmt.Errorf("stat session log: %w", err)
	}

	content := fmt.Sprintf(`# Session Log: %s

## %s

- Started session on topic: %s

`, topic, s.now().Format("2006-01-02"), topic)

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return StartResult{}, fmt.Errorf("write session log: %w", err)
	}

	return result, nil
}

func (s Store) Append(slug string, message string) (AppendResult, error) {
	filename, err := s.findLogFile(slug)
	if err != nil {
		return AppendResult{}, err
	}

	path := filepath.Join(s.workspace.LogsDir(), filename)
	fileContent, err := os.ReadFile(path)
	if err != nil {
		return AppendResult{}, fmt.Errorf("read session log: %w", err)
	}

	newContent := appendMessageToDateSection(string(fileContent), s.now().Format("2006-01-02"), message)

	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		return AppendResult{}, fmt.Errorf("write session log: %w", err)
	}

	return AppendResult{Slug: strings.TrimSuffix(filename, ".md"), Filename: filename, Path: path}, nil
}

func (s Store) Search(query string) (SearchResult, error) {
	if err := s.workspace.EnsureLogsDir(); err != nil {
		return SearchResult{}, err
	}

	files, err := os.ReadDir(s.workspace.LogsDir())
	if err != nil {
		return SearchResult{}, fmt.Errorf("read logs directory: %w", err)
	}

	var matches []Match
	queryLower := strings.ToLower(query)

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
			continue
		}

		fileContent, err := os.ReadFile(filepath.Join(s.workspace.LogsDir(), file.Name()))
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

	return SearchResult{Query: query, Matches: matches}, nil
}

func (s Store) Summarize(slug string) (SummaryResult, error) {
	filename, err := s.findLogFile(slug)
	if err != nil {
		return SummaryResult{}, err
	}

	path := filepath.Join(s.workspace.LogsDir(), filename)
	content, err := os.ReadFile(path)
	if err != nil {
		return SummaryResult{}, fmt.Errorf("read session log: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	var decisions []Decision
	var currentDate string

	for _, line := range lines {
		if strings.HasPrefix(line, "## ") && len(line) > 3 {
			trimmed := strings.TrimPrefix(line, "## ")
			if dateRegex.MatchString(trimmed) {
				currentDate = dateRegex.FindString(trimmed)
			}
		}

		if strings.Contains(strings.ToLower(line), "decision:") {
			decisions = append(decisions, Decision{Date: currentDate, Content: strings.TrimSpace(line)})
		}
	}

	return SummaryResult{Slug: strings.TrimSuffix(filename, ".md"), Filename: filename, Decisions: decisions}, nil
}

func (s Store) List() (ListResult, error) {
	if err := s.workspace.EnsureLogsDir(); err != nil {
		return ListResult{}, err
	}

	files, err := os.ReadDir(s.workspace.LogsDir())
	if err != nil {
		return ListResult{}, fmt.Errorf("read logs directory: %w", err)
	}

	var logs []LogInfo
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
			continue
		}
		info, err := file.Info()
		if err != nil {
			continue
		}
		logs = append(logs, LogInfo{
			Filename: file.Name(),
			Path:     filepath.Join(s.workspace.LogsDir(), file.Name()),
			Modified: info.ModTime(),
		})
	}

	sort.Slice(logs, func(i, j int) bool {
		return logs[i].Modified.After(logs[j].Modified)
	})

	return ListResult{Logs: logs}, nil
}

func (s Store) findLogFile(slug string) (string, error) {
	if strings.TrimSpace(slug) == "" {
		return "", fmt.Errorf("%w: %s", ErrLogNotFound, slug)
	}

	if err := s.workspace.EnsureLogsDir(); err != nil {
		return "", err
	}

	files, err := os.ReadDir(s.workspace.LogsDir())
	if err != nil {
		return "", fmt.Errorf("read logs directory: %w", err)
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

	return "", fmt.Errorf("%w: %s", ErrLogNotFound, slug)
}

func StartLog(topic string) {
	var buf strings.Builder
	StartLogTo(topic, &buf)
	fmt.Print(buf.String())
}

func StartLogTo(topic string, w io.Writer) {
	store, err := NewDefaultStore()
	if err != nil {
		fmt.Fprintf(w, "Error creating log file: %v\n", err)
		return
	}

	result, err := store.Start(topic)
	if err != nil {
		if errors.Is(err, ErrLogExists) {
			fmt.Fprintf(w, "Log file already exists: %s\n", result.Filename)
			fmt.Fprintln(w, "Use a different topic or append to existing log")
			return
		}
		fmt.Fprintf(w, "Error creating log file: %v\n", err)
		return
	}

	WriteStartResult(w, result)
}

func WriteStartResult(w io.Writer, result StartResult) {
	fmt.Fprintf(w, "Created session log: %s\n", result.Filename)
	fmt.Fprintf(w, "Use 'bots log append %s \"<message>\"' to add entries\n", result.Slug)
}

func AppendEntry(slug string, message string) {
	var buf strings.Builder
	AppendEntryTo(slug, message, &buf)
	fmt.Print(buf.String())
}

func AppendEntryTo(slug string, message string, w io.Writer) {
	store, err := NewDefaultStore()
	if err != nil {
		fmt.Fprintf(w, "Error reading log file: %v\n", err)
		return
	}

	result, err := store.Append(slug, message)
	if err != nil {
		if errors.Is(err, ErrLogNotFound) {
			fmt.Fprintf(w, "No log file found for slug: %s\n", slug)
			fmt.Fprintln(w, "Use 'bots log list' to see available logs")
			return
		}
		fmt.Fprintf(w, "Error writing to log file: %v\n", err)
		return
	}

	WriteAppendResult(w, result)
}

func WriteAppendResult(w io.Writer, result AppendResult) {
	fmt.Fprintf(w, "Appended to %s\n", result.Filename)
}

func SearchLogs(query string) {
	var buf strings.Builder
	SearchLogsTo(query, &buf)
	fmt.Print(buf.String())
}

func SearchLogsTo(query string, w io.Writer) {
	store, err := NewDefaultStore()
	if err != nil {
		fmt.Fprintf(w, "Error reading logs directory: %v\n", err)
		return
	}

	result, err := store.Search(query)
	if err != nil {
		fmt.Fprintf(w, "Error reading logs directory: %v\n", err)
		return
	}

	WriteSearchResult(w, result)
}

func WriteSearchResult(w io.Writer, result SearchResult) {
	if len(result.Matches) == 0 {
		fmt.Fprintln(w, "No matches found")
		return
	}

	fmt.Fprintf(w, "Found %d matches for '%s':\n\n", len(result.Matches), result.Query)
	for _, match := range result.Matches {
		fmt.Fprintf(w, "%s:%d: %s\n", match.File, match.Line, match.Content)
	}
}

func SummarizeLog(slug string) {
	var buf strings.Builder
	SummarizeLogTo(slug, &buf)
	fmt.Print(buf.String())
}

func SummarizeLogTo(slug string, w io.Writer) {
	store, err := NewDefaultStore()
	if err != nil {
		fmt.Fprintf(w, "Error reading log file: %v\n", err)
		return
	}

	result, err := store.Summarize(slug)
	if err != nil {
		if errors.Is(err, ErrLogNotFound) {
			fmt.Fprintf(w, "No log file found for slug: %s\n", slug)
			return
		}
		fmt.Fprintf(w, "Error reading log file: %v\n", err)
		return
	}

	WriteSummaryResult(w, result)
}

func WriteSummaryResult(w io.Writer, result SummaryResult) {
	fmt.Fprintf(w, "Summary of %s:\n\n", result.Filename)
	if len(result.Decisions) == 0 {
		fmt.Fprintln(w, "No decisions recorded")
		return
	}

	fmt.Fprintf(w, "Total decisions: %d\n\n", len(result.Decisions))
	for _, decision := range result.Decisions {
		if decision.Date != "" {
			fmt.Fprintf(w, "- [%s] %s\n", decision.Date, decision.Content)
		} else {
			fmt.Fprintf(w, "- %s\n", decision.Content)
		}
	}
}

func ListLogs() {
	var buf strings.Builder
	ListLogsTo(&buf)
	fmt.Print(buf.String())
}

func ListLogsTo(w io.Writer) {
	store, err := NewDefaultStore()
	if err != nil {
		fmt.Fprintf(w, "Error reading logs directory: %v\n", err)
		return
	}

	result, err := store.List()
	if err != nil {
		fmt.Fprintf(w, "Error reading logs directory: %v\n", err)
		return
	}

	WriteListResult(w, result)
}

func WriteListResult(w io.Writer, result ListResult) {
	if len(result.Logs) == 0 {
		fmt.Fprintln(w, "No session logs found")
		fmt.Fprintln(w, "Use 'bots log start <topic>' to create one")
		return
	}

	fmt.Fprintln(w, "Session logs:")
	for _, file := range result.Logs {
		fmt.Fprintf(w, "  %s (%s)\n", file.Filename, file.Modified.Format("2006-01-02"))
	}
}

var dateRegex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}`)

func appendMessageToDateSection(content string, date string, message string) string {
	entry := fmt.Sprintf("- %s", message)
	_, afterHeading, sectionEnd, found := findDateSectionBounds(content, date)
	if !found {
		trimmed := strings.TrimRight(content, "\r\n")
		if trimmed == "" {
			return fmt.Sprintf("## %s\n\n%s\n\n", date, entry)
		}
		return fmt.Sprintf("%s\n\n## %s\n\n%s\n\n", trimmed, date, entry)
	}

	sectionBody := content[afterHeading:sectionEnd]
	prefix := strings.TrimRight(content[:sectionEnd], "\r\n")
	separator := "\n"
	if strings.TrimSpace(sectionBody) == "" {
		separator = "\n\n"
	}

	return prefix + separator + entry + "\n\n" + content[sectionEnd:]
}

func findDateSectionBounds(content string, date string) (int, int, int, bool) {
	var sectionStart int
	var afterHeading int
	found := false
	inFence := false
	fenceChar := byte(0)
	fenceLen := 0

	for offset := 0; offset < len(content); {
		lineStart := offset
		lineEnd := strings.IndexByte(content[offset:], '\n')
		if lineEnd == -1 {
			lineEnd = len(content)
		} else {
			lineEnd += offset
		}
		nextLineStart := lineEnd
		if nextLineStart < len(content) && content[nextLineStart] == '\n' {
			nextLineStart++
		}

		line := strings.TrimSuffix(content[lineStart:lineEnd], "\r")
		if !inFence && strings.HasPrefix(line, "## ") {
			heading := strings.TrimSpace(strings.TrimPrefix(line, "## "))
			if found {
				return sectionStart, afterHeading, lineStart, true
			}
			if heading == date {
				sectionStart = lineStart
				afterHeading = nextLineStart
				found = true
			}
		}

		if char, length, ok := markdownFence(line); ok {
			if inFence {
				if char == fenceChar && length >= fenceLen {
					inFence = false
					fenceChar = 0
					fenceLen = 0
				}
			} else {
				inFence = true
				fenceChar = char
				fenceLen = length
			}
		}

		offset = nextLineStart
	}

	if found {
		return sectionStart, afterHeading, len(content), true
	}
	return 0, 0, 0, false
}

func markdownFence(line string) (byte, int, bool) {
	trimmed := strings.TrimLeft(line, " \t")
	if trimmed == "" {
		return 0, 0, false
	}

	char := trimmed[0]
	if char != '`' && char != '~' {
		return 0, 0, false
	}

	length := 0
	for length < len(trimmed) && trimmed[length] == char {
		length++
	}
	if length < 3 {
		return 0, 0, false
	}
	return char, length, true
}

func createSlug(topic string) string {
	slug := strings.ToLower(topic)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = slugRegex.ReplaceAllString(slug, "")
	slug = multiHyphenRegex.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = "session"
	}
	return slug
}

var slugRegex = regexp.MustCompile(`[^a-z0-9-]`)
var multiHyphenRegex = regexp.MustCompile(`-+`)
