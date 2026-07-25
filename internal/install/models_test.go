package install_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"angel-ai-opencode/internal/install"
)

// catalogFixture exercises unknown fields, an effort entry mixed with other
// reasoning options, a model with no effort entry, a non-tool-calling model,
// and a provider whose models are all ineligible.
const catalogFixture = `{
  "openai": {
    "id": "openai",
    "name": "OpenAI",
    "doc": "https://example.invalid",
    "models": {
      "gpt-5": {
        "id": "gpt-5",
        "name": "GPT-5",
        "tool_call": true,
        "reasoning": true,
        "reasoning_options": [
          {"type": "budget_tokens", "min": 1024},
          {"type": "effort", "values": ["low", "medium", "high"]}
        ],
        "cost": {"input": 1, "output": 2}
      },
      "gpt-legacy": {
        "id": "gpt-legacy",
        "name": "GPT Legacy",
        "tool_call": false,
        "reasoning_options": [{"type": "effort", "values": ["low"]}]
      }
    }
  },
  "anthropic": {
    "id": "anthropic",
    "name": "Anthropic",
    "models": {
      "claude-flat": {
        "id": "claude-flat",
        "name": "Claude Flat",
        "tool_call": true
      },
      "claude-sonnet": {
        "id": "claude-sonnet",
        "name": "Claude Sonnet",
        "tool_call": true,
        "reasoning_options": [{"type": "effort", "values": ["high", "max"]}]
      }
    }
  },
  "deadprovider": {
    "id": "deadprovider",
    "name": "Dead Provider",
    "models": {
      "nope": {"id": "nope", "name": "Nope", "tool_call": false}
    }
  }
}`

const authFixture = `{
  "anthropic": {"type": "oauth", "refresh": "SECRET-REFRESH", "access": "SECRET-ACCESS"},
  "openai": {"type": "api", "key": "SECRET-KEY"}
}`

func writeFixture(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("writing fixture %s: %v", name, err)
	}
	return path
}

func loadFixtureCatalog(t *testing.T, models, auth string) ([]install.ProviderOption, bool) {
	t.Helper()
	dir := t.TempDir()
	paths := install.ModelCatalogPaths{}
	if models != "" {
		paths.ModelsPath = writeFixture(t, dir, "models.json", models)
	} else {
		paths.ModelsPath = filepath.Join(dir, "missing-models.json")
	}
	if auth != "" {
		paths.AuthPath = writeFixture(t, dir, "auth.json", auth)
	} else {
		paths.AuthPath = filepath.Join(dir, "missing-auth.json")
	}
	return install.LoadModelCatalog(paths)
}

func providerIDs(providers []install.ProviderOption) []string {
	ids := make([]string, 0, len(providers))
	for _, provider := range providers {
		ids = append(ids, provider.ID)
	}
	return ids
}

func findProvider(t *testing.T, providers []install.ProviderOption, id string) install.ProviderOption {
	t.Helper()
	for _, provider := range providers {
		if provider.ID == id {
			return provider
		}
	}
	t.Fatalf("provider %q not offered; got %v", id, providerIDs(providers))
	return install.ProviderOption{}
}

func modelIDs(provider install.ProviderOption) []string {
	ids := make([]string, 0, len(provider.Models))
	for _, model := range provider.Models {
		ids = append(ids, model.ID)
	}
	return ids
}

func findModel(t *testing.T, provider install.ProviderOption, id string) install.ModelOption {
	t.Helper()
	for _, model := range provider.Models {
		if model.ID == id {
			return model
		}
	}
	t.Fatalf("model %q not offered by %q; got %v", id, provider.ID, modelIDs(provider))
	return install.ModelOption{}
}

func TestLoadModelCatalogParsesEffortAndDisplayName(t *testing.T) {
	providers, ok := loadFixtureCatalog(t, catalogFixture, authFixture)
	if !ok {
		t.Fatal("expected the catalog to be available")
	}

	model := findModel(t, findProvider(t, providers, "openai"), "gpt-5")
	if model.Name != "GPT-5" {
		t.Errorf("display name = %q, want %q", model.Name, "GPT-5")
	}
	if want := []string{"low", "medium", "high"}; !slices.Equal(model.Efforts, want) {
		t.Errorf("efforts = %v, want %v", model.Efforts, want)
	}
	if provider := findProvider(t, providers, "anthropic"); provider.Name != "Anthropic" {
		t.Errorf("provider name = %q, want %q", provider.Name, "Anthropic")
	}
}

func TestLoadModelCatalogModelWithoutEffortEntry(t *testing.T) {
	providers, ok := loadFixtureCatalog(t, catalogFixture, authFixture)
	if !ok {
		t.Fatal("expected the catalog to be available")
	}

	model := findModel(t, findProvider(t, providers, "anthropic"), "claude-flat")
	if len(model.Efforts) != 0 {
		t.Errorf("efforts = %v, want none", model.Efforts)
	}
}

func TestLoadModelCatalogUnavailableWithoutFailing(t *testing.T) {
	cases := map[string]string{
		"missing":       "",
		"invalid json":  "{not json",
		"not an object": `["openai"]`,
		"empty object":  `{}`,
	}
	for name, fixture := range cases {
		t.Run(name, func(t *testing.T) {
			providers, ok := loadFixtureCatalog(t, fixture, authFixture)
			if ok {
				t.Fatalf("expected the catalog to be unavailable, got %v", providerIDs(providers))
			}
			if len(providers) != 0 {
				t.Errorf("expected no providers, got %v", providerIDs(providers))
			}
		})
	}
}

func TestLoadModelCatalogFiltersByAuthAndToolCall(t *testing.T) {
	providers, ok := loadFixtureCatalog(t, catalogFixture, `{"anthropic": {"type": "oauth", "refresh": "SECRET"}}`)
	if !ok {
		t.Fatal("expected the catalog to be available")
	}

	if want := []string{"anthropic"}; !slices.Equal(providerIDs(providers), want) {
		t.Fatalf("providers = %v, want %v", providerIDs(providers), want)
	}
	provider := findProvider(t, providers, "anthropic")
	if want := []string{"claude-flat", "claude-sonnet"}; !slices.Equal(modelIDs(provider), want) {
		t.Errorf("models = %v, want %v", modelIDs(provider), want)
	}
}

func TestLoadModelCatalogFallsBackWhenAuthUnusable(t *testing.T) {
	for name, fixture := range map[string]string{"missing": "", "no keys": `{}`, "invalid json": "{not json"} {
		t.Run(name, func(t *testing.T) {
			providers, ok := loadFixtureCatalog(t, catalogFixture, fixture)
			if !ok {
				t.Fatal("expected the catalog to be available")
			}
			// Every provider is kept, except the one pruned for having no
			// tool-calling model, and the order is deterministic.
			if want := []string{"anthropic", "openai"}; !slices.Equal(providerIDs(providers), want) {
				t.Fatalf("providers = %v, want %v", providerIDs(providers), want)
			}
			provider := findProvider(t, providers, "openai")
			if want := []string{"gpt-5"}; !slices.Equal(modelIDs(provider), want) {
				t.Errorf("models = %v, want %v", modelIDs(provider), want)
			}
		})
	}
}
