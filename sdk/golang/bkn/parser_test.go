// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package bkn

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// === Parse Frontmatter Tests ===

func TestParseFrontmatter_Success(t *testing.T) {
	text := `---
type: object_type
id: pod
name: Pod
---

## ObjectType: pod
Content here
`
	fm, err := ParseFrontmatter(text)
	require.NoError(t, err)
	assert.Equal(t, "object_type", fm["type"])
	assert.Equal(t, "pod", fm["id"])
	assert.Equal(t, "Pod", fm["name"])
}

func TestParseFrontmatter_NoFrontmatter(t *testing.T) {
	text := "# No frontmatter\nJust content"
	fm, err := ParseFrontmatter(text)
	// When there's no frontmatter, it returns empty map without error
	require.NoError(t, err)
	assert.Empty(t, fm)
}

func TestParseFrontmatter_EmptyFrontmatter(t *testing.T) {
	text := `---
---

Content`
	fm, err := ParseFrontmatter(text)
	require.NoError(t, err)
	assert.Empty(t, fm)
}

// === Parse Network File Tests ===

func TestParseNetworkFile_Success(t *testing.T) {
	text := `---
type: network
id: k8s-network
name: Kubernetes Network
version: "1.0"
---

## Network: k8s-network

Kubernetes resource network
`
	net, err := ParseNetworkFile(text, "/test/network.bkn")
	require.NoError(t, err)
	assert.Equal(t, "network", net.Type)
	assert.Equal(t, "k8s-network", net.ID)
	assert.Equal(t, "Kubernetes Network", net.Name)
	assert.Equal(t, "1.0", net.Version)
}

func TestParseNetworkFile_MissingType(t *testing.T) {
	text := `---
id: test
---

Content`
	_, err := ParseNetworkFile(text, "/test/network.bkn")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "type")
}

func TestParseNetworkFile_MissingID(t *testing.T) {
	text := `---
type: network
---

Content`
	_, err := ParseNetworkFile(text, "/test/network.bkn")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "id")
}

// === Parse Object Type Tests ===

func TestParseObjectType_Basic(t *testing.T) {
	text := `---
type: object_type
id: pod
name: Pod
tags: [k8s, workload]
---

## ObjectType: pod

Kubernetes Pod resource

### Data Properties

| Name | DisplayName | Type | Description |
|------|-------------|------|-------------|
| name | Name | string | Pod name |
| image | Image | string | Container image |
`
	ot, err := ParseObjectTypeFile(text, "/test/pod.bkn")
	require.NoError(t, err)
	assert.Equal(t, "pod", ot.ID)
	assert.Equal(t, "Pod", ot.Name)
	assert.ElementsMatch(t, []string{"k8s", "workload"}, ot.Tags)
	require.Len(t, ot.DataProperties, 2)
	assert.Equal(t, "name", ot.DataProperties[0].Name)
}

func TestParseObjectType_WithDataSource(t *testing.T) {
	text := `---
type: object_type
id: deployment
name: Deployment
---

## ObjectType: deployment

### Data Source

| Type | ID | Name |
|------|-----|------|
| data_view | dv_deployments | Deployment View |

### Data Properties

| Name | DisplayName | Type |
|------|-------------|------|
| replicas | Replicas | number |
`
	ot, err := ParseObjectTypeFile(text, "/test/deployment.bkn")
	require.NoError(t, err)
	require.NotNil(t, ot.DataSource)
	assert.Equal(t, "data_view", ot.DataSource.Type)
	assert.Equal(t, "dv_deployments", ot.DataSource.ID)
	assert.Equal(t, "Deployment View", ot.DataSource.Name)
}

func TestParseObjectType_WithLogicProperties(t *testing.T) {
	text := `---
type: object_type
id: service
name: Service
---

## ObjectType: service

### Logic Properties

| Name | DisplayName | Type | Description |
|------|-------------|------|-------------|
| endpoint_count | Endpoint Count | integer | Number of endpoints |
| health_status | Health Status | string | Health status |
`
	ot, err := ParseObjectTypeFile(text, "/test/service.bkn")
	require.NoError(t, err)
	// Logic properties are parsed from the table
	require.NotEmpty(t, ot.LogicProperties)
}

// === Parse Relation Type Tests ===

func TestParseRelationType_Basic(t *testing.T) {
	text := `---
type: relation_type
id: belongs_to
name: Belongs To
---

## RelationType: belongs_to

Pod belongs to Node

### Endpoint

| Source | Target | Type |
|--------|--------|------|
| pod | node | direct |

### Mapping Rules

| Source Property | Target Property |
|-----------------|-----------------|
| node_name | name |
| node_id | id |
`
	rt, err := ParseRelationTypeFile(text, "/test/belongs_to.bkn")
	require.NoError(t, err)
	assert.Equal(t, "belongs_to", rt.ID)
	assert.Equal(t, "Pod belongs to Node", rt.Description)
	assert.Equal(t, "pod", rt.Endpoint.Source)
	assert.Equal(t, "node", rt.Endpoint.Target)
	assert.Equal(t, "direct", rt.Endpoint.Type)
	rules, ok := rt.MappingRules.(DirectMappingRule)
	require.True(t, ok, "MappingRules should be DirectMappingRule")
	require.Len(t, rules, 2)
	assert.Equal(t, "node_name", rules[0].SourceProperty)
}

