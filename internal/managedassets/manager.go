package managedassets

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"angel-ai-opencode/internal/assets"
	"angel-ai-opencode/internal/catalog"
	"angel-ai-opencode/internal/install"
)

// Apply installs a request and records the exact selection and resulting file
// bytes after the installer succeeds.
func Apply(request install.InstallationRequest) ([]string, error) {
	targets, err := install.ManagedFilePaths(request)
	if err != nil {
		return nil, err
	}
	report, digests, err := install.ApplyInstallationWithDigests(request)
	if err != nil {
		return report, err
	}
	state, err := stateFromRequest(request, targets, digests)
	if err != nil {
		return report, err
	}
	if err := save(request.ConfigDir, state); err != nil {
		return report, fmt.Errorf("assets were installed but managed state could not be saved; fix the target and rerun the installer: %w", err)
	}
	return report, nil
}

// Sync rebuilds the saved request against source. An apply stops when a
// recorded file has drifted. A dry run still returns the installer's plan and
// a DriftError so callers can show both the proposed changes and the blocker.
func Sync(source assets.Source, configDir string, dryRun bool) ([]string, error) {
	saved, err := load(configDir)
	if err != nil {
		return nil, err
	}
	request, lifecycleFindings, err := requestFromState(saved, source, configDir)
	if err != nil {
		return nil, err
	}
	findings := append(inspectFiles(saved, configDir, request.FileExpectations), lifecycleFindings...)
	if len(findings) > 0 && !dryRun {
		return nil, &DriftError{Findings: findings}
	}
	plan, err := install.PlanInstallation(request)
	if err != nil {
		return nil, err
	}
	if dryRun {
		if len(findings) > 0 {
			return plan, &DriftError{Findings: findings}
		}
		return plan, nil
	}
	return Apply(request)
}

func stateFromRequest(request install.InstallationRequest, targets []string, digests map[string]string) (state, error) {
	bundleDigest, err := assets.Digest(request.Assets)
	if err != nil {
		return state{}, fmt.Errorf("digesting asset bundle: %w", err)
	}
	savedSelection, err := selectionFromRequest(request)
	if err != nil {
		return state{}, err
	}
	files := make([]managedFile, 0, len(targets))
	for _, target := range targets {
		relative, err := managedRelativePath(request.ConfigDir, target)
		if err != nil {
			return state{}, err
		}
		// Digests come from the installer's transaction, not from re-reading
		// the target: a re-read after the installation lock is released could
		// record a concurrent installer's bytes as this baseline.
		digest, ok := digests[target]
		if !ok {
			return state{}, fmt.Errorf("installer reported no digest for managed file %s", target)
		}
		files = append(files, managedFile{Path: relative, Digest: digest})
	}
	return state{
		SchemaVersion: schemaVersion,
		BundleDigest:  bundleDigest,
		Selection:     savedSelection,
		Files:         sortedFiles(files),
	}, nil
}

func selectionFromRequest(request install.InstallationRequest) (selection, error) {
	grouped := map[string][]string{}
	seen := map[string]struct{}{}
	for _, item := range request.Items {
		source := path.Clean(item.Source)
		category, _, ok := strings.Cut(source, "/")
		if !ok || category == "" || source == "." || source == ".." || strings.HasPrefix(source, "../") {
			return selection{}, fmt.Errorf("asset source %q has no safe category", item.Source)
		}
		key := category + "\x00" + source
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		grouped[category] = append(grouped[category], source)
	}
	categories := make([]categorySelection, 0, len(grouped))
	for name, sources := range grouped {
		sort.Strings(sources)
		categories = append(categories, categorySelection{Name: name, Sources: sources})
	}
	sort.Slice(categories, func(i, j int) bool { return categories[i].Name < categories[j].Name })

	extras := make(map[string]bool, len(request.Extras))
	for name, selected := range request.Extras {
		extras[name] = selected
	}
	models := make(install.AgentModelAssignments, len(request.AgentModels))
	for name, assignment := range request.AgentModels {
		models[name] = assignment
	}
	return selection{Categories: categories, Extras: extras, AgentModels: models}, nil
}

