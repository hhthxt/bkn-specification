// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package test contains integration tests for the BKN SDK.
// These tests verify end-to-end workflows including file I/O,
// network loading, serialization, and tar operations.
package test

import (
	"archive/tar"
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/kweaver-ai/bkn-specification/sdk/golang/bkn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// allExampleDirs returns all example directories with network.bkn files.
// Test runs from sdk/golang/test/, examples are at ../../../examples
func allExampleDirs(t *testing.T) []string {
	t.Helper()

	examplesDir := filepath.Join("..", "..", "..", "examples")
	entries, err := os.ReadDir(examplesDir)
	require.NoError(t, err, "read examples dir")

	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			d := filepath.Join(examplesDir, e.Name())
			if _, err := os.Stat(filepath.Join(d, "network.bkn")); err == nil {
				dirs = append(dirs, d)
			}
		}
	}

	if len(dirs) == 0 {
		t.Skip("no example directories with network.bkn found")
	}
	return dirs
}

// tempDir creates a temporary directory for test files.
func tempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "bkn-test-*")
	require.NoError(t, err, "create temp dir")
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

// buildTarFromDir packs all files in dir into a tar buffer.
func buildTarFromDir(t *testing.T, dir string) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(dir, path)
		rel = filepath.ToSlash(rel)
		data, err := os.ReadFile(path)
		require.NoError(t, err, "read %s", path)
		tw.WriteHeader(&tar.Header{Name: rel, Size: int64(len(data)), Mode: 0644})
		tw.Write(data)
		return nil
	})
	tw.Close()
	return &buf
}

// === Core Workflow Tests ===

// TestLoadFromFile: 文件 → Model
func TestLoadFromFile(t *testing.T) {
	for _, dir := range allExampleDirs(t) {
		name := filepath.Base(dir)
		t.Run(name, func(t *testing.T) {
			net, err := bkn.LoadNetwork(dir)
			require.NoError(t, err, "load network")

			// Verify basic structure
			assert.Equal(t, "network", net.Root.Frontmatter.Type, "type should be network")
			assert.NotEmpty(t, net.Root.Frontmatter.ID, "network id should not be empty")

			// Verify at least some content was loaded (objects, relations, or actions)
			totalEntities := len(net.AllObjects()) + len(net.AllRelations()) + len(net.AllActions())
			assert.Greater(t, totalEntities, 0, "expected at least some entities (objects, relations, or actions)")
		})
	}
}

// TestLoadFromTar: Tar → Model
func TestLoadFromTar(t *testing.T) {
	for _, dir := range allExampleDirs(t) {
		name := filepath.Base(dir)
		t.Run(name, func(t *testing.T) {
			// Build tar from directory
			buf := buildTarFromDir(t, dir)

			// Load from tar
			net, err := bkn.LoadNetworkFromTar(buf)
			require.NoError(t, err, "load from tar")

			// Compare with file load
			fileNet, err := bkn.LoadNetwork(dir)
			require.NoError(t, err, "load from file")

			// Verify counts match
			assert.Equal(t, len(fileNet.AllObjects()), len(net.AllObjects()), "objects count should match")
			assert.Equal(t, fileNet.Root.Frontmatter.ID, net.Root.Frontmatter.ID, "root ID should match")
		})
	}
}

// TestWriteToTar: Model → Tar
func TestWriteToTar(t *testing.T) {
	for _, dir := range allExampleDirs(t) {
		name := filepath.Base(dir)
		t.Run(name, func(t *testing.T) {
			net, err := bkn.LoadNetwork(dir)
			require.NoError(t, err, "load network")

			// Write to tar
			var buf bytes.Buffer
			err = bkn.WriteNetworkToTar(net, &buf)
			require.NoError(t, err, "write to tar")

			// Reload from tar and compare
			reloaded, err := bkn.LoadNetworkFromTar(&buf)
			require.NoError(t, err, "reload from tar")

			// Verify root frontmatter
			assert.Equal(t, net.Root.Frontmatter.ID, reloaded.Root.Frontmatter.ID, "root ID should match")
			assert.Equal(t, len(net.AllObjects()), len(reloaded.AllObjects()), "objects count should match")
		})
	}
}

// TestRoundTrip_FileToTar: 文件→Model→Tar→Model
func TestRoundTrip_FileToTar(t *testing.T) {
	for _, dir := range allExampleDirs(t) {
		name := filepath.Base(dir)
		t.Run(name, func(t *testing.T) {
			// Step 1: Load from file
			original, err := bkn.LoadNetwork(dir)
			require.NoError(t, err, "load from file")

			// Step 2: Write to tar
			var buf bytes.Buffer
			err = bkn.WriteNetworkToTar(original, &buf)
			require.NoError(t, err, "write to tar")

			// Step 3: Load from tar
			result, err := bkn.LoadNetworkFromTar(&buf)
			require.NoError(t, err, "load from tar")

			// Step 4: Strict comparison (original vs result)
			assert.Equal(t, original.Root.Frontmatter.ID, result.Root.Frontmatter.ID, "root ID should match")
			assert.Equal(t, original.Root.Frontmatter.Type, result.Root.Frontmatter.Type, "root type should match")
			assert.Equal(t, len(original.AllObjects()), len(result.AllObjects()), "objects count should match")
			assert.Equal(t, len(original.AllRelations()), len(result.AllRelations()), "relations count should match")
			assert.Equal(t, len(original.AllActions()), len(result.AllActions()), "actions count should match")
		})
	}
}

