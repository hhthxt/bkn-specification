package bkn

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Column name aliases (Chinese -> English)
var columnAliases = map[string]string{
	"属性": "Property", "显示名": "Display Name", "显示名称": "Display Name",
	"类型": "Type", "约束": "Constraint", "描述": "Description", "说明": "Description",
	"主键": "Primary Key", "显示属性": "Display Key", "索引": "Index",
	"数据来源": "Data Source", "名称": "Name",
	"起点": "Source", "终点": "Target", "必须": "Required",
	"起点属性": "Source Property", "终点属性": "Target Property",
	"参数": "Parameter", "来源": "Source", "绑定": "Binding",
	"工具": "Tool ID", "绑定对象": "Bound Object", "行动类型": "Action Type",
	"对象": "Object", "检查": "Check", "条件": "Condition", "消息": "Message",
	"表达式": "Expression", "索引配置": "Index Config",
	"影响说明": "Impact Description",
}

// Section name aliases
var sectionAliases = map[string]string{
	"数据来源": "Data Source", "数据属性": "Data Properties", "属性覆盖": "Property Override",
	"逻辑属性": "Logic Properties", "业务语义": "Business Semantics",
	"关联定义": "Endpoints", "映射规则": "Mapping Rules", "映射视图": "Mapping View",
	"起点映射": "Source Mapping", "终点映射": "Target Mapping",
	"绑定对象": "Bound Object", "触发条件": "Trigger Condition", "前置条件": "Pre-conditions",
	"工具配置": "Tool Configuration", "参数绑定": "Parameter Binding",
	"调度配置": "Schedule", "影响范围": "Scope of Impact", "执行说明": "Execution Description",
}

var definitionRE = regexp.MustCompile(`(?m)^##\s+(Object|Relation|Action):\s*(\S+)`)
var sectionRE = regexp.MustCompile(`(?m)^###\s+(.+)$`)
var subSectionRE = regexp.MustCompile(`(?m)^####\s+(.+)$`)
var inlineMetaRE = regexp.MustCompile(`(?m)^-\s+\*\*(\w+)\*\*:\s*(.+)$`)
var displayNameRE = regexp.MustCompile(`(?m)^\*\*(.+?)\*\*(?:\s*-\s*(.*))?$`)
var headingRE = regexp.MustCompile(`(?m)^#{1,2}\s+(.+)$`)
var tableSepRE = regexp.MustCompile(`^\|?[\s:*-]+(\|[\s:*-]+)*\|?$`)
var yamlBlockRE = regexp.MustCompile("(?s)```yaml\\s*\\n(.+?)```")

func normalizeColumn(name string) string {
	name = strings.TrimSpace(name)
	if v, ok := columnAliases[name]; ok {
		return v
	}
	return name
}

func normalizeSection(name string) string {
	name = strings.TrimSpace(name)
	if v, ok := sectionAliases[name]; ok {
		return v
	}
	return name
}

func isYes(val string) bool {
	return strings.TrimSpace(strings.ToUpper(val)) == "YES"
}

func splitFrontmatter(text string) (fm string, body string) {
	text = strings.TrimPrefix(text, "\ufeff")
	if !strings.HasPrefix(text, "---") {
		return "", text
	}
	end := strings.Index(text[3:], "\n---")
	if end == -1 {
		return "", text
	}
	end += 3
	idx := strings.Index(text[end+3:], "\n")
	if idx == -1 {
		return strings.TrimSpace(text[3:end]), ""
	}
	fm = strings.TrimSpace(text[3:end])
	body = text[end+4+idx:]
	return fm, body
}