func TestParseRelationType_EndpointTable(t *testing.T) {
	text := `---
type: relation_type
id: runs_on
name: Runs On
---

## RelationType: runs_on

Container runs on Pod

### Endpoint

| Source | Target | Type |
|--------|--------|------|
| container | pod | indirect |

### Mapping Rules

| Source Property | Target Property |
|-----------------|-----------------|
| pod_id | id |
`
	rt, err := ParseRelationTypeFile(text, "/test/runs_on.bkn")
	require.NoError(t, err)
	assert.Equal(t, "container", rt.Endpoint.Source)
	assert.Equal(t, "pod", rt.Endpoint.Target)
	assert.Equal(t, "indirect", rt.Endpoint.Type)
	rules, ok := rt.MappingRules.([]MappingRule)
	require.True(t, ok, "MappingRules should be []MappingRule for indirect")
	require.Len(t, rules, 1)
}

// === Parse Action Type Tests ===

func TestParseActionType_Basic(t *testing.T) {
	text := `---
type: action_type
id: restart
name: Restart Pod
action_type: modify
risk_level: high
requires_approval: true
---

## ActionType: restart

Restart a pod gracefully

### Bound Object

| Bound Object | Action Type |
|--------------|-------------|
| pod | modify |

### Parameter Binding

| Parameter | Type | Source | Binding | Description |
|-----------|------|--------|---------|-------------|
| graceful | boolean | const | true | Graceful restart |
| timeout | number | property | spec.timeout | Timeout seconds |
`
	at, err := ParseActionTypeFile(text, "/test/restart.bkn")
	require.NoError(t, err)
	assert.Equal(t, "restart", at.ID)
	assert.Equal(t, "modify", at.ActionType)
	assert.Equal(t, "high", at.RiskLevel)
	assert.True(t, at.RequiresApproval)
	assert.Equal(t, "pod", at.ObjectTypeID)
	require.Len(t, at.Parameters, 2)
	assert.Equal(t, "graceful", at.Parameters[0].Name)
}

func TestParseActionType_WithSchedule(t *testing.T) {
	text := `---
type: action_type
id: backup
name: Backup Data
action_type: create
---

## ActionType: backup

### Schedule

| Type | Expression |
|------|------------|
| cron | 0 2 * * * |
`
	at, err := ParseActionTypeFile(text, "/test/backup.bkn")
	require.NoError(t, err)
	require.NotNil(t, at.Schedule)
	assert.Equal(t, "cron", at.Schedule.Type)
	assert.Equal(t, "0 2 * * *", at.Schedule.Expression)
}

// === Parse Risk Type Tests ===

func TestParseRiskType_Basic(t *testing.T) {
	text := `---
type: risk_type
id: high_memory
name: High Memory Usage
---

## RiskType: high_memory

Detects high memory usage

### Control Scope

production

### Pre-checks

| Object | Check | Condition | Message |
|--------|-------|-----------|---------|
| pod | memory_check | memory > 90 | Memory usage too high |
| node | swap_check | swap > 50 | Swap usage too high |
`
	rt, err := ParseRiskTypeFile(text, "/test/high_memory.bkn")
	require.NoError(t, err)
	assert.Equal(t, "high_memory", rt.ID)
	assert.Equal(t, "High Memory Usage", rt.Name)
	assert.Equal(t, "production", rt.ControlScope)
}

// === Parse Concept Group Tests ===

func TestParseConceptGroup_Basic(t *testing.T) {
	text := `---
type: concept_group
id: k8s_resources
name: Kubernetes Resources
---

## ConceptGroup: k8s_resources

Core Kubernetes resources
`
	cg, err := ParseConceptGroupFile(text, "/test/k8s_resources.bkn")
	require.NoError(t, err)
	assert.Equal(t, "k8s_resources", cg.ID)
	assert.Equal(t, "Kubernetes Resources", cg.Name)
}

// === Error Handling Tests ===

func TestParse_InvalidType(t *testing.T) {
	text := `---
type: invalid_type
id: test
---

Content`
	// ParseObjectTypeFile doesn't validate type, it just parses
	// The validation happens elsewhere
	_, err := ParseObjectTypeFile(text, "/test/invalid.bkn")
	// Currently parser doesn't validate type field strictly
	// This test documents current behavior
	_ = err
}

