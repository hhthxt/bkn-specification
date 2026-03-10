package bkn

import (
	"archive/tar"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// buildTarFromDir reads all files under dir and packs them into an in-memory tar.
// File paths inside the tar are relative to dir.
func buildTarFromDir(t *testing.T, dir string) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	filepath.Walk(dir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil || info.IsDir() {
			return walkErr
		}
		rel, _ := filepath.Rel(dir, path)
		rel = filepath.ToSlash(rel)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		tw.WriteHeader(&tar.Header{Name: rel, Size: int64(len(data)), Mode: 0644})
		tw.Write(data)
		return nil
	})
	if err := tw.Close(); err != nil {
		t.Fatalf("close tar: %v", err)
	}
	return &buf
}

// buildTarFromDirExcluding is like buildTarFromDir but skips files with the given base name.
func buildTarFromDirExcluding(t *testing.T, dir, excludeName string) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	filepath.Walk(dir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil || info.IsDir() || filepath.Base(path) == excludeName {
			return walkErr
		}
		rel, _ := filepath.Rel(dir, path)
		rel = filepath.ToSlash(rel)
		data, _ := os.ReadFile(path)
		tw.WriteHeader(&tar.Header{Name: rel, Size: int64(len(data)), Mode: 0644})
		tw.Write(data)
		return nil
	})
	tw.Close()
	return &buf
}

// allExampleDirs returns all example directories under examples/.
func allExampleDirs(t *testing.T) []string {
	t.Helper()
	root := repoRoot(t)
	examplesRoot := filepath.Join(root, "examples")
	entries, err := os.ReadDir(examplesRoot)
	if err != nil {
		t.Fatalf("read examples dir: %v", err)
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, filepath.Join(examplesRoot, e.Name()))
		}
	}
	if len(dirs) == 0 {
		t.Skip("no example directories found")
	}
	return dirs
}

func TestLoadNetworkFromTar(t *testing.T) {
	for _, dir := range allExampleDirs(t) {
		name := filepath.Base(dir)
		t.Run(name, func(t *testing.T) {
			buf := buildTarFromDir(t, dir)

			net, err := LoadNetworkFromTar(buf)
			if err != nil {
				t.Fatalf("load: %v", err)
			}
			if net.Root.Frontmatter.Type != "network" {
				t.Errorf("expected type network, got %q", net.Root.Frontmatter.Type)
			}
			if net.Root.Frontmatter.ID == "" {
				t.Error("expected non-empty network id")
			}

			// Compare with local FS load
			localNet, err := LoadNetwork(dir)
			if err != nil {
				t.Fatalf("local load: %v", err)
			}
			if len(net.AllObjects()) != len(localNet.AllObjects()) {
				t.Errorf("objects: tar=%d local=%d", len(net.AllObjects()), len(localNet.AllObjects()))
			}
			if len(net.AllRelations()) != len(localNet.AllRelations()) {
				t.Errorf("relations: tar=%d local=%d", len(net.AllRelations()), len(localNet.AllRelations()))
			}
			if len(net.AllActions()) != len(localNet.AllActions()) {
				t.Errorf("actions: tar=%d local=%d", len(net.AllActions()), len(localNet.AllActions()))
			}
			if len(net.AllRisks()) != len(localNet.AllRisks()) {
				t.Errorf("risks: tar=%d local=%d", len(net.AllRisks()), len(localNet.AllRisks()))
			}
		})
	}
}

func TestExtractTarToMemory_IncludesChecksum(t *testing.T) {
	for _, dir := range allExampleDirs(t) {
		name := filepath.Base(dir)
		t.Run(name, func(t *testing.T) {
			buf := buildTarFromDir(t, dir)

			mfs, _, err := ExtractTarToMemory(buf)
			if err != nil {
				t.Fatalf("extract: %v", err)
			}

			// All examples have CHECKSUM
			data, err := mfs.ReadFile("CHECKSUM")
			if err != nil {
				t.Fatalf("CHECKSUM not found in memory fs: %v", err)
			}
			if !strings.Contains(string(data), "sha256:") {
				t.Error("CHECKSUM should contain sha256 entries")
			}

			// All examples have SKILL.md
			_, err = mfs.ReadFile("SKILL.md")
			if err != nil {
				t.Fatalf("SKILL.md not found in memory fs: %v", err)
			}
		})
	}
}

func TestComputeChecksumFromTar(t *testing.T) {
	for _, dir := range allExampleDirs(t) {
		name := filepath.Base(dir)
		t.Run(name, func(t *testing.T) {
			// Compute via local FS
			fsys := NewOSFileSystem()
			localChecksums, err := ComputeNetworkChecksums(fsys, dir)
			if err != nil {
				t.Fatalf("local compute: %v", err)
			}

			// Compute via tar
			buf := buildTarFromDir(t, dir)
			tarChecksums, err := ComputeChecksumFromTar(buf)
			if err != nil {
				t.Fatalf("tar compute: %v", err)
			}

			// Results must match
			for key, localHash := range localChecksums {
				tarHash, ok := tarChecksums[key]
				if !ok {
					t.Errorf("tar checksums missing key %q", key)
					continue
				}
				if localHash != tarHash {
					t.Errorf("checksum mismatch for %q: local=%q tar=%q", key, localHash, tarHash)
				}
			}
			for key := range tarChecksums {
				if _, ok := localChecksums[key]; !ok {
					t.Errorf("tar has extra key %q not in local", key)
				}
			}
		})
	}
}

