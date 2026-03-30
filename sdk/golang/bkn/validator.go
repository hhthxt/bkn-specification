// Copyright The kweaver-ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package bkn

import (
	"fmt"
	"regexp"
	"strings"
)

// IDs allow lowercase letters, digits, underscores, and hyphens (matches common BKN examples e.g. k8s-network).
var idPattern = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)

// ValidationError represents a single validation problem.
type ValidationError struct {
	Table   string
	Row     *int
	Column  string
	Code    string
	Message string
}

// ValidationResult aggregates validation outcome.
type ValidationResult struct {
	Errors []ValidationError
}

// OK returns true if there are no errors.
func (r *ValidationResult) OK() bool {
	return len(r.Errors) == 0
}

func appendError(result *ValidationResult, table, column, code, message string) {
	result.Errors = append(result.Errors, ValidationError{
		Table:   table,
		Row:     nil,
		Column:  column,
		Code:    code,
		Message: message,
	})
}

// ValidateNetwork performs structural validation on a loaded BknNetwork (aligned with TypeScript SDK).
func ValidateNetwork(doc *BknNetwork) *ValidationResult {
	result := &ValidationResult{}

	// Root network.bkn
	if strings.TrimSpace(doc.Type) == "" {
		appendError(result, "network.bkn", "type", "missing_frontmatter_field", "frontmatter 'type' is required")
	}
	if strings.TrimSpace(doc.ID) == "" {
		appendError(result, "network.bkn", "id", "missing_frontmatter_field", "frontmatter 'id' is required")
	}
	if strings.TrimSpace(doc.Name) == "" {
		appendError(result, "network.bkn", "name", "missing_frontmatter_field", "frontmatter 'name' is required")
	}
	if doc.ID != "" && !idPattern.MatchString(strings.TrimSpace(doc.ID)) {
		appendError(result, "network.bkn", "id", "invalid_id", fmt.Sprintf("frontmatter id %q must match /^[a-z][a-z0-9_]*$/", doc.ID))
	}

	objectIDs := make(map[string]struct{})
	for _, ot := range doc.ObjectTypes {
		objectIDs[ot.ID] = struct{}{}
	}

	for _, ot := range doc.ObjectTypes {
		t := tableName("object_type", ot.ID)
		validateDefFrontmatter(result, t, ot.Type, ot.ID, ot.Name)
		if !ot.HasDataPropertiesSection {
			appendError(result, t, "", "missing_section", "ObjectType must include a ### Data Properties section")
		}
		if !ot.HasKeysSection {
			appendError(result, t, "", "missing_section", "ObjectType must include a ### Keys section")
		}
	}

	for _, rt := range doc.RelationTypes {
		t := tableName("relation_type", rt.ID)
		validateDefFrontmatter(result, t, rt.Type, rt.ID, rt.Name)
		src := strings.TrimSpace(rt.Endpoint.Source)
		tgt := strings.TrimSpace(rt.Endpoint.Target)
		if src == "" && tgt == "" {
			appendError(result, t, "", "empty_endpoint", "RelationType must have at least one endpoint row under ### Endpoint")
		}
		if src != "" {
			if _, ok := objectIDs[src]; !ok {
				appendError(result, t, "Source", "invalid_endpoint_ref", fmt.Sprintf("endpoint source %q is not a defined object type id", src))
			}
		}
		if tgt != "" {
			if _, ok := objectIDs[tgt]; !ok {
				appendError(result, t, "Target", "invalid_endpoint_ref", fmt.Sprintf("endpoint target %q is not a defined object type id", tgt))
			}
		}
	}

	for _, at := range doc.ActionTypes {
		t := tableName("action_type", at.ID)
		validateDefFrontmatter(result, t, at.Type, at.ID, at.Name)
		bo := strings.TrimSpace(at.BoundObject)
		if bo != "" {
			if _, ok := objectIDs[bo]; !ok {
				appendError(result, t, "Bound Object", "invalid_bound_object_ref", fmt.Sprintf("bound object %q is not a defined object type id", bo))
			}
		}
	}

	for _, r := range doc.RiskTypes {
		t := tableName("risk_type", r.ID)
		validateDefFrontmatter(result, t, r.Type, r.ID, r.Name)
	}

	for _, cg := range doc.ConceptGroups {
		t := tableName("concept_group", cg.ID)
		validateDefFrontmatter(result, t, cg.Type, cg.ID, cg.Name)
		for _, oid := range cg.ObjectTypes {
			oid = strings.TrimSpace(oid)
			if oid == "" {
				continue
			}
			if _, ok := objectIDs[oid]; !ok {
				appendError(result, t, "Object Types", "invalid_concept_group_ref", fmt.Sprintf("concept group lists unknown object type id %q", oid))
			}
		}
	}

	return result
}

func tableName(kind, id string) string {
	if strings.TrimSpace(id) == "" {
		return kind + ":<unknown>"
	}
	return fmt.Sprintf("%s:%s", kind, id)
}

func validateDefFrontmatter(result *ValidationResult, table, typ, id, name string) {
	if strings.TrimSpace(typ) == "" {
		appendError(result, table, "type", "missing_frontmatter_field", "frontmatter 'type' is required")
	}
	if strings.TrimSpace(id) == "" {
		appendError(result, table, "id", "missing_frontmatter_field", "frontmatter 'id' is required")
	}
	if strings.TrimSpace(name) == "" {
		appendError(result, table, "name", "missing_frontmatter_field", "frontmatter 'name' is required")
	}
	if strings.TrimSpace(id) != "" && !idPattern.MatchString(strings.TrimSpace(id)) {
		appendError(result, table, "id", "invalid_id", fmt.Sprintf("frontmatter id %q must match /^[a-z][a-z0-9_]*$/", id))
	}
}
