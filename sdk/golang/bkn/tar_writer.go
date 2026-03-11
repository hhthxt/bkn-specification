// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package bkn

import (
	"archive/tar"
	"fmt"
	"io"
	"strings"
	"time"
)

// WriteNetworkToTar serializes a BknDocument to a tar stream.
// The document is written as:
// - network.bkn (frontmatter only)
// - object_types/*.bkn for each ObjectType
// - relation_types/*.bkn for each RelationType
// - action_types/*.bkn for each ActionType
// - risk_types/*.bkn for each RiskType
// - concept_groups/*.bkn for each ConceptGroup
// - SKILL.md (auto-generated)
// - CHECKSUM (auto-generated)
func WriteNetworkToTar(doc *BknNetwork, w io.Writer) error {
	tw := tar.NewWriter(w)
	defer tw.Close()

	now := time.Now()
	mfs := NewMemoryFileSystem()

	// Write network.bkn (frontmatter only)
	rootContent := serializeFrontmatter(doc.BknNetworkFrontmatter)
	mfs.AddFile("network.bkn", []byte(rootContent))
	if err := writeTarEntry(tw, "network.bkn", []byte(rootContent), now); err != nil {
		return err
	}

	// Write ObjectTypes
	for _, ot := range doc.ObjectTypes {
		content := serializeObjectType(ot)
		path := "object_types/" + ot.ID + ".bkn"
		mfs.AddFile(path, []byte(content))
		if err := writeTarEntry(tw, path, []byte(content), now); err != nil {
			return err
		}
	}

	// Write RelationTypes
	for _, rt := range doc.RelationTypes {
		content := serializeRelationType(rt)
		path := "relation_types/" + rt.ID + ".bkn"
		mfs.AddFile(path, []byte(content))
		if err := writeTarEntry(tw, path, []byte(content), now); err != nil {
			return err
		}
	}

	// Write ActionTypes
	for _, at := range doc.ActionTypes {
		content := serializeActionType(at)
		path := "action_types/" + at.ID + ".bkn"
		mfs.AddFile(path, []byte(content))
		if err := writeTarEntry(tw, path, []byte(content), now); err != nil {
			return err
		}
	}

	// Write RiskTypes
	for _, rt := range doc.RiskTypes {
		content := serializeRiskType(rt)
		path := "risk_types/" + rt.ID + ".bkn"
		mfs.AddFile(path, []byte(content))
		if err := writeTarEntry(tw, path, []byte(content), now); err != nil {
			return err
		}
	}

	// Write ConceptGroups
	for _, cg := range doc.ConceptGroups {
		content := serializeConceptGroup(cg)
		path := "concept_groups/" + cg.ID + ".bkn"
		mfs.AddFile(path, []byte(content))
		if err := writeTarEntry(tw, path, []byte(content), now); err != nil {
			return err
		}
	}

	// Generate and write SKILL.md
	skillContent := generateSkillMd(doc)
	mfs.AddFile("SKILL.md", []byte(skillContent))
	if err := writeTarEntry(tw, "SKILL.md", []byte(skillContent), now); err != nil {
		return err
	}

	// Generate and write CHECKSUM
	checksumContent, err := GenerateChecksumFileWithFS(mfs, ".")
	if err != nil {
		return fmt.Errorf("failed to generate checksum: %w", err)
	}
	if err := writeTarEntry(tw, ChecksumFileName, []byte(checksumContent), now); err != nil {
		return err
	}

	return nil
}

// serializeFrontmatter serializes BknNetworkFrontmatter to YAML frontmatter string
func serializeFrontmatter(fm BknNetworkFrontmatter) string {
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.WriteString("type: network\n")
	sb.WriteString(fmt.Sprintf("id: %s\n", fm.ID))
	sb.WriteString(fmt.Sprintf("name: %s\n", fm.Name))
	sb.WriteString(fmt.Sprintf("tags: [%s]\n", strings.Join(fm.Tags, ", ")))
	sb.WriteString(fmt.Sprintf("description: %s\n", fm.Description))

	if fm.Version != "" {
		sb.WriteString(fmt.Sprintf("version: %s\n", fm.Version))
	}
	if fm.Branch != "" {
		sb.WriteString(fmt.Sprintf("branch: %s\n", fm.Branch))
	}
	if fm.BusinessDomain != "" {
		sb.WriteString(fmt.Sprintf("business_domain: %s\n", fm.BusinessDomain))
	}
	sb.WriteString("---\n")
	return sb.String()
}