func splitRow(row string) []string {
	row = strings.TrimSpace(row)
	if strings.HasPrefix(row, "|") {
		row = row[1:]
	}
	if strings.HasSuffix(row, "|") {
		row = row[:len(row)-1]
	}
	parts := strings.Split(row, "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func parseTable(lines []string) []map[string]string {
	var tableLines []string
	for _, line := range lines {
		s := strings.TrimSpace(line)
		if strings.HasPrefix(s, "|") {
			tableLines = append(tableLines, s)
		} else if len(tableLines) > 0 {
			break
		}
	}
	if len(tableLines) < 2 {
		return nil
	}
	headers := splitRow(tableLines[0])
	for i, h := range headers {
		headers[i] = normalizeColumn(h)
	}
	sepLine := strings.TrimSpace(tableLines[1])
	dataStart := 2
	if !tableSepRE.MatchString(sepLine) {
		dataStart = 1
	}
	var rows []map[string]string
	for _, line := range tableLines[dataStart:] {
		cells := splitRow(line)
		row := make(map[string]string)
		for i, h := range headers {
			if i < len(cells) {
				row[h] = cells[i]
			} else {
				row[h] = ""
			}
		}
		rows = append(rows, row)
	}
	return rows
}

func extractSections(body string, level string) map[string]string {
	var re *regexp.Regexp
	if level == "###" {
		re = sectionRE
	} else {
		re = subSectionRE
	}
	matches := re.FindAllStringSubmatchIndex(body, -1)
	sections := make(map[string]string)
	for i, m := range matches {
		title := normalizeSection(strings.TrimSpace(body[m[2]:m[3]]))
		start := m[1]
		end := len(body)
		if i+1 < len(matches) {
			end = matches[i+1][0]
		}
		sections[title] = strings.TrimSpace(body[start:end])
	}
	return sections
}

func extractFirstTableLines(text string) []string {
	var tableLines []string
	started := false
	for _, line := range strings.Split(text, "\n") {
		s := strings.TrimSpace(line)
		if strings.HasPrefix(s, "|") {
			tableLines = append(tableLines, s)
			started = true
		} else if started {
			break
		}
	}
	return tableLines
}

func parseTableColumns(tableLines []string) []string {
	if len(tableLines) == 0 {
		return nil
	}
	header := strings.TrimSpace(tableLines[0])
	if strings.HasPrefix(header, "|") {
		header = header[1:]
	}
	if strings.HasSuffix(header, "|") {
		header = header[:len(header)-1]
	}
	parts := strings.Split(header, "|")
	var cols []string
	for _, p := range parts {
		cols = append(cols, normalizeColumn(strings.TrimSpace(p)))
	}
	return cols
}

func parseInlineMeta(text string) (tags []string, owner string) {
	matches := inlineMetaRE.FindAllStringSubmatch(text, -1)
	for _, m := range matches {
		if len(m) >= 3 {
			key := strings.TrimSpace(m[1])
			val := strings.TrimSpace(m[2])
			if key == "Tags" {
				for _, t := range strings.Split(val, ",") {
					if s := strings.TrimSpace(t); s != "" {
						tags = append(tags, s)
					}
				}
			} else if key == "Owner" {
				owner = val
			}
		}
	}
	return tags, owner
}

func parseDisplayName(text string) (name, desc string) {
	m := displayNameRE.FindStringSubmatch(text)
	if len(m) >= 2 {
		name = strings.TrimSpace(m[1])
		if len(m) >= 3 {
			desc = strings.TrimSpace(m[2])
		}
	}
	return name, desc
}

func parseDataSource(sectionText string) *DataSource {
	rows := parseTable(strings.Split(sectionText, "\n"))
	if len(rows) == 0 {
		return nil
	}
	r := rows[0]
	return &DataSource{
		Type: r["Type"],
		ID:   r["ID"],
		Name: r["Name"],
	}
}

func parseDataProperties(sectionText string) []DataProperty {
	rows := parseTable(strings.Split(sectionText, "\n"))
	var props []DataProperty
	for _, row := range rows {
		props = append(props, DataProperty{
			Property:    row["Property"],
			DisplayName: row["Display Name"],
			Type:        row["Type"],
			Constraint:  row["Constraint"],
			Description: row["Description"],
			PrimaryKey:  isYes(row["Primary Key"]),
			DisplayKey:  isYes(row["Display Key"]),
			Index:       isYes(row["Index"]),
		})
	}
	return props
}

func parsePropertyOverrides(sectionText string) []PropertyOverride {
	rows := parseTable(strings.Split(sectionText, "\n"))
	var out []PropertyOverride
	for _, row := range rows {
		out = append(out, PropertyOverride{
			Property:    row["Property"],
			DisplayName: row["Display Name"],
			IndexConfig: row["Index Config"],
			Constraint:  row["Constraint"],
			Description: row["Description"],
		})
	}
	return out
}

func parseLogicProperties(sectionText string) []LogicProperty {
	subs := extractSections(sectionText, "####")
	var props []LogicProperty
	for propName, content := range subs {
		lp := LogicProperty{Name: propName}
		if m := regexp.MustCompile(`-\s+\*\*Type\*\*:\s*(\S+)`).FindStringSubmatch(content); len(m) >= 2 {
			lp.LPType = strings.TrimSpace(m[1])
		}
		if m := regexp.MustCompile(`(?s)-\s+\*\*Source\*\*:\s*(.+?)(?:\((.+?)\))?\s*$`).FindStringSubmatch(content); len(m) >= 2 {
			lp.Source = strings.TrimSpace(m[1])
			if len(m) >= 3 {
				lp.SourceType = strings.TrimSpace(m[2])
			}
		}
		if m := regexp.MustCompile(`(?s)-\s+\*\*Description\*\*:\s*(.+)$`).FindStringSubmatch(content); len(m) >= 2 {
			lp.Description = strings.TrimSpace(m[1])
		}
		rows := parseTable(strings.Split(content, "\n"))
		for _, row := range rows {
			lp.Parameters = append(lp.Parameters, LogicPropertyParameter{
				Parameter:   row["Parameter"],
				Type:        row["Type"],
				Source:      row["Source"],
				Binding:     row["Binding"],
				Description: row["Description"],
			})
		}
		props = append(props, lp)
	}
	return props
}

func parseEndpoints(sectionText string) []Endpoint {
	rows := parseTable(strings.Split(sectionText, "\n"))
	var out []Endpoint
	for _, row := range rows {
		out = append(out, Endpoint{
			Source:   row["Source"],
			Target:   row["Target"],
			Type:     row["Type"],
			Required: row["Required"],
			Min:      row["Min"],
			Max:      row["Max"],
		})
	}
	return out
}

func parseMappingRules(sectionText string) []MappingRule {
	rows := parseTable(strings.Split(sectionText, "\n"))
	var out []MappingRule
	for _, row := range rows {
		out = append(out, MappingRule{
			SourceProperty: row["Source Property"],
			TargetProperty: row["Target Property"],
		})
	}
	return out
}

func parseObjectBlock(blockID, blockText string) BknObject {
	name, desc := parseDisplayName(blockText)
	tags, owner := parseInlineMeta(blockText)
	sections := extractSections(blockText, "###")

	obj := BknObject{
		ID:          blockID,
		Name:        name,
		Description: desc,
		Tags:        tags,
		Owner:       owner,
	}
	if s, ok := sections["Data Source"]; ok {
		obj.DataSource = parseDataSource(s)
	}
	if s, ok := sections["Data Properties"]; ok {
		obj.DataProperties = parseDataProperties(s)
	}
	if s, ok := sections["Property Override"]; ok {
		obj.PropertyOverrides = parsePropertyOverrides(s)
	}
	if s, ok := sections["Logic Properties"]; ok {
		obj.LogicProperties = parseLogicProperties(s)
	}
	if s, ok := sections["Business Semantics"]; ok {
		obj.BusinessSemantics = s
	}
	return obj
}

func parseRelationBlock(blockID, blockText string) Relation {
	name, desc := parseDisplayName(blockText)
	tags, owner := parseInlineMeta(blockText)
	sections := extractSections(blockText, "###")

	rel := Relation{
		ID:          blockID,
		Name:        name,
		Description: desc,
		Tags:        tags,
		Owner:       owner,
	}
	if s, ok := sections["Endpoints"]; ok {
		rel.Endpoints = parseEndpoints(s)
	}
	if s, ok := sections["Mapping Rules"]; ok {
		rel.MappingRules = parseMappingRules(s)
	}
	if s, ok := sections["Business Semantics"]; ok {
		rel.BusinessSemantics = s
	}
	return rel
}

func parseActionBlock(blockID, blockText string) Action {
	name, desc := parseDisplayName(blockText)
	sections := extractSections(blockText, "###")

	action := Action{
		ID:          blockID,
		Name:        name,
		Description: desc,
	}
	boundRows := parseTable(strings.Split(blockText, "\n"))
	for _, row := range boundRows {
		if _, ok := row["Bound Object"]; ok {
			action.BoundObject = row["Bound Object"]
			action.ActionType = row["Action Type"]
			break
		}
	}
	if s, ok := sections["Trigger Condition"]; ok {
		if m := yamlBlockRE.FindStringSubmatch(s); len(m) >= 2 {
			action.TriggerCondition = strings.TrimSpace(m[1])
		} else {
			action.TriggerCondition = s
		}
	}
	if s, ok := sections["Pre-conditions"]; ok {
		rows := parseTable(strings.Split(s, "\n"))
		for _, row := range rows {
			action.PreConditions = append(action.PreConditions, PreCondition{
				Object:    row["Object"],
				Check:     row["Check"],
				Condition: row["Condition"],
				Message:   row["Message"],
			})
		}
	}
	if s, ok := sections["Tool Configuration"]; ok {
		rows := parseTable(strings.Split(s, "\n"))
		if len(rows) > 0 {
			r := rows[0]
			toolType := r["Type"]
			toolID := r["Tool ID"]
			if toolID == "" {
				toolID = r["MCP"]
			}
			action.ToolConfig = &ToolConfig{Type: toolType, ToolID: toolID}
		}
	}
	if s, ok := sections["Parameter Binding"]; ok {
		rows := parseTable(strings.Split(s, "\n"))
		for _, row := range rows {
			action.ParameterBinding = append(action.ParameterBinding, row)
		}
	}
	if s, ok := sections["Schedule"]; ok {
		rows := parseTable(strings.Split(s, "\n"))
		if len(rows) > 0 {
			r := rows[0]
			action.Schedule = &Schedule{
				Type:       r["Type"],
				Expression: r["Expression"],
			}
		}
	}
	if s, ok := sections["Scope of Impact"]; ok {
		rows := parseTable(strings.Split(s, "\n"))
		for _, row := range rows {
			action.ScopeOfImpact = append(action.ScopeOfImpact, row)
		}
	}
	if s, ok := sections["Execution Description"]; ok {
		action.ExecutionDescription = s
	}
	return action
}

// ParseFrontmatter parses the YAML frontmatter of a .bkn file.
func ParseFrontmatter(text string) (*Frontmatter, error) {
	fmStr, _ := splitFrontmatter(text)
	if fmStr == "" {
		return &Frontmatter{}, nil
	}
	var data map[string]any
	if err := yaml.Unmarshal([]byte(fmStr), &data); err != nil {
		return nil, err
	}
	if data == nil {
		data = make(map[string]any)
	}

	fm := &Frontmatter{
		Type:        strVal(data, "type"),
		ID:          strVal(data, "id"),
		Name:        strVal(data, "name"),
		Version:     strVal(data, "version"),
		Description: strVal(data, "description"),
		Network:     strVal(data, "network"),
		Namespace:   strVal(data, "namespace"),
		Owner:       strVal(data, "owner"),
		SpecVersion: strVal(data, "spec_version"),
		RiskLevel:   strVal(data, "risk_level"),
		Object:      strVal(data, "object"),
		Relation:    strVal(data, "relation"),
		Source:      strVal(data, "source"),
	}
	if v, ok := data["tags"].([]any); ok {
		for _, t := range v {
			fm.Tags = append(fm.Tags, fmt.Sprint(t))
		}
	}
	if v, ok := data["includes"].([]any); ok {
		for _, i := range v {
			fm.Includes = append(fm.Includes, fmt.Sprint(i))
		}
	}
	if v, ok := data["enabled"].(bool); ok {
		fm.Enabled = &v
	}
	if v, ok := data["requires_approval"].(bool); ok {
		fm.RequiresApproval = &v
	}

	known := map[string]bool{
		"type": true, "id": true, "name": true, "version": true, "tags": true,
		"description": true, "includes": true, "network": true, "namespace": true,
		"owner": true, "spec_version": true, "enabled": true, "risk_level": true,
		"requires_approval": true, "object": true, "relation": true, "source": true,
	}
	fm.Extra = make(map[string]any)
	for k, v := range data {
		if !known[k] {
			fm.Extra[k] = v
		}
	}
	return fm, nil
}

func strVal(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		return fmt.Sprint(v)
	}
	return ""
}

