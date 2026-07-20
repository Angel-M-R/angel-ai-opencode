package install

import (
	"fmt"
	"path/filepath"
)

// ExtraOption is a standalone toggle applied at the end of the wizard,
// outside the assets/ catalog: it merges into tui.json (never
// opencode.json) and optionally copies its own plugin files.
type ExtraOption struct {
	Key         string
	Label       string
	Description string
}

// ExtraOptions is the fixed set of end-of-install toggles. Unlike the
// assets/ catalog, these are hardcoded because each one bundles a
// multi-file copy and/or a tui.json-specific merge that plain file scanning
// can't express.
var ExtraOptions = []ExtraOption{
	{
		Key:         "angel-logo",
		Label:       "Logo Angel AI",
		Description: "ASCII logo propio + estado de los MCP en el footer de la TUI",
	},
	{
		Key:         "theme",
		Label:       "Tema one-dark-pro",
		Description: "Activa one-dark-pro como tema de la TUI (tui.json)",
	},
	{
		Key:         "subagent-statusline",
		Label:       "Subagent statusline",
		Description: "Plugin de terceros (npm): actividad de los workers en la sidebar",
	},
}

// PlanExtras describes what ApplyExtras would do, one line per action.
func PlanExtras(selected map[string]bool, configDir string) []string {
	var lines []string
	if selected["angel-logo"] {
		for _, name := range angelLogoFiles {
			lines = append(lines, fmt.Sprintf("copiar %s → %s", name, filepath.Join(configDir, "tui-plugins", name)))
		}
		lines = append(lines, "merge  angel-logo → tui.json")
	}
	if selected["theme"] {
		lines = append(lines, "merge  theme → tui.json")
	}
	if selected["subagent-statusline"] {
		lines = append(lines, "merge  subagent-statusline → tui.json")
	}
	return lines
}

var angelLogoFiles = []string{"angel-logo.tsx", "mcp-footer-state.ts"}

// ApplyExtras installs the toggles selected by key. Files are copied into
// <configDir>/tui-plugins/; all merges target <configDir>/tui.json.
func ApplyExtras(selected map[string]bool, assetsDir, configDir string) ([]string, error) {
	var done []string
	var patches []map[string]any

	if selected["angel-logo"] {
		for _, name := range angelLogoFiles {
			src := filepath.Join(assetsDir, "tui-plugins", name)
			dest := filepath.Join(configDir, "tui-plugins", name)
			if err := copyFile(src, dest); err != nil {
				return done, fmt.Errorf("installing %s: %w", name, err)
			}
			done = append(done, "instalado "+dest)
		}
		patches = append(patches, map[string]any{
			"plugin": []any{filepath.Join(configDir, "tui-plugins", "angel-logo.tsx")},
		})
	}
	if selected["theme"] {
		patches = append(patches, map[string]any{"theme": "one-dark-pro"})
	}
	if selected["subagent-statusline"] {
		patches = append(patches, map[string]any{
			"plugin": []any{"opencode-subagent-statusline"},
		})
	}
	if len(patches) == 0 {
		return done, nil
	}

	tuiPath := filepath.Join(configDir, "tui.json")
	lines, err := mergeJSON(tuiPath, "https://opencode.ai/tui.json", patches)
	done = append(done, lines...)
	return done, err
}
