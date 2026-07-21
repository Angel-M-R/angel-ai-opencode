package install

const (
	codegraphOptionKey = "codegraph"
	openSpecOptionKey  = "openspec"
)

// ExtraOption is a standalone integration or UI toggle applied at the end of
// the wizard, outside the assets/ catalog.
type ExtraOption struct {
	Key         string
	Label       string
	Description string
}

// ExtraOptions is the fixed set of end-of-install toggles. Unlike the assets/
// catalog, these are hardcoded because each one requires behavior that plain
// file scanning cannot express.
var ExtraOptions = []ExtraOption{
	{
		Key:         codegraphOptionKey,
		Label:       "CodeGraph",
		Description: "Instala el CLI, registra el MCP local y añade sus reglas a AGENTS.md",
	},
	{
		Key:         openSpecOptionKey,
		Label:       "OpenSpec",
		Description: "Instala o actualiza el CLI global de OpenSpec",
	},
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

var angelLogoFiles = []string{"angel-logo.tsx", "mcp-footer-state.ts"}
