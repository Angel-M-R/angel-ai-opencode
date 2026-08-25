package managedassets

import (
	"errors"

	"angel-ai-opencode/internal/assets"
)

// FindingCode is a stable, machine-checkable doctor result.
type FindingCode string

const (
	FindingStateMissing  FindingCode = "state_missing"
	FindingBundleChanged FindingCode = "bundle_changed"
	FindingFileMissing   FindingCode = "file_missing"
	FindingFileModified  FindingCode = "file_modified"
	FindingRetiredFile   FindingCode = "retired_file"
)

// Finding describes one mismatch without changing the target.
type Finding struct {
	Code     FindingCode
	Path     string
	Expected string
	Actual   string
	Detail   string
}

// DoctorReport is healthy only when state, bundle and every managed file agree.
type DoctorReport struct {
	Healthy  bool
	Findings []Finding
}

// Doctor compares the target with its saved state and the supplied asset
// bundle. It performs no write, package inspection or process execution.
func Doctor(source assets.Source, configDir string) (DoctorReport, error) {
	saved, err := load(configDir)
	if errors.Is(err, errStateNotFound) {
		return DoctorReport{Findings: []Finding{{Code: FindingStateMissing, Path: statePath(configDir)}}}, nil
	}
	if err != nil {
		return DoctorReport{}, err
	}
	bundleDigest, err := assets.Digest(source)
	if err != nil {
		return DoctorReport{}, err
	}
	var findings []Finding
	if bundleDigest != saved.BundleDigest {
		findings = append(findings, Finding{
			Code: FindingBundleChanged, Expected: saved.BundleDigest, Actual: bundleDigest,
		})
	}
	request, lifecycleFindings, err := requestFromState(saved, source, configDir)
	if err != nil {
		return DoctorReport{}, err
	}
	findings = append(findings, inspectFiles(saved, configDir, request.FileExpectations)...)
	findings = append(findings, lifecycleFindings...)
	return DoctorReport{Healthy: len(findings) == 0, Findings: findings}, nil
}
