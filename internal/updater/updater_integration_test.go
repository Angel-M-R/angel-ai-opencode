package updater

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type integrationFileSystem struct {
	FileSystem
	executable      string
	replaceThenFail bool
	replacementSeen bool
}

func (fileSystem *integrationFileSystem) Executable() (string, error) {
	return fileSystem.executable, nil
}

func (fileSystem *integrationFileSystem) Rename(oldPath, newPath string) error {
	if fileSystem.replaceThenFail && !fileSystem.replacementSeen && newPath == fileSystem.executable && strings.Contains(filepath.Base(oldPath), ".update-") {
		fileSystem.replacementSeen = true
		if err := fileSystem.FileSystem.Rename(oldPath, newPath); err != nil {
			return err
		}
		return errors.New("injected replacement failure")
	}
	return fileSystem.FileSystem.Rename(oldPath, newPath)
}

type recordingProcess struct {
	args        []string
	environment []string
	execErr     error
	execCalls   int
	execPath    string
	execArgs    []string
	execEnv     []string
}

func (process *recordingProcess) Args() []string {
	return append([]string(nil), process.args...)
}

func (process *recordingProcess) Environ() []string {
	return append([]string(nil), process.environment...)
}

func (process *recordingProcess) Exec(path string, arguments, environment []string) error {
	process.execCalls++
	process.execPath = path
	process.execArgs = append([]string(nil), arguments...)
	process.execEnv = append([]string(nil), environment...)
	return process.execErr
}

type countingHTTPClient struct {
	requests int
}

func (client *countingHTTPClient) Do(*http.Request) (*http.Response, error) {
	client.requests++
	return nil, errors.New("unexpected HTTP request")
}

func TestUpdaterRejectsChecksumMismatchAndCleansTemporaryState(t *testing.T) {
	oldBinary := []byte("old executable")
	newBinary := []byte("new executable")
	executable := temporaryExecutable(t, oldBinary)
	wrongDigest := sha256.Sum256([]byte("different artifact"))
	server, client, manifestURL := updateFixture(t, newBinary, hex.EncodeToString(wrongDigest[:]))
	defer server.Close()

	process := &recordingProcess{args: []string{"angel-ai"}}
	var output bytes.Buffer
	updater := New(Config{
		HTTP:        client,
		FileSystem:  &integrationFileSystem{FileSystem: osFileSystem{}, executable: executable},
		Process:     process,
		Output:      &output,
		ManifestURL: manifestURL,
	})
	if err := updater.Run("v1.0.0", false); err != nil {
		t.Fatal(err)
	}

	assertExecutableBytes(t, executable, oldBinary)
	assertOnlyExecutableRemains(t, executable)
	if process.execCalls != 0 {
		t.Fatalf("exec calls = %d, want 0", process.execCalls)
	}
	if !strings.Contains(output.String(), "warning:") || !strings.Contains(output.String(), "SHA-256 mismatch") {
		t.Fatalf("warning output = %q", output.String())
	}
}

func TestUpdaterAtomicallyReplacesAndRelaunchesWithOriginalProcessState(t *testing.T) {
	oldBinary := []byte("old executable")
	newBinary := []byte("new executable")
	executable := temporaryExecutable(t, oldBinary)
	server, client, manifestURL := updateFixture(t, newBinary, "")
	defer server.Close()

	arguments := []string{"angel-ai", "update", "--future-compatible-argument"}
	process := &recordingProcess{
		args:        arguments,
		environment: []string{"HOME=/tmp/home", relaunchMarker + "=stale"},
	}
	var output bytes.Buffer
	fileSystem := &integrationFileSystem{FileSystem: osFileSystem{}, executable: executable}
	updater := New(Config{
		HTTP:        client,
		FileSystem:  fileSystem,
		Process:     process,
		Output:      &output,
		ManifestURL: manifestURL,
	})
	if err := updater.Run("v1.0.0", true); err != nil {
		t.Fatal(err)
	}

	assertExecutableBytes(t, executable, newBinary)
	info, err := os.Stat(executable)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o111 == 0 {
		t.Fatalf("replacement permissions = %o, want executable", info.Mode().Perm())
	}
	if process.execCalls != 1 || process.execPath != executable {
		t.Fatalf("exec calls/path = %d %q", process.execCalls, process.execPath)
	}
	if !reflect.DeepEqual(process.execArgs, arguments) {
		t.Fatalf("exec arguments = %#v, want %#v", process.execArgs, arguments)
	}
	if !reflect.DeepEqual(process.execEnv, []string{"HOME=/tmp/home", relaunchMarker + "=1"}) {
		t.Fatalf("exec environment = %#v", process.execEnv)
	}
	if output.Len() != 0 {
		t.Fatalf("pre-relaunch output = %q", output.String())
	}

	backupPath := updater.backupPath(executable)
	assertExecutableBytes(t, backupPath, oldBinary)
	noNetwork := &countingHTTPClient{}
	relaunchOutput := &bytes.Buffer{}
	relaunched := New(Config{
		HTTP:       noNetwork,
		FileSystem: fileSystem,
		Process: &recordingProcess{
			args:        arguments,
			environment: process.execEnv,
		},
		Output: relaunchOutput,
	})
	if err := relaunched.Run("v1.1.0", true); err != nil {
		t.Fatal(err)
	}
	if noNetwork.requests != 0 {
		t.Fatalf("relaunch network requests = %d, want 0", noNetwork.requests)
	}
	if relaunchOutput.String() != "angel-ai updated to v1.1.0\n" {
		t.Fatalf("relaunch output = %q", relaunchOutput.String())
	}
	if _, err := os.Stat(backupPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("backup remains after relaunch cleanup: %v", err)
	}
	assertExecutableBytes(t, executable, newBinary)
}