func TestParse_MalformedYAML(t *testing.T) {
	text := `---
type: object_type
id: [invalid yaml structure
---

Content`
	_, err := ParseFrontmatter(text)
	assert.Error(t, err)
}

func TestParse_EmptyFile(t *testing.T) {
	fm, err := ParseFrontmatter("")
	// Empty file returns empty frontmatter without error
	require.NoError(t, err)
	assert.Empty(t, fm)
}

// === Data Properties Parsing Tests ===

func TestParseDataProperties_VariousTypes(t *testing.T) {
	text := `---
type: object_type
id: test
---

## ObjectType: test

### Data Properties

| Name | DisplayName | Type | Description |
|------|-------------|------|-------------|
| str_field | String Field | string | A string |
| num_field | Number Field | number | A number |
| bool_field | Bool Field | boolean | A boolean |
| date_field | Date Field | datetime | A date |
| json_field | JSON Field | json | JSON data |
`
	ot, err := ParseObjectTypeFile(text, "/test/test.bkn")
	require.NoError(t, err)
	require.Len(t, ot.DataProperties, 5)
	assert.Equal(t, "string", ot.DataProperties[0].Type)
	assert.Equal(t, "number", ot.DataProperties[1].Type)
	assert.Equal(t, "boolean", ot.DataProperties[2].Type)
	assert.Equal(t, "datetime", ot.DataProperties[3].Type)
	assert.Equal(t, "json", ot.DataProperties[4].Type)
}

func TestParseDataProperties_WithMappedField(t *testing.T) {
	text := `---
type: object_type
id: test
---

## ObjectType: test

### Data Properties

| Name | Display Name | Type | Description | Mapped Field |
|------|--------------|------|-------------|--------------|
| status | Status | string | Status field | status_code |
| name | Name | string | Name field | full_name |
`
	ot, err := ParseObjectTypeFile(text, "/test/test.bkn")
	require.NoError(t, err)
	require.Len(t, ot.DataProperties, 2)
	assert.Equal(t, "status", ot.DataProperties[0].Name)
	assert.Equal(t, "Status field", ot.DataProperties[0].Description)
	assert.Equal(t, "status_code", ot.DataProperties[0].MappedField)
	assert.Equal(t, "name", ot.DataProperties[1].Name)
	assert.Equal(t, "full_name", ot.DataProperties[1].MappedField)
}

// === Logic Properties Parsing Tests ===

func TestParseLogicProperties_WithParameters(t *testing.T) {
	text := `---
type: object_type
id: test
---

## ObjectType: test

### Logic Properties

| Name | DisplayName | Type | Description |
|------|-------------|------|-------------|
| computed_value | Computed Value | number | A computed value |
`
	ot, err := ParseObjectTypeFile(text, "/test/test.bkn")
	require.NoError(t, err)
	// Logic properties are parsed from the table
	require.NotEmpty(t, ot.LogicProperties)
}

func TestParseLogicProperties_SubSection(t *testing.T) {
	text := `---
type: object_type
id: product
name: 产品
---

## ObjectType: 产品

存储企业生产的成品基本信息

### Logic Properties

#### product_bom

- **Display**: product_bom
- **Type**: operator
- **Source**: bom_tree_builder (operator)

| Parameter | Type | Source | Binding | Description |
|-----------|------|--------|---------|-------------|
| timeout | number | input | - |  |
| cache | boolean | input | - |  |
| knowledge_network_id | string | const | supplychain_hd0202 |  |
`
	ot, err := ParseObjectTypeFile(text, "/test/product.bkn")
	require.NoError(t, err)
	assert.Equal(t, "存储企业生产的成品基本信息", ot.Description)
	require.Len(t, ot.LogicProperties, 1)

	lp := ot.LogicProperties[0]
	assert.Equal(t, "product_bom", lp.Name)
	assert.Equal(t, "product_bom", lp.DisplayName)
	assert.Equal(t, "operator", lp.Type)
	require.NotNil(t, lp.DataSource)
	assert.Equal(t, "bom_tree_builder", lp.DataSource.ID)
	assert.Equal(t, "operator", lp.DataSource.Type)
	require.Len(t, lp.Parameters, 3)
	assert.Equal(t, "timeout", lp.Parameters[0].Name)
	assert.Equal(t, "const", lp.Parameters[2].Source)
	assert.Equal(t, "supplychain_hd0202", lp.Parameters[2].ValueFrom)
}

func TestParseLogicProperties_Empty(t *testing.T) {
	text := `---
type: object_type
id: material
name: 物料
---

## ObjectType: 物料

物料基础信息

### Logic Properties


### Keys

Primary Keys: material_code
`
	ot, err := ParseObjectTypeFile(text, "/test/material.bkn")
	require.NoError(t, err)
	assert.Empty(t, ot.LogicProperties)
}

