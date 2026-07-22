package install

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const (
	codegraphRegistryPackage   = "@colbymchenry/codegraph"
	openSpecRegistryPackage    = "@fission-ai/openspec"
	openSpecPackage            = openSpecRegistryPackage + "@latest"
	openSpecMinimumNodeVersion = "20.19.0"
)

var semanticVersionPattern = regexp.MustCompile(
	`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$`,
)

type globalCLIDescriptor struct {
	optionKey          string
	displayName        string
	registryPackage    string
	installSpec        string
	executable         string
	versionArgs        []string
	minimumNodeVersion string
}

var openSpecGlobalCLI = globalCLIDescriptor{
	optionKey:          openSpecOptionKey,
	displayName:        "OpenSpec",
	registryPackage:    openSpecRegistryPackage,
	installSpec:        openSpecPackage,
	executable:         "openspec",
	versionArgs:        []string{"--version"},
	minimumNodeVersion: openSpecMinimumNodeVersion,
}

var globalCLIDescriptors = []globalCLIDescriptor{
	{
		optionKey:       codegraphOptionKey,
		displayName:     "CodeGraph",
		registryPackage: codegraphRegistryPackage,
		installSpec:     codegraphPackage,
		executable:      "codegraph",
		versionArgs:     []string{"--version"},
	},
	openSpecGlobalCLI,
}

func selectedGlobalCLIs(extras map[string]bool) []globalCLIDescriptor {
	selected := make([]globalCLIDescriptor, 0, len(globalCLIDescriptors))
	for _, descriptor := range globalCLIDescriptors {
		if extras[descriptor.optionKey] {
			selected = append(selected, descriptor)
		}
	}
	return selected
}

type globalCLICommands struct {
	lookPath func(string) (string, error)
	run      func(string, ...string) ([]byte, error)
}

var systemGlobalCLICommands = globalCLICommands{
	lookPath: exec.LookPath,
	run: func(path string, args ...string) ([]byte, error) {
		return exec.Command(path, args...).CombinedOutput()
	},
}

type globalPackageManager struct {
	name                     string
	executable               string
	globalBin                string
	install                  []string
	probePackageRegistration func(globalCLIDescriptor, globalCLICommands) globalCLIPackageRegistration
	probeLatestVersion       func(globalCLIDescriptor, globalCLICommands) globalCLIRegistryVersion
}

type globalCLIPackageRegistrationState string

const (
	globalCLIPackageRegistered              globalCLIPackageRegistrationState = "registered"
	globalCLIPackageUnregistered            globalCLIPackageRegistrationState = "unregistered"
	globalCLIPackageRegistrationUnavailable globalCLIPackageRegistrationState = "unavailable"
	globalCLIPackageRegistrationMalformed   globalCLIPackageRegistrationState = "malformed"
)

type globalCLIPackageRegistration struct {
	state  globalCLIPackageRegistrationState
	detail string
}

type globalCLIRegistryVersionState string

const (
	globalCLIRegistryVersionAvailable   globalCLIRegistryVersionState = "available"
	globalCLIRegistryVersionUnavailable globalCLIRegistryVersionState = "unavailable"
	globalCLIRegistryVersionMalformed   globalCLIRegistryVersionState = "malformed"
)

type globalCLIRegistryVersion struct {
	state   globalCLIRegistryVersionState
	version semanticVersion
	detail  string
}

type globalCLIInspectionDisposition string

const (
	globalCLIInstall            globalCLIInspectionDisposition = "install"
	globalCLICurrent            globalCLIInspectionDisposition = "current"
	globalCLIOutdated           globalCLIInspectionDisposition = "outdated"
	globalCLIAhead              globalCLIInspectionDisposition = "ahead"
	globalCLIRegistryUnverified globalCLIInspectionDisposition = "registry-unverified"
)

type globalCLIInspection struct {
	descriptor       globalCLIDescriptor
	registration     globalCLIPackageRegistration
	executablePath   string
	installedVersion *semanticVersion
	registryVersion  *semanticVersion
	disposition      globalCLIInspectionDisposition
	warning          string
}

type globalCLIInspectionSnapshot struct {
	manager     globalPackageManager
	inspections []globalCLIInspection
}

