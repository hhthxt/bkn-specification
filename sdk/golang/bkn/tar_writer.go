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
// 自动生成 SKILL.md 和 CHECKSUM 文件。
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

	// 生成并写入 SKILL.md
	skillContent := generateSkillMd(network)
	mfs.AddFile("SKILL.md", []byte(skillContent))
	if err := writeTarEntry(tw, "SKILL.md", []byte(skillContent), now); err != nil {
		return err
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

// generateSkillMd generates SKILL.md content from BknNetwork.
func generateSkillMd(network *BknNetwork) string {
	fm := network.Root.Frontmatter

	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("# %s - Agent 使用指南\n\n", fm.Name))
	sb.WriteString(fmt.Sprintf("> **网络ID**: %s  \n", fm.ID))
	sb.WriteString(fmt.Sprintf("> **版本**: %s  \n", fm.Version))
	if len(fm.Tags) > 0 {
		sb.WriteString(fmt.Sprintf("> **标签**: %s  \n", strings.Join(fm.Tags, ", ")))
	}
	sb.WriteString("\n")

	// Overview
	sb.WriteString("## 网络概览\n\n")
	if fm.Description != "" {
		sb.WriteString(fm.Description)
		sb.WriteString("\n\n")
	}

	// Collect objects, relations, actions
	var objects, relations, actions []*BknDocument
	for i := range network.Includes {
		doc := &network.Includes[i]
		docType := strings.ToLower(strings.TrimSpace(doc.Frontmatter.Type))
		switch docType {
		case "object_type":
			objects = append(objects, doc)
		case "relation_type":
			relations = append(relations, doc)
		case "action_type":
			actions = append(actions, doc)
		}
	}

	// Objects table
	if len(objects) > 0 {
		sb.WriteString("### 核心对象\n\n")
		sb.WriteString("| 对象 | 文件路径 | 说明 |\n")
		sb.WriteString("|------|----------|------|\n")
		for _, obj := range objects {
			id := obj.Frontmatter.ID
			name := obj.Frontmatter.Name
			if name == "" {
				name = id
			}
			path := docToTarPath(obj, "")
			sb.WriteString(fmt.Sprintf("| %s | `%s` | %s |\n", name, path, id))
		}
		sb.WriteString("\n")
	}

	// Relations table
	if len(relations) > 0 {
		sb.WriteString("### 核心关系\n\n")
		sb.WriteString("| 关系 | 文件路径 | 说明 |\n")
		sb.WriteString("|------|----------|------|\n")
		for _, rel := range relations {
			id := rel.Frontmatter.ID
			name := rel.Frontmatter.Name
			if name == "" {
				name = id
			}
			path := docToTarPath(rel, "")
			sb.WriteString(fmt.Sprintf("| %s | `%s` | %s |\n", name, path, id))
		}
		sb.WriteString("\n")
	}

	// Actions table
	if len(actions) > 0 {
		sb.WriteString("### 可用行动\n\n")
		sb.WriteString("| 行动 | 文件路径 | 说明 |\n")
		sb.WriteString("|------|----------|------|\n")
		for _, act := range actions {
			id := act.Frontmatter.ID
			name := act.Frontmatter.Name
			if name == "" {
				name = id
			}
			path := docToTarPath(act, "")
			sb.WriteString(fmt.Sprintf("| %s | `%s` | %s |\n", name, path, id))
		}
		sb.WriteString("\n")
	}

	// Directory structure
	sb.WriteString("## 目录结构\n\n")
	sb.WriteString("```\n")
	sb.WriteString(".\n")
	sb.WriteString("├── network.bkn\n")
	sb.WriteString("├── SKILL.md\n")
	sb.WriteString("├── CHECKSUM\n")
	if len(objects) > 0 {
		sb.WriteString("├── object_types/\n")
	}
	if len(relations) > 0 {
		sb.WriteString("├── relation_types/\n")
	}
	if len(actions) > 0 {
		sb.WriteString("└── action_types/\n")
	}
	sb.WriteString("```\n\n")

	// Usage suggestions
	sb.WriteString("## 使用建议\n\n")
	sb.WriteString("### 查询场景\n\n")
	sb.WriteString("1. **获取所有对象定义**\n")
	sb.WriteString("   - 查看 `object_types/` 目录下的文件\n\n")
	sb.WriteString("2. **查找关系定义**\n")
	sb.WriteString("   - 查看 `relation_types/` 目录下的文件\n\n")
	if len(actions) > 0 {
		sb.WriteString("### 运维场景\n\n")
		sb.WriteString("1. **执行运维操作**\n")
		sb.WriteString("   - 查看 `action_types/` 目录下的行动定义\n")
		sb.WriteString("   - 了解触发条件和参数绑定\n\n")
	}

	// Index tables
	sb.WriteString("## 索引表\n\n")
	sb.WriteString("### 按类型索引\n\n")
	if len(objects) > 0 {
		sb.WriteString("- **对象定义**: `object_types/`\n")
	}
	if len(relations) > 0 {
		sb.WriteString("- **关系定义**: `relation_types/`\n")
	}
	if len(actions) > 0 {
		sb.WriteString("- **行动定义**: `action_types/`\n")
	}
	sb.WriteString("\n")

	sb.WriteString("## 注意事项\n\n")
	sb.WriteString("1. 本网络由 BKN SDK 自动生成 SKILL.md\n")
	sb.WriteString("2. 所有定义遵循 BKN 规范\n")
	sb.WriteString("3. 使用 CHECKSUM 文件验证网络完整性\n")

	return sb.String()
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