// TestRoundTrip_TarToTar: Tar→Model→Tar→Model
func TestRoundTrip_TarToTar(t *testing.T) {
	for _, dir := range allExampleDirs(t) {
		name := filepath.Base(dir)
		t.Run(name, func(t *testing.T) {
			// Step 1: Build initial tar
			buf1 := buildTarFromDir(t, dir)

			// Step 2: Load from tar
			original, err := bkn.LoadNetworkFromTar(buf1)
			require.NoError(t, err, "load from tar")

			// Step 3: Write to new tar
			var buf2 bytes.Buffer
			err = bkn.WriteNetworkToTar(original, &buf2)
			require.NoError(t, err, "write to tar")

			// Step 4: Load from new tar
			result, err := bkn.LoadNetworkFromTar(&buf2)
			require.NoError(t, err, "reload from tar")

			// Step 5: Strict comparison
			assert.Equal(t, original.Root.Frontmatter.ID, result.Root.Frontmatter.ID, "root ID should match")
			assert.Equal(t, len(original.AllObjects()), len(result.AllObjects()), "objects count should match")
		})
	}
}

// === Edge Case Tests ===

// TestEmptyNetwork: 空网络处理
func TestEmptyNetwork(t *testing.T) {
	tempDir := tempDir(t)

	// Create minimal network.bkn
	content := `---
type: network
id: empty-net
name: Empty Network
---

# Empty Network
`
	err := os.WriteFile(filepath.Join(tempDir, "network.bkn"), []byte(content), 0644)
	require.NoError(t, err, "write network.bkn")

	// Load should succeed
	net, err := bkn.LoadNetwork(tempDir)
	require.NoError(t, err, "load empty network")

	// Verify empty structure
	assert.Equal(t, "empty-net", net.Root.Frontmatter.ID, "network id should match")
	assert.Empty(t, net.AllObjects(), "should have 0 objects")
	assert.Empty(t, net.AllRelations(), "should have 0 relations")
	assert.Empty(t, net.AllActions(), "should have 0 actions")
}

// TestMissingRootFile: 目录无 network.bkn
func TestMissingRootFile(t *testing.T) {
	tempDir := tempDir(t)

	// Create a non-network file
	err := os.WriteFile(filepath.Join(tempDir, "readme.txt"), []byte("hello"), 0644)
	require.NoError(t, err, "write readme")

	// Load should fail
	_, err = bkn.LoadNetwork(tempDir)
	assert.Error(t, err, "expected error for missing root file")
}

// TestInvalidFrontmatter: 无效 YAML
func TestInvalidFrontmatter(t *testing.T) {
	tempDir := tempDir(t)

	// Create invalid network.bkn (malformed YAML)
	content := `---
type: network
id: test
name: Test
  invalid_yaml_here: [
---

# Test
`
	err := os.WriteFile(filepath.Join(tempDir, "network.bkn"), []byte(content), 0644)
	require.NoError(t, err, "write network.bkn")

	// Load should fail
	_, err = bkn.LoadNetwork(tempDir)
	assert.Error(t, err, "expected error for invalid frontmatter")
}

// TestCircularInclude: 循环包含（只有 network 类型可以有 includes）
func TestCircularInclude(t *testing.T) {
	tempDir := tempDir(t)

	// Create network.bkn that includes a.bkn
	networkContent := `---
type: network
id: circular-test
includes:
  - a.bkn
---

# Test
`
	err := os.WriteFile(filepath.Join(tempDir, "network.bkn"), []byte(networkContent), 0644)
	require.NoError(t, err, "write network.bkn")

	// Create a.bkn (network type) that includes b.bkn
	aContent := `---
type: network
id: type-a
includes:
  - b.bkn
---

# Type A
`
	err = os.WriteFile(filepath.Join(tempDir, "a.bkn"), []byte(aContent), 0644)
	require.NoError(t, err, "write a.bkn")

	// Create b.bkn (network type) that includes a.bkn (circular)
	bContent := `---
type: network
id: type-b
includes:
  - a.bkn
---

# Type B
`
	err = os.WriteFile(filepath.Join(tempDir, "b.bkn"), []byte(bContent), 0644)
	require.NoError(t, err, "write b.bkn")

	// Load should fail with circular error
	_, err = bkn.LoadNetwork(tempDir)
	require.Error(t, err, "expected error for circular include")
	assert.Contains(t, err.Error(), "circular", "error should mention circular")
}

// TestMissingInclude: include 文件不存在
func TestMissingInclude(t *testing.T) {
	tempDir := tempDir(t)

	// Create network.bkn that includes non-existent file
	content := `---
type: network
id: missing-include-test
includes:
  - nonexistent.bkn
---

# Test
`
	err := os.WriteFile(filepath.Join(tempDir, "network.bkn"), []byte(content), 0644)
	require.NoError(t, err, "write network.bkn")

	// Load should fail
	_, err = bkn.LoadNetwork(tempDir)
	require.Error(t, err, "expected error for missing include")
	assert.Contains(t, err.Error(), "not found", "error should mention not found")
}
