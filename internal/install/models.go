package install

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
)

// ModelOption is a single assignable model as offered by the wizard.
type ModelOption struct {
	ID      string
	Name    string
	Efforts []string
}

// ProviderOption is a provider with the models the user may assign from it.
type ProviderOption struct {
	ID     string
	Name   string
	Models []ModelOption
}

// ModelCatalogPaths locates the two files the catalog is derived from. Both are
// injectable so tests never touch the real user environment.
type ModelCatalogPaths struct {
	// ModelsPath is OpenCode's own model cache, normally
	// ~/.cache/opencode/models.json.
	ModelsPath string
	// AuthPath is OpenCode's credential store, normally
	// ~/.local/share/opencode/auth.json. Only its top-level keys are read.
	AuthPath string
}

// DefaultModelCatalogPaths resolves the catalog inputs from the user's home
// directory. An unresolvable home yields empty paths, which LoadModelCatalog
// reports as an unavailable catalog rather than as an error.
func DefaultModelCatalogPaths() ModelCatalogPaths {
	home, err := os.UserHomeDir()
	if err != nil {
		return ModelCatalogPaths{}
	}
	return ModelCatalogPaths{
		ModelsPath: filepath.Join(home, ".cache", "opencode", "models.json"),
		AuthPath:   filepath.Join(home, ".local", "share", "opencode", "auth.json"),
	}
}

// rawModelCatalog mirrors only the fields of models.json this installer needs.
// Every other field is ignored so that upstream schema growth cannot break
// parsing.
type rawModelCatalog map[string]rawProvider

type rawProvider struct {
	Name   string              `json:"name"`
	Models map[string]rawModel `json:"models"`
}

type rawModel struct {
	Name             string             `json:"name"`
	ToolCall         bool               `json:"tool_call"`
	ReasoningOptions []rawReasoningOpts `json:"reasoning_options"`
}

type rawReasoningOpts struct {
	Type   string   `json:"type"`
	Values []string `json:"values"`
}

// LoadModelCatalog reads OpenCode's model cache and returns the providers and
// models the user can realistically assign to an agent: providers they are
// authenticated against, and models that support tool calling.
//
// It never returns an error. A missing, unreadable, or invalid catalog — or one
// left with no eligible provider after filtering — reports the catalog as
// unavailable (ok == false) so the installer can skip the feature instead of
// failing the install.
func LoadModelCatalog(paths ModelCatalogPaths) ([]ProviderOption, bool) {
	raw, ok := readModelCatalogFile(paths.ModelsPath)
	if !ok {
		return nil, false
	}
	providers := filterModelCatalog(raw, readAuthProviderIDs(paths.AuthPath))
	if len(providers) == 0 {
		return nil, false
	}
	return providers, true
}

func readModelCatalogFile(path string) (rawModelCatalog, bool) {
	if path == "" {
		return nil, false
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	var raw rawModelCatalog
	if err := json.Unmarshal(content, &raw); err != nil {
		return nil, false
	}
	if raw == nil {
		return nil, false
	}
	return raw, true
}

// readAuthProviderIDs reads only the top-level keys of auth.json. The decoded
// values are credentials and are discarded without ever being copied into a
// struct field, a log line, or an error message. A missing, unreadable, or
// invalid file yields no ids, which callers treat as "do not filter".
func readAuthProviderIDs(path string) map[string]bool {
	if path == "" {
		return nil
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var entries map[string]json.RawMessage
	if err := json.Unmarshal(content, &entries); err != nil {
		return nil
	}
	if len(entries) == 0 {
		return nil
	}
	ids := make(map[string]bool, len(entries))
	for id := range entries {
		ids[id] = true
	}
	return ids
}

// filterModelCatalog keeps only tool-calling models, restricts providers to the
// authenticated ids when there are any, drops providers left with no model, and
// orders everything deterministically for a stable picker.
func filterModelCatalog(raw rawModelCatalog, authenticated map[string]bool) []ProviderOption {
	var providers []ProviderOption
	for id, provider := range raw {
		if id == "" {
			continue
		}
		if len(authenticated) > 0 && !authenticated[id] {
			continue
		}
		models := eligibleModels(provider.Models)
		if len(models) == 0 {
			continue
		}
		name := provider.Name
		if name == "" {
			name = id
		}
		providers = append(providers, ProviderOption{ID: id, Name: name, Models: models})
	}
	sort.Slice(providers, func(i, j int) bool { return providers[i].ID < providers[j].ID })
	return providers
}

func eligibleModels(raw map[string]rawModel) []ModelOption {
	var models []ModelOption
	for id, model := range raw {
		if !model.ToolCall || id == "" {
			continue
		}
		name := model.Name
		if name == "" {
			name = id
		}
		models = append(models, ModelOption{
			ID:      id,
			Name:    name,
			Efforts: effortValues(model.ReasoningOptions),
		})
	}
	sort.Slice(models, func(i, j int) bool { return models[i].ID < models[j].ID })
	return models
}

// effortValues returns the values of the reasoning option whose type is
// "effort", in the order given. A model without such an entry has no efforts.
func effortValues(options []rawReasoningOpts) []string {
	for _, option := range options {
		if option.Type != "effort" {
			continue
		}
		if len(option.Values) == 0 {
			return nil
		}
		return append([]string(nil), option.Values...)
	}
	return nil
}