func (inspection globalCLIInspection) reportLine() string {
	if inspection.disposition == globalCLIInstall {
		return "INSTALAR   " + inspection.descriptor.installSpec
	}
	line := fmt.Sprintf(
		"CLI         %s: %s (installed %s",
		inspection.descriptor.displayName,
		inspection.disposition,
		inspectionVersion(inspection.installedVersion),
	)
	if inspection.registryVersion != nil {
		line += "; registry latest " + inspection.registryVersion.String()
	}
	line += ")"
	if inspection.warning != "" {
		line += "; WARNING: " + inspection.warning
	}
	return line
}

func inspectionVersion(version *semanticVersion) string {
	if version == nil {
		return "unknown"
	}
	return version.String()
}

type semanticVersion struct {
	major      uint64
	minor      uint64
	patch      uint64
	prerelease string
	build      string
}

func parseSemanticVersion(raw string) (semanticVersion, error) {
	value := strings.TrimSpace(raw)
	value = strings.TrimPrefix(value, "v")
	matches := semanticVersionPattern.FindStringSubmatch(value)
	if matches == nil || hasInvalidNumericPrerelease(matches[4]) {
		return semanticVersion{}, fmt.Errorf("invalid semantic version %q", raw)
	}
	parts := make([]uint64, 3)
	for index := range parts {
		part, err := strconv.ParseUint(matches[index+1], 10, 64)
		if err != nil {
			return semanticVersion{}, fmt.Errorf("invalid semantic version %q", raw)
		}
		parts[index] = part
	}
	return semanticVersion{
		major: parts[0], minor: parts[1], patch: parts[2],
		prerelease: matches[4], build: matches[5],
	}, nil
}

func hasInvalidNumericPrerelease(prerelease string) bool {
	for _, identifier := range strings.Split(prerelease, ".") {
		if len(identifier) > 1 && identifier[0] == '0' && isDecimal(identifier) {
			return true
		}
	}
	return false
}

