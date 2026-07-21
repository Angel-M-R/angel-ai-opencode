package install

import (
	"fmt"
	"strings"
)

const codegraphPackage = "@colbymchenry/codegraph@latest"

func removeManagedBlock(content, startMarker, endMarker string) (string, error) {
	for {
		start := strings.Index(content, startMarker)
		if start < 0 {
			return content, nil
		}
		endOffset := strings.Index(content[start+len(startMarker):], endMarker)
		if endOffset < 0 {
			return "", fmt.Errorf("unterminated managed block %q in AGENTS.md", startMarker)
		}
		end := start + len(startMarker) + endOffset + len(endMarker)
		before := strings.TrimRight(content[:start], "\n")
		after := strings.TrimLeft(content[end:], "\n")
		switch {
		case before == "":
			content = after
		case after == "":
			content = before
		default:
			content = before + "\n\n" + after
		}
	}
}
