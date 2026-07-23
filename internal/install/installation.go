package install

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"angel-ai-opencode/internal/assets"
	"angel-ai-opencode/internal/catalog"
)

// InstallationRequest is the complete desired installer selection. Planning
// and applying consume the same request so they cannot disagree about extras or
// about files shared by more than one feature, such as AGENTS.md.
type InstallationRequest struct {
	Items     []catalog.Item
	Extras    map[string]bool
	Assets    assets.Source
	ConfigDir string
}

type preparedFile struct {
	path            string
	content         []byte
	perm            os.FileMode
	exists          bool
	unchanged       bool
	fullReplacement bool
	jsonObject      map[string]any
}

type preparedInstallation struct {
	files      []preparedFile
	globalCLIs []globalCLIDescriptor
}

var prepareInstallationForApply = prepareInstallation

// PlanInstallation inspects the destination and describes the exact changes
// ApplyInstallation would make without mutating the machine.
func PlanInstallation(request InstallationRequest) ([]string, error) {
	if err := preflightSelectedExtras(request.Extras, systemGlobalCLICommands.lookPath); err != nil {
		return nil, err
	}
	prepared, err := prepareInstallation(request)
	if err != nil {
		return nil, err
	}
	snapshot, err := preflightGlobalCLIs(prepared.globalCLIs, systemGlobalCLICommands)
	if err != nil {
		return nil, err
	}
	lines := make([]string, 0, len(prepared.files)+len(snapshot.inspections))
	for _, file := range prepared.files {
		lines = append(lines, file.planLine())
	}
	for _, inspection := range snapshot.inspections {
		lines = append(lines, inspection.reportLine())
	}
	return lines, nil
}

// ApplyInstallation validates the complete desired state before performing any
// package installation or file write, then applies only changed files.
func ApplyInstallation(request InstallationRequest) ([]string, error) {
	if err := preflightSelectedExtras(request.Extras, systemGlobalCLICommands.lookPath); err != nil {
		return nil, err
	}
	prepared, err := prepareInstallationForApply(request)
	if err != nil {
		return nil, err
	}
	snapshot, err := preflightGlobalCLIs(prepared.globalCLIs, systemGlobalCLICommands)
	if err != nil {
		return nil, err
	}
	reprepareAfterCLIs := len(snapshot.inspections) > 0
	var done []string
	for _, inspection := range snapshot.inspections {
		line, err := applyGlobalCLIInspection(inspection, snapshot.manager, systemGlobalCLICommands)
		if err != nil {
			return done, err
		}
		done = append(done, line)
	}
	if reprepareAfterCLIs {
		prepared, err = prepareInstallationForApply(request)
		if err != nil {
			return done, err
		}
	}
	for _, file := range prepared.files {
		result, err := reconcileFile(file)
		if err != nil {
			return done, err
		}
		done = append(done, fileResultLines(file.path, result)...)
	}
	return done, nil
}

func (file preparedFile) contentMatches(content []byte) bool {
	if file.jsonObject == nil {
		return bytes.Equal(content, file.content)
	}
	var current map[string]any
	if err := json.Unmarshal(content, &current); err != nil {
		return false
	}
	return reflect.DeepEqual(current, file.jsonObject)
}

func prepareInstallation(request InstallationRequest) (preparedInstallation, error) {
	prepared := preparedInstallation{globalCLIs: selectedGlobalCLIs(request.Extras)}
	var fragments []map[string]any
	var globalAgents *catalog.Item

	for _, item := range request.Items {
		switch item.Kind {
		case catalog.MergeJSON:
			patch, err := readAssetJSONObject(request.Assets, item.Source)
			if err != nil {
				return preparedInstallation{}, fmt.Errorf("parsing fragment %s: %w", item.Name, err)
			}
			fragments = append(fragments, patch)
		case catalog.CopyDir:
			files, err := prepareDirectory(request.Assets, item.Source, filepath.Join(request.ConfigDir, item.Dest))
			if err != nil {
				return preparedInstallation{}, fmt.Errorf("preparing %s: %w", item.Name, err)
			}
			prepared.files = append(prepared.files, files...)
		default:
			if filepath.Clean(item.Dest) == "AGENTS.md" {
				copy := item
				globalAgents = &copy
				continue
			}
			file, err := prepareSourceFile(request.Assets, item.Source, filepath.Join(request.ConfigDir, item.Dest), false)
			if err != nil {
				return preparedInstallation{}, fmt.Errorf("preparing %s: %w", item.Name, err)
			}
			prepared.files = append(prepared.files, file)
		}
	}

	codegraphSelected, codegraphSpecified := request.Extras[codegraphOptionKey]
	var codegraphObject map[string]any
	if codegraphSpecified && codegraphSelected {
		patch, err := readAssetJSONObject(request.Assets, "integrations/codegraph/mcp.json")
		if err != nil {
			return preparedInstallation{}, fmt.Errorf("reading CodeGraph MCP config: %w", err)
		}
		mcp, ok := patch["mcp"].(map[string]any)
		if !ok {
			return preparedInstallation{}, fmt.Errorf("CodeGraph MCP config has no mcp object")
		}
		codegraphObject, ok = mcp["codegraph"].(map[string]any)
		if !ok {
			return preparedInstallation{}, fmt.Errorf("CodeGraph MCP config has no codegraph object")
		}
	}

	opencodeFile, ok, err := prepareJSONObject(
		filepath.Join(request.ConfigDir, "opencode.json"),
		"https://opencode.ai/config.json",
		fragments,
		codegraphSpecified,
		codegraphSelected,
		codegraphObject,
	)
	if err != nil {
		return preparedInstallation{}, err
	}
	if ok {
		prepared.files = append(prepared.files, opencodeFile)
	}

	if err := prepareCMUXExtra(&prepared, request); err != nil {
		return preparedInstallation{}, err
	}
	if err := prepareUIExtras(&prepared, request); err != nil {
		return preparedInstallation{}, err
	}
	agentsFile, ok, err := prepareAgentsFile(request, globalAgents, codegraphSpecified, codegraphSelected)
	if err != nil {
		return preparedInstallation{}, err
	}
	if ok {
		prepared.files = append(prepared.files, agentsFile)
	}

	sort.SliceStable(prepared.files, func(i, j int) bool {
		if prepared.files[i].fullReplacement != prepared.files[j].fullReplacement {
			return prepared.files[i].fullReplacement
		}
		return prepared.files[i].path < prepared.files[j].path
	})
	return prepared, nil
}

