package checkpoint

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Read reads and displays the current CHECKPOINTS.md content
func Read() {
	checkpointFile := getCheckpointFile()
	content, err := os.ReadFile(checkpointFile)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No CHECKPOINTS.md found")
			fmt.Println("Use 'bots checkpoint update' to create initial checkpoint")
			return
		}
		fmt.Fprintf(os.Stderr, "Error reading checkpoint file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(content))
}

// Update updates a specific section in the checkpoint file
func Update(section string, content string) {
	checkpointFile := getCheckpointFile()

	// Read existing content or start fresh
	var existingContent string
	if data, err := os.ReadFile(checkpointFile); err == nil {
		existingContent = string(data)
	}

	// Check if section exists
	sectionHeader := fmt.Sprintf("## %s", section)
	sectionIndex := strings.Index(existingContent, sectionHeader)

	var newContent string
	if sectionIndex == -1 {
		// Section doesn't exist, append it
		if existingContent != "" && !strings.HasSuffix(existingContent, "\n\n") {
			existingContent += "\n\n"
		}
		newContent = existingContent + sectionHeader + "\n\n" + content + "\n"
	} else {
		// Section exists, replace it
		lines := strings.Split(existingContent, "\n")
		var newLines []string
		inSection := false
		skipUntilNextSection := false

		for _, line := range lines {
			if strings.HasPrefix(line, "## ") && line != sectionHeader {
				// Found a different section header
				if skipUntilNextSection {
					newLines = append(newLines, sectionHeader)
					newLines = append(newLines, "")
					newLines = append(newLines, strings.Split(content, "\n")...)
					newLines = append(newLines, "")
					skipUntilNextSection = false
				}
				inSection = false
			}

			if line == sectionHeader {
				inSection = true
				skipUntilNextSection = true
				continue
			}

			if inSection && strings.HasPrefix(line, "## ") {
				// Reached next section, insert our content
				newLines = append(newLines, sectionHeader)
				newLines = append(newLines, "")
				newLines = append(newLines, strings.Split(content, "\n")...)
				newLines = append(newLines, "")
				inSection = false
				skipUntilNextSection = false
			}

			if !skipUntilNextSection && !inSection {
				newLines = append(newLines, line)
			}
		}

		// If we were still in section at EOF, append the new content
		if skipUntilNextSection {
			newLines = append(newLines, sectionHeader)
			newLines = append(newLines, "")
			newLines = append(newLines, strings.Split(content, "\n")...)
			newLines = append(newLines, "")
		}

		newContent = strings.Join(newLines, "\n")
	}

	if err := os.WriteFile(checkpointFile, []byte(newContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing checkpoint file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Updated checkpoint section: %s\n", section)
}

// List lists all checkpoint sections
func List() {
	checkpointFile := getCheckpointFile()
	content, err := os.ReadFile(checkpointFile)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No CHECKPOINTS.md found")
			return
		}
		fmt.Fprintf(os.Stderr, "Error reading checkpoint file: %v\n", err)
		os.Exit(1)
	}

	lines := strings.Split(string(content), "\n")
	var sections []string

	for _, line := range lines {
		if strings.HasPrefix(line, "## ") {
			sections = append(sections, strings.TrimPrefix(line, "## "))
		}
	}

	if len(sections) == 0 {
		fmt.Println("No checkpoint sections found")
		return
	}

	fmt.Println("Checkpoint sections:")
	for _, section := range sections {
		fmt.Printf("  - %s\n", section)
	}
}

// Helper functions

func getCheckpointFile() string {
	// Look for .bots/CHECKPOINTS.md in current directory or parents
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
			// Reached root, create default structure
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
