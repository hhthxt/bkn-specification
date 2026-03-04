package bkn

import (
	"errors"
	"fmt"
	"strings"
)

// ToBkndOptions holds options for serializing to .bknd format.
type ToBkndOptions struct {
	ObjectID   string
	RelationID string
	Rows       []map[string]interface{}
	Network    string
	Source     string
	Columns    []string
}

func escapeCell(val interface{}) string {
	if val == nil {
		return ""
	}
	s := strings.TrimSpace(fmt.Sprint(val))
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return s
}

func rowToStringMap(row map[string]string) map[string]interface{} {
	out := make(map[string]interface{})
	for k, v := range row {
		out[k] = v
	}
	return out
}

// ToBknd serializes structured data to .bknd Markdown format.
func ToBknd(opts ToBkndOptions) (string, error) {
	if opts.ObjectID != "" && opts.RelationID != "" {
		return "", errors.New("specify either object_id or relation_id, not both")
	}
	if opts.ObjectID == "" && opts.RelationID == "" {
		return "", errors.New("specify either object_id or relation_id")
	}
	if opts.Rows == nil {
		opts.Rows = []map[string]interface{}{}
	}

	targetID := opts.RelationID
	isRelation := opts.RelationID != ""
	if targetID == "" {
		targetID = opts.ObjectID
	}

	columns := opts.Columns
	if columns == nil && len(opts.Rows) > 0 {
		// Infer from first row
		first := opts.Rows[0]
		for k := range first {
			columns = append(columns, k)
		}
	}
	if columns == nil {
		columns = []string{}
	}

	var fmLines []string
	fmLines = append(fmLines, "---", "type: data", "network: "+opts.Network)
	if isRelation {
		fmLines = append(fmLines, "relation: "+opts.RelationID)
	} else {
		fmLines = append(fmLines, "object: "+opts.ObjectID)
	}
	if opts.Source != "" {
		fmLines = append(fmLines, "source: "+opts.Source)
	}
	fmLines = append(fmLines, "---", "", "# "+targetID, "")

	if len(columns) == 0 {
		return strings.Join(fmLines, "\n"), nil
	}

	header := "| " + strings.Join(columns, " | ") + " |"
	sepParts := make([]string, len(columns))
	for i := range sepParts {
		sepParts[i] = "---"
	}
	sep := "|" + strings.Join(sepParts, "|") + "|"
	tableLines := []string{header, sep}
	for _, row := range opts.Rows {
		var cells []string
		for _, c := range columns {
			v, ok := row[c]
			if !ok {
				v = ""
			}
			cells = append(cells, escapeCell(v))
		}
		tableLines = append(tableLines, "| "+strings.Join(cells, " | ")+" |")
	}
	fmLines = append(fmLines, strings.Join(tableLines, "\n"))
	return strings.Join(fmLines, "\n"), nil
}

// ToBkndFromTable serializes a DataTable to .bknd Markdown format.
func ToBkndFromTable(table *DataTable, network, source string) (string, error) {
	net := network
	if net == "" {
		net = table.Network
	}
	cols := table.Columns
	if len(cols) == 0 && len(table.Rows) > 0 {
		for k := range table.Rows[0] {
			cols = append(cols, k)
		}
	}
	rows := make([]map[string]interface{}, len(table.Rows))
	for i, r := range table.Rows {
		rows[i] = rowToStringMap(r)
	}
	if table.IsRelation {
		return ToBknd(ToBkndOptions{
			RelationID: table.ObjectOrRelation,
			Rows:       rows,
			Network:    net,
			Source:     source,
			Columns:    cols,
		})
	}
	return ToBknd(ToBkndOptions{
		ObjectID: table.ObjectOrRelation,
		Rows:     rows,
		Network:  net,
		Source:   source,
		Columns:  cols,
	})
}
