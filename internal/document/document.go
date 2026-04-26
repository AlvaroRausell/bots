package document

import "strings"

// UpsertSection replaces a Markdown section body or appends the section if absent.
// The header must be the full heading line, for example "## Status".
func UpsertSection(content, header, body string) string {
	level := headingLevel(header)
	if level == 0 {
		return content
	}

	start, ok := findHeading(content, header)
	if !ok {
		return appendSection(content, header, body)
	}

	lines := strings.Split(content, "\n")
	end := sectionEnd(lines, start+1, level)
	replacement := sectionLines(header, body)

	newLines := make([]string, 0, len(lines)-end+start+len(replacement))
	newLines = append(newLines, lines[:start]...)
	newLines = append(newLines, replacement...)
	newLines = append(newLines, lines[end:]...)

	return ensureFinalNewline(strings.Join(newLines, "\n"))
}

// ListSections returns the titles of headings at the requested level.
func ListSections(content string, level int) []string {
	lines := strings.Split(content, "\n")
	inCodeBlock := false
	var sections []string

	for _, line := range lines {
		if isFence(line) {
			inCodeBlock = !inCodeBlock
			continue
		}
		if inCodeBlock {
			continue
		}
		if headingLevel(line) == level {
			sections = append(sections, strings.TrimSpace(strings.TrimLeft(line, "#")))
		}
	}

	return sections
}

// SectionBody returns the body of a Markdown section without surrounding blank lines.
func SectionBody(content, header string) (string, bool) {
	level := headingLevel(header)
	if level == 0 {
		return "", false
	}

	start, ok := findHeading(content, header)
	if !ok {
		return "", false
	}

	lines := strings.Split(content, "\n")
	end := sectionEnd(lines, start+1, level)
	body := strings.Join(lines[start+1:end], "\n")
	return strings.Trim(body, "\n"), true
}

func appendSection(content, header, body string) string {
	base := strings.TrimRight(content, "\n")
	section := strings.Join(sectionLines(header, body), "\n")
	if base == "" {
		return ensureFinalNewline(section)
	}
	return ensureFinalNewline(base + "\n\n" + section)
}

func sectionLines(header, body string) []string {
	trimmedBody := strings.Trim(body, "\n")
	lines := []string{header, ""}
	if trimmedBody != "" {
		lines = append(lines, strings.Split(trimmedBody, "\n")...)
	}
	lines = append(lines, "")
	return lines
}

func findHeading(content, header string) (int, bool) {
	lines := strings.Split(content, "\n")
	inCodeBlock := false

	for i, line := range lines {
		if isFence(line) {
			inCodeBlock = !inCodeBlock
			continue
		}
		if !inCodeBlock && line == header {
			return i, true
		}
	}

	return -1, false
}

func sectionEnd(lines []string, start, level int) int {
	inCodeBlock := false
	for i := start; i < len(lines); i++ {
		line := lines[i]
		if isFence(line) {
			inCodeBlock = !inCodeBlock
			continue
		}
		if inCodeBlock {
			continue
		}
		lineLevel := headingLevel(line)
		if lineLevel > 0 && lineLevel <= level {
			return i
		}
	}
	return len(lines)
}

func headingLevel(line string) int {
	if line == "" || line[0] != '#' {
		return 0
	}
	level := 0
	for level < len(line) && line[level] == '#' {
		level++
	}
	if level == len(line) || line[level] != ' ' {
		return 0
	}
	return level
}

func isFence(line string) bool {
	return strings.HasPrefix(strings.TrimSpace(line), "```")
}

func ensureFinalNewline(content string) string {
	if strings.HasSuffix(content, "\n") {
		return content
	}
	return content + "\n"
}
