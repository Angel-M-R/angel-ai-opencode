// Package catalog scans the assets/ directory and turns it into the list of
// installable items the wizard shows. Adding content never requires touching
// Go code: drop a file into the matching assets/ subdirectory.
package catalog

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Kind describes how an Item is installed.
type Kind int

const (
	// CopyFile copies a single file into the opencode config dir.
	CopyFile Kind = iota
	// CopyDir copies a whole directory (used for skills).
	CopyDir
	// MergeJSON deep-merges a JSON fragment into opencode.json.
	MergeJSON
)

// Item is one selectable thing in the wizard.
type Item struct {
	Name   string // display name, e.g. "angel-orchestrator"
	Source string // absolute path inside assets/
	Dest   string // path relative to the opencode config dir; empty for MergeJSON
	Kind   Kind
}

// Category groups items into one wizard step.
type Category struct {
	Name  string // assets subdirectory
	Title string // step title shown in the TUI
	Items []Item
}

var sections = []struct {
	dir     string
	title   string
	destDir string
	kind    Kind
}{
	{"agents", "Agentes", "agents", CopyFile},
	{"commands", "Comandos", "commands", CopyFile},
	{"skills", "Skills", "skills", CopyDir},
	{"plugins", "Plugins", "plugins", CopyFile},
	{"themes", "Themes", "themes", CopyFile},
	{"agents-md", "Reglas globales (AGENTS.md)", "", CopyFile},
	{"fragments", "Config (opencode.json)", "", MergeJSON},
}

// Load builds the categories from assetsDir. Missing subdirectories are
// skipped so a trimmed-down assets tree still works.
func Load(assetsDir string) ([]Category, error) {
	info, err := os.Stat(assetsDir)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("assets directory not found: %s", assetsDir)
	}

	var categories []Category
	for _, section := range sections {
		dir := filepath.Join(assetsDir, section.dir)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		category := Category{Name: section.dir, Title: section.title}
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			if section.kind == CopyDir && !entry.IsDir() {
				continue
			}
			if section.kind != CopyDir && entry.IsDir() {
				continue
			}
			item := Item{
				Name:   strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())),
				Source: filepath.Join(dir, entry.Name()),
				Kind:   section.kind,
			}
			if section.kind == CopyDir {
				item.Name = entry.Name()
			}
			if section.kind != MergeJSON {
				item.Dest = filepath.Join(section.destDir, entry.Name())
			}
			category.Items = append(category.Items, item)
		}
		sort.Slice(category.Items, func(i, j int) bool {
			return category.Items[i].Name < category.Items[j].Name
		})
		if len(category.Items) > 0 {
			categories = append(categories, category)
		}
	}
	if len(categories) == 0 {
		return nil, fmt.Errorf("no installable assets under %s", assetsDir)
	}
	return categories, nil
}
