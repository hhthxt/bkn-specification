// // Copyright The kweaver.ai Authors.
// //
// // Licensed under the Apache License, Version 2.0.
// // See the LICENSE file in the project root for details.

package bkn

// import (
// 	"archive/tar"
// 	"bytes"
// 	"strings"
// 	"testing"
// 	"time"
// )

// // --- Unit Tests for Tar Operations ---

// func TestWriteTarEntry(t *testing.T) {
// 	var buf bytes.Buffer
// 	tw := tar.NewWriter(&buf)
// 	defer tw.Close()

// 	content := []byte("test content")
// 	err := writeTarEntry(tw, "test.bkn", content, time.Now())
// 	if err != nil {
// 		t.Fatalf("writeTarEntry: %v", err)
// 	}
// 	tw.Close()

// 	// Verify tar can be read
// 	tr := tar.NewReader(&buf)
// 	hdr, err := tr.Next()
// 	if err != nil {
// 		t.Fatalf("read tar header: %v", err)
// 	}
// 	if hdr.Name != "test.bkn" {
// 		t.Errorf("expected name test.bkn, got %q", hdr.Name)
// 	}
// 	if hdr.Size != int64(len(content)) {
// 		t.Errorf("expected size %d, got %d", len(content), hdr.Size)
// 	}
// }

// func TestExtractTarToMemory_SingleFile(t *testing.T) {
// 	var buf bytes.Buffer
// 	tw := tar.NewWriter(&buf)

// 	// Write a single file
// 	content := []byte("---\ntype: network\nid: test\n---\n")
// 	tw.WriteHeader(&tar.Header{Name: "network.bkn", Size: int64(len(content)), Mode: 0644})
// 	tw.Write(content)
// 	tw.Close()

// 	mfs, _, err := ExtractTarToMemory(&buf)
// 	if err != nil {
// 		t.Fatalf("extract: %v", err)
// 	}

// 	data, err := mfs.ReadFile("network.bkn")
// 	if err != nil {
// 		t.Fatalf("read file: %v", err)
// 	}
// 	if string(data) != string(content) {
// 		t.Errorf("content mismatch")
// 	}
// }

// func TestExtractTarToMemory_MultipleFiles(t *testing.T) {
// 	var buf bytes.Buffer
// 	tw := tar.NewWriter(&buf)

// 	files := map[string]string{
// 		"network.bkn":          "---\ntype: network\nid: test\n---\n",
// 		"object_types/pod.bkn": "---\ntype: object_type\nid: pod\n---\n",
// 		"SKILL.md":             "# Test Network\n",
// 	}

// 	for name, content := range files {
// 		data := []byte(content)
// 		tw.WriteHeader(&tar.Header{Name: name, Size: int64(len(data)), Mode: 0644})
// 		tw.Write(data)
// 	}
// 	tw.Close()

// 	mfs, _, err := ExtractTarToMemory(&buf)
// 	if err != nil {
// 		t.Fatalf("extract: %v", err)
// 	}

// 	for name := range files {
// 		_, err := mfs.ReadFile(name)
// 		if err != nil {
// 			t.Errorf("file %q not found: %v", name, err)
// 		}
// 	}
// }

// func TestExtractTarToMemory_EmptyTar(t *testing.T) {
// 	var buf bytes.Buffer
// 	tw := tar.NewWriter(&buf)
// 	tw.Close()

// 	// Empty tar should return error because no root network file found
// 	_, _, err := ExtractTarToMemory(&buf)
// 	if err == nil {
// 		t.Error("expected error for empty tar (no root network file)")
// 	}
// }

// // --- Unit Tests for MemoryFileSystem ---

// func TestMemoryFileSystem_AddFile(t *testing.T) {
// 	mfs := NewMemoryFileSystem()
// 	mfs.AddFile("test.bkn", []byte("content"))

// 	data, err := mfs.ReadFile("test.bkn")
// 	if err != nil {
// 		t.Fatalf("read file: %v", err)
// 	}
// 	if string(data) != "content" {
// 		t.Errorf("expected 'content', got %q", string(data))
// 	}
// }

