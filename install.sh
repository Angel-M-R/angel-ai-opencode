#!/bin/sh

set -eu

manifest_base_url="https://github.com/Angel-M-R/angel-ai-opencode/releases/latest/download"
supported_platforms="Darwin/arm64, Linux/amd64, and Linux/arm64"
install_dir="$HOME/.local/bin"
install_path="$install_dir/angel-ai"
manifest_temp=""
artifact_temp=""

fail() {
	printf 'error: %s\n' "$*" >&2
	exit 1
}

cleanup() {
	if [ -n "$artifact_temp" ]; then
		rm -f -- "$artifact_temp"
	fi
	if [ -n "$manifest_temp" ]; then
		rm -f -- "$manifest_temp"
	fi
}

trap cleanup 0 HUP INT TERM

compute_artifact_checksum() {
	case "$checksum_tool" in
		shasum)
			shasum -a 256 "$1"
			;;
		*)
			sha256sum "$1"
			;;
	esac
}

read_manifest_fields() {
	case "$json_tool" in
		plutil)
			if ! manifest_xml=$(plutil -convert xml1 -o - "$manifest_temp" 2>/dev/null); then
				return 1
			fi
			if [ "$(printf '%s\n' "$manifest_xml" | grep -c '<key>')" != 3 ]; then
				return 2
			fi
			version=$(plutil -extract version raw -o - "$manifest_temp" 2>/dev/null) || return 3
			artifact_url=$(plutil -extract artifact_url raw -o - "$manifest_temp" 2>/dev/null) || return 4
			expected_sha256=$(plutil -extract sha256 raw -o - "$manifest_temp" 2>/dev/null) || return 5
			;;
		python3)
			fields=""
			field_status=0
			fields=$(python3 -c '
import json, sys

try:
    with open(sys.argv[1], encoding="utf-8") as handle:
        document = json.load(handle)
except Exception:
    sys.exit(1)
if not isinstance(document, dict):
    sys.exit(1)
expected = ("version", "artifact_url", "sha256")
if sorted(document) != sorted(expected):
    sys.exit(2)
values = [document[key] for key in expected]
for value in values:
    if not isinstance(value, str):
        sys.exit(3)
sys.stdout.write("".join(value + "\n" for value in values))
' "$manifest_temp") || field_status=$?
			case "$field_status" in
				0) ;;
				1) return 1 ;;
				*) return 2 ;;
			esac
			if [ "$(printf '%s\n' "$fields" | grep -c '')" != 3 ]; then
				return 3
			fi
			version=$(printf '%s\n' "$fields" | sed -n '1p')
			artifact_url=$(printf '%s\n' "$fields" | sed -n '2p')
			expected_sha256=$(printf '%s\n' "$fields" | sed -n '3p')
			;;
		*)
			return 1
			;;
	esac
}

operating_system=$(uname -s 2>/dev/null) || fail "unable to detect the operating system; angel-ai supports only $supported_platforms"
architecture=$(uname -m 2>/dev/null) || fail "unable to detect the host architecture; angel-ai supports only $supported_platforms"
case "$operating_system/$architecture" in
	Darwin/arm64) platform=darwin-arm64 ;;
	Linux/x86_64|Linux/amd64) platform=linux-amd64 ;;
	Linux/arm64|Linux/aarch64) platform=linux-arm64 ;;
	*) fail "unsupported platform $operating_system/$architecture; angel-ai supports only $supported_platforms. No files were installed." ;;
esac
manifest_url="$manifest_base_url/manifest-$platform.json"

for command_name in curl grep mktemp sed mkdir chmod mv rm; do
	command -v "$command_name" >/dev/null 2>&1 || fail "required command '$command_name' was not found; install it and run the installer again"
done

if command -v shasum >/dev/null 2>&1; then
	checksum_tool=shasum
elif command -v sha256sum >/dev/null 2>&1; then
	checksum_tool=sha256sum
else
	fail "required command 'shasum' or 'sha256sum' was not found; install one of them and run the installer again"
fi

if command -v plutil >/dev/null 2>&1; then
	json_tool=plutil
elif command -v python3 >/dev/null 2>&1; then
	json_tool=python3
else
	fail "required command 'plutil' or 'python3' was not found; install one of them and run the installer again"
fi

manifest_temp=$(mktemp "${TMPDIR:-/tmp}/angel-ai-manifest.XXXXXX") || fail "unable to create a temporary manifest file"
if ! curl --fail --location --silent --show-error --proto '=https' --proto-redir '=https' --output "$manifest_temp" "$manifest_url"; then
	fail "unable to download the latest-release manifest from $manifest_url"
fi

read_status=0
read_manifest_fields || read_status=$?
case "$read_status" in
	0) ;;
	1) fail "latest-release manifest is not valid JSON" ;;
	2) fail "latest-release manifest must contain exactly version, artifact_url, and sha256" ;;
	*) fail "latest-release manifest is missing a valid version, artifact_url, or sha256" ;;
esac

if ! printf '%s\n' "$version" | LC_ALL=C grep -Eq '^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$'; then
	fail "latest-release manifest version must be stable vMAJOR.MINOR.PATCH"
fi
case "$artifact_url" in
	https://*) ;;
	*) fail "latest-release manifest artifact_url must be an absolute HTTPS URL without credentials" ;;
esac
artifact_authority=${artifact_url#https://}
artifact_authority=${artifact_authority%%/*}
artifact_authority=${artifact_authority%%\?*}
artifact_authority=${artifact_authority%%\#*}
if [ -z "$artifact_authority" ]; then
	fail "latest-release manifest artifact_url must be an absolute HTTPS URL without credentials"
fi
case "$artifact_authority" in
	*@*) fail "latest-release manifest artifact_url must be an absolute HTTPS URL without credentials" ;;
esac
if ! printf '%s\n' "$artifact_url" | LC_ALL=C grep -Eq '^[!-~]+$'; then
	fail "latest-release manifest artifact_url must be an absolute HTTPS URL without credentials"
fi
if ! printf '%s\n' "$expected_sha256" | LC_ALL=C grep -Eq '^[0-9a-f]{64}$'; then
	fail "latest-release manifest sha256 must be 64 lowercase hexadecimal characters"
fi

mkdir -p "$install_dir" || fail "unable to create installation directory $install_dir"
artifact_temp=$(mktemp "$install_dir/.angel-ai.XXXXXX") || fail "unable to create a temporary artifact in $install_dir"
if ! curl --fail --location --silent --show-error --proto '=https' --proto-redir '=https' --output "$artifact_temp" "$artifact_url"; then
	fail "unable to download angel-ai $version from $artifact_url"
fi

checksum_output=$(compute_artifact_checksum "$artifact_temp") || fail "unable to compute the downloaded artifact checksum"
actual_sha256=${checksum_output%% *}
if [ "$actual_sha256" != "$expected_sha256" ]; then
	fail "downloaded artifact checksum verification failed"
fi
chmod 0755 "$artifact_temp" || fail "unable to make the verified artifact executable"
mv -f "$artifact_temp" "$install_path" || fail "unable to atomically install angel-ai at $install_path; any existing installation was left unchanged"
artifact_temp=""

printf 'Installed angel-ai %s at %s\n' "$version" "$install_path"
case ":${PATH-}:" in
	*":$install_dir:"*|*":$install_dir/:"*)
		printf 'angel-ai is ready to use.\n'
		;;
	*)
		printf 'Add angel-ai to PATH manually by running this command:\n'
		printf '%s\n' 'export PATH="$HOME/.local/bin:$PATH"'
		printf 'Add that command to your shell profile yourself if you want it applied to future sessions.\n'
		;;
esac
