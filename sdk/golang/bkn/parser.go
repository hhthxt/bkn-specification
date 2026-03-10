package bkn

import (
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
	"端点":   "Endpoint", "凭据引用": "Secret Ref",
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
	"管控范围": "Control Scope", "管控策略": "Control Policy",
	"前置检查": "Pre-checks", "回滚方案": "Rollback Plan", "审计要求": "Audit Requirements",
}

var definitionRE = regexp.MustCompile(`(?m)^##\s+(Object|Relation|Action|Risk):\s*(\S+)`)
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
	row = strings.TrimPrefix(row, "|")
	row = strings.TrimSuffix(row, "|")
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
	header = strings.TrimPrefix(header, "|")
	header = strings.TrimSuffix(header, "|")
	parts := strings.Split(header, "|")
	var cols []string
	for _, p := range parts {
		cols = append(cols, normalizeColumn(strings.TrimSpace(p)))
	}
	return cols
}

func parseDataSource(sectionText string) *ResourceInfo {
	rows := parseTable(strings.Split(sectionText, "\n"))
	if len(rows) == 0 {
		return nil
	}
	r := rows[0]
	return &ResourceInfo{
		Type: r["Type"],
		ID:   r["ID"],
		Name: r["Name"],
	}
}

func parseDataProperties(sectionText string) []*DataProperty {
	rows := parseTable(strings.Split(sectionText, "\n"))
	var props []*DataProperty
	for _, row := range rows {
		props = append(props, &DataProperty{
			Name:        row["Name"],
			DisplayName: row["Display Name"],
			Type:        row["Type"],
			Description: row["Description"],
		})
	}
	return props
}

func parseLogicProperties(sectionText string) []*LogicProperty {
	rows := parseTable(strings.Split(sectionText, "\n"))
	var props []*LogicProperty
	for _, row := range rows {
		props = append(props, &LogicProperty{
			Name:        row["Name"],
			DisplayName: row["Display Name"],
			Type:        row["Type"],
			Description: row["Description"],
		})
	}
	return props
}

func parseKeys(sectionText string) (pks []string, dk string, ik string) {
	rows := parseTable(strings.Split(sectionText, "\n"))
	if len(rows) == 0 {
		return nil, "", ""
	}
	r := rows[0]
	pks = strings.Split(r["Primary Keys"], ",")
	dk = r["Display Key"]
	ik = r["Incremental Key"]
	return pks, dk, ik
}

// ParseFrontmatter parses the YAML frontmatter of a .bkn file.
func ParseFrontmatter(text string) (map[string]any, error) {
	fmStr, _ := splitFrontmatter(text)
	if fmStr == "" {
		return map[string]any{}, nil
	}
	var data map[string]any
	if err := yaml.Unmarshal([]byte(fmStr), &data); err != nil {
		return nil, err
	}
	if data == nil {
		data = make(map[string]any)
	}

	return data, nil
}

func strVal(m map[string]any, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprint(v)
	}
	return ""
}

// strSliceVal safely extracts a string slice from a map value.
// YAML unmarshals arrays as []interface{}, so we need to convert each element.
func strSliceVal(m map[string]any, key string) []string {
	if v, ok := m[key]; ok && v != nil {
		switch val := v.(type) {
		case []string:
			return val
		case []interface{}:
			result := make([]string, 0, len(val))
			for _, item := range val {
				if item != nil {
					result = append(result, fmt.Sprint(item))
				}
			}
			return result
		case string:
			// Handle single string as single-element slice
			return []string{val}
		}
	}
	return nil
}

// ParseNetworkFile parses a network.bkn file (type: network).
// Network files contain only frontmatter, no body definitions.
func ParseNetworkFile(text string, sourcePath string) (*BknNetwork, error) {
	fmData, err := ParseFrontmatter(text)
	if err != nil {
		return nil, err
	}

	network := &BknNetwork{
		BknNetworkFrontmatter: BknNetworkFrontmatter{
			Type:           strVal(fmData, "type"),
			ID:             strVal(fmData, "id"),
			Name:           strVal(fmData, "name"),
			Tags:           strSliceVal(fmData, "tags"),
			Description:    strVal(fmData, "description"),
			Version:        strVal(fmData, "version"),
			Branch:         strVal(fmData, "branch"),
			BusinessDomain: strVal(fmData, "business_domain"),
		},
	}

	return network, nil
}