// ParseBody parses the Markdown body of a .bkn file into lists of definitions.
func ParseBody(text string) ([]BknObject, []Relation, []Action) {
	_, body := splitFrontmatter(text)
	matches := definitionRE.FindAllStringSubmatchIndex(body, -1)
	var objects []BknObject
	var relations []Relation
	var actions []Action

	for i, m := range matches {
		defType := body[m[2]:m[3]]
		defID := body[m[4]:m[5]]
		start := m[1]
		end := len(body)
		if i+1 < len(matches) {
			end = matches[i+1][0]
		}
		blockText := body[start:end]
		hrSplit := regexp.MustCompile(`(?m)^\s*---\s*$`).Split(blockText, 2)
		blockText = hrSplit[0]

		switch defType {
		case "Object":
			objects = append(objects, parseObjectBlock(defID, blockText))
		case "Relation":
			relations = append(relations, parseRelationBlock(defID, blockText))
		case "Action":
			actions = append(actions, parseActionBlock(defID, blockText))
		}
	}
	return objects, relations, actions
}

// ParseDataTables parses .bknd body into DataTable list.
func ParseDataTables(text string, fm *Frontmatter, sourcePath string) ([]DataTable, error) {
	if fm == nil {
		var err error
		fm, err = ParseFrontmatter(text)
		if err != nil {
			return nil, err
		}
	}
	_, body := splitFrontmatter(text)

	hasObject := strings.TrimSpace(fm.Object) != ""
	hasRelation := strings.TrimSpace(fm.Relation) != ""
	if hasObject && hasRelation {
		return nil, errors.New("type: data frontmatter must have exactly one of object or relation, got both")
	}
	if !hasObject && !hasRelation {
		return nil, errors.New("type: data frontmatter must have exactly one of object or relation, got neither")
	}

	if !headingRE.MatchString(body) {
		return nil, errors.New("type: data body must have a heading (# or ##) followed by a table")
	}
	m := headingRE.FindStringSubmatchIndex(body)
	tableText := body[m[1]:]
	rawTableLines := extractFirstTableLines(tableText)
	rows := parseTable(rawTableLines)
	columns := parseTableColumns(rawTableLines)

	if len(rawTableLines) < 2 || len(columns) == 0 {
		return nil, errors.New("type: data body must have a valid GFM table (header + separator + rows)")
	}

	isRelation := hasRelation
	objectOrRelation := fm.Relation
	if !isRelation {
		objectOrRelation = fm.Object
	}

	return []DataTable{{
		ObjectOrRelation: objectOrRelation,
		IsRelation:       isRelation,
		Columns:          columns,
		Rows:             rows,
		SourcePath:       sourcePath,
		Network:          fm.Network,
	}}, nil
}

