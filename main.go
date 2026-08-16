// angel is a personal installer for opencode configuration: it reads the
// editable content under assets/ and installs the selection into
// ~/.config/opencode through a step-by-step TUI wizard.
package main

import (
	"context"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"angel-ai-opencode/internal/assets"
	"angel-ai-opencode/internal/catalog"
	"angel-ai-opencode/internal/install"
	"angel-ai-opencode/internal/openspecbootstrap"
	"angel-ai-opencode/internal/tui"
	"angel-ai-opencode/internal/updater"
	"angel-ai-opencode/internal/verifiertasks"
)

//go:embed all:assets
var embeddedAssetTree embed.FS

// version is set from a stable release tag through -ldflags. Local builds use
// dev so they cannot update themselves accidentally.
var version = "dev"

type rootOptions struct {
	assetsDir string
	configDir string
	all       bool
	dryRun    bool
}

type updatePolicy interface {
	Run(currentVersion string, forced bool) error
}

type cliDependencies struct {
	stdout                io.Writer
	stdin                 io.Reader
	runInstaller          func(rootOptions) error
	newUpdatePolicy       func() updatePolicy
	workingDirectory      func() (string, error)
	captureVerifierTasks  func(context.Context, verifiertasks.ResolveRequest) (verifiertasks.Result, error)
	completeVerifierTasks func(context.Context, verifiertasks.ResolveRequest, verifiertasks.CompleteRequest) (verifiertasks.Result, error)
	runOpenSpecBootstrap  func(context.Context, openspecbootstrap.Request) (openspecbootstrap.Result, error)
}

func main() {
	if err := runCLI(os.Args[1:], defaultCLIDependencies()); err != nil {
		fmt.Fprintln(os.Stderr, "angel:", err)
		os.Exit(1)
	}
}

func defaultCLIDependencies() cliDependencies {
	verifierTasks := verifiertasks.NewService()
	return cliDependencies{
		stdout: os.Stdout,
		stdin:  os.Stdin,
		runInstaller: func(options rootOptions) error {
			return run(options.assetsDir, options.configDir, options.all, options.dryRun)
		},
		newUpdatePolicy: func() updatePolicy {
			return updater.New(updater.Config{Output: os.Stdout})
		},
		workingDirectory:      os.Getwd,
		captureVerifierTasks:  verifierTasks.Capture,
		completeVerifierTasks: verifierTasks.Complete,
		runOpenSpecBootstrap:  openspecbootstrap.NewService().Run,
	}
}

func runCLI(args []string, dependencies cliDependencies) error {
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		switch args[0] {
		case "version":
			return runVersionCommand(args[1:], dependencies)
		case "update":
			return runUpdateCommand(args[1:], dependencies)
		case "verifier-tasks":
			return runVerifierTasksCommand(args[1:], dependencies)
		case "openspec-bootstrap":
			return runOpenSpecBootstrapCommand(args[1:], dependencies)
		default:
			return fmt.Errorf("unknown command %q", args[0])
		}
	}
	return runRootCommand(args, dependencies)
}

func runVerifierTasksCommand(args []string, dependencies cliDependencies) error {
	if len(args) == 0 {
		return fmt.Errorf("verifier-tasks: phase must be snapshot or complete")
	}
	phase := args[0]
	flags := flag.NewFlagSet("angel-ai verifier-tasks "+phase, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	change := flags.String("change", "", "active OpenSpec change")
	store := flags.String("store", "", "explicit OpenSpec store id")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("verifier-tasks %s: unexpected argument %q", phase, flags.Arg(0))
	}
	if strings.TrimSpace(*change) == "" {
		return fmt.Errorf("verifier-tasks %s: --change is required", phase)
	}
	if dependencies.workingDirectory == nil {
		return fmt.Errorf("verifier-tasks: working-directory resolver is unavailable")
	}
	directory, err := dependencies.workingDirectory()
	if err != nil {
		return fmt.Errorf("verifier-tasks: resolving working directory: %w", err)
	}
	resolve := verifiertasks.ResolveRequest{Change: *change, Store: *store, WorkingDirectory: directory}
	var result verifiertasks.Result
	switch phase {
	case "snapshot":
		if dependencies.captureVerifierTasks == nil {
			return fmt.Errorf("verifier-tasks: snapshot operation is unavailable")
		}
		result, err = dependencies.captureVerifierTasks(context.Background(), resolve)
	case "complete":
		if dependencies.completeVerifierTasks == nil {
			return fmt.Errorf("verifier-tasks: completion operation is unavailable")
		}
		if dependencies.stdin == nil {
			return fmt.Errorf("verifier-tasks complete: JSON stdin is required")
		}
		var request verifiertasks.CompleteRequest
		decoder := json.NewDecoder(dependencies.stdin)
		decoder.DisallowUnknownFields()
		if decodeErr := decoder.Decode(&request); decodeErr != nil {
			return fmt.Errorf("verifier-tasks complete: decoding JSON stdin: %w", decodeErr)
		}
		if decodeErr := ensureJSONEnd(decoder); decodeErr != nil {
			return fmt.Errorf("verifier-tasks complete: decoding JSON stdin: %w", decodeErr)
		}
		result, err = dependencies.completeVerifierTasks(context.Background(), resolve, request)
	default:
		return fmt.Errorf("verifier-tasks: unknown phase %q", phase)
	}
	if encodeErr := json.NewEncoder(dependencies.stdout).Encode(result); encodeErr != nil {
		return encodeErr
	}
	// Logical rejections and conflicts are complete structured results and return
	// a nil operation error. Non-nil errors are reserved for malformed input or
	// infrastructure and encoding failures, which map to a non-zero CLI exit.
	return err
}