// serializeObjectType serializes BknObjectType to BKN format
func serializeObjectType(ot *BknObjectType) string {
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.WriteString("type: object_type\n")
	sb.WriteString(fmt.Sprintf("id: %s\n", ot.ID))
	sb.WriteString(fmt.Sprintf("name: %s\n", ot.Name))
	sb.WriteString(fmt.Sprintf("tags: [%s]\n", strings.Join(ot.Tags, ", ")))
	sb.WriteString(fmt.Sprintf("description: %s\n", ot.Description))
	sb.WriteString("---\n\n")

	// Data Source
	sb.WriteString("### Data Source\n\n")
	sb.WriteString("| Type | ID | Name |\n")
	sb.WriteString("|------|-----|------|\n")
	if ot.DataSource != nil {
		sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n",
			ot.DataSource.Type, ot.DataSource.ID, ot.DataSource.Name))
	}
	sb.WriteString("\n")

	// Data Properties
	sb.WriteString("### Data Properties\n\n")
	sb.WriteString("| Property | Display Name | Type | Description |\n")
	sb.WriteString("|----------|--------------|------|-------------|\n")
	if len(ot.DataProperties) > 0 {
		for _, dp := range ot.DataProperties {
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
				dp.Name, dp.DisplayName, dp.Type, dp.Description))
		}
	}
	sb.WriteString("\n")

	// Logic Properties
	sb.WriteString("### Logic Properties\n\n")
	for _, lp := range ot.LogicProperties {
		sb.WriteString(fmt.Sprintf("#### %s\n\n", lp.Name))
		if lp.Type != "" {
			sb.WriteString(fmt.Sprintf("- **Type**: %s\n", lp.Type))
		}
		if lp.DataSource != nil {
			sb.WriteString(fmt.Sprintf("- **Source**: %s (%s)\n", lp.DataSource.Type, lp.DataSource.Name))
		}
		if lp.Description != "" {
			sb.WriteString(fmt.Sprintf("- **Description**: %s\n", lp.Description))
		}
		if len(lp.Parameters) > 0 {
			sb.WriteString("\n| Parameter | Type | Source | Binding | Description |\n")
			sb.WriteString("|-----------|------|--------|---------|-------------|\n")
			for _, p := range lp.Parameters {
				sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
					p.Name, p.Type, p.Source, p.Operation, p.Description))
			}
		}
	}
	sb.WriteString("\n")

	// Keys section
	sb.WriteString("### Keys\n\n")
	sb.WriteString(fmt.Sprintf("Primary Key: %s\n", strings.Join(ot.PrimaryKeys, ", ")))
	sb.WriteString(fmt.Sprintf("Display Key: %s\n", ot.DisplayKey))
	sb.WriteString(fmt.Sprintf("Incremental Key: %s\n", ot.IncrementalKey))
	sb.WriteString("\n")

	return sb.String()
}

// serializeRelationType serializes BknRelationType to BKN format
func serializeRelationType(rt *BknRelationType) string {
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.WriteString("type: relation_type\n")
	sb.WriteString(fmt.Sprintf("id: %s\n", rt.ID))
	sb.WriteString(fmt.Sprintf("name: %s\n", rt.Name))
	sb.WriteString(fmt.Sprintf("tags: [%s]\n", strings.Join(rt.Tags, ", ")))
	sb.WriteString(fmt.Sprintf("description: %s\n", rt.Description))
	sb.WriteString(fmt.Sprintf("source_object_type: %s\n", rt.SourceObjectTypeID))
	sb.WriteString(fmt.Sprintf("target_object_type: %s\n", rt.TargetObjectTypeID))
	sb.WriteString("---\n\n")

	// Mapping Rules
	sb.WriteString("### Mapping Rules\n\n")
	if rt.RelationType != "" {
		sb.WriteString(fmt.Sprintf("type: %s\n\n", rt.RelationType))
	}

	return sb.String()
}

// serializeActionType serializes BknActionType to BKN format
func serializeActionType(at *BknActionType) string {
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.WriteString("type: action_type\n")
	sb.WriteString(fmt.Sprintf("id: %s\n", at.ID))
	sb.WriteString(fmt.Sprintf("name: %s\n", at.Name))
	sb.WriteString(fmt.Sprintf("tags: [%s]\n", strings.Join(at.Tags, ", ")))
	sb.WriteString(fmt.Sprintf("description: %s\n", at.Description))
	sb.WriteString(fmt.Sprintf("action_type: %s\n", at.ActionType))
	sb.WriteString(fmt.Sprintf("enabled: %v\n", at.Enabled))
	sb.WriteString(fmt.Sprintf("risk_level: %s\n", at.RiskLevel))
	sb.WriteString(fmt.Sprintf("requires_approval: %v\n", at.RequiresApproval))
	sb.WriteString("---\n\n")

	// Bound Object
	sb.WriteString("### Bound Object\n\n")
	sb.WriteString("| Bound Object | Action Type |\n")
	sb.WriteString("|--------------|-------------|\n")
	if at.ObjectTypeID != "" {
		sb.WriteString(fmt.Sprintf("| %s | %s |\n", at.ObjectTypeID, at.ActionType))
	}
	sb.WriteString("\n")

	// Parameter Binding
	sb.WriteString("### Parameter Binding\n\n")
	sb.WriteString("| Parameter | Type | Source | Binding | Description |\n")
	sb.WriteString("|-----------|------|--------|---------|-------------|\n")
	for _, p := range at.Parameters {
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
			p.Name, p.Type, p.Source, p.ValueFrom, p.Description))
	}
	sb.WriteString("\n")

	// Schedule
	sb.WriteString("### Schedule\n\n")
	sb.WriteString("| Type | Expression |\n")
	sb.WriteString("|------|------------|\n")
	if at.Schedule != nil && at.Schedule.Type != "" {
		sb.WriteString(fmt.Sprintf("| %s | %s |\n", at.Schedule.Type, at.Schedule.Expression))
	}
	sb.WriteString("\n")

	return sb.String()
}

