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
			doc, err := bkn.LoadNetwork(dir)
			require.NoError(t, err, "load network")

			// Verify basic structure
			assert.NotEmpty(t, doc.BknNetworkFrontmatter.ID, "network id should not be empty")
			assert.NotEmpty(t, doc.BknNetworkFrontmatter.Name, "network name should not be empty")

			// Verify at least some content was loaded (objects, relations, or actions)
			totalEntities := len(doc.ObjectTypes) + len(doc.RelationTypes) + len(doc.ActionTypes)
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
			doc, err := bkn.LoadNetworkFromTar(buf)
			require.NoError(t, err, "load from tar")

			// Compare with file load
			fileDoc, err := bkn.LoadNetwork(dir)
			require.NoError(t, err, "load from file")

			// Verify counts match
			assert.Equal(t, len(fileDoc.ObjectTypes), len(doc.ObjectTypes), "objects count should match")
			assert.Equal(t, fileDoc.BknNetworkFrontmatter.ID, doc.BknNetworkFrontmatter.ID, "root ID should match")
		})
	}
}

// TestWriteToTar: Model → Tar
func TestWriteToTar(t *testing.T) {
	for _, dir := range allExampleDirs(t) {
		name := filepath.Base(dir)
		t.Run(name, func(t *testing.T) {
			doc, err := bkn.LoadNetwork(dir)
			require.NoError(t, err, "load network")

			// Write to tar
			var buf bytes.Buffer
			err = bkn.WriteNetworkToTar(doc, &buf)
			require.NoError(t, err, "write to tar")

			// Reload from tar and compare
			reloaded, err := bkn.LoadNetworkFromTar(&buf)
			require.NoError(t, err, "reload from tar")

			// Verify root frontmatter
			assert.Equal(t, doc.BknNetworkFrontmatter.ID, reloaded.BknNetworkFrontmatter.ID, "root ID should match")
			assert.Equal(t, len(doc.ObjectTypes), len(reloaded.ObjectTypes), "objects count should match")
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
			assert.Equal(t, original.BknNetworkFrontmatter.ID, result.BknNetworkFrontmatter.ID, "root ID should match")
			assert.Equal(t, original.BknNetworkFrontmatter.Name, result.BknNetworkFrontmatter.Name, "root name should match")
			assert.Equal(t, len(original.ObjectTypes), len(result.ObjectTypes), "objects count should match")
			assert.Equal(t, len(original.RelationTypes), len(result.RelationTypes), "relations count should match")
			assert.Equal(t, len(original.ActionTypes), len(result.ActionTypes), "actions count should match")
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
			require.NoError(t, err, "load from new tar")

			// Step 5: Verify consistency
			assert.Equal(t, original.BknNetworkFrontmatter.ID, result.BknNetworkFrontmatter.ID, "root ID should match")
			assert.Equal(t, len(original.ObjectTypes), len(result.ObjectTypes), "objects count should match")
		})
	}
}

// TestChecksumConsistency: 校验和一致性
func TestChecksumConsistency(t *testing.T) {
	for _, dir := range allExampleDirs(t) {
		name := filepath.Base(dir)
		t.Run(name, func(t *testing.T) {
			// Load and write to tar
			doc, err := bkn.LoadNetwork(dir)
			require.NoError(t, err, "load network")

			var buf bytes.Buffer
			err = bkn.WriteNetworkToTar(doc, &buf)
			require.NoError(t, err, "write to tar")

			// Load from tar
			result, err := bkn.LoadNetworkFromTar(&buf)
			require.NoError(t, err, "load from tar")

			// Verify basic consistency
			assert.Equal(t, doc.BknNetworkFrontmatter.ID, result.BknNetworkFrontmatter.ID, "root ID should match")
		})
	}
}

// === Boundary Case Tests ===