// runOpenSpecBootstrapCommand prepares the working directory (and optional
// explicit store) for the OpenSpec workflow and emits one structured JSON
// result. A logical block is a complete result with exit 0; non-zero exits
// are reserved for malformed input and infrastructure failures.
func runOpenSpecBootstrapCommand(args []string, dependencies cliDependencies) error {
	flags := flag.NewFlagSet("angel-ai openspec-bootstrap", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	store := flags.String("store", "", "explicit OpenSpec store id")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("openspec-bootstrap: unexpected argument %q", flags.Arg(0))
	}
	if dependencies.runOpenSpecBootstrap == nil {
		return fmt.Errorf("openspec-bootstrap: operation is unavailable")
	}
	if dependencies.workingDirectory == nil {
		return fmt.Errorf("openspec-bootstrap: working-directory resolver is unavailable")
	}
	directory, err := dependencies.workingDirectory()
	if err != nil {
		return fmt.Errorf("openspec-bootstrap: resolving working directory: %w", err)
	}
	result, err := dependencies.runOpenSpecBootstrap(context.Background(), openspecbootstrap.Request{
		WorkingDirectory: directory,
		Store:            *store,
	})
	if err != nil {
		return err
	}
	return json.NewEncoder(dependencies.stdout).Encode(result)
}

func ensureJSONEnd(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("multiple JSON values are not allowed")
		}
		return err
	}
	return nil
}

func runRootCommand(args []string, dependencies cliDependencies) error {
	flags := flag.NewFlagSet("angel-ai", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	options := rootOptions{}
	flags.StringVar(&options.assetsDir, "assets", "", "assets directory override (default: embedded assets)")
	flags.StringVar(&options.configDir, "target", "", "opencode config directory (default: ~/.config/opencode)")
	flags.BoolVar(&options.all, "all", false, "install everything without the TUI")
	flags.BoolVar(&options.dryRun, "dry-run", false, "with --all, print the plan without installing")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unknown command %q", flags.Arg(0))
	}

	if !options.all {
		if err := runUpdatePolicyFailOpen(false, dependencies); err != nil {
			return err
		}
	}
	return dependencies.runInstaller(options)
}

func runVersionCommand(args []string, dependencies cliDependencies) error {
	flags := flag.NewFlagSet("angel-ai version", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("version: unexpected argument %q", flags.Arg(0))
	}
	_, err := fmt.Fprintln(dependencies.stdout, version)
	return err
}

func runUpdateCommand(args []string, dependencies cliDependencies) error {
	flags := flag.NewFlagSet("angel-ai update", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("update: unexpected argument %q", flags.Arg(0))
	}
	return runUpdatePolicyFailOpen(true, dependencies)
}

func runUpdatePolicyFailOpen(forced bool, dependencies cliDependencies) error {
	if err := runUpdatePolicy(forced, dependencies); err != nil {
		_, warningErr := fmt.Fprintf(dependencies.stdout, "warning: update failed: %v\n", err)
		return warningErr
	}
	return nil
}

func runUpdatePolicy(forced bool, dependencies cliDependencies) error {
	if version == "dev" {
		if forced {
			_, err := fmt.Fprintln(dependencies.stdout, "self-update is disabled for dev builds")
			return err
		}
		return nil
	}
	return dependencies.newUpdatePolicy().Run(version, forced)
}

func run(assetsDir, configDir string, all, dryRun bool) error {
	assetSource, err := sourceForAssets(assetsDir)
	if err != nil {
		return err
	}
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		configDir = filepath.Join(home, ".config", "opencode")
	}

	categories, err := catalog.Load(assetSource)
	if err != nil {
		return err
	}

	if !all {
		return tui.Run(categories, assetSource, configDir)
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
			Items: items, Extras: extras, Assets: assetSource, ConfigDir: configDir,
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
		Items: items, Extras: extras, Assets: assetSource, ConfigDir: configDir,
	})
	for _, line := range report {
		fmt.Println(line)
	}
	return err
}

func sourceForAssets(directory string) (assets.Source, error) {
	if directory != "" {
		return assets.Directory(directory), nil
	}
	embedded, err := fs.Sub(embeddedAssetTree, "assets")
	if err != nil {
		return assets.Source{}, fmt.Errorf("opening embedded assets: %w", err)
	}
	return assets.Embedded(embedded), nil
}
