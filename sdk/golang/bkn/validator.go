package bkn

import (
	"regexp"
	"strconv"
	"strings"
)

// ValidationError represents a single validation problem.
type ValidationError struct {
	Table   string
	Row     *int
	Column  string
	Code    string
	Message string
}

func (e ValidationError) String() string {
	loc := e.Table
	if e.Row != nil {
		loc += " row " + strconv.Itoa(*e.Row)
	}
	if e.Column != "" {
		loc += " [" + e.Column + "]"
	}
	return loc + ": " + e.Code + " - " + e.Message
}

// ValidationResult aggregates validation outcome.
type ValidationResult struct {
	Errors []ValidationError
}

// OK returns true if there are no errors.
func (r *ValidationResult) OK() bool {
	return len(r.Errors) == 0
}

func parseConstraints(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ";")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func tryFloat(val string) *float64 {
	f, err := strconv.ParseFloat(strings.TrimSpace(val), 64)
	if err != nil {
		return nil
	}
	return &f
}

func checkCell(value string, prop DataProperty, tableName string, rowIdx int, errors *[]ValidationError) {
	val := strings.TrimSpace(value)
	constraints := parseConstraints(prop.Constraint)
	col := prop.Property

	for _, cst := range constraints {
		switch {
		case cst == "not_null":
			if val == "" {
				*errors = append(*errors, ValidationError{tableName, &rowIdx, col, "not_null", "value must not be empty"})
			}
		case strings.HasPrefix(cst, "regex:"):
			pattern := strings.TrimSpace(cst[6:])
			if val != "" {
				matched, _ := regexp.MatchString(pattern, val)
				if !matched {
					*errors = append(*errors, ValidationError{tableName, &rowIdx, col, "regex", "'" + value + "' does not match /" + pattern + "/"})
				}
			}
		case strings.HasPrefix(cst, "in(") && strings.HasSuffix(cst, ")"):
			inner := cst[3 : len(cst)-1]
			allowed := strings.Split(inner, ",")
			for i := range allowed {
				allowed[i] = strings.TrimSpace(allowed[i])
			}
			if val != "" {
				found := false
				for _, a := range allowed {
					if val == a {
						found = true
						break
					}
				}
				if !found {
					*errors = append(*errors, ValidationError{tableName, &rowIdx, col, "in", "'" + value + "' not in allowed set"})
				}
			}
		case strings.HasPrefix(cst, "not_in(") && strings.HasSuffix(cst, ")"):
			inner := cst[7 : len(cst)-1]
			forbidden := strings.Split(inner, ",")
			for i := range forbidden {
				forbidden[i] = strings.TrimSpace(forbidden[i])
			}
			if val != "" {
				for _, f := range forbidden {
					if val == f {
						*errors = append(*errors, ValidationError{tableName, &rowIdx, col, "not_in", "'" + value + "' is forbidden"})
						break
					}
				}
			}
		case strings.HasPrefix(cst, "range(") && strings.HasSuffix(cst, ")"):
			inner := cst[6 : len(cst)-1]
			parts := strings.Split(inner, ",")
			if len(parts) == 2 && val != "" {
				lo := tryFloat(parts[0])
				hi := tryFloat(parts[1])
				v := tryFloat(val)
				if lo != nil && hi != nil && v != nil && (*v < *lo || *v > *hi) {
					*errors = append(*errors, ValidationError{tableName, &rowIdx, col, "range", "value not in range"})
				}
			}
		case strings.HasPrefix(cst, ">="):
			threshold := tryFloat(strings.TrimSpace(cst[2:]))
			v := tryFloat(val)
			if threshold != nil && v != nil && *v < *threshold {
				*errors = append(*errors, ValidationError{tableName, &rowIdx, col, ">=", "value below threshold"})
			}
		case strings.HasPrefix(cst, "<="):
			threshold := tryFloat(strings.TrimSpace(cst[2:]))
			v := tryFloat(val)
			if threshold != nil && v != nil && *v > *threshold {
				*errors = append(*errors, ValidationError{tableName, &rowIdx, col, "<=", "value above threshold"})
			}
		case strings.HasPrefix(cst, ">") && !strings.HasPrefix(cst, ">="):
			threshold := tryFloat(strings.TrimSpace(cst[1:]))
			v := tryFloat(val)
			if threshold != nil && v != nil && *v <= *threshold {
				*errors = append(*errors, ValidationError{tableName, &rowIdx, col, ">", "value not above threshold"})
			}
		case strings.HasPrefix(cst, "<") && !strings.HasPrefix(cst, "<="):
			threshold := tryFloat(strings.TrimSpace(cst[1:]))
			v := tryFloat(val)
			if threshold != nil && v != nil && *v >= *threshold {
				*errors = append(*errors, ValidationError{tableName, &rowIdx, col, "<", "value not below threshold"})
			}
		case strings.HasPrefix(cst, "== "):
			expected := strings.TrimSpace(cst[3:])
			if val != "" && val != expected {
				*errors = append(*errors, ValidationError{tableName, &rowIdx, col, "==", "'" + value + "' != '" + expected + "'"})
			}
		case strings.HasPrefix(cst, "!= "):
			forbidden := strings.TrimSpace(cst[3:])
			if val == forbidden {
				*errors = append(*errors, ValidationError{tableName, &rowIdx, col, "!=", "value must not be '" + forbidden + "'"})
			}
		}
	}

	propType := strings.ToLower(strings.TrimSpace(prop.Type))
	if val != "" {
		switch propType {
		case "bool":
			vl := strings.ToLower(val)
			if vl != "true" && vl != "false" {
				*errors = append(*errors, ValidationError{tableName, &rowIdx, col, "type_bool", "'" + value + "' is not a valid bool"})
			}
		case "int32", "int64", "integer", "float32", "float64", "float":
			if tryFloat(val) == nil {
				*errors = append(*errors, ValidationError{tableName, &rowIdx, col, "type_numeric", "'" + value + "' is not a valid " + propType})
			}
		default:
			if strings.HasPrefix(propType, "decimal") && tryFloat(val) == nil {
				*errors = append(*errors, ValidationError{tableName, &rowIdx, col, "type_numeric", "'" + value + "' is not a valid " + propType})
			}
		}
	}
}

