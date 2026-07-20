// angel is a personal installer for opencode configuration: it reads the
// editable content under assets/ and installs the selection into
// ~/.config/opencode through a step-by-step TUI wizard.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"angel-ai-opencode/internal/catalog"
	"angel-ai-opencode/internal/install"
	"angel-ai-opencode/internal/tui"
)

func main() {
	assets := flag.String("assets", "", "assets directory (default: assets/ next to the binary, then ./assets)")
	target := flag.String("target", "", "opencode config directory (default: ~/.config/opencode)")
	all := flag.Bool("all", false, "install everything without the TUI")
	dryRun := flag.Bool("dry-run", false, "with --all, print the plan without installing")
	flag.Parse()

	if err := run(*assets, *target, *all, *dryRun); err != nil {
		fmt.Fprintln(os.Stderr, "angel:", err)
		os.Exit(1)
	}
}

func run(assetsDir, configDir string, all, dryRun bool) error {
	if assetsDir == "" {
		assetsDir = findAssets()
	}
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		configDir = filepath.Join(home, ".config", "opencode")
	}

	categories, err := catalog.Load(assetsDir)
	if err != nil {
		return err
	}

	if !all {
		return tui.Run(categories, assetsDir, configDir)
	}

	var items []catalog.Item
	for _, category := range categories {
		items = append(items, category.Items...)
	}
	extras := make(map[string]bool, len(install.ExtraOptions))
	for _, extra := range install.ExtraOptions {
		extras[extra.Key] = true
	}

	if dryRun {
		plan, err := install.PlanInstallation(install.InstallationRequest{
			Items: items, Extras: extras, AssetsDir: assetsDir, ConfigDir: configDir,
		})
		if err != nil {
			return err
		}
		for _, line := range plan {
			fmt.Println(line)
		}
		return nil
	}
	report, err := install.ApplyInstallation(install.InstallationRequest{
		Items: items, Extras: extras, AssetsDir: assetsDir, ConfigDir: configDir,
	})
	for _, line := range report {
		fmt.Println(line)
	}
	return err
}

// findAssets prefers assets/ next to the binary so `angel` works from
// anywhere, and falls back to ./assets for `go run .` in the repo.
func findAssets() string {
	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), "assets")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}
	return "assets"
}