func requestFromState(saved state, source assets.Source, configDir string) (install.InstallationRequest, []Finding, error) {
	categories, err := catalog.Load(source)
	if err != nil {
		return install.InstallationRequest{}, nil, err
	}
	available := make(map[string]map[string]catalog.Item, len(categories))
	for _, category := range categories {
		items := make(map[string]catalog.Item, len(category.Items))
		for _, item := range category.Items {
			items[item.Source] = item
		}
		available[category.Name] = items
	}
	// A saved category or asset the current bundle no longer offers is a
	// normal lifecycle event, not a dead end: drop it from the rebuilt
	// selection so the retired-file scan below can name its installed files
	// and the inventory can advance once they are gone.
	var selected []catalog.Item
	for _, category := range saved.Selection.Categories {
		items := available[category.Name]
		for _, sourcePath := range category.Sources {
			item, ok := items[sourcePath]
			if !ok {
				continue
			}
			selected = append(selected, item)
		}
	}
	extras := make(map[string]bool, len(saved.Selection.Extras))
	for name, value := range saved.Selection.Extras {
		extras[name] = value
	}
	models := make(install.AgentModelAssignments, len(saved.Selection.AgentModels))
	for name, assignment := range saved.Selection.AgentModels {
		models[name] = assignment
	}
	request := install.InstallationRequest{
		Items:       selected,
		Extras:      extras,
		Assets:      source,
		ConfigDir:   configDir,
		AgentModels: models,
	}
	targets, err := install.ManagedFilePaths(request)
	if err != nil {
		return install.InstallationRequest{}, nil, err
	}
	previous := make(map[string]string, len(saved.Files))
	for _, file := range saved.Files {
		previous[file.Path] = file.Digest
	}
	current := make(map[string]struct{}, len(targets))
	request.FileExpectations = make(map[string]install.FileExpectation, len(targets))
	for _, target := range targets {
		relative, err := managedRelativePath(configDir, target)
		if err != nil {
			return install.InstallationRequest{}, nil, err
		}
		current[relative] = struct{}{}
		if digest, exists := previous[relative]; exists {
			request.FileExpectations[target] = install.FileExpectation{SHA256: digest}
		} else {
			request.FileExpectations[target] = install.FileExpectation{Absent: true}
		}
	}
	var findings []Finding
	for _, file := range saved.Files {
		if _, exists := current[file.Path]; exists {
			continue
		}
		fullPath := filepath.Join(configDir, filepath.FromSlash(file.Path))
		_, err := os.Lstat(fullPath)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		detail := "the current bundle no longer produces this managed file; sync leaves it in place"
		if err != nil {
			detail = fmt.Sprintf("checking retired managed file: %v", err)
		}
		findings = append(findings, Finding{
			Code: FindingRetiredFile, Path: file.Path, Expected: file.Digest,
			Detail: detail,
		})
	}
	return request, findings, nil
}

func managedRelativePath(configDir, target string) (string, error) {
	relative, err := filepath.Rel(configDir, target)
	if err != nil {
		return "", err
	}
	relative = filepath.ToSlash(relative)
	if !safeRelativePath(relative) || relative == stateFileName {
		return "", fmt.Errorf("managed path %q is outside the installation target", target)
	}
	return relative, nil
}

// DriftError lists files whose current bytes no longer match managed state.
type DriftError struct {
	Findings []Finding
}

func (err *DriftError) Error() string {
	paths := make([]string, 0, len(err.Findings))
	for _, finding := range err.Findings {
		paths = append(paths, finding.Path)
	}
	return "managed asset sync is blocked: " + strings.Join(paths, ", ")
}

func inspectFiles(saved state, configDir string, current map[string]install.FileExpectation) []Finding {
	var findings []Finding
	for _, managed := range saved.Files {
		fullPath := filepath.Join(configDir, filepath.FromSlash(managed.Path))
		if _, stillManaged := current[fullPath]; !stillManaged {
			continue
		}
		digest, err := digestFile(fullPath)
		switch {
		case errors.Is(err, os.ErrNotExist):
			findings = append(findings, Finding{Code: FindingFileMissing, Path: managed.Path, Expected: managed.Digest})
		case err != nil:
			findings = append(findings, Finding{Code: FindingFileModified, Path: managed.Path, Expected: managed.Digest, Detail: err.Error()})
		case digest != managed.Digest:
			findings = append(findings, Finding{Code: FindingFileModified, Path: managed.Path, Expected: managed.Digest, Actual: digest})
		}
	}
	return findings
}
