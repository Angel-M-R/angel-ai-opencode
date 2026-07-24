#!/bin/sh

set -eu

manifest_url="https://github.com/Angel-M-R/angel-ai-opencode/releases/latest/download/manifest.json"
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

operating_system=$(uname -s 2>/dev/null) || fail "unable to detect the operating system; angel-ai supports only macOS Apple Silicon (Darwin/arm64)"
architecture=$(uname -m 2>/dev/null) || fail "unable to detect the host architecture; angel-ai supports only macOS Apple Silicon (Darwin/arm64)"
if [ "$operating_system" != "Darwin" ] || [ "$architecture" != "arm64" ]; then
	fail "unsupported platform $operating_system/$architecture; angel-ai supports only macOS Apple Silicon (Darwin/arm64). No files were installed."
fi

for command_name in curl plutil grep mktemp shasum mkdir chmod mv rm; do
	command -v "$command_name" >/dev/null 2>&1 || fail "required command '$command_name' was not found; install it and run the installer again"
done

manifest_temp=$(mktemp "${TMPDIR:-/tmp}/angel-ai-manifest.XXXXXX") || fail "unable to create a temporary manifest file"
if ! curl --fail --location --silent --show-error --proto '=https' --proto-redir '=https' --output "$manifest_temp" "$manifest_url"; then
	fail "unable to download the latest-release manifest from $manifest_url"
fi

if ! manifest_xml=$(plutil -convert xml1 -o - "$manifest_temp" 2>/dev/null); then
	fail "latest-release manifest is not valid JSON"
fi
manifest_key_count=$(printf '%s\n' "$manifest_xml" | grep -c '<key>')
if [ "$manifest_key_count" -ne 3 ]; then
	fail "latest-release manifest must contain exactly version, artifact_url, and sha256"
fi
if ! version=$(plutil -extract version raw -o - "$manifest_temp" 2>/dev/null); then
	fail "latest-release manifest is missing a valid version"
fi
if ! artifact_url=$(plutil -extract artifact_url raw -o - "$manifest_temp" 2>/dev/null); then
	fail "latest-release manifest is missing a valid artifact_url"
fi
if ! expected_sha256=$(plutil -extract sha256 raw -o - "$manifest_temp" 2>/dev/null); then
	fail "latest-release manifest is missing a valid sha256"
fi

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

actual_sha256=$(shasum -a 256 "$artifact_temp") || fail "unable to compute the downloaded artifact checksum"
actual_sha256=${actual_sha256%% *}
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
