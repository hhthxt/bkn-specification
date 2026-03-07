package bkn

import "strings"

// DeleteTarget specifies a single deletion target by type and id.
type DeleteTarget struct {
	Type string // "object", "relation", or "action"
	ID   string
}

// DeletePlan is the result of planning a delete operation.
type DeletePlan struct {
	Targets   []DeleteTarget // Validated targets that exist in the network
	NotFound  []DeleteTarget // Targets not found in the network
}

// OK returns true if all targets exist in the network.
func (p *DeletePlan) OK() bool {
	return len(p.NotFound) == 0
}

// PlanDelete validates that targets exist in the network.
// dryRun is for API compatibility; the SDK does not persist changes.
// Actual deletion is performed by the consumer (e.g. backend API).
func PlanDelete(network *BknNetwork, targets []DeleteTarget, dryRun bool) *DeletePlan {
	plan := &DeletePlan{}
	for _, t := range targets {
		typ := strings.ToLower(strings.TrimSpace(t.Type))
		exists := false
		switch typ {
		case "object":
			for _, o := range network.AllObjects() {
				if o.ID == t.ID {
					exists = true
					break
				}
			}
		case "relation":
			for _, r := range network.AllRelations() {
				if r.ID == t.ID {
					exists = true
					break
				}
			}
		case "action":
			for _, a := range network.AllActions() {
				if a.ID == t.ID {
					exists = true
					break
				}
			}
		default:
			// Invalid type, treat as not found
			plan.NotFound = append(plan.NotFound, t)
			continue
		}
		if exists {
			plan.Targets = append(plan.Targets, t)
		} else {
			plan.NotFound = append(plan.NotFound, t)
		}
	}
	return plan
}

// NetworkWithout returns a new BknNetwork with the given targets removed (in-memory simulation).
// Targets that do not exist are ignored.
func NetworkWithout(network *BknNetwork, targets []DeleteTarget) *BknNetwork {
	targetSet := make(map[string]bool)
	for _, t := range targets {
		key := strings.ToLower(strings.TrimSpace(t.Type)) + ":" + t.ID
		targetSet[key] = true
	}

	out := &BknNetwork{
		Root:     *copyDocumentExcluding(&network.Root, targetSet),
		Includes: make([]BknDocument, len(network.Includes)),
	}
	for i := range network.Includes {
		out.Includes[i] = *copyDocumentExcluding(&network.Includes[i], targetSet)
	}
	return out
}

func copyDocumentExcluding(doc *BknDocument, targetSet map[string]bool) *BknDocument {
	out := &BknDocument{
		Frontmatter: doc.Frontmatter,
		SourcePath:  doc.SourcePath,
	}
	for _, o := range doc.Objects {
		if !targetSet["object:"+o.ID] {
			out.Objects = append(out.Objects, o)
		}
	}
	for _, r := range doc.Relations {
		if !targetSet["relation:"+r.ID] {
			out.Relations = append(out.Relations, r)
		}
	}
	for _, a := range doc.Actions {
		if !targetSet["action:"+a.ID] {
			out.Actions = append(out.Actions, a)
		}
	}
	out.Connections = append(out.Connections, doc.Connections...)
	out.DataTables = append(out.DataTables, doc.DataTables...)
	return out
}