// func TestMemoryFileSystem_ReadFile_NotFound(t *testing.T) {
// 	mfs := NewMemoryFileSystem()
// 	_, err := mfs.ReadFile("nonexistent.bkn")
// 	if err == nil {
// 		t.Error("expected error for nonexistent file")
// 	}
// }

// func TestMemoryFileSystem_Stat(t *testing.T) {
// 	mfs := NewMemoryFileSystem()
// 	mfs.AddFile("a.bkn", []byte("a"))
// 	mfs.AddFile("b.bkn", []byte("b"))

// 	// Test file stat
// 	info, err := mfs.Stat("a.bkn")
// 	if err != nil {
// 		t.Fatalf("stat file: %v", err)
// 	}
// 	if info.IsDir() {
// 		t.Error("expected file, got dir")
// 	}

// 	// Test nonexistent file
// 	_, err = mfs.Stat("nonexistent.bkn")
// 	if err == nil {
// 		t.Error("expected error for nonexistent file")
// 	}
// }

// // --- Unit Tests for GenerateSkillMd ---

// func TestGenerateSkillMd_Basic(t *testing.T) {
// 	net := &BknNetwork{
// 		Root: BknDocument{
// 			Frontmatter: Frontmatter{
// 				Type:        "network",
// 				ID:          "test-net",
// 				Name:        "Test Network",
// 				Description: "A test network",
// 				Version:     "1.0.0",
// 			},
// 		},
// 		Includes: []BknDocument{
// 			{
// 				Frontmatter: Frontmatter{Type: "object_type", ID: "pod", Name: "Pod"},
// 				Objects:     []BknObject{{ID: "pod", Name: "Pod"}},
// 			},
// 		},
// 	}

// 	skill := generateSkillMd(net)
// 	if !strings.Contains(skill, "# Test Network") {
// 		t.Error("SKILL.md should contain network name as title")
// 	}
// 	if !strings.Contains(skill, "test-net") {
// 		t.Error("SKILL.md should contain network ID")
// 	}
// 	if !strings.Contains(skill, "1.0.0") {
// 		t.Error("SKILL.md should contain version")
// 	}
// 	if !strings.Contains(skill, "pod") {
// 		t.Error("SKILL.md should contain object type")
// 	}
// }

// func TestGenerateSkillMd_EmptyNetwork(t *testing.T) {
// 	net := &BknNetwork{
// 		Root: BknDocument{
// 			Frontmatter: Frontmatter{
// 				Type: "network",
// 				ID:   "empty-net",
// 			},
// 		},
// 	}

// 	skill := generateSkillMd(net)
// 	if !strings.Contains(skill, "empty-net") {
// 		t.Error("SKILL.md should contain network ID")
// 	}
// 	// SKILL.md should have network overview section
// 	if !strings.Contains(skill, "## 网络概览") {
// 		t.Error("SKILL.md should have network overview section")
// 	}
// }

// // --- Unit Tests for Checksum from Tar ---

// func TestComputeChecksumFromTar_Empty(t *testing.T) {
// 	var buf bytes.Buffer
// 	tw := tar.NewWriter(&buf)
// 	tw.Close()

// 	// Empty tar should return error because no root network file found
// 	_, err := ComputeChecksumFromTar(&buf)
// 	if err == nil {
// 		t.Error("expected error for empty tar (no root network file)")
// 	}
// }

// func TestVerifyChecksumFromTar_NoChecksumFile(t *testing.T) {
// 	var buf bytes.Buffer
// 	tw := tar.NewWriter(&buf)

// 	content := []byte("---\ntype: network\nid: test\n---\n")
// 	tw.WriteHeader(&tar.Header{Name: "network.bkn", Size: int64(len(content)), Mode: 0644})
// 	tw.Write(content)
// 	tw.Close()

// 	ok, errs := VerifyChecksumFromTar(&buf)
// 	if ok {
// 		t.Error("verify should fail without CHECKSUM file")
// 	}
// 	if len(errs) == 0 {
// 		t.Error("expected error messages when CHECKSUM is missing")
// 	}
// }
