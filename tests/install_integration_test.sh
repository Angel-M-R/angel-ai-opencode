#!/bin/sh

set -u

test_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
installer="$test_dir/../install.sh"
official_manifest_url="https://github.com/Angel-M-R/angel-ai-opencode/releases/latest/download/manifest.json"
artifact_url="https://downloads.example.test/angel-ai"
passed=0
failed=0

fail_test() {
	printf '    %s\n' "$*" >&2
	return 1
}

assert_equal() {
	actual=$1
	expected=$2
	message=$3
	[ "$actual" = "$expected" ] || fail_test "$message: got '$actual', want '$expected'"
}

assert_contains() {
	haystack=$1
	needle=$2
	message=$3
	case "$haystack" in
		*"$needle"*) return 0 ;;
		*) fail_test "$message: missing '$needle'" ;;
	esac
}

assert_not_contains() {
	haystack=$1
	needle=$2
	message=$3
	case "$haystack" in
		*"$needle"*) fail_test "$message: unexpectedly contained '$needle'" ;;
		*) return 0 ;;
	esac
}

setup_case() {
	case_root=$(mktemp -d "${TMPDIR:-/tmp}/angel-ai-installer-test.XXXXXX") || return 1
	case_home="$case_root/home"
	mock_bin="$case_root/mock-bin"
	case_tmp="$case_root/tmp"
	manifest_file="$case_root/manifest.json"
	artifact_file="$case_root/artifact"
	curl_log="$case_root/curl.log"
	mkdir -p "$case_home" "$mock_bin" "$case_tmp" || return 1
	printf 'new angel-ai binary\n' >"$artifact_file" || return 1

	cat >"$mock_bin/uname" <<'EOF'
#!/bin/sh
case "${1-}" in
	-s) printf '%s\n' "${TEST_UNAME_S:-Darwin}" ;;
	-m) printf '%s\n' "${TEST_UNAME_M:-arm64}" ;;
	*) exit 2 ;;
esac
EOF
	cat >"$mock_bin/curl" <<EOF
#!/bin/sh
output=""
url=""
while [ "\$#" -gt 0 ]; do
	case "\$1" in
		--output)
			output=\$2
			shift 2
			;;
		--*) shift ;;
		*) url=\$1; shift ;;
	esac
done
printf '%s\n' "\$url" >>"\$TEST_CURL_LOG"
if [ "\$url" = "$official_manifest_url" ]; then
	[ "\${TEST_MANIFEST_MODE:-ok}" = "ok" ] || exit 22
	cp "\$TEST_MANIFEST_FILE" "\$output"
	exit \$?
fi
if [ "\$url" = "$artifact_url" ]; then
	[ "\${TEST_ARTIFACT_MODE:-ok}" = "ok" ] || exit 22
	cp "\$TEST_ARTIFACT_FILE" "\$output"
	exit \$?
fi
exit 22
EOF
	chmod 0755 "$mock_bin/uname" "$mock_bin/curl" || return 1
	test_uname_s=Darwin
	test_uname_m=arm64
	test_manifest_mode=ok
	test_artifact_mode=ok
	write_valid_manifest
}

cleanup_case() {
	rm -rf "$case_root"
}

artifact_checksum() {
	digest=$(shasum -a 256 "$artifact_file") || return 1
	printf '%s\n' "${digest%% *}"
}

write_valid_manifest() {
	checksum=$(artifact_checksum) || return 1
	printf '{"version":"v1.2.3","artifact_url":"%s","sha256":"%s"}\n' "$artifact_url" "$checksum" >"$manifest_file"
}

run_installer() {
	installer_path=$1
	if installer_output=$(HOME="$case_home" PATH="$installer_path" TMPDIR="$case_tmp" TEST_UNAME_S="$test_uname_s" TEST_UNAME_M="$test_uname_m" TEST_MANIFEST_MODE="$test_manifest_mode" TEST_ARTIFACT_MODE="$test_artifact_mode" TEST_MANIFEST_FILE="$manifest_file" TEST_ARTIFACT_FILE="$artifact_file" TEST_CURL_LOG="$curl_log" /bin/sh "$installer" 2>&1); then
		installer_status=0
	else
		installer_status=$?
	fi
}

