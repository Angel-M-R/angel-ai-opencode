package catalog

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadGroupsNestedOpenSpecSkills(t *testing.T) {
	assets := filepath.Join("..", "..", "assets")
	categories, err := Load(assets)
	if err != nil {
		t.Fatal(err)
	}

	var openSpecItems []Item
	for _, category := range categories {
		if category.Name != "skills" {
			continue
		}
		for _, item := range category.Items {
			if item.Name == "openspec" || strings.HasPrefix(item.Name, "openspec-") {
				openSpecItems = append(openSpecItems, item)
			}
		}
	}
	if len(openSpecItems) != 1 {
		t.Fatalf("OpenSpec catalog items = %#v, want one grouped bundle", openSpecItems)
	}
	item := openSpecItems[0]
	if item.Name != "openspec" || item.Source != filepath.Join(assets, "skills", "openspec") ||
		item.Dest != filepath.Join("skills", "openspec") || item.Kind != CopyDir {
		t.Fatalf("OpenSpec catalog item = %#v, want nested CopyDir bundle", item)
	}
}
