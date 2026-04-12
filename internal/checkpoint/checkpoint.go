package checkpoint

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func Read() {
	var buf strings.Builder
	ReadTo(&buf)
	fmt.Print(buf.String())
}

func ReadTo(w io.Writer) {
	checkpointFile := getCheckpointFile()
	content, err := os.ReadFile(checkpointFile)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintln(w, "No CHECKPOINTS.md found")
			fmt.Fprintln(w, "Use 'bots checkpoint update' to create initial checkpoint")
			return
		}
		fmt.Fprintf(w, "Error reading checkpoint file: %v\n", err)
		return
	}

	fmt.Fprintln(w, string(content))
}

func Update(section string, content string) {
	var buf strings.Builder
	UpdateTo(section, content, &buf)
	fmt.Print(buf.String())
}

func UpdateTo(section string, content string, w io.Writer) {
	checkpointFile := getCheckpointFile()

	var existingContent string
	if data, err := os.ReadFile(checkpointFile); err == nil {
		existingContent = string(data)
	}

	sectionHeader := fmt.Sprintf("## %s", section)

	var newContent string
	if !strings.Contains(existingContent, sectionHeader+"\n") && !strings.HasSuffix(existingContent, sectionHeader) {
		if existingContent != "" && !strings.HasSuffix(existingContent, "\n\n") {
			if strings.HasSuffix(existingContent, "\n") && !strings.HasSuffix(existingContent, "\n\n") {
				existingContent += "\n"
			} else if !strings.HasSuffix(existingContent, "\n") {
				existingContent += "\n\n"
			}
		}
		newContent = existingContent + sectionHeader + "\n\n" + content + "\n"
	} else {
		newContent = replaceSection(existingContent, sectionHeader, content)
	}

	if err := os.WriteFile(checkpointFile, []byte(newContent), 0644); err != nil {
		fmt.Fprintf(w, "Error writing checkpoint file: %v\n", err)
		return
	}

	fmt.Fprintf(w, "Updated checkpoint section: %s\n", section)
}

func replaceSection(existingContent, sectionHeader, content string) string {
	lines := strings.Split(existingContent, "\n")
	var newLines []string
	inCodeBlock := false
	inTargetSection := false
	contentInserted := false

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		if strings.HasPrefix(line, "```") {
			inCodeBlock = !inCodeBlock
		}

		if !inCodeBlock && line == sectionHeader {
			inTargetSection = true
			contentInserted = false
			continue
		}

		if inTargetSection && !inCodeBlock && strings.HasPrefix(line, "## ") && line != sectionHeader {
			if !contentInserted {
				newLines = append(newLines, sectionHeader)
				newLines = append(newLines, "")
				newLines = append(newLines, strings.Split(content, "\n")...)
				newLines = append(newLines, "")
				contentInserted = true
			}
			inTargetSection = false
			newLines = append(newLines, line)
			continue
		}

		if !inTargetSection {
			newLines = append(newLines, line)
		}
	}

	if inTargetSection && !contentInserted {
		newLines = append(newLines, sectionHeader)
		newLines = append(newLines, "")
		newLines = append(newLines, strings.Split(content, "\n")...)
		newLines = append(newLines, "")
	}

	return strings.Join(newLines, "\n")
}

func List() {
	var buf strings.Builder
	ListTo(&buf)
	fmt.Print(buf.String())
}

func ListTo(w io.Writer) {
	checkpointFile := getCheckpointFile()
	content, err := os.ReadFile(checkpointFile)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintln(w, "No CHECKPOINTS.md found")
			return
		}
		fmt.Fprintf(w, "Error reading checkpoint file: %v\n", err)
		return
	}

	lines := strings.Split(string(content), "\n")
	inCodeBlock := false
	var sections []string

	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			inCodeBlock = !inCodeBlock
		}
		if !inCodeBlock && strings.HasPrefix(line, "## ") {
			sections = append(sections, strings.TrimPrefix(line, "## "))
		}
	}

	if len(sections) == 0 {
		fmt.Fprintln(w, "No checkpoint sections found")
		return
	}

	fmt.Fprintln(w, "Checkpoint sections:")
	for _, section := range sections {
		fmt.Fprintf(w, "  - %s\n", section)
	}
}

func getCheckpointFile() string {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting working directory: %v\n", err)
		os.Exit(1)
	}

	for {
		botsDir := filepath.Join(dir, ".bots")
		checkpointFile := filepath.Join(botsDir, "CHECKPOINTS.md")
		if _, err := os.Stat(checkpointFile); err == nil {
			return checkpointFile
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			defaultBotsDir := ".bots"
			if err := os.MkdirAll(defaultBotsDir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating .bots directory: %v\n", err)
				os.Exit(1)
			}
			return filepath.Join(defaultBotsDir, "CHECKPOINTS.md")
		}
		dir = parent
	}
}