run_installer_from_stdin() {
	installer_path=$1
	if installer_output=$(HOME="$case_home" PATH="$installer_path" TMPDIR="$case_tmp" TEST_UNAME_S="$test_uname_s" TEST_UNAME_M="$test_uname_m" TEST_MANIFEST_MODE="$test_manifest_mode" TEST_ARTIFACT_MODE="$test_artifact_mode" TEST_MANIFEST_FILE="$manifest_file" TEST_ARTIFACT_FILE="$artifact_file" TEST_CURL_LOG="$curl_log" /bin/sh <"$installer" 2>&1); then
		installer_status=0
	else
		installer_status=$?
	fi
}

write_existing_binary() {
	mkdir -p "$case_home/.local/bin" || return 1
	printf 'old binary\n' >"$case_home/.local/bin/angel-ai" || return 1
	chmod 0755 "$case_home/.local/bin/angel-ai"
}

assert_installed_bytes() {
	installed="$case_home/.local/bin/angel-ai"
	[ -x "$installed" ] || fail_test "installed binary is not executable"
	cmp -s "$installed" "$artifact_file" || fail_test "installed binary does not match downloaded artifact"
}

assert_existing_bytes() {
	expected=$1
	installed="$case_home/.local/bin/angel-ai"
	[ -x "$installed" ] || fail_test "existing binary is no longer executable"
	actual=$(cat "$installed") || return 1
	assert_equal "$actual" "$expected" "existing binary changed"
}

assert_no_artifact_temps() {
	temp_count=$(find "$case_home/.local/bin" -name '.angel-ai.*' -type f 2>/dev/null | wc -l | tr -d ' ')
	assert_equal "$temp_count" "0" "artifact temporary files were not cleaned"
}