func TestUpdaterRelaunchMarkerCompletesAutomaticTUIFlowWithoutLooping(t *testing.T) {
	executable := temporaryExecutable(t, []byte("new executable"))
	fileSystem := &integrationFileSystem{FileSystem: osFileSystem{}, executable: executable}
	updater := New(Config{FileSystem: fileSystem})
	backupPath := updater.backupPath(executable)
	if err := os.WriteFile(backupPath, []byte("old executable"), 0o751); err != nil {
		t.Fatal(err)
	}

	noNetwork := &countingHTTPClient{}
	var output bytes.Buffer
	relaunched := New(Config{
		HTTP:       noNetwork,
		FileSystem: fileSystem,
		Process: &recordingProcess{
			args:        []string{"angel-ai", "--target", "/tmp/config"},
			environment: []string{relaunchMarker + "=1"},
		},
		Output: &output,
	})
	if err := relaunched.Run("v1.1.0", false); err != nil {
		t.Fatal(err)
	}
	if noNetwork.requests != 0 {
		t.Fatalf("relaunch network requests = %d, want 0", noNetwork.requests)
	}
	if output.Len() != 0 {
		t.Fatalf("automatic relaunch output = %q", output.String())
	}
	if _, err := os.Stat(backupPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("backup remains after automatic relaunch cleanup: %v", err)
	}
}

func TestUpdaterReplacementFailureRestoresCurrentExecutable(t *testing.T) {
	oldBinary := []byte("old executable")
	newBinary := []byte("new executable")
	executable := temporaryExecutable(t, oldBinary)
	server, client, manifestURL := updateFixture(t, newBinary, "")
	defer server.Close()

	fileSystem := &integrationFileSystem{
		FileSystem:      osFileSystem{},
		executable:      executable,
		replaceThenFail: true,
	}
	process := &recordingProcess{args: []string{"angel-ai"}}
	var output bytes.Buffer
	updater := New(Config{
		HTTP:        client,
		FileSystem:  fileSystem,
		Process:     process,
		Output:      &output,
		ManifestURL: manifestURL,
	})
	if err := updater.Run("v1.0.0", false); err != nil {
		t.Fatal(err)
	}

	assertExecutableBytes(t, executable, oldBinary)
	assertOnlyExecutableRemains(t, executable)
	if process.execCalls != 0 {
		t.Fatalf("exec calls = %d, want 0", process.execCalls)
	}
	if !strings.Contains(output.String(), "warning:") || !strings.Contains(output.String(), "replacing executable") {
		t.Fatalf("warning output = %q", output.String())
	}
}

func TestUpdaterRelaunchFailureRollsBackImmediately(t *testing.T) {
	oldBinary := []byte("old executable")
	newBinary := []byte("new executable")
	executable := temporaryExecutable(t, oldBinary)
	server, client, manifestURL := updateFixture(t, newBinary, "")
	defer server.Close()

	relaunchError := errors.New("exec format error")
	process := &recordingProcess{args: []string{"angel-ai", "--target", "/tmp/config"}, execErr: relaunchError}
	var output bytes.Buffer
	updater := New(Config{
		HTTP:        client,
		FileSystem:  &integrationFileSystem{FileSystem: osFileSystem{}, executable: executable},
		Process:     process,
		Output:      &output,
		ManifestURL: manifestURL,
	})
	if err := updater.Run("v1.0.0", false); err != nil {
		t.Fatal(err)
	}

	assertExecutableBytes(t, executable, oldBinary)
	assertOnlyExecutableRemains(t, executable)
	if process.execCalls != 1 {
		t.Fatalf("exec calls = %d, want 1", process.execCalls)
	}
	if !strings.Contains(output.String(), "warning:") || !strings.Contains(output.String(), "relaunching replacement") {
		t.Fatalf("warning output = %q", output.String())
	}
}

func updateFixture(t *testing.T, artifact []byte, manifestDigest string) (*httptest.Server, *http.Client, string) {
	t.Helper()
	if manifestDigest == "" {
		digest := sha256.Sum256(artifact)
		manifestDigest = hex.EncodeToString(digest[:])
	}
	serverURL := ""
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/manifest.json":
			_, _ = fmt.Fprintf(response, `{"version":"v1.1.0","artifact_url":%q,"sha256":%q}`, serverURL+"/angel-ai", manifestDigest)
		case "/angel-ai":
			_, _ = response.Write(artifact)
		default:
			http.NotFound(response, request)
		}
	}))
	serverURL = server.URL
	return server, server.Client(), server.URL + "/manifest.json"
}

func temporaryExecutable(t *testing.T, content []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "angel-ai")
	if err := os.WriteFile(path, content, 0o751); err != nil {
		t.Fatal(err)
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatal(err)
	}
	return resolved
}

func assertExecutableBytes(t *testing.T, path string, want []byte) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("%s bytes = %q, want %q", path, got, want)
	}
}

func assertOnlyExecutableRemains(t *testing.T, executable string) {
	t.Helper()
	entries, err := os.ReadDir(filepath.Dir(executable))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != filepath.Base(executable) {
		names := make([]string, 0, len(entries))
		for _, entry := range entries {
			names = append(names, entry.Name())
		}
		t.Fatalf("temporary state remains: %v", names)
	}
}
