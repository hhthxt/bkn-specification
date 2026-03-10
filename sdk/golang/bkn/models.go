package bkn

// BknNetworkFrontmatter is YAML frontmatter metadata for a .bkn file.
type BknNetworkFrontmatter struct {
	Type        string   `yaml:"type"`
	ID          string   `yaml:"id"`
	Name        string   `yaml:"name"`
	Tags        []string `yaml:"tags"`
	Description string   `yaml:"description"`

	Version        string `yaml:"version"`
	Branch         string `yaml:"branch"`
	BusinessDomain string `yaml:"business_domain"`
}

// BknDocument is a parsed network.bkn file: frontmatter + body definitions.
type BknNetwork struct {
	BknNetworkFrontmatter

	ObjectTypes   []*BknObjectType
	RelationTypes []*BknRelationType
	ActionTypes   []*BknActionType
	RiskTypes     []*BknRiskType
	ConceptGroups []*BknConceptGroup
}

// BknObjectTypeFrontmatter is YAML frontmatter metadata for a .bkn file.
type BknObjectTypeFrontmatter struct {
	Type        string   `yaml:"type"`
	ID          string   `yaml:"id"`
	Name        string   `yaml:"name"`
	Tags        []string `yaml:"tags"`
	Description string   `yaml:"description"`
}

// BknObjectType represents an object type definition.
type BknObjectType struct {
	BknObjectTypeFrontmatter

	DataSource      *ResourceInfo
	DataProperties  []*DataProperty
	LogicProperties []*LogicProperty

	// Keys section
	PrimaryKeys    []string
	DisplayKey     string
	IncrementalKey string
}

// ResourceInfo represents a data source reference.
type ResourceInfo struct {
	Type string
	ID   string
	Name string
}

// DataProperty is a ### Data Properties table row.
type DataProperty struct {
	Name        string
	DisplayName string
	Type        string
	Description string

	MappedField *Field
}

// LogicProperty represents a logic property definition.
type LogicProperty struct {
	Name        string
	DisplayName string
	Type        string
	Description string

	DataSource   *ResourceInfo
	Parameters   []Parameter
	AnalysisDims []Field
}

type Field struct {
	Name        string
	Type        string
	DisplayName string
	Description string
}

// Parameter represents a parameter binding.
type Parameter struct {
	Name        string
	Type        string
	Source      string // property, const, etc.
	Operation   string
	ValueFrom   string
	Value       any
	IfSystemGen bool
	Description string
}

// BknRelationTypeFrontmatter is YAML frontmatter metadata for a .bkn file.
type BknRelationTypeFrontmatter struct {
	Type        string   `yaml:"type"`
	ID          string   `yaml:"id"`
	Name        string   `yaml:"name"`
	Tags        []string `yaml:"tags"`
	Description string   `yaml:"description"`

	SourceObjectTypeID string `yaml:"source_object_type"`
	TargetObjectTypeID string `yaml:"target_object_type"`
}

// BknRelationType represents a relation type definition.
type BknRelationType struct {
	BknRelationTypeFrontmatter

	// Mapping Rules
	RelationType string // direct/data_view.
	MappingRules any
}

// MappingRule represents a property mapping between source and target.
type MappingRule struct {
	SourceProperty string
	TargetProperty string
}

// DirectMappingRule represents a direct mapping rule.
type DirectMappingRule []MappingRule

// InDirectMappingRule represents a non-direct mapping rule.
type InDirectMappingRule struct {
	BackingDataSource  *ResourceInfo
	SourceMappingRules []MappingRule
	TargetMappingRules []MappingRule
}

// BknActionTypeFrontmatter is YAML frontmatter metadata for a .bkn file.
type BknActionTypeFrontmatter struct {
	Type        string   `yaml:"type"`
	ID          string   `yaml:"id"`
	Name        string   `yaml:"name"`
	Tags        []string `yaml:"tags"`
	Description string   `yaml:"description"`

	ActionType       string `yaml:"action_type"`
	Enabled          bool   `yaml:"enabled"`
	RiskLevel        string `yaml:"risk_level"`
	RequiresApproval bool   `yaml:"requires_approval"`
}

// BknActionType represents an action type definition.
type BknActionType struct {
	BknActionTypeFrontmatter

	// Bound Object
	ObjectTypeID string
	BoundObject  string

	// Trigger Condition
	TriggerCondition *CondCfg
	Condition        *CondCfg

	// Pre-conditions
	PreConditions []*PreCondition

	// Scope of Impact
	ScopeOfImpact []*ImpactEntry
	Affect        *ActionAffect

	// Tool Configuration
	ToolConfig   *ToolConfiguration
	ActionSource ActionSource

	// Parameter Binding
	Parameters []Parameter

	// Schedule
	Schedule *Schedule
}

// CondCfg represents a condition configuration.
type CondCfg struct {
	ObjectTypeID string
	Field        string
	Operation    string
	SubConds     []*CondCfg
	ValueFrom    string
	Value        any
}

// PreCondition represents a pre-condition check.
type PreCondition struct {
	Object    string
	Check     string
	Condition string
	Message   string
}

// ImpactEntry represents a scope of impact entry.
type ImpactEntry struct {
	Object      string
	Description string
}

// ToolConfiguration represents tool configuration.
type ToolConfiguration struct {
	Type     string // tool, mcp, etc.
	BoxID    string
	ToolID   string
	McpID    string
	ToolName string
}

// Schedule represents an action schedule.
type Schedule struct {
	Type       string // FIX_RATE, CRON, etc.
	Expression string
}

// ActionAffect represents action affect.
type ActionAffect struct {
	ObjectTypeID string
	Description  string
}

// ActionSource represents action source.
type ActionSource struct {
	Type string
	// type 为 tool
	BoxID  string
	ToolID string
	// type 为 mcp
	McpID    string
	ToolName string
}

// BknRiskTypeFrontmatter is YAML frontmatter metadata for a .bkn file.
type BknRiskTypeFrontmatter struct {
	Type        string   `yaml:"type"`
	ID          string   `yaml:"id"`
	Name        string   `yaml:"name"`
	Tags        []string `yaml:"tags"`
	Description string   `yaml:"description"`
}

// BknRiskType represents a risk type definition.
type BknRiskType struct {
	BknRiskTypeFrontmatter

	ControlScope      string
	ControlPolicy     string
	PreChecks         []*CondCfg
	RollbackPlan      string
	AuditRequirements string
}

// BknConceptGroupFrontmatter is YAML frontmatter metadata for a .bkn file.
type BknConceptGroupFrontmatter struct {
	Type        string   `yaml:"type"`
	ID          string   `yaml:"id"`
	Name        string   `yaml:"name"`
	Tags        []string `yaml:"tags"`
	Description string   `yaml:"description"`
}

// BknConceptGroup represents a concept group definition.
type BknConceptGroup struct {
	BknConceptGroupFrontmatter

	ObjectTypes []string
}
