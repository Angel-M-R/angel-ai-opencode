package install

import "fmt"

const (
	codegraphOptionKey = "codegraph"
	openSpecOptionKey  = "openspec"
	cmuxOptionKey      = "cmux"
)

// ExtraOption is a standalone integration or UI toggle applied at the end of
// the wizard, outside the assets/ catalog.
type ExtraOption struct {
	Key             string
	Label           string
	Description     string
	DefaultSelected bool
}

// ExtraOptions is the fixed set of end-of-install toggles. Unlike the assets/
// catalog, these are hardcoded because each one requires behavior that plain
// file scanning cannot express.
var ExtraOptions = []ExtraOption{
	{
		Key:             codegraphOptionKey,
		Label:           "CodeGraph",
		Description:     "Instala el CLI, registra el MCP local y añade sus reglas a AGENTS.md",
		DefaultSelected: true,
	},
	{
		Key:             openSpecOptionKey,
		Label:           "OpenSpec",
		Description:     "Instala o actualiza el CLI global de OpenSpec",
		DefaultSelected: true,
	},
	{
		Key:             "angel-logo",
		Label:           "Logo Angel AI",
		Description:     "ASCII logo propio + estado de los MCP en el footer de la TUI",
		DefaultSelected: true,
	},
	{
		Key:             "theme",
		Label:           "Tema one-dark-pro",
		Description:     "Activa one-dark-pro como tema de la TUI (tui.json)",
		DefaultSelected: true,
	},
	{
		Key:             "subagent-statusline",
		Label:           "Subagent statusline",
		Description:     "Plugin de terceros (npm): actividad de los workers en la sidebar",
		DefaultSelected: true,
	},
	{
		Key:             cmuxOptionKey,
		Label:           "cmux",
		Description:     "Notificaciones y Feed de cmux para sesiones de OpenCode",
		DefaultSelected: false,
	},
}

var angelLogoFiles = []string{"angel-logo.tsx", "mcp-footer-state.ts"}
var cmuxPluginFiles = []string{"cmux-session.js", "cmux-feed.js"}

type executableLookup func(string) (string, error)

func preflightSelectedExtras(extras map[string]bool, lookPath executableLookup) error {
	if !extras[cmuxOptionKey] {
		return nil
	}
	if _, err := lookPath("cmux"); err != nil {
		return fmt.Errorf("cmux extra requires cmux to be available on PATH: %w", err)
	}
	return nil
}
