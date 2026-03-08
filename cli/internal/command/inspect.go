package command

import (
	"fmt"
	"os"
	"strings"

	bkn "github.com/kweaver-ai/bkn-specification/sdk/golang/bkn"
	"github.com/spf13/cobra"
)

type fileInspectResult struct {
	Path        string                 `json:"path"`
	Frontmatter bkn.Frontmatter        `json:"frontmatter"`
	Title       string                 `json:"title,omitempty"`
	Description string                 `json:"description,omitempty"`
	Counts      map[string]int         `json:"counts"`
	Details     *networkInspectDetails `json:"details,omitempty"`
}

type networkInspectResult struct {
	Path        string                 `json:"path"`
	Root        bkn.Frontmatter        `json:"root_frontmatter"`
	Title       string                 `json:"title,omitempty"`
	Description string                 `json:"description,omitempty"`
	Counts      map[string]int         `json:"counts"`
	Details     *networkInspectDetails `json:"details,omitempty"`
}

type networkInspectDetails struct {
	Objects     []namedDefinition `json:"objects,omitempty"`
	Relations   []namedDefinition `json:"relations,omitempty"`
	Actions     []namedDefinition `json:"actions,omitempty"`
	Risks       []namedDefinition `json:"risks,omitempty"`
	Connections []namedDefinition `json:"connections,omitempty"`
}

type namedDefinition struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

func newInspectCommand(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inspect",
		Short: "Inspect BKN files and networks",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newInspectFileCommand(opts), newInspectNetworkCommand(opts))
	return cmd
}

func newInspectFileCommand(opts *Options) *cobra.Command {
	var verbose bool
	cmd := &cobra.Command{
		Use:   "file <path>",
		Short: "Inspect a single BKN file",
		Example: "  bkn inspect file examples/risk/actions.bkn\n" +
			"  bkn inspect file examples/risk/actions.bkn --verbose\n" +
			"  bkn inspect file examples/risk/data/risk_scenario.bknd --format json",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			doc, err := bkn.Load(args[0])
			if err != nil {
				return err
			}
			title, description := readDocumentOverview(doc.SourcePath)
			result := fileInspectResult{
				Path:        args[0],
				Frontmatter: doc.Frontmatter,
				Title:       title,
				Description: description,
				Counts: map[string]int{
					"objects":     len(doc.Objects),
					"relations":   len(doc.Relations),
					"actions":     len(doc.Actions),
					"risks":       len(doc.Risks),
					"connections": len(doc.Connections),
					"data_tables": len(doc.DataTables),
				},
			}
			lines := []string{
				fmt.Sprintf("File: %s", result.Path),
				fmt.Sprintf("Type: %s", result.Frontmatter.Type),
				fmt.Sprintf("ID: %s", result.Frontmatter.ID),
				fmt.Sprintf("Objects: %d", result.Counts["objects"]),
				fmt.Sprintf("Relations: %d", result.Counts["relations"]),
				fmt.Sprintf("Actions: %d", result.Counts["actions"]),
				fmt.Sprintf("Risks: %d", result.Counts["risks"]),
				fmt.Sprintf("Connections: %d", result.Counts["connections"]),
				fmt.Sprintf("DataTables: %d", result.Counts["data_tables"]),
			}
			if verbose {
				result.Details = &networkInspectDetails{
					Objects:     collectObjectDetails(doc.Objects),
					Relations:   collectRelationDetails(doc.Relations),
					Actions:     collectActionDetails(doc.Actions),
					Risks:       collectRiskDetails(doc.Risks),
					Connections: collectConnectionDetails(doc.Connections),
				}
				if result.Frontmatter.Name != "" {
					lines = append(lines, fmt.Sprintf("Name: %s", result.Frontmatter.Name))
				}
				if result.Frontmatter.Version != "" {
					lines = append(lines, fmt.Sprintf("Version: %s", result.Frontmatter.Version))
				}
				if result.Frontmatter.Network != "" {
					lines = append(lines, fmt.Sprintf("Network: %s", result.Frontmatter.Network))
				}
				if result.Frontmatter.Namespace != "" {
					lines = append(lines, fmt.Sprintf("Namespace: %s", result.Frontmatter.Namespace))
				}
				if result.Frontmatter.Owner != "" {
					lines = append(lines, fmt.Sprintf("Owner: %s", result.Frontmatter.Owner))
				}
				if result.Title != "" {
					lines = append(lines, fmt.Sprintf("Title: %s", result.Title))
				}
				if result.Description != "" {
					lines = append(lines, fmt.Sprintf("Description: %s", result.Description))
				}
				appendDefinitionLines := func(label string, items []namedDefinition) {
					if len(items) == 0 {
						return
					}
					lines = append(lines, label+":")
					for _, item := range items {
						if item.Name != "" {
							lines = append(lines, fmt.Sprintf("  - %s (%s)", item.ID, item.Name))
						} else {
							lines = append(lines, "  - "+item.ID)
						}
					}
				}
				appendDefinitionLines("Objects", result.Details.Objects)
				appendDefinitionLines("Relations", result.Details.Relations)
				appendDefinitionLines("Actions", result.Details.Actions)
				appendDefinitionLines("Risks", result.Details.Risks)
				appendDefinitionLines("Connections", result.Details.Connections)
			}
			return emit(cmd, opts.Format, strings.Join(lines, "\n"), result)
		},
	}
	cmd.Flags().BoolVar(&verbose, "verbose", false, "Include file metadata and definition details")
	return cmd
}