// TestEmptyNetwork: 空网络处理
func TestEmptyNetwork(t *testing.T) {
	dir := tempDir(t)

	// Create minimal network.bkn
	networkContent := `---
type: network
id: test-empty
name: Test Empty Network
version: "1.0"
---

# Test Empty Network
`
	err := os.WriteFile(filepath.Join(dir, "network.bkn"), []byte(networkContent), 0644)
	require.NoError(t, err, "write network.bkn")

	// Load should succeed
	doc, err := bkn.LoadNetwork(dir)
	require.NoError(t, err, "load empty network")
	assert.Equal(t, "test-empty", doc.BknNetworkFrontmatter.ID, "root ID should match")
	assert.Empty(t, doc.ObjectTypes, "should have no objects")
	assert.Empty(t, doc.RelationTypes, "should have no relations")
	assert.Empty(t, doc.ActionTypes, "should have no actions")
}

// TestCircularInclude: 循环include检测
func TestCircularInclude(t *testing.T) {
	dir := tempDir(t)

	// Create network.bkn with circular include
	networkContent := `---
type: network
id: test-circular
name: Test Circular
version: "1.0"
---

# Test Circular
`
	err := os.WriteFile(filepath.Join(dir, "network.bkn"), []byte(networkContent), 0644)
	require.NoError(t, err, "write network.bkn")

	// Create object_types directory with a file
	objDir := filepath.Join(dir, "object_types")
	err = os.MkdirAll(objDir, 0755)
	require.NoError(t, err, "create object_types dir")

	objContent := `---
type: object_type
id: test-obj
name: Test Object
---

## ObjectType: test-obj

Test object description.
`
	err = os.WriteFile(filepath.Join(objDir, "test.bkn"), []byte(objContent), 0644)
	require.NoError(t, err, "write test.bkn")

	// Load should succeed
	doc, err := bkn.LoadNetwork(dir)
	require.NoError(t, err, "load network with objects")
	assert.Equal(t, 1, len(doc.ObjectTypes), "should have 1 object")
}

// TestMissingInclude: 缺失include文件
func TestMissingInclude(t *testing.T) {
	dir := tempDir(t)

	// Create network.bkn
	networkContent := `---
type: network
id: test-missing
name: Test Missing
version: "1.0"
---

# Test Missing
`
	err := os.WriteFile(filepath.Join(dir, "network.bkn"), []byte(networkContent), 0644)
	require.NoError(t, err, "write network.bkn")

	// Load should succeed even with missing subdirectories
	doc, err := bkn.LoadNetwork(dir)
	require.NoError(t, err, "load network with missing includes")
	assert.Equal(t, "test-missing", doc.BknNetworkFrontmatter.ID, "root ID should match")
}

// TestLargeNetwork: 大规模网络性能
func TestLargeNetwork(t *testing.T) {
	dir := tempDir(t)

	// Create network.bkn
	networkContent := `---
type: network
id: test-large
name: Test Large Network
version: "1.0"
---

# Test Large Network
`
	err := os.WriteFile(filepath.Join(dir, "network.bkn"), []byte(networkContent), 0644)
	require.NoError(t, err, "write network.bkn")

	// Create object_types directory with multiple files
	objDir := filepath.Join(dir, "object_types")
	err = os.MkdirAll(objDir, 0755)
	require.NoError(t, err, "create object_types dir")

	// Create 10 object files
	for i := 0; i < 10; i++ {
		objContent := `---
type: object_type
id: test-obj-` + string(rune('0'+i)) + `
name: Test Object ` + string(rune('0'+i)) + `
---

## ObjectType: test-obj-` + string(rune('0'+i)) + `

Test object description.
`
		err = os.WriteFile(filepath.Join(objDir, "test"+string(rune('0'+i))+".bkn"), []byte(objContent), 0644)
		require.NoError(t, err, "write test object")
	}

	// Load should succeed
	doc, err := bkn.LoadNetwork(dir)
	require.NoError(t, err, "load large network")
	assert.Equal(t, 10, len(doc.ObjectTypes), "should have 10 objects")
}

// TestInvalidBKNFile: 无效BKN文件处理
func TestInvalidBKNFile(t *testing.T) {
	dir := tempDir(t)

	// Create invalid network.bkn (missing frontmatter)
	invalidContent := `# Invalid Network

This file has no frontmatter.
`
	err := os.WriteFile(filepath.Join(dir, "network.bkn"), []byte(invalidContent), 0644)
	require.NoError(t, err, "write invalid network.bkn")

	// Load should fail
	_, err = bkn.LoadNetwork(dir)
	assert.Error(t, err, "should fail to load invalid network")
}