// ParseObjectTypeFile parses an object_type definition file.
func ParseObjectTypeFile(text string, sourcePath string) (*BknObjectType, error) {
	fmData, err := ParseFrontmatter(text)
	if err != nil {
		return nil, err
	}

	obj := &BknObjectType{
		BknObjectTypeFrontmatter: BknObjectTypeFrontmatter{
			Type:        "object_type",
			ID:          strVal(fmData, "id"),
			Name:        strVal(fmData, "name"),
			Tags:        strSliceVal(fmData, "tags"),
			Description: strVal(fmData, "description"),
		},
	}

	sections := extractSections(text, "###")
	if s, ok := sections["Data Source"]; ok {
		obj.DataSource = parseDataSource(s)
	}
	if s, ok := sections["Data Properties"]; ok {
		obj.DataProperties = parseDataProperties(s)
	}
	if s, ok := sections["Logic Properties"]; ok {
		obj.LogicProperties = parseLogicProperties(s)
	}
	if s, ok := sections["Keys"]; ok {
		pks, dk, ik := parseKeys(s)
		obj.PrimaryKeys = pks
		obj.DisplayKey = dk
		obj.IncrementalKey = ik
	}

	return obj, nil
}

// ParseRelationTypeFile parses a relation_type definition file.
func ParseRelationTypeFile(text string, sourcePath string) (*BknRelationType, error) {
	fmData, err := ParseFrontmatter(text)
	if err != nil {
		return nil, err
	}

	rel := &BknRelationType{
		BknRelationTypeFrontmatter: BknRelationTypeFrontmatter{
			Type:        "relation_type",
			ID:          strVal(fmData, "id"),
			Name:        strVal(fmData, "name"),
			Tags:        strSliceVal(fmData, "tags"),
			Description: strVal(fmData, "description"),
		},
	}

	return rel, nil
}

// ParseActionTypeFile parses an action_type definition file.
func ParseActionTypeFile(text string, sourcePath string) (*BknActionType, error) {
	fmData, err := ParseFrontmatter(text)
	if err != nil {
		return nil, err
	}

	act := &BknActionType{
		BknActionTypeFrontmatter: BknActionTypeFrontmatter{
			Type:        "action_type",
			ID:          strVal(fmData, "id"),
			Name:        strVal(fmData, "name"),
			Tags:        strSliceVal(fmData, "tags"),
			Description: strVal(fmData, "description"),
		},
	}

	return act, nil
}

// ParseRiskTypeFile parses a risk_type definition file.
func ParseRiskTypeFile(text string, sourcePath string) (*BknRiskType, error) {
	fmData, err := ParseFrontmatter(text)
	if err != nil {
		return nil, err
	}

	risk := &BknRiskType{
		BknRiskTypeFrontmatter: BknRiskTypeFrontmatter{
			Type:        "risk_type",
			ID:          strVal(fmData, "id"),
			Name:        strVal(fmData, "name"),
			Tags:        strSliceVal(fmData, "tags"),
			Description: strVal(fmData, "description"),
		},
	}

	sections := extractSections(text, "###")

	if s, ok := sections["Control Scope"]; ok {
		risk.ControlScope = s
	}
	if s, ok := sections["Control Policy"]; ok {
		risk.ControlPolicy = s
	}

	return risk, nil
}

func ParseConceptGroupFile(text string, sourcePath string) (*BknConceptGroup, error) {
	fmData, err := ParseFrontmatter(text)
	if err != nil {
		return nil, err
	}

	cg := &BknConceptGroup{
		BknConceptGroupFrontmatter: BknConceptGroupFrontmatter{
			Type:        "concept_group",
			ID:          strVal(fmData, "id"),
			Name:        strVal(fmData, "name"),
			Tags:        strSliceVal(fmData, "tags"),
			Description: strVal(fmData, "description"),
		},
	}

	return cg, nil
}
