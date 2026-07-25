package install

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// configurableAgents are the Angel AI agents whose model and reasoning effort
// the installer may write. Any other agent in opencode.json belongs to the user
// and is never read or written by this feature.
var configurableAgents = []string{
	"angel-orchestrator",
	"openspec-planner",
	"openspec-implementer",
	"openspec-verifier",
	"review-correctness",
	"review-security-risk",
	"review-simplicity",
}

// ConfigurableAgents returns the fixed list of agents the wizard offers, in
// presentation order.
func ConfigurableAgents() []string {
	return append([]string(nil), configurableAgents...)
}

// AgentModelAssignment is the choice made for one agent, in the exact shape it
// is written to opencode.json. Variant is empty when the chosen model has no
// reasoning effort levels; it is still written so a stale effort cannot survive
// a model change.
type AgentModelAssignment struct {
	Model   string
	Variant string
}

// AgentModelAssignments maps an agent name to its assignment. A nil or empty
// map means the user made no choice, which must leave opencode.json exactly as
// the installer would have left it before this feature existed.
type AgentModelAssignments map[string]AgentModelAssignment

// AgentModelSelection is an assignment split into the components the wizard
// picks one at a time.
type AgentModelSelection struct {
	Provider string
	Model    string
	Effort   string
}

// Assignment renders the selection as it is written to opencode.json.
func (selection AgentModelSelection) Assignment() AgentModelAssignment {
	return AgentModelAssignment{
		Model:   selection.Provider + "/" + selection.Model,
		Variant: selection.Effort,
	}
}

// wellFormed reports whether Model carries both halves of the
// "<provider>/<model>" pair opencode expects. A half-built value is dropped
// rather than written, so a malformed model string can never reach
// opencode.json.
func (assignment AgentModelAssignment) wellFormed() bool {
	provider, model, ok := strings.Cut(assignment.Model, "/")
	return ok && provider != "" && model != ""
}

// agentModelsPatch builds the opencode.json fragment for the given
// assignments: {"agent": {"<name>": {"model": ..., "variant": ...}}}. Only
// assigned agents get an entry, and an entry always carries both keys. No
// assignment at all yields no patch, so the deep-merge sees nothing.
func agentModelsPatch(assignments AgentModelAssignments) (map[string]any, bool) {
	agents := map[string]any{}
	for name, assignment := range assignments {
		if name == "" || !assignment.wellFormed() {
			continue
		}
		agents[name] = map[string]any{
			"model":   assignment.Model,
			"variant": assignment.Variant,
		}
	}
	if len(agents) == 0 {
		return nil, false
	}
	return map[string]any{"agent": agents}, true
}

// LoadAgentModelSelections reads the assignments already present in
// opencode.json for the configurable agents so the wizard can open on the
// user's previous choices. A missing, unreadable, or malformed file — or an
// entry that does not carry a "<provider>/<model>" string — simply yields no
// selection for that agent, never an error.
func LoadAgentModelSelections(configDir string) map[string]AgentModelSelection {
	raw, err := os.ReadFile(filepath.Join(configDir, "opencode.json"))
	if err != nil {
		return nil
	}
	var config struct {
		Agent map[string]struct {
			Model   string `json:"model"`
			Variant string `json:"variant"`
		} `json:"agent"`
	}
	if err := json.Unmarshal(raw, &config); err != nil {
		return nil
	}
	selections := map[string]AgentModelSelection{}
	for _, name := range configurableAgents {
		entry, ok := config.Agent[name]
		if !ok {
			continue
		}
		provider, model, ok := strings.Cut(entry.Model, "/")
		if !ok || provider == "" || model == "" {
			continue
		}
		selections[name] = AgentModelSelection{
			Provider: provider,
			Model:    model,
			Effort:   entry.Variant,
		}
	}
	if len(selections) == 0 {
		return nil
	}
	return selections
}
