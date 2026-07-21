package install

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const (
	openSpecPackage            = "@fission-ai/openspec@latest"
	openSpecMinimumNodeVersion = "20.19.0"
)

var semanticVersionPattern = regexp.MustCompile(
	`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$`,
)

type globalCLIDescriptor struct {
	optionKey          string
	displayName        string
	executable         string
	packageSpec        string
	minimumNodeVersion string
}

var openSpecGlobalCLI = globalCLIDescriptor{
	optionKey:          openSpecOptionKey,
	displayName:        "OpenSpec",
	executable:         "openspec",
	packageSpec:        openSpecPackage,
	minimumNodeVersion: openSpecMinimumNodeVersion,
}

var globalCLIDescriptors = []globalCLIDescriptor{
	{
		optionKey:   codegraphOptionKey,
		displayName: "CodeGraph",
		executable:  "codegraph",
		packageSpec: codegraphPackage,
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
	name       string
	executable string
	globalBin  string
	install    []string
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
	}, nil
}

func preflightGlobalCLIs(
	descriptors []globalCLIDescriptor,
	commands globalCLICommands,
) (globalPackageManager, error) {
	if len(descriptors) == 0 {
		return globalPackageManager{}, nil
	}
	manager, err := selectGlobalPackageManager(commands)
	if err != nil {
		return globalPackageManager{}, err
	}
	if err := validateGlobalCLIRuntimes(descriptors, commands); err != nil {
		return globalPackageManager{}, err
	}
	return manager, nil
}

func installGlobalCLI(
	descriptor globalCLIDescriptor,
	manager globalPackageManager,
	commands globalCLICommands,
) (string, error) {
	args := append(append([]string{}, manager.install...), descriptor.packageSpec)
	output, err := commands.run(manager.executable, args...)
	if err != nil {
		detail := strings.TrimSpace(string(output))
		if detail == "" {
			return "", fmt.Errorf("installing %s with %s: %w", descriptor.displayName, manager.name, err)
		}
		return "", fmt.Errorf("installing %s with %s: %w: %s", descriptor.displayName, manager.name, err, detail)
	}
	if _, err := commands.lookPath(descriptor.executable); err != nil {
		return "", fmt.Errorf(
			"%s package command succeeded but %s is not available on PATH",
			descriptor.displayName,
			descriptor.executable,
		)
	}
	return "instalado  " + descriptor.packageSpec, nil
}
