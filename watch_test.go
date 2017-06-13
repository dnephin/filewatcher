package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEndToEndWithExcludes(t *testing.T) {
	path, excludes, cleanup := setupTestDirectory(t)
	defer cleanup()

	shutdown := make(chan struct{}, 1)
	opts := &options{
		dirs: []string{path},
		exclude: excludes,
		command: echoCommand(path),
	}
	go func() {
		require.NoError(t, run(opts, shutdown))
	}()
	// TODO: better way to wait for setup
	time.Sleep(200 * time.Millisecond)

	writeFile(t, path, "new")
	time.Sleep(200 * time.Millisecond)
	shutdown <- struct{}{}

	content, err := ioutil.ReadFile(filepath.Join(path, "output"))
	require.NoError(t, err)
	events := strings.Split(strings.TrimSpace(string(content)), "\n")
	assert.Len(t, events, 1)
}

func setupTestDirectory(t *testing.T) (string, []string, func()) {
	path, err := ioutil.TempDir("", "test-filewatcher")
	require.NoError(t, err)

	goodDir := mkDir(t, path, "good")
	ignoreDir := mkDir(t, path, "ignore")
	writeFile(t, path, "foo.ign")
	writeFile(t, path, "file0.txt")
	writeFile(t, goodDir, "file1.txt")
	writeFile(t, goodDir, "foo.ign")
	writeFile(t, ignoreDir, "file2.txt")

	excludes := []string{ignoreDir, "*.ign", "output"}
	return path, excludes, func() { require.NoError(t, os.RemoveAll(path)) }
}

func mkDir(t *testing.T, path string, name string) string {
	fullPath := filepath.Join(path, name)
	require.NoError(t, os.Mkdir(fullPath, 0x755))
	return fullPath
}

func writeFile(t *testing.T, path string, name string) {
	fullPath := filepath.Join(path, name)
	require.NoError(t, ioutil.WriteFile(fullPath, []byte("content"), 0x644))
}

func echoCommand(path string) []string {
	filename := filepath.Join(path, "output")
	return []string{"bash", "-c", "echo '${filepath}' >> " + filename}
}