// serializeRiskType serializes BknRiskType to BKN format
func serializeRiskType(rt *BknRiskType) string {
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.WriteString("type: risk_type\n")
	sb.WriteString(fmt.Sprintf("id: %s\n", rt.ID))
	sb.WriteString(fmt.Sprintf("name: %s\n", rt.Name))
	sb.WriteString(fmt.Sprintf("tags: [%s]\n", strings.Join(rt.Tags, ", ")))
	sb.WriteString(fmt.Sprintf("description: %s\n", rt.Description))
	sb.WriteString("---\n\n")

	sb.WriteString("### Control Scope\n\n")
	if rt.ControlScope != "" {
		sb.WriteString(rt.ControlScope)
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	sb.WriteString("### Control Policy\n\n")
	if rt.ControlPolicy != "" {
		sb.WriteString(rt.ControlPolicy)
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	sb.WriteString("### Rollback Plan\n\n")
	if rt.RollbackPlan != "" {
		sb.WriteString(rt.RollbackPlan)
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	sb.WriteString("### Audit Requirements\n\n")
	if rt.AuditRequirements != "" {
		sb.WriteString(rt.AuditRequirements)
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	return sb.String()
}

// serializeConceptGroup serializes BknConceptGroup to BKN format
func serializeConceptGroup(cg *BknConceptGroup) string {
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.WriteString("type: concept_group\n")
	sb.WriteString(fmt.Sprintf("id: %s\n", cg.ID))
	sb.WriteString(fmt.Sprintf("name: %s\n", cg.Name))
	sb.WriteString(fmt.Sprintf("tags: [%s]\n", strings.Join(cg.Tags, ", ")))
	sb.WriteString(fmt.Sprintf("description: %s\n", cg.Description))
	sb.WriteString("---\n\n")

	return sb.String()
}

// generateSkillMd generates SKILL.md content from BknNetwork.
func generateSkillMd(doc *BknNetwork) string {
	fm := doc.BknNetworkFrontmatter

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

	// Objects table
	if len(doc.ObjectTypes) > 0 {
		sb.WriteString("### 核心对象\n\n")
		sb.WriteString("| 对象 | 文件路径 | 说明 |\n")
		sb.WriteString("|------|----------|------|\n")
		for _, ot := range doc.ObjectTypes {
			path := "object_types/" + ot.ID + ".bkn"
			sb.WriteString(fmt.Sprintf("| %s | `%s` | %s |\n", ot.Name, path, ot.Description))
		}
		sb.WriteString("\n")
	}

	// Relations table
	if len(doc.RelationTypes) > 0 {
		sb.WriteString("### 核心关系\n\n")
		sb.WriteString("| 关系 | 文件路径 | 说明 |\n")
		sb.WriteString("|------|----------|------|\n")
		for _, rt := range doc.RelationTypes {
			path := "relation_types/" + rt.ID + ".bkn"
			sb.WriteString(fmt.Sprintf("| %s | `%s` | %s |\n", rt.Name, path, rt.Description))
		}
		sb.WriteString("\n")
	}

	// Actions table
	if len(doc.ActionTypes) > 0 {
		sb.WriteString("### 可用行动\n\n")
		sb.WriteString("| 行动 | 文件路径 | 说明 |\n")
		sb.WriteString("|------|----------|------|\n")
		for _, at := range doc.ActionTypes {
			path := "action_types/" + at.ID + ".bkn"
			sb.WriteString(fmt.Sprintf("| %s | `%s` | %s |\n", at.Name, path, at.Description))
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
	if len(doc.ObjectTypes) > 0 {
		sb.WriteString("├── object_types/\n")
	}
	if len(doc.RelationTypes) > 0 {
		sb.WriteString("├── relation_types/\n")
	}
	if len(doc.ActionTypes) > 0 {
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
	if len(doc.ActionTypes) > 0 {
		sb.WriteString("### 运维场景\n\n")
		sb.WriteString("1. **执行运维操作**\n")
		sb.WriteString("   - 查看 `action_types/` 目录下的行动定义\n")
		sb.WriteString("   - 了解触发条件和参数绑定\n\n")
	}

	// Index tables
	sb.WriteString("## 索引表\n\n")
	sb.WriteString("### 按类型索引\n\n")
	if len(doc.ObjectTypes) > 0 {
		sb.WriteString("- **对象定义**: `object_types/`\n")
	}
	if len(doc.RelationTypes) > 0 {
		sb.WriteString("- **关系定义**: `relation_types/`\n")
	}
	if len(doc.ActionTypes) > 0 {
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
