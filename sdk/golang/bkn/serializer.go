package bkn

import (
	"fmt"
	"strings"
)

// Serialize converts a BknDocument back to .bkn file content (frontmatter + body).
func Serialize(doc *BknDocument) string {
	var b strings.Builder

	// Frontmatter
	b.WriteString("---\n")
	writeFrontmatter(&b, &doc.Frontmatter)
	b.WriteString("---\n")

	// Body: definitions
	for _, obj := range doc.Objects {
		b.WriteString("\n")
		writeObject(&b, &obj)
	}
	for _, rel := range doc.Relations {
		b.WriteString("\n")
		writeRelation(&b, &rel)
	}
	for _, act := range doc.Actions {
		b.WriteString("\n")
		writeAction(&b, &act)
	}
	for _, risk := range doc.Risks {
		b.WriteString("\n")
		writeRisk(&b, &risk)
	}

	return b.String()
}

func writeFrontmatter(b *strings.Builder, fm *Frontmatter) {
	writeField(b, "type", fm.Type)
	writeField(b, "id", fm.ID)
	writeField(b, "name", fm.Name)
	writeField(b, "version", fm.Version)
	writeField(b, "branch", fm.Branch)
	writeField(b, "description", fm.Description)
	writeField(b, "network", fm.Network)
	writeField(b, "namespace", fm.Namespace)
	writeField(b, "owner", fm.Owner)
	writeField(b, "author", fm.Author)
	writeField(b, "status", fm.Status)
	writeField(b, "spec_version", fm.SpecVersion)
	writeField(b, "risk_level", fm.RiskLevel)
	writeField(b, "object", fm.Object)
	writeField(b, "relation", fm.Relation)
	writeField(b, "source", fm.Source)
	writeField(b, "created_at", fm.CreatedAt)
	writeField(b, "updated_at", fm.UpdatedAt)
	if fm.Enabled != nil {
		fmt.Fprintf(b, "enabled: %v\n", *fm.Enabled)
	}
	if fm.RequiresApproval != nil {
		fmt.Fprintf(b, "requires_approval: %v\n", *fm.RequiresApproval)
	}
	if len(fm.Tags) > 0 {
		fmt.Fprintf(b, "tags: [%s]\n", strings.Join(fm.Tags, ", "))
	}
	if len(fm.Capabilities) > 0 {
		fmt.Fprintf(b, "capabilities: [%s]\n", strings.Join(fm.Capabilities, ", "))
	}
	if len(fm.Includes) > 0 {
		b.WriteString("includes:\n")
		for _, inc := range fm.Includes {
			fmt.Fprintf(b, "  - %s\n", inc)
		}
	}
}

func writeField(b *strings.Builder, key, val string) {
	if val != "" {
		fmt.Fprintf(b, "%s: %s\n", key, val)
	}
}

func writeObject(b *strings.Builder, obj *BknObject) {
	fmt.Fprintf(b, "## Object: %s\n", obj.ID)
	if obj.Name != "" {
		fmt.Fprintf(b, "**%s**", obj.Name)
		if obj.Description != "" {
			fmt.Fprintf(b, " - %s", obj.Description)
		}
		b.WriteString("\n")
	}
	writeInlineMeta(b, obj.Tags, obj.Owner)

	if obj.DataSource != nil {
		b.WriteString("\n### Data Source\n\n")
		b.WriteString("| Type | ID | Name |\n")
		b.WriteString("|------|------|------|\n")
		fmt.Fprintf(b, "| %s | %s | %s |\n", obj.DataSource.Type, obj.DataSource.ID, obj.DataSource.Name)
	}
	if len(obj.DataProperties) > 0 {
		b.WriteString("\n### Data Properties\n\n")
		b.WriteString("| Property | Display Name | Type | Constraint | Description | Primary Key | Display Key | Index |\n")
		b.WriteString("|----------|-------------|------|-----------|-------------|------------|------------|-------|\n")
		for _, dp := range obj.DataProperties {
			fmt.Fprintf(b, "| %s | %s | %s | %s | %s | %s | %s | %s |\n",
				dp.Property, dp.DisplayName, dp.Type, dp.Constraint, dp.Description,
				boolToYes(dp.PrimaryKey), boolToYes(dp.DisplayKey), boolToYes(dp.Index))
		}
	}
	if obj.BusinessSemantics != "" {
		fmt.Fprintf(b, "\n### Business Semantics\n\n%s\n", obj.BusinessSemantics)
	}
}

