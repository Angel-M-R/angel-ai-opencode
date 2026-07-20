package install

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const codegraphPackage = "@colbymchenry/codegraph@latest"

func applyCodegraph(assetsDir, configDir string) ([]string, error) {
	integrationDir := filepath.Join(assetsDir, "integrations", "codegraph")

	rawPatch, err := os.ReadFile(filepath.Join(integrationDir, "mcp.json"))
	if err != nil {
		return nil, fmt.Errorf("reading CodeGraph MCP config: %w", err)
	}
	var patch map[string]any
	if err := json.Unmarshal(rawPatch, &patch); err != nil {
		return nil, fmt.Errorf("parsing CodeGraph MCP config: %w", err)
	}
	guidance, err := os.ReadFile(filepath.Join(integrationDir, "AGENTS.md"))
	if err != nil {
		return nil, fmt.Errorf("reading CodeGraph agent guidance: %w", err)
	}
	agentsPath := filepath.Join(configDir, "AGENTS.md")
	renderedGuidance, _, err := renderCodegraphGuidance(agentsPath, string(guidance))
	if err != nil {
		return nil, err
	}

	installLine, err := ensureCodegraphInstalled()
	if err != nil {
		return nil, err
	}
	done := []string{installLine}

	opencodePath := filepath.Join(configDir, "opencode.json")
	lines, err := mergeJSON(opencodePath, "https://opencode.ai/config.json", []map[string]any{patch})
	done = append(done, lines...)
	if err != nil {
		return done, err
	}
	done = append(done, "merge     codegraph MCP")

	if err := writeCodegraphGuidance(agentsPath, renderedGuidance); err != nil {
		return done, err
	}
	done = append(done, "actualizado "+agentsPath)
	return done, nil
}

func removeCodegraph(configDir string) ([]string, error) {
	agentsPath := filepath.Join(configDir, "AGENTS.md")
	renderedGuidance, guidanceChanged, err := renderCodegraphGuidance(agentsPath, "")
	if err != nil {
		return nil, err
	}

	done, mcpChanged, err := removeCodegraphMCP(filepath.Join(configDir, "opencode.json"))
	if err != nil {
		return done, err
	}
	if mcpChanged {
		done = append(done, "eliminado  codegraph MCP")
	}
	if guidanceChanged {
		if err := writeCodegraphGuidance(agentsPath, renderedGuidance); err != nil {
			return done, err
		}
		done = append(done, "eliminado  CodeGraph de "+agentsPath)
	}
	return done, nil
}

func removeCodegraphMCP(path string) ([]string, bool, error) {
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var config map[string]any
	if err := json.Unmarshal(raw, &config); err != nil {
		return nil, false, fmt.Errorf("parsing %s: %w", path, err)
	}
	mcp, ok := config["mcp"].(map[string]any)
	if !ok {
		return nil, false, nil
	}
	if _, ok := mcp["codegraph"]; !ok {
		return nil, false, nil
	}
	delete(mcp, "codegraph")
	if len(mcp) == 0 {
		delete(config, "mcp")
	}
	lines, err := writeJSON(path, config, raw)
	return lines, err == nil, err
}

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

func renderCodegraphGuidance(path, guidance string) ([]byte, bool, error) {
	raw, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, false, fmt.Errorf("reading %s: %w", path, err)
	}

	content := string(raw)
	original := content
	for _, markers := range [][2]string{
		{"<!-- codegraph-guidance -->", "<!-- /codegraph-guidance -->"},
		{"<!-- gentle-ai:codegraph-guidance -->", "<!-- /gentle-ai:codegraph-guidance -->"},
	} {
		content, err = removeManagedBlock(content, markers[0], markers[1])
		if err != nil {
			return nil, false, err
		}
	}
	if guidance == "" && content == original {
		return nil, false, nil
	}
	content = strings.TrimSpace(content)
	guidance = strings.TrimSpace(guidance)
	if content != "" && guidance != "" {
		content += "\n\n"
	}
	content += guidance
	if content != "" {
		content += "\n"
	}
	return []byte(content), content != original, nil
}

func writeCodegraphGuidance(path string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
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