var validBknTypes = map[string]bool{
	"network": true, "object": true, "relation": true, "action": true,
	"fragment": true, "data": true, "delete": true,
}

// Parse parses a complete .bkn/.bknd/.md file into a BknDocument.
// Content must have YAML frontmatter with a valid type field.
func Parse(text string, sourcePath string) (*BknDocument, error) {
	fm, err := ParseFrontmatter(text)
	if err != nil {
		return nil, err
	}
	fmStr, _ := splitFrontmatter(text)
	if strings.TrimSpace(fmStr) == "" {
		hint := ""
		if strings.HasSuffix(strings.ToLower(sourcePath), ".md") {
			hint = " .md files used as BKN must start with YAML frontmatter (--- ... ---)."
		}
		return nil, errors.New("BKN file must have YAML frontmatter with a valid type" + hint)
	}
	typeVal := strings.TrimSpace(fm.Type)
	if typeVal == "" {
		return nil, errors.New("BKN frontmatter must include a valid 'type' field (network, object, relation, action, fragment, data, or delete)")
	}
	if !validBknTypes[typeVal] {
		return nil, fmt.Errorf("invalid BKN type: %q; valid types: network, object, relation, action, fragment, data, delete", typeVal)
	}
	if fm.Type == "data" {
		tables, err := ParseDataTables(text, fm, sourcePath)
		if err != nil {
			return nil, err
		}
		return &BknDocument{
			Frontmatter: *fm,
			DataTables:  tables,
			SourcePath:  sourcePath,
		}, nil
	}
	objects, relations, actions := ParseBody(text)
	return &BknDocument{
		Frontmatter: *fm,
		Objects:     objects,
		Relations:   relations,
		Actions:     actions,
		SourcePath:  sourcePath,
	}, nil
}
