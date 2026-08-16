package catalog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	assetfs "angel-ai-opencode/internal/assets"
)

func TestLoadDoesNotIncludeVendoredOpenSpecSkills(t *testing.T) {
	assets := filepath.Join("..", "..", "assets")
	categories, err := Load(assetfs.Directory(assets))
	if err != nil {
		t.Fatal(err)
	}

	for _, category := range categories {
		if category.Name != "skills" {
			continue
		}
		for _, item := range category.Items {
			if item.Name == "openspec" || strings.HasPrefix(item.Name, "openspec-") {
				t.Fatalf("OpenSpec skill %q is vendored in the installer catalog", item.Name)
			}
		}
	}
}

func TestThemesReferenceDefinedColors(t *testing.T) {
	themeDir := filepath.Join("..", "..", "assets", "themes")
	entries, err := os.ReadDir(themeDir)
	if err != nil {
		t.Fatal(err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		t.Run(entry.Name(), func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join(themeDir, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			var document struct {
				Defs  map[string]any `json:"defs"`
				Theme map[string]any `json:"theme"`
			}
			if err := json.Unmarshal(raw, &document); err != nil {
				t.Fatal(err)
			}
			assertThemeColorsDefined(t, document.Theme, document.Defs, "theme")
		})
	}
}

func assertThemeColorsDefined(t *testing.T, value any, defs map[string]any, path string) {
	t.Helper()
	switch value := value.(type) {
	case string:
		if value == "none" || strings.HasPrefix(value, "#") {
			return
		}
		if _, ok := defs[value]; !ok {
			t.Errorf("%s references undefined color %q", path, value)
		}
	case map[string]any:
		for key, child := range value {
			assertThemeColorsDefined(t, child, defs, path+"."+key)
		}
	case float64:
		// Numeric ANSI-256 colors are valid theme values.
	case nil:
		t.Errorf("%s contains a null color", path)
	default:
		t.Errorf("%s contains unsupported color value %T", path, value)
	}
}
