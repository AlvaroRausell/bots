package checkpoint

import (
	"fmt"
	"io"
	"os"
	"strings"

	"bots/internal/document"
	"bots/internal/workspace"
)

// Store manages the checkpoint document in a project workspace.
type Store struct {
	workspace workspace.Workspace
}

type ReadResult struct {
	Path    string
	Content string
	Found   bool
}

type UpdateResult struct {
	Path    string
	Section string
}

type ListResult struct {
	Path     string
	Found    bool
	Sections []string
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

func (s Store) Read() (ReadResult, error) {
	path := s.workspace.CheckpointFile()
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ReadResult{Path: path, Found: false}, nil
		}
		return ReadResult{}, fmt.Errorf("read checkpoint: %w", err)
	}

	return ReadResult{Path: path, Content: string(content), Found: true}, nil
}

func (s Store) Update(section string, content string) (UpdateResult, error) {
	if err := os.MkdirAll(s.workspace.BotsDir(), 0755); err != nil {
		return UpdateResult{}, fmt.Errorf("create project state directory: %w", err)
	}

	path := s.workspace.CheckpointFile()
	var existingContent string
	if data, err := os.ReadFile(path); err == nil {
		existingContent = string(data)
	} else if !os.IsNotExist(err) {
		return UpdateResult{}, fmt.Errorf("read checkpoint: %w", err)
	}

	sectionHeader := fmt.Sprintf("## %s", section)
	newContent := document.UpsertSection(existingContent, sectionHeader, content)

	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		return UpdateResult{}, fmt.Errorf("write checkpoint: %w", err)
	}

	return UpdateResult{Path: path, Section: section}, nil
}

func (s Store) List() (ListResult, error) {
	path := s.workspace.CheckpointFile()
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ListResult{Path: path, Found: false}, nil
		}
		return ListResult{}, fmt.Errorf("read checkpoint: %w", err)
	}

	return ListResult{Path: path, Found: true, Sections: document.ListSections(string(content), 2)}, nil
}

func Read() {
	var buf strings.Builder
	ReadTo(&buf)
	fmt.Print(buf.String())
}

func ReadTo(w io.Writer) {
	store, err := NewDefaultStore()
	if err != nil {
		fmt.Fprintf(w, "Error reading checkpoint file: %v\n", err)
		return
	}

	result, err := store.Read()
	if err != nil {
		fmt.Fprintf(w, "Error reading checkpoint file: %v\n", err)
		return
	}

	WriteReadResult(w, result)
}

func WriteReadResult(w io.Writer, result ReadResult) {
	if !result.Found {
		fmt.Fprintln(w, "No CHECKPOINTS.md found")
		fmt.Fprintln(w, "Use 'bots checkpoint update' to create initial checkpoint")
		return
	}
	fmt.Fprint(w, result.Content)
	if !strings.HasSuffix(result.Content, "\n") {
		fmt.Fprintln(w)
	}
}

func Update(section string, content string) {
	var buf strings.Builder
	UpdateTo(section, content, &buf)
	fmt.Print(buf.String())
}

func UpdateTo(section string, content string, w io.Writer) {
	store, err := NewDefaultStore()
	if err != nil {
		fmt.Fprintf(w, "Error writing checkpoint file: %v\n", err)
		return
	}

	result, err := store.Update(section, content)
	if err != nil {
		fmt.Fprintf(w, "Error writing checkpoint file: %v\n", err)
		return
	}

	WriteUpdateResult(w, result)
}

func WriteUpdateResult(w io.Writer, result UpdateResult) {
	fmt.Fprintf(w, "Updated checkpoint section: %s\n", result.Section)
}

func List() {
	var buf strings.Builder
	ListTo(&buf)
	fmt.Print(buf.String())
}

func ListTo(w io.Writer) {
	store, err := NewDefaultStore()
	if err != nil {
		fmt.Fprintf(w, "Error reading checkpoint file: %v\n", err)
		return
	}

	result, err := store.List()
	if err != nil {
		fmt.Fprintf(w, "Error reading checkpoint file: %v\n", err)
		return
	}

	WriteListResult(w, result)
}

func WriteListResult(w io.Writer, result ListResult) {
	if !result.Found {
		fmt.Fprintln(w, "No CHECKPOINTS.md found")
		return
	}

	if len(result.Sections) == 0 {
		fmt.Fprintln(w, "No checkpoint sections found")
		return
	}

	fmt.Fprintln(w, "Checkpoint sections:")
	for _, section := range result.Sections {
		fmt.Fprintf(w, "  - %s\n", section)
	}
}
