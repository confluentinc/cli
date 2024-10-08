package testserver

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

var packagesRoutes = []route{
	{"/confluent-cli/binaries/", handleConfluentCliBinaries},
	{"/confluent-cli/binaries/{version}/confluent_checksums.txt", handleConfluentCliBinariesVersionConfluentChecksumsTxt},
	{"/confluent-cli/binaries/{version}/{binary}", handleConfluentCliBinariesVersionBinary},
	{"/confluent-cli/release-notes/{version}/release-notes.rst", handleConfluentCliReleaseNotesVersionReleaseNotesRst},
}

func NewPackagesRouter(t *testing.T) *mux.Router {
	router := mux.NewRouter()
	router.Use(defaultHeaderMiddleware)

	for _, route := range packagesRoutes {
		router.HandleFunc(route.path, route.handler(t))
	}

	return router
}

// Handler for: "/confluent-cli/binaries/"
func handleConfluentCliBinaries(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		content := `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 3.2 Final//EN">
<html>
  <head>
    <title>Index of confluent-cli/binaries</title>
  </head>
  <body>
    <h1>Index of confluent-cli/binaries</h1>
    <hr/>
    <pre>
      <a href="/confluent-cli/">../</a>
      <a href="/confluent-cli/binaries/0.0.0/">0.0.0/</a>
      <a href="/confluent-cli/binaries/0.1.0/">0.1.0/</a>
      <a href="/confluent-cli/binaries/1.0.0/">1.0.0/</a>
      <a href="/confluent-cli/binaries/2.0.0/">2.0.0/</a>
      <a href="/confluent-cli/binaries/0.1.1/">0.1.1/</a>
    </pre>
    <hr/>
  </body>
</html>`

		_, err := w.Write([]byte(content))
		require.NoError(t, err)
	}
}

// Handler for: "/confluent-cli/binaries/{version}/confluent_checksums.txt"
func handleConfluentCliBinariesVersionConfluentChecksumsTxt(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		binary := "confluent"
		if runtime.GOOS == "windows" {
			binary += ".exe"
		}

		out, err := os.ReadFile(filepath.Join("test", "bin", binary))
		require.NoError(t, err)

		hash := sha256.Sum256(out)

		content := fmt.Sprintf(`%[1]s confluent_darwin_amd64
%[1]s confluent_darwin_arm64
%[1]s confluent_linux_amd64
%[1]s confluent_linux_arm64
%[1]s confluent_windows_amd64.exe`, hex.EncodeToString(hash[:]))

		_, err = w.Write([]byte(content))
		require.NoError(t, err)
	}
}

// Handler for: "/confluent-cli/binaries/{version}/{binary}"
func handleConfluentCliBinariesVersionBinary(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		binary := "confluent"
		if runtime.GOOS == "windows" {
			binary += ".exe"
		}

		out, err := os.ReadFile(filepath.Join("test", "bin", binary))
		require.NoError(t, err)

		_, err = w.Write(out)
		require.NoError(t, err)
	}
}

// Handler for: "/confluent-cli/release-notes/{version}/release-notes.rst"
func handleConfluentCliReleaseNotesVersionReleaseNotesRst(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		content := map[string]string{
			"0.1.0": `[1/1/1970] Confluent CLI v0.1.0 Release Notes
=============================================

New Features
------------
- New Feature #1

`,
			"0.1.1": `[1/2/1970] Confluent CLI v0.1.1 Release Notes
=============================================

Bug Fixes
---------
- Bug Fix #1

`,
			"1.0.0": `[1/3/1970] Confluent CLI v1.0.0 Release Notes
=============================================

Breaking Changes
----------------
- Breaking Change #1

`,
			"2.0.0": `[1/4/1970] Confluent CLI v2.0.0 Release Notes
=============================================

Breaking Changes
----------------
- Breaking Change #2

`,
		}

		version := mux.Vars(r)["version"]
		_, err := w.Write([]byte(content[version]))
		require.NoError(t, err)
	}
}
