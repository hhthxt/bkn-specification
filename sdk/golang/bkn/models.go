package bkn

// Frontmatter is YAML frontmatter metadata for a .bkn file.
type Frontmatter struct {
	Type             string   `yaml:"type"`
	ID               string   `yaml:"id"`
	Name             string   `yaml:"name"`
	Version          string   `yaml:"version"`
	Tags             []string `yaml:"tags"`
	Description      string   `yaml:"description"`
	Includes         []string `yaml:"includes"`
	Network          string   `yaml:"network"`
	Namespace        string   `yaml:"namespace"`
	Owner            string   `yaml:"owner"`
	SpecVersion      string   `yaml:"spec_version"`
	Enabled          *bool    `yaml:"enabled"`
	RiskLevel        string   `yaml:"risk_level"`
	RequiresApproval *bool    `yaml:"requires_approval"`
	Object           string   `yaml:"object"`
	Relation         string   `yaml:"relation"`
	Source           string   `yaml:"source"`
	Extra            map[string]any
}

// DataSource is a ### Data Source table row.
type DataSource struct {
	Type string
	ID   string
	Name string
}

// DataProperty is a ### Data Properties table row.
type DataProperty struct {
	Property    string
	DisplayName string
	Type        string
	Constraint  string
	Description string
	PrimaryKey  bool
	DisplayKey  bool
	Index       bool
}

// PropertyOverride is a ### Property Override table row.
type PropertyOverride struct {
	Property    string
	DisplayName string
	IndexConfig string
	Constraint  string
	Description string
}

// LogicPropertyParameter is a parameter row inside a Logic Property sub-section.
type LogicPropertyParameter struct {
	Parameter   string
	Type        string
	Source      string
	Binding     string
	Description string
}

// LogicProperty is a #### {property_name} under ### Logic Properties.
type LogicProperty struct {
	Name        string
	LPType      string // metric | operator
	Source      string
	SourceType  string
	Description string
	Parameters  []LogicPropertyParameter
}

// Endpoint is a ### Endpoints table row (relations).
type Endpoint struct {
	Source   string
	Target   string
	Type     string // direct | data_view
	Required string
	Min      string
	Max      string
}

// MappingRule is a ### Mapping Rules table row.
type MappingRule struct {
	SourceProperty string
	TargetProperty string
}

// ToolConfig is a ### Tool Configuration table row.
type ToolConfig struct {
	Type   string // tool | mcp
	ToolID string
}

// PreCondition is a ### Pre-conditions table row.
type PreCondition struct {
	Object    string
	Check     string
	Condition string
	Message   string
}

// Schedule is a ### Schedule table row.
type Schedule struct {
	Type       string // FIX_RATE | CRON
	Expression string
}

// BknObject is a ## Object: {id} block.
type BknObject struct {
	ID                string
	Name              string
	Description       string
	Tags              []string
	Owner             string
	DataSource        *DataSource
	DataProperties    []DataProperty
	PropertyOverrides []PropertyOverride
	LogicProperties   []LogicProperty
	BusinessSemantics string
}

// Relation is a ## Relation: {id} block.
type Relation struct {
	ID                string
	Name              string
	Description       string
	Tags              []string
	Owner             string
	Endpoints         []Endpoint
	MappingRules      []MappingRule
	BusinessSemantics string
}

// Action is a ## Action: {id} block.
type Action struct {
	ID                   string
	Name                 string
	Description          string
	BoundObject          string
	ActionType           string
	TriggerCondition     string
	PreConditions        []PreCondition
	ToolConfig           *ToolConfig
	ParameterBinding     []map[string]string
	Schedule             *Schedule
	ScopeOfImpact        []map[string]string
	ExecutionDescription string
}

// BknDocument is a parsed .bkn file: frontmatter + body definitions.
type BknDocument struct {
	Frontmatter Frontmatter
	Objects     []BknObject
	Relations   []Relation
	Actions     []Action
	SourcePath  string
}

// BknNetwork is an aggregated network: root document + all included documents.
type BknNetwork struct {
	Root     BknDocument
	Includes []BknDocument
}

// AllObjects returns all objects from root and included documents.
func (n *BknNetwork) AllObjects() []BknObject {
	var out []BknObject
	out = append(out, n.Root.Objects...)
	for _, doc := range n.Includes {
		out = append(out, doc.Objects...)
	}
	return out
}

// AllRelations returns all relations from root and included documents.
func (n *BknNetwork) AllRelations() []Relation {
	var out []Relation
	out = append(out, n.Root.Relations...)
	for _, doc := range n.Includes {
		out = append(out, doc.Relations...)
	}
	return out
}

// AllActions returns all actions from root and included documents.
func (n *BknNetwork) AllActions() []Action {
	var out []Action
	out = append(out, n.Root.Actions...)
	for _, doc := range n.Includes {
		out = append(out, doc.Actions...)
	}
	return out
}