// ValidateDataTable validates a DataTable against its Object schema.
// If schema is nil and network is provided, schema is looked up by table.ObjectOrRelation.
func ValidateDataTable(table *DataTable, schema *BknObject, network *BknNetwork) *ValidationResult {
	result := &ValidationResult{}
	tableName := table.ObjectOrRelation
	if tableName == "" {
		tableName = table.SourcePath
	}
	if table.IsRelation {
		return result
	}
	if schema == nil && network != nil {
		for _, e := range network.AllObjects() {
			if e.ID == table.ObjectOrRelation {
				schema = &e
				break
			}
		}
	}
	if schema == nil {
		result.Errors = append(result.Errors, ValidationError{
			Table: tableName, Column: "", Code: "no_schema",
			Message: "no Object schema found for '" + table.ObjectOrRelation + "'",
		})
		return result
	}

	schemaProps := make(map[string]DataProperty)
	for _, dp := range schema.DataProperties {
		schemaProps[dp.Property] = dp
	}
	schemaPropNames := make(map[string]bool)
	for k := range schemaProps {
		schemaPropNames[k] = true
	}

	for _, c := range table.Columns {
		if !schemaPropNames[c] {
			result.Errors = append(result.Errors, ValidationError{
				Table: tableName, Column: c, Code: "extra_column",
				Message: "column '" + c + "' not defined in Object schema",
			})
		}
	}
	for p := range schemaPropNames {
		found := false
		for _, c := range table.Columns {
			if c == p {
				found = true
				break
			}
		}
		if !found {
			result.Errors = append(result.Errors, ValidationError{
				Table: tableName, Column: p, Code: "missing_column",
				Message: "schema property '" + p + "' not present in data",
			})
		}
	}

	var pkProps []DataProperty
	for _, dp := range schema.DataProperties {
		if dp.PrimaryKey {
			pkProps = append(pkProps, dp)
		}
	}
	pkSeen := make(map[string][]int)

	for rowIdx, row := range table.Rows {
		idx := rowIdx + 1
		for colName, dp := range schemaProps {
			v, ok := row[colName]
			if !ok {
				continue
			}
			checkCell(v, dp, tableName, idx, &result.Errors)
		}
		if len(pkProps) > 0 {
			var pkParts []string
			for _, dp := range pkProps {
				pkParts = append(pkParts, row[dp.Property])
			}
			pkKey := strings.Join(pkParts, "|")
			pkSeen[pkKey] = append(pkSeen[pkKey], idx)
		}
	}

	for pkKey, rows := range pkSeen {
		if len(rows) > 1 {
			var pkCols []string
			for _, dp := range pkProps {
				pkCols = append(pkCols, dp.Property)
			}
			result.Errors = append(result.Errors, ValidationError{
				Table: tableName, Row: &rows[0], Column: strings.Join(pkCols, ", "),
				Code: "pk_duplicate", Message: "duplicate primary key '" + pkKey + "'",
			})
		}
	}
	return result
}

// ValidateNetworkData validates all DataTables in a network against their Object schemas.
func ValidateNetworkData(network *BknNetwork) *ValidationResult {
	result := &ValidationResult{}
	for _, table := range network.AllDataTables() {
		tr := ValidateDataTable(&table, nil, network)
		result.Errors = append(result.Errors, tr.Errors...)
	}
	return result
}
