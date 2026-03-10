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
)

// LoadNetworkFromTar 从 tar 包直接加载 BKN 网络
// 无需写入本地文件系统，完全在内存中处理
func LoadNetworkFromTar(tarReader io.Reader) (*BknNetwork, error) {
	// 1. 解压 tar 包到内存文件系统
	mfs, rootFile, err := ExtractTarToMemory(tarReader)
	if err != nil {
		return nil, fmt.Errorf("failed to extract tar: %w", err)
	}

	// 2. 使用内存文件系统加载网络
	return LoadNetworkWithFS(mfs, rootFile)
}

// ExtractTarToMemory 将 tar 包解压到内存文件系统
// 返回内存文件系统和根文件路径
func ExtractTarToMemory(reader io.Reader) (*MemoryFileSystem, string, error) {
	mfs := NewMemoryFileSystem()
	tr := tar.NewReader(reader)

	var rootFile string
	rootCandidates := []string{"network.bkn", "network.md", "index.bkn", "index.md"}

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, "", fmt.Errorf("failed to read tar header: %w", err)
		}

		// 跳过目录
		if header.Typeflag == tar.TypeDir {
			continue
		}

		// 只处理支持的文件类型（.bkn, .md）以及 CHECKSUM 和 SKILL.md
		ext := strings.ToLower(filepath.Ext(header.Name))
		base := filepath.Base(header.Name)
		if !supportedExtensions[ext] && base != checksumFilename && base != "SKILL.md" {
			continue
		}

		// 读取文件内容
		content := make([]byte, header.Size)
		if _, err := io.ReadFull(tr, content); err != nil {
			return nil, "", fmt.Errorf("failed to read file %s: %w", header.Name, err)
		}

		// 标准化路径（使用 / 作为分隔符）
		path := filepath.ToSlash(header.Name)
		mfs.AddFile(path, content)

		// 检查是否是根文件候选
		for _, candidate := range rootCandidates {
			if strings.EqualFold(base, candidate) {
				rootFile = path
				break
			}
		}
	}

	if rootFile == "" {
		// 如果没有找到标准根文件，尝试查找 type: network 的文件
		for path, content := range mfs.files {
			doc, err := Parse(string(content), path)
			if err != nil {
				continue
			}
			if strings.EqualFold(strings.TrimSpace(doc.Frontmatter.Type), "network") {
				rootFile = path
				break
			}
		}
	}

	if rootFile == "" {
		return nil, "", fmt.Errorf("no root network file found in tar")
	}

	return mfs, rootFile, nil
}