func (file preparedFile) planLine() string {
	switch {
	case file.unchanged:
		return "SIN CAMBIOS " + file.path
	case !file.exists:
		return "CREAR       " + file.path
	case file.fullReplacement:
		return "REEMPLAZAR  " + file.path + " (reemplazo completo; se creará backup)"
	default:
		return "ACTUALIZAR  " + file.path + " (se creará backup)"
	}
}

func prepareSourceFile(source assets.Source, sourcePath, target string, fullReplacement bool) (preparedFile, error) {
	content, err := source.ReadFile(sourcePath)
	if err != nil {
		return preparedFile{}, err
	}
	mode, err := source.FileMode(sourcePath)
	if err != nil {
		return preparedFile{}, err
	}
	return prepareFile(target, content, mode, fullReplacement)
}

func prepareDirectory(source assets.Source, sourceRoot, target string) ([]preparedFile, error) {
	var files []preparedFile
	err := source.WalkDir(sourceRoot, func(sourcePath string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		relative := strings.TrimPrefix(sourcePath, sourceRoot+"/")
		file, err := prepareSourceFile(source, sourcePath, filepath.Join(target, filepath.FromSlash(relative)), false)
		if err != nil {
			return err
		}
		files = append(files, file)
		return nil
	})
	return files, err
}

func prepareFile(path string, content []byte, perm os.FileMode, fullReplacement bool) (preparedFile, error) {
	file := preparedFile{
		path: path, content: append([]byte(nil), content...), perm: perm,
		fullReplacement: fullReplacement,
	}
	existing, err := os.ReadFile(path)
	switch {
	case err == nil:
		file.exists = true
		file.unchanged = bytes.Equal(existing, content)
	case os.IsNotExist(err):
	case err != nil:
		return preparedFile{}, err
	}
	return file, nil
}

func prepareCMUXExtra(prepared *preparedInstallation, request InstallationRequest) error {
	if !request.Extras[cmuxOptionKey] {
		return nil
	}
	for _, name := range cmuxPluginFiles {
		file, err := prepareSourceFile(
			request.Assets,
			path.Join("integrations", "cmux", name),
			filepath.Join(request.ConfigDir, "plugins", name),
			false,
		)
		if err != nil {
			return fmt.Errorf("preparing cmux plugin %s: %w", name, err)
		}
		prepared.files = append(prepared.files, file)
	}
	return nil
}

func readAssetJSONObject(source assets.Source, sourcePath string) (map[string]any, error) {
	raw, err := source.ReadFile(sourcePath)
	if err != nil {
		return nil, err
	}
	var object map[string]any
	if err := json.Unmarshal(raw, &object); err != nil {
		return nil, err
	}
	if object == nil {
		return nil, fmt.Errorf("JSON root is not an object")
	}
	return object, nil
}