test_unsupported_platforms_reject_before_download() (
	setup_case || exit 1
	trap cleanup_case 0
	for platform in 'Darwin x86_64' 'Linux arm64' 'MINGW64_NT arm64' 'Darwin riscv64'; do
		test_uname_s=${platform% *}
		test_uname_m=${platform#* }
		rm -f "$curl_log"
		run_installer "$mock_bin:/usr/bin:/bin"
		[ "$installer_status" -ne 0 ] || fail_test "$platform unexpectedly succeeded" || exit 1
		assert_contains "$installer_output" "unsupported platform $test_uname_s/$test_uname_m" "$platform rejection was not actionable" || exit 1
		[ ! -e "$case_home/.local/bin/angel-ai" ] || fail_test "$platform installed a binary" || exit 1
		[ ! -s "$curl_log" ] || fail_test "$platform reached the network before rejection" || exit 1
	done
)

test_first_install_on_supported_platform() (
	setup_case || exit 1
	trap cleanup_case 0
	run_installer_from_stdin "$mock_bin:/usr/bin:/bin"
	assert_equal "$installer_status" "0" "supported first install failed: $installer_output" || exit 1
	assert_installed_bytes || exit 1
	assert_contains "$installer_output" "Installed angel-ai v1.2.3" "success output omitted the installed version" || exit 1
	assert_no_artifact_temps || exit 1
)

test_existing_binary_is_atomically_replaced() (
	setup_case || exit 1
	trap cleanup_case 0
	write_existing_binary || exit 1
	run_installer "$mock_bin:/usr/bin:/bin"
	assert_equal "$installer_status" "0" "replacement failed: $installer_output" || exit 1
	assert_installed_bytes || exit 1
	assert_no_artifact_temps || exit 1
)

test_manifest_download_failure_preserves_existing_binary() (
	setup_case || exit 1
	trap cleanup_case 0
	write_existing_binary || exit 1
	test_manifest_mode=fail
	run_installer "$mock_bin:/usr/bin:/bin"
	[ "$installer_status" -ne 0 ] || fail_test "manifest download failure unexpectedly succeeded" || exit 1
	assert_contains "$installer_output" "unable to download the latest-release manifest" "manifest download error was not reported" || exit 1
	assert_existing_bytes "old binary" || exit 1
)

test_invalid_manifests_preserve_existing_binary() (
	setup_case || exit 1
	trap cleanup_case 0
	write_existing_binary || exit 1
	checksum=$(artifact_checksum) || exit 1

	invalid_manifests=$(cat <<EOF
not-json
{"version":"1.2.3","artifact_url":"$artifact_url","sha256":"$checksum"}
{"version":"v1.2.3","artifact_url":"http://downloads.example.test/angel-ai","sha256":"$checksum"}
{"version":"v1.2.3","artifact_url":"$artifact_url","sha256":"ABCDEF"}
{"version":"v1.2.3","artifact_url":"$artifact_url","sha256":"$checksum","extra":"value"}
EOF
)
	old_ifs=$IFS
	IFS='
'
	for invalid_manifest in $invalid_manifests; do
		printf '%s\n' "$invalid_manifest" >"$manifest_file"
		run_installer "$mock_bin:/usr/bin:/bin"
		[ "$installer_status" -ne 0 ] || fail_test "invalid manifest unexpectedly succeeded: $invalid_manifest" || exit 1
		assert_existing_bytes "old binary" || exit 1
	done
	IFS=$old_ifs
)

test_artifact_download_failure_preserves_existing_binary() (
	setup_case || exit 1
	trap cleanup_case 0
	write_existing_binary || exit 1
	test_artifact_mode=fail
	run_installer "$mock_bin:/usr/bin:/bin"
	[ "$installer_status" -ne 0 ] || fail_test "artifact download failure unexpectedly succeeded" || exit 1
	assert_contains "$installer_output" "unable to download angel-ai v1.2.3" "artifact download error was not reported" || exit 1
	assert_existing_bytes "old binary" || exit 1
	assert_no_artifact_temps || exit 1
)

test_checksum_failure_preserves_existing_binary() (
	setup_case || exit 1
	trap cleanup_case 0
	write_existing_binary || exit 1
	printf '{"version":"v1.2.3","artifact_url":"%s","sha256":"%064d"}\n' "$artifact_url" 0 >"$manifest_file"
	run_installer "$mock_bin:/usr/bin:/bin"
	[ "$installer_status" -ne 0 ] || fail_test "checksum mismatch unexpectedly succeeded" || exit 1
	assert_contains "$installer_output" "checksum verification failed" "checksum mismatch was not reported" || exit 1
	assert_existing_bytes "old binary" || exit 1
	assert_no_artifact_temps || exit 1
)

test_path_present_reports_ready_without_guidance() (
	setup_case || exit 1
	trap cleanup_case 0
	run_installer "$case_home/.local/bin:$mock_bin:/usr/bin:/bin"
	assert_equal "$installer_status" "0" "PATH-present install failed: $installer_output" || exit 1
	assert_contains "$installer_output" "angel-ai is ready to use." "PATH-present output omitted ready status" || exit 1
	assert_not_contains "$installer_output" 'export PATH="$HOME/.local/bin:$PATH"' "PATH-present output included unnecessary export guidance" || exit 1
)

test_path_missing_prints_exact_guidance_without_profile_writes() (
	setup_case || exit 1
	trap cleanup_case 0
	printf 'zsh sentinel\n' >"$case_home/.zshrc"
	printf 'bash sentinel\n' >"$case_home/.bash_profile"
	printf 'profile sentinel\n' >"$case_home/.profile"
	run_installer "$mock_bin:/usr/bin:/bin"
	assert_equal "$installer_status" "0" "PATH-missing install failed: $installer_output" || exit 1
	export_count=$(printf '%s\n' "$installer_output" | grep -Fxc 'export PATH="$HOME/.local/bin:$PATH"')
	assert_equal "$export_count" "1" "PATH-missing output did not contain the exact export command once" || exit 1
	assert_contains "$installer_output" "manually" "PATH-missing output omitted manual guidance" || exit 1
	assert_equal "$(cat "$case_home/.zshrc")" "zsh sentinel" ".zshrc was modified" || exit 1
	assert_equal "$(cat "$case_home/.bash_profile")" "bash sentinel" ".bash_profile was modified" || exit 1
	assert_equal "$(cat "$case_home/.profile")" "profile sentinel" ".profile was modified" || exit 1
)

run_test() {
	test_name=$1
	shift
	if "$@"; then
		printf 'ok - %s\n' "$test_name"
		passed=$((passed + 1))
	else
		printf 'not ok - %s\n' "$test_name"
		failed=$((failed + 1))
	fi
}

run_test "unsupported hosts reject before download" test_unsupported_platforms_reject_before_download
run_test "supported first install" test_first_install_on_supported_platform
run_test "existing binary replacement" test_existing_binary_is_atomically_replaced
run_test "manifest download failure preservation" test_manifest_download_failure_preserves_existing_binary
run_test "invalid manifest preservation" test_invalid_manifests_preserve_existing_binary
run_test "artifact download failure preservation" test_artifact_download_failure_preserves_existing_binary
run_test "checksum failure preservation" test_checksum_failure_preserves_existing_binary
run_test "PATH-present output" test_path_present_reports_ready_without_guidance
run_test "PATH-missing output and profile preservation" test_path_missing_prints_exact_guidance_without_profile_writes

printf '%s passed; %s failed\n' "$passed" "$failed"
[ "$failed" -eq 0 ]