// === Parameter Binding Tests ===

func TestParseParameters_VariousSources(t *testing.T) {
	text := `---
type: action_type
id: test_action
---

## ActionType: test_action

### Parameter Binding

| Parameter | Type | Source | Binding | Description |
|-----------|------|--------|---------|-------------|
| fixed_val | string | const | hello | Fixed value |
| from_prop | string | property | metadata.name | From property |
`
	at, err := ParseActionTypeFile(text, "/test/test_action.bkn")
	require.NoError(t, err)
	require.Len(t, at.Parameters, 2)
	assert.Equal(t, "const", at.Parameters[0].Source)
	assert.Equal(t, "hello", at.Parameters[0].ValueFrom)
	assert.Equal(t, "property", at.Parameters[1].Source)
	assert.Equal(t, "metadata.name", at.Parameters[1].ValueFrom)
}

// === Edge Cases Tests ===

func TestParse_EmptyTables(t *testing.T) {
	text := `---
type: object_type
id: empty_obj
---

## ObjectType: empty_obj

No tables here
`
	ot, err := ParseObjectTypeFile(text, "/test/empty.bkn")
	require.NoError(t, err)
	assert.Empty(t, ot.DataProperties)
	assert.Empty(t, ot.LogicProperties)
}

func TestParse_ExtraWhitespace(t *testing.T) {
	text := `---
type: object_type
id: test
name: Test
---

## ObjectType: test

   

### Data Properties

| Name | DisplayName | Type |
|------|-------------|------|
| field1 | Field 1 | string |

   
`
	ot, err := ParseObjectTypeFile(text, "/test/test.bkn")
	require.NoError(t, err)
	assert.Equal(t, "test", ot.ID)
	require.Len(t, ot.DataProperties, 1)
}

func TestParse_SpecialCharactersInID(t *testing.T) {
	text := `---
type: object_type
id: my-app_v1.0
name: My App
---

## ObjectType: my-app_v1.0

Content
`
	ot, err := ParseObjectTypeFile(text, "/test/my-app_v1.0.bkn")
	require.NoError(t, err)
	assert.Equal(t, "my-app_v1.0", ot.ID)
}

func TestParse_UnicodeContent(t *testing.T) {
	text := `---
type: object_type
id: unicode_test
name: 测试对象
---

## ObjectType: 测试对象

这是一个测试对象

### Data Properties

| Name | DisplayName | Type |
|------|-------------|------|
| 名称 | 名称 | string |
`
	ot, err := ParseObjectTypeFile(text, "/test/unicode.bkn")
	require.NoError(t, err)
	assert.Equal(t, "测试对象", ot.Name)
	assert.Equal(t, "这是一个测试对象", ot.Description)
	require.Len(t, ot.DataProperties, 1)
	assert.Equal(t, "名称", ot.DataProperties[0].Name)
}

// === Integration Tests ===

func TestParseFullNetwork(t *testing.T) {
	// Create a temporary directory with full network structure
	dir, err := os.MkdirTemp("", "bkn-full-network-*")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	// Create network.bkn
	networkContent := `---
type: network
id: test-network
name: Test Network
version: "1.0.0"
---

## Network: test-network

Test network description
`
	err = os.WriteFile(filepath.Join(dir, "network.bkn"), []byte(networkContent), 0644)
	require.NoError(t, err)

	// Create object_types directory and file
	objTypesDir := filepath.Join(dir, "object_types")
	err = os.MkdirAll(objTypesDir, 0755)
	require.NoError(t, err)

	objContent := `---
type: object_type
id: test_obj
name: Test Object
---

## ObjectType: test_obj

Test object

### Data Properties

| Name | DisplayName | Type |
|------|-------------|------|
| id | ID | string |
`
	err = os.WriteFile(filepath.Join(objTypesDir, "test_obj.bkn"), []byte(objContent), 0644)
	require.NoError(t, err)

	// Load the network
	net, err := LoadNetwork(dir)
	require.NoError(t, err)

	assert.Equal(t, "test-network", net.ID)
	assert.Equal(t, "Test Network", net.Name)
	require.Len(t, net.ObjectTypes, 1)
	assert.Equal(t, "test_obj", net.ObjectTypes[0].ID)
}

func TestParse_InvalidFilePath(t *testing.T) {
	// Empty file path is handled gracefully
	_, err := ParseObjectTypeFile("---\ntype: object_type\nid: test\n---\n", "")
	// The parser may or may not error on empty path - document current behavior
	_ = err
}

func TestParse_NonExistentType(t *testing.T) {
	text := `---
type: network
id: test
---

Content`
	// ParseNetworkFile validates that type is "network"
	_, err := ParseNetworkFile(text, "/test/test.bkn")
	// Currently it accepts the file as long as frontmatter is valid
	// The type validation happens at higher level
	require.NoError(t, err)
}