func TestVerifyChecksumFromTar(t *testing.T) {
	for _, dir := range allExampleDirs(t) {
		name := filepath.Base(dir)
		t.Run(name, func(t *testing.T) {
			// Generate fresh CHECKSUM from the tar (avoids stale stored CHECKSUM)
			bufGen := buildTarFromDirExcluding(t, dir, "CHECKSUM")
			checksumContent, err := GenerateChecksumFromTar(bufGen)
			if err != nil {
				t.Fatalf("generate: %v", err)
			}

			// Build tar with fresh CHECKSUM, verify passes
			bufVerify := buildTarFromDirReplaceChecksum(t, dir, checksumContent)
			ok, errs := VerifyChecksumFromTar(bufVerify)
			if !ok {
				t.Errorf("verify should pass, got errors: %v", errs)
			}

			// Verify without CHECKSUM should fail
			bufNoCk := buildTarFromDirExcluding(t, dir, "CHECKSUM")
			ok2, errs2 := VerifyChecksumFromTar(bufNoCk)
			if ok2 {
				t.Error("verify without CHECKSUM should fail")
			}
			if len(errs2) == 0 {
				t.Error("expected error messages when CHECKSUM is missing")
			}
		})
	}
}

// buildTarFromDirReplaceChecksum builds a tar from dir, replacing the CHECKSUM file content.
func buildTarFromDirReplaceChecksum(t *testing.T, dir, checksumContent string) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	filepath.Walk(dir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil || info.IsDir() {
			return walkErr
		}
		rel, _ := filepath.Rel(dir, path)
		rel = filepath.ToSlash(rel)

		var data []byte
		if filepath.Base(path) == "CHECKSUM" {
			data = []byte(checksumContent)
		} else {
			data, _ = os.ReadFile(path)
		}
		tw.WriteHeader(&tar.Header{Name: rel, Size: int64(len(data)), Mode: 0644})
		tw.Write(data)
		return nil
	})
	tw.Close()
	return &buf
}

func TestDiffNetworksFromTar(t *testing.T) {
	dirs := allExampleDirs(t)
	if len(dirs) < 2 {
		t.Skip("need at least 2 example directories for cross-diff")
	}

	// Diff between two different example networks
	oldBuf := buildTarFromDir(t, dirs[0])
	newBuf := buildTarFromDir(t, dirs[1])

	result, err := DiffNetworksFromTar(oldBuf, newBuf)
	if err != nil {
		t.Fatalf("diff: %v", err)
	}

	// Different networks should have changes
	if !result.HasChanges() {
		t.Error("expected changes between different networks")
	}

	// Self-diff should have no changes
	selfOld := buildTarFromDir(t, dirs[0])
	selfNew := buildTarFromDir(t, dirs[0])
	selfResult, err := DiffNetworksFromTar(selfOld, selfNew)
	if err != nil {
		t.Fatalf("self-diff: %v", err)
	}
	if selfResult.HasChanges() {
		t.Errorf("self-diff should have no changes, got creates=%d updates=%d deletes=%d",
			len(selfResult.Creates()), len(selfResult.Updates()), len(selfResult.Deletes()))
	}
}

func TestWriteNetworkToTar(t *testing.T) {
	for _, dir := range allExampleDirs(t) {
		name := filepath.Base(dir)
		t.Run(name, func(t *testing.T) {
			buf := buildTarFromDir(t, dir)
			net, err := LoadNetworkFromTar(buf)
			if err != nil {
				t.Fatalf("load: %v", err)
			}

			// Write to tar
			var outBuf bytes.Buffer
			if err := WriteNetworkToTar(net, &outBuf); err != nil {
				t.Fatalf("write: %v", err)
			}

			// Verify CHECKSUM is included and valid
			outCopy := make([]byte, outBuf.Len())
			copy(outCopy, outBuf.Bytes())
			ok, errs := VerifyChecksumFromTar(bytes.NewReader(outCopy))
			if !ok {
				t.Errorf("CHECKSUM verify should pass after WriteNetworkToTar, got errors: %v", errs)
			}

			// Re-load from the written tar (round-trip)
			net2, err := LoadNetworkFromTar(&outBuf)
			if err != nil {
				t.Fatalf("re-load: %v", err)
			}

			if net2.Root.Frontmatter.ID != net.Root.Frontmatter.ID {
				t.Errorf("id mismatch: original=%q round-trip=%q",
					net.Root.Frontmatter.ID, net2.Root.Frontmatter.ID)
			}
			if len(net2.AllObjects()) != len(net.AllObjects()) {
				t.Errorf("objects: original=%d round-trip=%d",
					len(net.AllObjects()), len(net2.AllObjects()))
			}
			if len(net2.AllRelations()) != len(net.AllRelations()) {
				t.Errorf("relations: original=%d round-trip=%d",
					len(net.AllRelations()), len(net2.AllRelations()))
			}
			if len(net2.AllActions()) != len(net.AllActions()) {
				t.Errorf("actions: original=%d round-trip=%d",
					len(net.AllActions()), len(net2.AllActions()))
			}
			if len(net2.AllRisks()) != len(net.AllRisks()) {
				t.Errorf("risks: original=%d round-trip=%d",
					len(net.AllRisks()), len(net2.AllRisks()))
			}
		})
	}
}