func isDecimal(value string) bool {
	if value == "" {
		return false
	}
	for _, character := range value {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}

func (version semanticVersion) String() string {
	value := fmt.Sprintf("%d.%d.%d", version.major, version.minor, version.patch)
	if version.prerelease != "" {
		value += "-" + version.prerelease
	}
	if version.build != "" {
		value += "+" + version.build
	}
	return value
}

func compareSemanticVersions(left, right semanticVersion) int {
	leftCore := []uint64{left.major, left.minor, left.patch}
	rightCore := []uint64{right.major, right.minor, right.patch}
	for index := range leftCore {
		if leftCore[index] < rightCore[index] {
			return -1
		}
		if leftCore[index] > rightCore[index] {
			return 1
		}
	}
	if left.prerelease == "" && right.prerelease == "" {
		return 0
	}
	if left.prerelease == "" {
		return 1
	}
	if right.prerelease == "" {
		return -1
	}
	return comparePrereleases(left.prerelease, right.prerelease)
}

func comparePrereleases(left, right string) int {
	leftParts := strings.Split(left, ".")
	rightParts := strings.Split(right, ".")
	for index := 0; index < len(leftParts) && index < len(rightParts); index++ {
		leftNumeric := isDecimal(leftParts[index])
		rightNumeric := isDecimal(rightParts[index])
		switch {
		case leftNumeric && rightNumeric:
			if len(leftParts[index]) < len(rightParts[index]) {
				return -1
			}
			if len(leftParts[index]) > len(rightParts[index]) {
				return 1
			}
		case leftNumeric:
			return -1
		case rightNumeric:
			return 1
		}
		if leftParts[index] < rightParts[index] {
			return -1
		}
		if leftParts[index] > rightParts[index] {
			return 1
		}
	}
	if len(leftParts) < len(rightParts) {
		return -1
	}
	if len(leftParts) > len(rightParts) {
		return 1
	}
	return 0
}

func validateGlobalCLIRuntimes(descriptors []globalCLIDescriptor, commands globalCLICommands) error {
	for _, descriptor := range descriptors {
		if descriptor.minimumNodeVersion == "" {
			continue
		}
		if err := validateNodeVersion(descriptor, commands); err != nil {
			return err
		}
	}
	return nil
}

func validateNodeVersion(descriptor globalCLIDescriptor, commands globalCLICommands) error {
	minimum, err := parseSemanticVersion(descriptor.minimumNodeVersion)
	if err != nil {
		return fmt.Errorf("invalid minimum Node.js version for %s: %w", descriptor.displayName, err)
	}
	node, err := commands.lookPath("node")
	if err != nil {
		return fmt.Errorf(
			"%s requires Node.js >=%s, but node is not available on PATH",
			descriptor.displayName,
			minimum,
		)
	}
	output, err := commands.run(node, "--version")
	if err != nil {
		detail := strings.TrimSpace(string(output))
		if detail == "" {
			return fmt.Errorf("checking Node.js version for %s: %w", descriptor.displayName, err)
		}
		return fmt.Errorf("checking Node.js version for %s: %w: %s", descriptor.displayName, err, detail)
	}
	versionOutput := strings.TrimSpace(string(output))
	if versionOutput == "" {
		return fmt.Errorf(
			"%s requires Node.js >=%s, but node --version returned no version",
			descriptor.displayName,
			minimum,
		)
	}
	version, err := parseSemanticVersion(versionOutput)
	if err != nil {
		return fmt.Errorf(
			"%s requires Node.js >=%s, but node --version returned malformed version %q",
			descriptor.displayName,
			minimum,
			versionOutput,
		)
	}
	if compareSemanticVersions(version, minimum) < 0 {
		return fmt.Errorf(
			"%s requires Node.js >=%s, but found Node.js %s",
			descriptor.displayName,
			minimum,
			version,
		)
	}
	return nil
}

func selectGlobalPackageManager(commands globalCLICommands) (globalPackageManager, error) {
	if npm, err := commands.lookPath("npm"); err == nil {
		return globalPackageManager{
			name:       "npm",
			executable: npm,
			install:    []string{"install", "--global"},
			probePackageRegistration: func(descriptor globalCLIDescriptor, commands globalCLICommands) globalCLIPackageRegistration {
				return probeNPMPackageRegistration(npm, descriptor, commands)
			},
			probeLatestVersion: func(descriptor globalCLIDescriptor, commands globalCLICommands) globalCLIRegistryVersion {
				return probeRegistryLatestVersion(npm, "view", descriptor, commands)
			},
		}, nil
	}

	pnpm, err := commands.lookPath("pnpm")
	if err != nil {
		return globalPackageManager{}, fmt.Errorf("global CLI installation requires npm or pnpm on PATH")
	}
	output, err := commands.run(pnpm, "bin", "-g")
	if err != nil {
		detail := strings.TrimSpace(string(output))
		if detail == "" {
			return globalPackageManager{}, fmt.Errorf("validating pnpm global bin directory: %w", err)
		}
		return globalPackageManager{}, fmt.Errorf("validating pnpm global bin directory: %w: %s", err, detail)
	}
	globalBin := strings.TrimSpace(string(output))
	if globalBin == "" {
		return globalPackageManager{}, fmt.Errorf("validating pnpm global bin directory: pnpm bin -g returned an empty path")
	}
	return globalPackageManager{
		name:       "pnpm",
		executable: pnpm,
		globalBin:  globalBin,
		install:    []string{"add", "--global"},
		probePackageRegistration: func(descriptor globalCLIDescriptor, commands globalCLICommands) globalCLIPackageRegistration {
			return probePNPMPackageRegistration(pnpm, descriptor, commands)
		},
		probeLatestVersion: func(descriptor globalCLIDescriptor, commands globalCLICommands) globalCLIRegistryVersion {
			return probeRegistryLatestVersion(pnpm, "view", descriptor, commands)
		},
	}, nil
}

func probeNPMPackageRegistration(
	executable string,
	descriptor globalCLIDescriptor,
	commands globalCLICommands,
) globalCLIPackageRegistration {
	output, commandErr := commands.run(
		executable,
		"list", "--global", "--depth=0", "--json", descriptor.registryPackage,
	)
	registered, parseErr := parseNPMPackageRegistration(output, descriptor.registryPackage)
	return packageRegistrationProbeResult(registered, parseErr, commandErr, output)
}

func probePNPMPackageRegistration(
	executable string,
	descriptor globalCLIDescriptor,
	commands globalCLICommands,
) globalCLIPackageRegistration {
	output, commandErr := commands.run(
		executable,
		"list", "--global", "--depth=0", "--json", descriptor.registryPackage,
	)
	registered, parseErr := parsePNPMPackageRegistration(output, descriptor.registryPackage)
	return packageRegistrationProbeResult(registered, parseErr, commandErr, output)
}

func packageRegistrationProbeResult(
	registered bool,
	parseErr error,
	commandErr error,
	output []byte,
) globalCLIPackageRegistration {
	if parseErr == nil {
		if registered {
			return globalCLIPackageRegistration{state: globalCLIPackageRegistered}
		}
		return globalCLIPackageRegistration{state: globalCLIPackageUnregistered}
	}
	if commandErr != nil {
		return globalCLIPackageRegistration{
			state:  globalCLIPackageRegistrationUnavailable,
			detail: commandFailureDetail(commandErr, output),
		}
	}
	return globalCLIPackageRegistration{
		state:  globalCLIPackageRegistrationMalformed,
		detail: parseErr.Error(),
	}
}

func parseNPMPackageRegistration(output []byte, packageName string) (bool, error) {
	var listing *struct {
		Dependencies map[string]json.RawMessage `json:"dependencies"`
	}
	if err := json.Unmarshal(output, &listing); err != nil {
		return false, fmt.Errorf("invalid npm package-list JSON: %w", err)
	}
	if listing == nil {
		return false, fmt.Errorf("invalid npm package-list JSON: expected an object")
	}
	entry, ok := listing.Dependencies[packageName]
	if !ok {
		return false, nil
	}
	if !validListedPackage(entry) {
		return false, fmt.Errorf("invalid npm package-list entry for %s", packageName)
	}
	return true, nil
}

func parsePNPMPackageRegistration(output []byte, packageName string) (bool, error) {
	var listings *[]struct {
		Dependencies         map[string]json.RawMessage `json:"dependencies"`
		DevDependencies      map[string]json.RawMessage `json:"devDependencies"`
		OptionalDependencies map[string]json.RawMessage `json:"optionalDependencies"`
	}
	if err := json.Unmarshal(output, &listings); err != nil {
		return false, fmt.Errorf("invalid pnpm package-list JSON: %w", err)
	}
	if listings == nil {
		return false, fmt.Errorf("invalid pnpm package-list JSON: expected an array")
	}
	for _, listing := range *listings {
		for _, dependencies := range []map[string]json.RawMessage{
			listing.Dependencies,
			listing.DevDependencies,
			listing.OptionalDependencies,
		} {
			if entry, ok := dependencies[packageName]; ok {
				if !validListedPackage(entry) {
					return false, fmt.Errorf("invalid pnpm package-list entry for %s", packageName)
				}
				return true, nil
			}
		}
	}
	return false, nil
}

func validListedPackage(raw json.RawMessage) bool {
	var value map[string]any
	return len(raw) > 0 && json.Unmarshal(raw, &value) == nil && value != nil
}

func probeRegistryLatestVersion(
	executable string,
	viewCommand string,
	descriptor globalCLIDescriptor,
	commands globalCLICommands,
) globalCLIRegistryVersion {
	output, commandErr := commands.run(
		executable,
		viewCommand, descriptor.registryPackage+"@latest", "version", "--json",
	)
	if commandErr != nil {
		return globalCLIRegistryVersion{
			state:  globalCLIRegistryVersionUnavailable,
			detail: commandFailureDetail(commandErr, output),
		}
	}
	var rawVersion string
	if err := json.Unmarshal(output, &rawVersion); err != nil || strings.TrimSpace(rawVersion) == "" {
		detail := "registry latest response did not contain a JSON version string"
		if err != nil {
			detail = fmt.Sprintf("invalid registry latest JSON: %v", err)
		}
		return globalCLIRegistryVersion{state: globalCLIRegistryVersionMalformed, detail: detail}
	}
	version, err := parseSemanticVersion(rawVersion)
	if err != nil {
		return globalCLIRegistryVersion{state: globalCLIRegistryVersionMalformed, detail: err.Error()}
	}
	return globalCLIRegistryVersion{state: globalCLIRegistryVersionAvailable, version: version}
}

func commandFailureDetail(commandErr error, output []byte) string {
	detail := strings.TrimSpace(string(output))
	if detail == "" {
		return commandErr.Error()
	}
	return fmt.Sprintf("%v: %s", commandErr, detail)
}

func inspectGlobalCLIs(
	descriptors []globalCLIDescriptor,
	manager globalPackageManager,
	commands globalCLICommands,
) (globalCLIInspectionSnapshot, error) {
	snapshot := globalCLIInspectionSnapshot{
		manager:     manager,
		inspections: make([]globalCLIInspection, 0, len(descriptors)),
	}
	for _, descriptor := range descriptors {
		inspection, err := inspectGlobalCLI(descriptor, manager, commands)
		if err != nil {
			return globalCLIInspectionSnapshot{}, err
		}
		snapshot.inspections = append(snapshot.inspections, inspection)
	}
	return snapshot, nil
}

func inspectGlobalCLI(
	descriptor globalCLIDescriptor,
	manager globalPackageManager,
	commands globalCLICommands,
) (globalCLIInspection, error) {
	if manager.probePackageRegistration == nil || manager.probeLatestVersion == nil {
		return globalCLIInspection{}, fmt.Errorf(
			"%s package manager has no CLI inspection probes",
			manager.name,
		)
	}
	registration := manager.probePackageRegistration(descriptor, commands)
	if registration.state == globalCLIPackageRegistrationUnavailable ||
		registration.state == globalCLIPackageRegistrationMalformed {
		return globalCLIInspection{}, fmt.Errorf(
			"inspecting %s package registration for %s with %s: %s",
			descriptor.registryPackage,
			descriptor.displayName,
			manager.name,
			registration.detail,
		)
	}

	executablePath, executableErr := commands.lookPath(descriptor.executable)
	if executableErr != nil {
		if registration.state == globalCLIPackageRegistered {
			return globalCLIInspection{}, registeredWithoutExecutableError(descriptor, manager)
		}
		latest := manager.probeLatestVersion(descriptor, commands)
		if latest.state != globalCLIRegistryVersionAvailable {
			return globalCLIInspection{}, absentCLIRegistryError(descriptor, manager, latest)
		}
		version := latest.version
		return globalCLIInspection{
			descriptor:      descriptor,
			registration:    registration,
			registryVersion: &version,
			disposition:     globalCLIInstall,
		}, nil
	}

	installedVersion, err := inspectExecutableVersion(descriptor, manager, executablePath, commands)
	if err != nil {
		return globalCLIInspection{}, err
	}
	inspection := globalCLIInspection{
		descriptor:       descriptor,
		registration:     registration,
		executablePath:   executablePath,
		installedVersion: &installedVersion,
	}
	latest := manager.probeLatestVersion(descriptor, commands)
	if latest.state != globalCLIRegistryVersionAvailable {
		inspection.disposition = globalCLIRegistryUnverified
		inspection.warning = fmt.Sprintf(
			"%s %s is working at %s, but %s could not verify registry latest: %s",
			manager.name,
			descriptor.displayName,
			installedVersion,
			manager.name,
			latest.detail,
		)
		return inspection, nil
	}
	registryVersion := latest.version
	inspection.registryVersion = &registryVersion
	switch compareSemanticVersions(installedVersion, registryVersion) {
	case -1:
		inspection.disposition = globalCLIOutdated
	case 1:
		inspection.disposition = globalCLIAhead
	default:
		inspection.disposition = globalCLICurrent
	}
	return inspection, nil
}

func inspectExecutableVersion(
	descriptor globalCLIDescriptor,
	manager globalPackageManager,
	executablePath string,
	commands globalCLICommands,
) (semanticVersion, error) {
	output, commandErr := commands.run(executablePath, descriptor.versionArgs...)
	if commandErr != nil {
		return semanticVersion{}, executableVersionRecoveryError(
			descriptor,
			manager,
			fmt.Sprintf("version command failed: %s", commandFailureDetail(commandErr, output)),
		)
	}
	versionOutput := strings.TrimSpace(string(output))
	if versionOutput == "" {
		return semanticVersion{}, executableVersionRecoveryError(
			descriptor,
			manager,
			"version command returned no version",
		)
	}
	version, err := parseSemanticVersion(versionOutput)
	if err != nil {
		return semanticVersion{}, executableVersionRecoveryError(
			descriptor,
			manager,
			fmt.Sprintf("version command returned uninterpretable output %q", versionOutput),
		)
	}
	return version, nil
}

func registeredWithoutExecutableError(
	descriptor globalCLIDescriptor,
	manager globalPackageManager,
) error {
	return fmt.Errorf(
		"%s registers %s globally for %s, but %s is not available on PATH; repair the %s global registration or PATH linkage outside this installer, then rerun; no package cleanup was performed",
		manager.name,
		descriptor.registryPackage,
		descriptor.displayName,
		descriptor.executable,
		manager.name,
	)
}

func executableVersionRecoveryError(
	descriptor globalCLIDescriptor,
	manager globalPackageManager,
	problem string,
) error {
	return fmt.Errorf(
		"%s executable %s cannot be validated with %s: %s; repair the %s executable or PATH outside this installer, then rerun; no package cleanup was performed",
		descriptor.displayName,
		descriptor.executable,
		manager.name,
		problem,
		descriptor.displayName,
	)
}

func absentCLIRegistryError(
	descriptor globalCLIDescriptor,
	manager globalPackageManager,
	latest globalCLIRegistryVersion,
) error {
	return fmt.Errorf(
		"cannot install %s with %s because registry latest for %s is %s: %s; verify %s registry access and rerun; no package changes were performed",
		descriptor.displayName,
		manager.name,
		descriptor.registryPackage,
		latest.state,
		latest.detail,
		manager.name,
	)
}

func preflightGlobalCLIs(
	descriptors []globalCLIDescriptor,
	commands globalCLICommands,
) (globalCLIInspectionSnapshot, error) {
	if len(descriptors) == 0 {
		return globalCLIInspectionSnapshot{}, nil
	}
	manager, err := selectGlobalPackageManager(commands)
	if err != nil {
		return globalCLIInspectionSnapshot{}, err
	}
	if err := validateGlobalCLIRuntimes(descriptors, commands); err != nil {
		return globalCLIInspectionSnapshot{}, err
	}
	return inspectGlobalCLIs(descriptors, manager, commands)
}

func installGlobalCLI(
	descriptor globalCLIDescriptor,
	manager globalPackageManager,
	commands globalCLICommands,
) (string, error) {
	if _, err := installGlobalCLIExecutable(descriptor, manager, commands); err != nil {
		return "", err
	}
	return "instalado  " + descriptor.installSpec, nil
}

func applyGlobalCLIInspection(
	inspection globalCLIInspection,
	manager globalPackageManager,
	commands globalCLICommands,
) (string, error) {
	switch inspection.disposition {
	case globalCLICurrent, globalCLIOutdated, globalCLIAhead, globalCLIRegistryUnverified:
		return inspection.reportLine(), nil
	case globalCLIInstall:
		if inspection.registration.state != globalCLIPackageUnregistered ||
			inspection.executablePath != "" || inspection.registryVersion == nil {
			return "", fmt.Errorf(
				"refusing unvalidated %s installation state for %s",
				inspection.disposition,
				inspection.descriptor.displayName,
			)
		}
	default:
		return "", fmt.Errorf(
			"unsupported CLI inspection disposition %q for %s",
			inspection.disposition,
			inspection.descriptor.displayName,
		)
	}

	executablePath, err := installGlobalCLIExecutable(inspection.descriptor, manager, commands)
	if err != nil {
		return "", err
	}
	version, err := inspectExecutableVersion(inspection.descriptor, manager, executablePath, commands)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("instalado  %s (version %s)", inspection.descriptor.installSpec, version), nil
}

func installGlobalCLIExecutable(
	descriptor globalCLIDescriptor,
	manager globalPackageManager,
	commands globalCLICommands,
) (string, error) {
	args := append(append([]string{}, manager.install...), descriptor.installSpec)
	output, err := commands.run(manager.executable, args...)
	if err != nil {
		detail := strings.TrimSpace(string(output))
		if detail == "" {
			return "", fmt.Errorf("installing %s with %s: %w", descriptor.displayName, manager.name, err)
		}
		return "", fmt.Errorf("installing %s with %s: %w: %s", descriptor.displayName, manager.name, err, detail)
	}
	executablePath, err := commands.lookPath(descriptor.executable)
	if err != nil {
		return "", fmt.Errorf(
			"%s package command succeeded but %s is not available on PATH",
			descriptor.displayName,
			descriptor.executable,
		)
	}
	return executablePath, nil
}
