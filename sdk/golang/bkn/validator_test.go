// Copyright The kweaver-ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package bkn

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testExamplesDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	// sdk/golang/bkn/validator_test.go -> ../../../examples
	return filepath.Join(filepath.Dir(file), "..", "..", "..", "examples")
}

func TestValidateNetwork_ValidK8s(t *testing.T) {
	dir := filepath.Join(testExamplesDir(t), "k8s-network")
	net, err := LoadNetwork(dir)
	require.NoError(t, err)
	res := ValidateNetwork(net)
	assert.True(t, res.OK(), "errors: %+v", res.Errors)
}

func TestValidateNetwork_ValidRestoreKnowledge(t *testing.T) {
	dir := filepath.Join(testExamplesDir(t), "restore-knowledge")
	net, err := LoadNetwork(dir)
	require.NoError(t, err)
	res := ValidateNetwork(net)
	assert.True(t, res.OK(), "errors: %+v", res.Errors)
}

func TestValidateNetwork_MissingID(t *testing.T) {
	net := &BknNetwork{
		BknNetworkFrontmatter: BknNetworkFrontmatter{
			Type: "knowledge_network",
			ID:   "net",
			Name: "Net",
		},
		ObjectTypes: []*BknObjectType{
			{
				BknObjectTypeFrontmatter: BknObjectTypeFrontmatter{
					Type: "object_type",
					ID:   "",
					Name: "X",
				},
				HasDataPropertiesSection: true,
				HasKeysSection:           true,
			},
		},
	}
	res := ValidateNetwork(net)
	assert.False(t, res.OK())
	var found bool
	for _, e := range res.Errors {
		if e.Code == "missing_frontmatter_field" && e.Column == "id" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected missing_frontmatter_field on id, got %+v", res.Errors)
}

func TestValidateNetwork_InvalidIDFormat(t *testing.T) {
	net := &BknNetwork{
		BknNetworkFrontmatter: BknNetworkFrontmatter{
			Type: "knowledge_network",
			ID:   "net",
			Name: "Net",
		},
		ObjectTypes: []*BknObjectType{
			{
				BknObjectTypeFrontmatter: BknObjectTypeFrontmatter{
					Type: "object_type",
					ID:   "Bad_Upper",
					Name: "Bad",
				},
				HasDataPropertiesSection: true,
				HasKeysSection:           true,
				DataProperties: []*DataProperty{{Name: "k", Type: "string"}},
				PrimaryKeys:    []string{"k"},
				DisplayKey:     "k",
			},
		},
	}
	res := ValidateNetwork(net)
	assert.False(t, res.OK())
	var found bool
	for _, e := range res.Errors {
		if e.Code == "invalid_id" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected invalid_id, got %+v", res.Errors)
}

func TestValidateNetwork_InvalidEndpointRef(t *testing.T) {
	net := &BknNetwork{
		BknNetworkFrontmatter: BknNetworkFrontmatter{
			Type: "knowledge_network",
			ID:   "net",
			Name: "Net",
		},
		ObjectTypes: []*BknObjectType{
			{
				BknObjectTypeFrontmatter: BknObjectTypeFrontmatter{
					Type: "object_type",
					ID:   "a",
					Name: "A",
				},
				HasDataPropertiesSection: true,
				HasKeysSection:           true,
			},
		},
		RelationTypes: []*BknRelationType{
			{
				BknRelationTypeFrontmatter: BknRelationTypeFrontmatter{
					Type: "relation_type",
					ID:   "r1",
					Name: "R",
				},
				Endpoint: Endpoint{
					Source: "a",
					Target: "missing_obj",
					Type:   "direct",
				},
			},
		},
	}
	res := ValidateNetwork(net)
	assert.False(t, res.OK())
	var found bool
	for _, e := range res.Errors {
		if e.Code == "invalid_endpoint_ref" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected invalid_endpoint_ref, got %+v", res.Errors)
}

func TestValidateNetwork_InvalidBoundObjectRef(t *testing.T) {
	net := &BknNetwork{
		BknNetworkFrontmatter: BknNetworkFrontmatter{
			Type: "knowledge_network",
			ID:   "net",
			Name: "Net",
		},
		ObjectTypes: []*BknObjectType{
			{
				BknObjectTypeFrontmatter: BknObjectTypeFrontmatter{
					Type: "object_type",
					ID:   "only_one",
					Name: "O",
				},
				HasDataPropertiesSection: true,
				HasKeysSection:           true,
			},
		},
		ActionTypes: []*BknActionType{
			{
				BknActionTypeFrontmatter: BknActionTypeFrontmatter{
					Type: "action_type",
					ID:   "act1",
					Name: "Act",
				},
				BoundObject: "ghost",
			},
		},
	}
	res := ValidateNetwork(net)
	assert.False(t, res.OK())
	var found bool
	for _, e := range res.Errors {
		if e.Code == "invalid_bound_object_ref" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected invalid_bound_object_ref, got %+v", res.Errors)
}
