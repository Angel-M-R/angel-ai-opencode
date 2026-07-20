package install

import (
	"fmt"
	"os/exec"
	"strings"
)

const codegraphPackage = "@colbymchenry/codegraph@latest"

func ensureCodegraphInstalled() (string, error) {
	if path, err := exec.LookPath("codegraph"); err == nil {
		return "disponible " + path, nil
	}

	npm, err := exec.LookPath("npm")
	if err != nil {
		return "", fmt.Errorf("CodeGraph is selected but neither codegraph nor npm is available on PATH")
	}
	cmd := exec.Command(npm, "install", "--global", codegraphPackage)
	output, err := cmd.CombinedOutput()
	if err != nil {
		detail := strings.TrimSpace(string(output))
		if detail == "" {
			return "", fmt.Errorf("installing CodeGraph with npm: %w", err)
		}
		return "", fmt.Errorf("installing CodeGraph with npm: %w: %s", err, detail)
	}
	if _, err := exec.LookPath("codegraph"); err != nil {
		return "", fmt.Errorf("CodeGraph was installed but codegraph is not available on PATH")
	}
	return "instalado  " + codegraphPackage, nil
}

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
