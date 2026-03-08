package command

import (
	"fmt"
	"strings"

	bkn "github.com/kweaver-ai/bkn-specification/sdk/golang/bkn"
	"github.com/spf13/cobra"
)

type fileInspectResult struct {
	Path        string          `json:"path"`
	Frontmatter bkn.Frontmatter `json:"frontmatter"`
	Counts      map[string]int  `json:"counts"`
}

type networkInspectResult struct {
	Path   string          `json:"path"`
	Root   bkn.Frontmatter `json:"root_frontmatter"`
	Counts map[string]int  `json:"counts"`
}

func newInspectCommand(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inspect",
		Short: "Inspect BKN files and networks",
	}
	cmd.AddCommand(newInspectFileCommand(opts), newInspectNetworkCommand(opts))
	return cmd
}

func newInspectFileCommand(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "file <path>",
		Short: "Inspect a single BKN file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			doc, err := bkn.Load(args[0])
			if err != nil {
				return err
			}
			result := fileInspectResult{
				Path:        args[0],
				Frontmatter: doc.Frontmatter,
				Counts: map[string]int{
					"objects":     len(doc.Objects),
					"relations":   len(doc.Relations),
					"actions":     len(doc.Actions),
					"risks":       len(doc.Risks),
					"connections": len(doc.Connections),
					"data_tables": len(doc.DataTables),
				},
			}
			text := strings.Join([]string{
				fmt.Sprintf("File: %s", result.Path),
				fmt.Sprintf("Type: %s", result.Frontmatter.Type),
				fmt.Sprintf("ID: %s", result.Frontmatter.ID),
				fmt.Sprintf("Objects: %d", result.Counts["objects"]),
				fmt.Sprintf("Relations: %d", result.Counts["relations"]),
				fmt.Sprintf("Actions: %d", result.Counts["actions"]),
				fmt.Sprintf("Risks: %d", result.Counts["risks"]),
				fmt.Sprintf("Connections: %d", result.Counts["connections"]),
				fmt.Sprintf("DataTables: %d", result.Counts["data_tables"]),
			}, "\n")
			return emit(cmd, opts.Format, text, result)
		},
	}
}

func newInspectNetworkCommand(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "network <path>",
		Short: "Inspect a BKN network with includes resolved",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			network, err := bkn.LoadNetwork(args[0])
			if err != nil {
				return err
			}
			result := networkInspectResult{
				Path: args[0],
				Root: network.Root.Frontmatter,
				Counts: map[string]int{
					"objects":     len(network.AllObjects()),
					"relations":   len(network.AllRelations()),
					"actions":     len(network.AllActions()),
					"risks":       len(network.AllRisks()),
					"connections": len(network.AllConnections()),
					"data_tables": len(network.AllDataTables()),
				},
			}
			text := strings.Join([]string{
				fmt.Sprintf("Network: %s", result.Path),
				fmt.Sprintf("Type: %s", result.Root.Type),
				fmt.Sprintf("ID: %s", result.Root.ID),
				fmt.Sprintf("Objects: %d", result.Counts["objects"]),
				fmt.Sprintf("Relations: %d", result.Counts["relations"]),
				fmt.Sprintf("Actions: %d", result.Counts["actions"]),
				fmt.Sprintf("Risks: %d", result.Counts["risks"]),
				fmt.Sprintf("Connections: %d", result.Counts["connections"]),
				fmt.Sprintf("DataTables: %d", result.Counts["data_tables"]),
			}, "\n")
			return emit(cmd, opts.Format, text, result)
		},
	}
}