func newInspectNetworkCommand(opts *Options) *cobra.Command {
	var verbose bool
	cmd := &cobra.Command{
		Use:   "network <path>",
		Short: "Inspect a BKN network with includes resolved",
		Example: "  bkn inspect network examples/risk/index.bkn\n" +
			"  bkn inspect network examples/risk/index.bkn --verbose\n" +
			"  bkn inspect network examples/k8s-modular/index.bkn --format json",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			network, err := bkn.LoadNetwork(args[0])
			if err != nil {
				return err
			}
			title, description := readDocumentOverview(network.Root.SourcePath)
			result := networkInspectResult{
				Path:        args[0],
				Root:        network.Root.Frontmatter,
				Title:       title,
				Description: description,
				Counts: map[string]int{
					"objects":     len(network.AllObjects()),
					"relations":   len(network.AllRelations()),
					"actions":     len(network.AllActions()),
					"risks":       len(network.AllRisks()),
					"connections": len(network.AllConnections()),
					"data_tables": len(network.AllDataTables()),
				},
			}
			lines := []string{
				fmt.Sprintf("Network: %s", result.Path),
				fmt.Sprintf("Type: %s", result.Root.Type),
				fmt.Sprintf("ID: %s", result.Root.ID),
				fmt.Sprintf("Objects: %d", result.Counts["objects"]),
				fmt.Sprintf("Relations: %d", result.Counts["relations"]),
				fmt.Sprintf("Actions: %d", result.Counts["actions"]),
				fmt.Sprintf("Risks: %d", result.Counts["risks"]),
				fmt.Sprintf("Connections: %d", result.Counts["connections"]),
				fmt.Sprintf("DataTables: %d", result.Counts["data_tables"]),
			}
			if verbose {
				result.Details = &networkInspectDetails{
					Objects:     collectObjectDetails(network.AllObjects()),
					Relations:   collectRelationDetails(network.AllRelations()),
					Actions:     collectActionDetails(network.AllActions()),
					Risks:       collectRiskDetails(network.AllRisks()),
					Connections: collectConnectionDetails(network.AllConnections()),
				}
				if result.Root.Name != "" {
					lines = append(lines, fmt.Sprintf("Name: %s", result.Root.Name))
				}
				if result.Root.Version != "" {
					lines = append(lines, fmt.Sprintf("Version: %s", result.Root.Version))
				}
				if result.Root.Namespace != "" {
					lines = append(lines, fmt.Sprintf("Namespace: %s", result.Root.Namespace))
				}
				if result.Root.Owner != "" {
					lines = append(lines, fmt.Sprintf("Owner: %s", result.Root.Owner))
				}
				if result.Title != "" {
					lines = append(lines, fmt.Sprintf("Title: %s", result.Title))
				}
				if result.Description != "" {
					lines = append(lines, fmt.Sprintf("Description: %s", result.Description))
				}
				if len(result.Root.Includes) > 0 {
					lines = append(lines, "Includes:")
					for _, include := range result.Root.Includes {
						lines = append(lines, "  - "+include)
					}
				}
				appendDefinitionLines := func(label string, items []namedDefinition) {
					if len(items) == 0 {
						return
					}
					lines = append(lines, label+":")
					for _, item := range items {
						if item.Name != "" {
							lines = append(lines, fmt.Sprintf("  - %s (%s)", item.ID, item.Name))
						} else {
							lines = append(lines, "  - "+item.ID)
						}
					}
				}
				appendDefinitionLines("Objects", result.Details.Objects)
				appendDefinitionLines("Relations", result.Details.Relations)
				appendDefinitionLines("Actions", result.Details.Actions)
				appendDefinitionLines("Risks", result.Details.Risks)
				appendDefinitionLines("Connections", result.Details.Connections)
			}
			return emit(cmd, opts.Format, strings.Join(lines, "\n"), result)
		},
	}
	cmd.Flags().BoolVar(&verbose, "verbose", false, "Include network metadata and definition details")
	return cmd
}

func collectObjectDetails(items []bkn.BknObject) []namedDefinition {
	out := make([]namedDefinition, 0, len(items))
	for _, item := range items {
		out = append(out, namedDefinition{ID: item.ID, Name: item.Name})
	}
	return out
}

func collectRelationDetails(items []bkn.Relation) []namedDefinition {
	out := make([]namedDefinition, 0, len(items))
	for _, item := range items {
		out = append(out, namedDefinition{ID: item.ID, Name: item.Name})
	}
	return out
}

func collectActionDetails(items []bkn.Action) []namedDefinition {
	out := make([]namedDefinition, 0, len(items))
	for _, item := range items {
		out = append(out, namedDefinition{ID: item.ID, Name: item.Name})
	}
	return out
}

func collectRiskDetails(items []bkn.Risk) []namedDefinition {
	out := make([]namedDefinition, 0, len(items))
	for _, item := range items {
		out = append(out, namedDefinition{ID: item.ID, Name: item.Name})
	}
	return out
}

func collectConnectionDetails(items []bkn.Connection) []namedDefinition {
	out := make([]namedDefinition, 0, len(items))
	for _, item := range items {
		out = append(out, namedDefinition{ID: item.ID, Name: item.Name})
	}
	return out
}

func readDocumentOverview(path string) (string, string) {
	if strings.TrimSpace(path) == "" {
		return "", ""
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", ""
	}
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	if strings.HasPrefix(text, "---") {
		if end := strings.Index(text[4:], "\n---\n"); end >= 0 {
			text = text[4+end+5:]
		}
	}
	lines := strings.Split(text, "\n")
	var (
		title       string
		description []string
	)
	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "---" {
			break
		}
		if strings.HasPrefix(line, "## ") {
			break
		}
		if strings.HasPrefix(line, "# ") {
			if title == "" {
				title = strings.TrimSpace(strings.TrimPrefix(line, "# "))
			}
			continue
		}
		if line != "" {
			description = append(description, line)
		}
	}
	return title, strings.Join(description, " ")
}
