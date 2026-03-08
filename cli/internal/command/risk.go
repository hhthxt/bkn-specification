package command

import (
	"fmt"
	"strings"

	"github.com/kweaver-ai/bkn-specification/cli/internal/input"
	bkn "github.com/kweaver-ai/bkn-specification/sdk/golang/bkn"
	"github.com/spf13/cobra"
)

type riskEvalResult struct {
	ActionID string         `json:"action_id"`
	Context  map[string]any `json:"context,omitempty"`
	Result   bkn.RiskResult `json:"result"`
}

func newRiskCommand(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "risk",
		Short: "Run risk-related operations",
	}
	cmd.AddCommand(newRiskEvalCommand(opts))
	return cmd
}

func newRiskEvalCommand(opts *Options) *cobra.Command {
	var (
		networkPath string
		actionID    string
		rulesPath   string
		contextRaw  []string
	)
	cmd := &cobra.Command{
		Use:   "eval",
		Short: "Evaluate action risk using external rule data",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(networkPath) == "" {
				return fmt.Errorf("--network is required")
			}
			if strings.TrimSpace(actionID) == "" {
				return fmt.Errorf("--action is required")
			}
			if strings.TrimSpace(rulesPath) == "" {
				return fmt.Errorf("--rules is required")
			}
			network, err := bkn.LoadNetwork(networkPath)
			if err != nil {
				return err
			}
			context, err := input.ParseKeyValuePairs(contextRaw)
			if err != nil {
				return err
			}
			var rules []map[string]any
			if err := input.ReadJSON(rulesPath, &rules); err != nil {
				return err
			}
			result := bkn.EvaluateRisk(network, actionID, context, rules)
			payload := riskEvalResult{
				ActionID: actionID,
				Context:  context,
				Result:   result,
			}
			text := fmt.Sprintf("Decision: %s", result.Decision)
			if result.RiskLevel != nil {
				text += fmt.Sprintf("\nRiskLevel: %d", *result.RiskLevel)
			}
			if result.Reason != "" {
				text += fmt.Sprintf("\nReason: %s", result.Reason)
			}
			return emit(cmd, opts.Format, text, payload)
		},
	}
	cmd.Flags().StringVar(&networkPath, "network", "", "Path to the root network file")
	cmd.Flags().StringVar(&actionID, "action", "", "Action ID to evaluate")
	cmd.Flags().StringVar(&rulesPath, "rules", "", "Path to a JSON rules array ('-' for stdin)")
	cmd.Flags().StringArrayVar(&contextRaw, "context", nil, "Context key=value pairs (repeatable)")
	return cmd
}
