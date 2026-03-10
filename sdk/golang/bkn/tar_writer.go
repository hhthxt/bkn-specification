// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package bkn

import (
	"archive/tar"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"
)

// typeToSubdir maps BKN frontmatter types to standard subdirectory names.
var typeToSubdir = map[string]string{
	"object_type":   "object_types",
	"relation_type": "relation_types",
	"action_type":   "action_types",
	"risk_type":     "risk_types",
}

// WriteNetworkToTar 将 BknNetwork 序列化为 tar 流写入 w。
// 根文档写为 network.bkn，子文档保持原始路径（若有 SourcePath）或按 type 写入对应子目录，
// 并自动生成 CHECKSUM 文件。
func WriteNetworkToTar(network *BknNetwork, w io.Writer) error {
	tw := tar.NewWriter(w)
	defer tw.Close()

	now := time.Now()
	mfs := NewMemoryFileSystem()

	// 确定根目录（用于计算子文档相对路径）
	rootPath := "network.bkn"
	rootDir := ""
	if network.Root.SourcePath != "" {
		rootDir = filepath.Dir(network.Root.SourcePath)
	}

	// 写入根文档
	rootContent := []byte(Serialize(&network.Root))
	mfs.AddFile(rootPath, rootContent)
	if err := writeTarEntry(tw, rootPath, rootContent, now); err != nil {
		return err
	}

	// 写入子文档
	for i := range network.Includes {
		doc := &network.Includes[i]
		content := []byte(Serialize(doc))
		path := docToTarPath(doc, rootDir)
		mfs.AddFile(path, content)
		if err := writeTarEntry(tw, path, content, now); err != nil {
			return err
		}
	}

	// 生成并写入 CHECKSUM
	checksumContent, err := GenerateChecksumFileWithFS(mfs, ".")
	if err != nil {
		return fmt.Errorf("failed to generate checksum: %w", err)
	}
	if err := writeTarEntry(tw, checksumFilename, []byte(checksumContent), now); err != nil {
		return err
	}

	return nil
}

// docToTarPath determines the tar entry path for a document.
// If the document has a SourcePath, compute relative path from rootDir.
// Otherwise fall back to type+id convention.
func docToTarPath(doc *BknDocument, rootDir string) string {
	if doc.SourcePath != "" && rootDir != "" {
		rel, err := filepath.Rel(rootDir, doc.SourcePath)
		if err == nil {
			return filepath.ToSlash(rel)
		}
	}

	docType := strings.ToLower(strings.TrimSpace(doc.Frontmatter.Type))
	id := doc.Frontmatter.ID
	if id == "" {
		id = "unnamed"
	}

	subdir, ok := typeToSubdir[docType]
	if !ok {
		subdir = docType
	}
	return subdir + "/" + id + ".bkn"
}

func writeTarEntry(tw *tar.Writer, name string, data []byte, modTime time.Time) error {
	header := &tar.Header{
		Name:    name,
		Size:    int64(len(data)),
		Mode:    0644,
		ModTime: modTime,
	}
	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	_, err := tw.Write(data)
	return err
}