func prepareJSONObject(
	path, defaultSchema string,
	patches []map[string]any,
	codegraphSpecified, codegraphSelected bool,
	codegraphObject map[string]any,
) (preparedFile, bool, error) {
	raw, err := os.ReadFile(path)
	exists := err == nil
	if err != nil && !os.IsNotExist(err) {
		return preparedFile{}, false, err
	}
	if !exists && len(patches) == 0 && (!codegraphSpecified || !codegraphSelected) {
		return preparedFile{}, false, nil
	}

	original := map[string]any{}
	if exists {
		if err := json.Unmarshal(raw, &original); err != nil {
			return preparedFile{}, false, fmt.Errorf("parsing %s: %w", path, err)
		}
	}
	config, err := cloneJSONObject(original)
	if err != nil {
		return preparedFile{}, false, err
	}
	if len(patches) > 0 || (codegraphSpecified && codegraphSelected) {
		if _, ok := config["$schema"]; !ok {
			config["$schema"] = defaultSchema
		}
	}
	for _, patch := range patches {
		merge(config, patch)
	}
	if codegraphSpecified {
		mcp, _ := config["mcp"].(map[string]any)
		if codegraphSelected {
			if mcp == nil {
				mcp = map[string]any{}
				config["mcp"] = mcp
			}
			cloned, err := cloneJSONObject(codegraphObject)
			if err != nil {
				return preparedFile{}, false, err
			}
			mcp["codegraph"] = cloned
		} else if mcp != nil {
			delete(mcp, "codegraph")
			if len(mcp) == 0 {
				delete(config, "mcp")
			}
		}
	}

	encoded, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return preparedFile{}, false, err
	}
	file, err := prepareFile(path, append(encoded, '\n'), 0o644, false)
	if err != nil {
		return preparedFile{}, false, err
	}
	file.jsonObject = config
	file.unchanged = exists && reflect.DeepEqual(original, config)
	return file, true, nil
}

func cloneJSONObject(source map[string]any) (map[string]any, error) {
	if source == nil {
		return map[string]any{}, nil
	}
	raw, err := json.Marshal(source)
	if err != nil {
		return nil, err
	}
	var cloned map[string]any
	if err := json.Unmarshal(raw, &cloned); err != nil {
		return nil, err
	}
	return cloned, nil
}

func prepareUIExtras(prepared *preparedInstallation, request InstallationRequest) error {
	var patches []map[string]any
	if request.Extras["angel-logo"] {
		for _, name := range angelLogoFiles {
			file, err := prepareSourceFile(
				request.Assets,
				path.Join("tui-plugins", name),
				filepath.Join(request.ConfigDir, "tui-plugins", name),
				false,
			)
			if err != nil {
				return fmt.Errorf("preparing %s: %w", name, err)
			}
			prepared.files = append(prepared.files, file)
		}
		patches = append(patches, map[string]any{
			"plugin": []any{filepath.Join(request.ConfigDir, "tui-plugins", "angel-logo.tsx")},
		})
	}
	if request.Extras["theme"] {
		patches = append(patches, map[string]any{"theme": "one-dark-pro"})
	}
	if request.Extras["subagent-statusline"] {
		patches = append(patches, map[string]any{"plugin": []any{"opencode-subagent-statusline"}})
	}
	file, ok, err := prepareJSONObject(
		filepath.Join(request.ConfigDir, "tui.json"),
		"https://opencode.ai/tui.json",
		patches,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}
	if ok {
		prepared.files = append(prepared.files, file)
	}
	return nil
}

func prepareAgentsFile(
	request InstallationRequest,
	globalAgents *catalog.Item,
	codegraphSpecified, codegraphSelected bool,
) (preparedFile, bool, error) {
	path := filepath.Join(request.ConfigDir, "AGENTS.md")
	if globalAgents != nil {
		content, err := request.Assets.ReadFile(globalAgents.Source)
		if err != nil {
			return preparedFile{}, false, err
		}
		if codegraphSpecified && codegraphSelected {
			guidance, err := request.Assets.ReadFile("integrations/codegraph/AGENTS.md")
			if err != nil {
				return preparedFile{}, false, err
			}
			content = joinDocumentParts(string(content), string(guidance))
		}
		file, err := prepareFile(path, content, 0o644, true)
		return file, true, err
	}
	if !codegraphSpecified {
		return preparedFile{}, false, nil
	}
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return preparedFile{}, false, err
	}
	if os.IsNotExist(err) && !codegraphSelected {
		return preparedFile{}, false, nil
	}
	updated, err := removeManagedBlock(
		string(existing),
		"<!-- codegraph-guidance -->",
		"<!-- /codegraph-guidance -->",
	)
	if err != nil {
		return preparedFile{}, false, err
	}
	if codegraphSelected {
		guidance, err := request.Assets.ReadFile("integrations/codegraph/AGENTS.md")
		if err != nil {
			return preparedFile{}, false, err
		}
		updated = string(joinDocumentParts(updated, string(guidance)))
	} else if updated != "" {
		updated = strings.TrimSpace(updated) + "\n"
	}
	file, err := prepareFile(path, []byte(updated), 0o644, false)
	return file, true, err
}

func joinDocumentParts(parts ...string) []byte {
	var nonempty []string
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			nonempty = append(nonempty, trimmed)
		}
	}
	if len(nonempty) == 0 {
		return nil
	}
	return []byte(strings.Join(nonempty, "\n\n") + "\n")
}