func writeRelation(b *strings.Builder, rel *Relation) {
	fmt.Fprintf(b, "## Relation: %s\n", rel.ID)
	if rel.Name != "" {
		fmt.Fprintf(b, "**%s**", rel.Name)
		if rel.Description != "" {
			fmt.Fprintf(b, " - %s", rel.Description)
		}
		b.WriteString("\n")
	}
	writeInlineMeta(b, rel.Tags, rel.Owner)

	if len(rel.Endpoints) > 0 {
		b.WriteString("\n### Endpoints\n\n")
		b.WriteString("| Source | Target | Type | Required | Min | Max |\n")
		b.WriteString("|--------|--------|------|----------|-----|-----|\n")
		for _, ep := range rel.Endpoints {
			fmt.Fprintf(b, "| %s | %s | %s | %s | %s | %s |\n",
				ep.Source, ep.Target, ep.Type, ep.Required, ep.Min, ep.Max)
		}
	}
	if len(rel.MappingRules) > 0 {
		b.WriteString("\n### Mapping Rules\n\n")
		b.WriteString("| Source Property | Target Property |\n")
		b.WriteString("|----------------|----------------|\n")
		for _, mr := range rel.MappingRules {
			fmt.Fprintf(b, "| %s | %s |\n", mr.SourceProperty, mr.TargetProperty)
		}
	}
	if rel.BusinessSemantics != "" {
		fmt.Fprintf(b, "\n### Business Semantics\n\n%s\n", rel.BusinessSemantics)
	}
}

func writeAction(b *strings.Builder, act *Action) {
	fmt.Fprintf(b, "## Action: %s\n", act.ID)
	if act.Name != "" {
		fmt.Fprintf(b, "**%s**", act.Name)
		if act.Description != "" {
			fmt.Fprintf(b, " - %s", act.Description)
		}
		b.WriteString("\n")
	}
	if act.BoundObject != "" || act.ActionType != "" {
		b.WriteString("\n| Bound Object | Action Type |\n")
		b.WriteString("|-------------|-------------|\n")
		fmt.Fprintf(b, "| %s | %s |\n", act.BoundObject, act.ActionType)
	}
	if act.TriggerCondition != "" {
		fmt.Fprintf(b, "\n### Trigger Condition\n\n```yaml\n%s\n```\n", act.TriggerCondition)
	}
	if len(act.PreConditions) > 0 {
		b.WriteString("\n### Pre-conditions\n\n")
		b.WriteString("| Object | Check | Condition | Message |\n")
		b.WriteString("|--------|-------|-----------|--------|\n")
		for _, pc := range act.PreConditions {
			fmt.Fprintf(b, "| %s | %s | %s | %s |\n", pc.Object, pc.Check, pc.Condition, pc.Message)
		}
	}
	if act.ToolConfig != nil {
		b.WriteString("\n### Tool Configuration\n\n")
		b.WriteString("| Type | Tool ID |\n")
		b.WriteString("|------|--------|\n")
		fmt.Fprintf(b, "| %s | %s |\n", act.ToolConfig.Type, act.ToolConfig.ToolID)
	}
	if act.ExecutionDescription != "" {
		fmt.Fprintf(b, "\n### Execution Description\n\n%s\n", act.ExecutionDescription)
	}
}

func writeRisk(b *strings.Builder, risk *Risk) {
	fmt.Fprintf(b, "## Risk: %s\n", risk.ID)
	if risk.Name != "" {
		fmt.Fprintf(b, "**%s**", risk.Name)
		if risk.Description != "" {
			fmt.Fprintf(b, " - %s", risk.Description)
		}
		b.WriteString("\n")
	}
	writeInlineMeta(b, risk.Tags, risk.Owner)

	if risk.ControlScope != "" {
		fmt.Fprintf(b, "\n### Control Scope\n\n%s\n", risk.ControlScope)
	}
	if risk.ControlPolicy != "" {
		fmt.Fprintf(b, "\n### Control Policy\n\n%s\n", risk.ControlPolicy)
	}
	if len(risk.PreChecks) > 0 {
		b.WriteString("\n### Pre-checks\n\n")
		b.WriteString("| Object | Check | Condition | Message |\n")
		b.WriteString("|--------|-------|-----------|--------|\n")
		for _, pc := range risk.PreChecks {
			fmt.Fprintf(b, "| %s | %s | %s | %s |\n", pc.Object, pc.Check, pc.Condition, pc.Message)
		}
	}
	if risk.RollbackPlan != "" {
		fmt.Fprintf(b, "\n### Rollback Plan\n\n%s\n", risk.RollbackPlan)
	}
	if risk.AuditRequirements != "" {
		fmt.Fprintf(b, "\n### Audit Requirements\n\n%s\n", risk.AuditRequirements)
	}
}

func writeInlineMeta(b *strings.Builder, tags []string, owner string) {
	if len(tags) > 0 {
		fmt.Fprintf(b, "- **Tags**: %s\n", strings.Join(tags, ", "))
	}
	if owner != "" {
		fmt.Fprintf(b, "- **Owner**: %s\n", owner)
	}
}

func boolToYes(v bool) string {
	if v {
		return "YES"
	}
	return ""
}
