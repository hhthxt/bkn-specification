package command

import (
	"fmt"
	"os"
	"strings"

	"github.com/kweaver-ai/bkn-specification/cli/internal/input"
	bkn "github.com/kweaver-ai/bkn-specification/sdk/golang/bkn"
	"github.com/spf13/cobra"
)

type dataResult struct {
	ObjectID   string `json:"object_id,omitempty"`
	RelationID string `json:"relation_id,omitempty"`
	Out        string `json:"out,omitempty"`
	Content    string `json:"content,omitempty"`
}

func newDataCommand(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "data",
		Short: "Work with BKN data files",
	}
	cmd.AddCommand(newToBkndCommand(opts))
	return cmd
}

func newToBkndCommand(opts *Options) *cobra.Command {
	var (
		objectID   string
		relationID string
		inPath     string
		outPath    string
		network    string
		source     string
		columnsRaw string
	)
	cmd := &cobra.Command{
		Use:   "to-bknd",
		Short: "Serialize JSON rows to .bknd",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(inPath) == "" {
				return fmt.Errorf("--in is required")
			}
			var rows []map[string]any
			if err := input.ReadJSON(inPath, &rows); err != nil {
				return err
			}
			var columns []string
			if strings.TrimSpace(columnsRaw) != "" {
				for _, column := range strings.Split(columnsRaw, ",") {
					column = strings.TrimSpace(column)
					if column != "" {
						columns = append(columns, column)
					}
				}
			}
			content, err := bkn.ToBknd(bkn.ToBkndOptions{
				ObjectID:   objectID,
				RelationID: relationID,
				Rows:       rows,
				Network:    network,
				Source:     source,
				Columns:    columns,
			})
			if err != nil {
				return err
			}
			if outPath != "" {
				if err := os.WriteFile(outPath, []byte(content), 0644); err != nil {
					return err
				}
			}
			payload := dataResult{
				ObjectID:   objectID,
				RelationID: relationID,
				Out:        outPath,
			}
			text := content
			if outPath != "" {
				text = fmt.Sprintf("Wrote %s", outPath)
			} else {
				payload.Content = content
			}
			return emit(cmd, opts.Format, text, payload)
		},
	}
	cmd.Flags().StringVar(&objectID, "object", "", "Object ID for the output data file")
	cmd.Flags().StringVar(&relationID, "relation", "", "Relation ID for the output data file")
	cmd.Flags().StringVar(&inPath, "in", "", "Path to a JSON array file ('-' for stdin)")
	cmd.Flags().StringVar(&outPath, "out", "", "Write output to a file instead of stdout")
	cmd.Flags().StringVar(&network, "network", "", "Network ID to embed in frontmatter")
	cmd.Flags().StringVar(&source, "source", "", "Optional source field for frontmatter")
	cmd.Flags().StringVar(&columnsRaw, "columns", "", "Optional comma-separated column order")
	return cmd
}
