package verifiertasks

import (
	"fmt"
	"regexp"
	"strings"
)

const OwnerMarker = "<!-- owner: openspec-verifier -->"

var (
	headingPattern = regexp.MustCompile(`^(#{1,6})[ \t]+(.+?)[ \t]*#*[ \t]*$`)
	taskPattern    = regexp.MustCompile(`^(\s*[-*+]\s+\[)([ xX])(\]\s+)(.+)$`)
	taskIDPattern  = regexp.MustCompile(`^([0-9]+(?:\.[0-9]+)*)(?:[.)])?\s+.+$`)
)

// TaskIdentity binds evidence to the complete, exact task text rather than to
// a possibly reused numeric identifier alone.
type TaskIdentity struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type TaskState struct {
	TaskIdentity
	Pending bool `json:"pending"`
	Line    int  `json:"line"`
}

type MarkerStructure struct {
	Present      bool        `json:"present"`
	SectionTitle string      `json:"sectionTitle,omitempty"`
	SectionLine  int         `json:"sectionLine,omitempty"`
	MarkerLine   int         `json:"markerLine,omitempty"`
	SectionEnd   int         `json:"sectionEnd,omitempty"`
	Tasks        []TaskState `json:"tasks,omitempty"`
}

type parsedDocument struct {
	Marker         MarkerStructure
	Pending        []TaskIdentity
	pendingOffsets map[TaskIdentity]int
}

type sourceLine struct {
	text  string
	start int
}

type heading struct {
	level int
	title string
	line  int
}

func parseDocument(content []byte) (parsedDocument, error) {
	lines := splitLines(content)
	var headings []heading
	var topSections []heading
	var exactMarkers []int
	for index, line := range lines {
		trimmed := strings.TrimSpace(line.text)
		if match := headingPattern.FindStringSubmatch(line.text); match != nil {
			h := heading{level: len(match[1]), title: strings.TrimSpace(match[2]), line: index}
			headings = append(headings, h)
			if h.level == 2 {
				topSections = append(topSections, h)
			}
		}
		if trimmed == OwnerMarker {
			exactMarkers = append(exactMarkers, index)
			continue
		}
		lower := strings.ToLower(trimmed)
		if strings.HasPrefix(lower, "<!--") && strings.Contains(lower, "openspec-verifier") {
			return parsedDocument{}, fmt.Errorf("malformed verifier owner marker on line %d", index+1)
		}
	}

	if len(exactMarkers) == 0 {
		return parsedDocument{Marker: MarkerStructure{}}, nil
	}
	if len(exactMarkers) != 1 {
		return parsedDocument{}, fmt.Errorf("duplicate verifier owner markers")
	}
	if len(topSections) == 0 {
		return parsedDocument{}, fmt.Errorf("verifier owner marker is not inside a named top-level task section")
	}

	markerLine := exactMarkers[0]
	sectionIndex := -1
	for index, section := range topSections {
		if section.line < markerLine {
			sectionIndex = index
		}
	}
	if sectionIndex < 0 {
		return parsedDocument{}, fmt.Errorf("verifier owner marker is misplaced")
	}
	section := topSections[sectionIndex]
	if sectionIndex != len(topSections)-1 {
		return parsedDocument{}, fmt.Errorf("verifier owner marker must be in the final task section")
	}
	for _, h := range headings {
		if h.line > section.line && h.line < markerLine {
			return parsedDocument{}, fmt.Errorf("verifier owner marker must not be nested")
		}
	}
	for line := section.line + 1; line < markerLine; line++ {
		if strings.TrimSpace(lines[line].text) != "" {
			return parsedDocument{}, fmt.Errorf("verifier owner marker must be the first nonblank line of its section")
		}
	}

	sectionEnd := len(lines)
	marker := MarkerStructure{
		Present:      true,
		SectionTitle: section.title,
		SectionLine:  section.line + 1,
		MarkerLine:   markerLine + 1,
		SectionEnd:   sectionEnd,
	}
	parsed := parsedDocument{Marker: marker, pendingOffsets: make(map[TaskIdentity]int)}
	seenIDs := make(map[string]struct{})
	for lineIndex := markerLine + 1; lineIndex < sectionEnd; lineIndex++ {
		line := lines[lineIndex]
		match := taskPattern.FindStringSubmatch(line.text)
		if match == nil {
			continue
		}
		idMatch := taskIDPattern.FindStringSubmatch(match[4])
		if idMatch == nil {
			return parsedDocument{}, fmt.Errorf("marked task on line %d has no stable numeric identity", lineIndex+1)
		}
		identity := TaskIdentity{ID: idMatch[1], Text: match[4]}
		if _, exists := seenIDs[identity.ID]; exists {
			return parsedDocument{}, fmt.Errorf("duplicate marked task identity %q", identity.ID)
		}
		seenIDs[identity.ID] = struct{}{}
		pending := match[2] == " "
		parsed.Marker.Tasks = append(parsed.Marker.Tasks, TaskState{
			TaskIdentity: identity,
			Pending:      pending,
			Line:         lineIndex + 1,
		})
		if pending {
			parsed.Pending = append(parsed.Pending, identity)
			parsed.pendingOffsets[identity] = line.start + len(match[1])
		}
	}
	if len(parsed.Marker.Tasks) == 0 {
		return parsedDocument{}, fmt.Errorf("marked verifier section contains no named tasks")
	}
	return parsed, nil
}

func splitLines(content []byte) []sourceLine {
	if len(content) == 0 {
		return nil
	}
	lines := make([]sourceLine, 0, strings.Count(string(content), "\n")+1)
	start := 0
	for start < len(content) {
		end := start
		for end < len(content) && content[end] != '\n' {
			end++
		}
		textEnd := end
		if textEnd > start && content[textEnd-1] == '\r' {
			textEnd--
		}
		lines = append(lines, sourceLine{text: string(content[start:textEnd]), start: start})
		if end == len(content) {
			break
		}
		start = end + 1
	}
	return lines
}
