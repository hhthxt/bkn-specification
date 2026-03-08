package command

import (
	"fmt"
	"strings"

	"github.com/kweaver-ai/bkn-specification/cli/internal/input"
	bkn "github.com/kweaver-ai/bkn-specification/sdk/golang/bkn"
	"github.com/spf13/cobra"
)

type deletePlanResult struct {
	OK       bool               `json:"ok"`
	Targets  []bkn.DeleteTarget `json:"targets"`
	NotFound []bkn.DeleteTarget `json:"not_found"`
}

type deleteSimulateResult struct {
	Plan   deletePlanResult `json:"plan"`
	Before map[string]int   `json:"before"`
	After  map[string]int   `json:"after"`
}

func newDeleteCommand(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Plan or simulate definition deletion",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newDeletePlanCommand(opts), newDeleteSimulateCommand(opts))
	return cmd
}

func newDeletePlanCommand(opts *Options) *cobra.Command {
	var networkPath string
	var targetRaw []string
	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Plan deletion targets",
		Example: "  bkn delete plan --network examples/k8s-modular/index.bkn --target object:pod\n" +
			"  bkn delete plan --network index.bkn --target object:pod --target action:restart_pod --format json",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(networkPath) == "" {
				return fmt.Errorf("--network is required")
			}
			if len(targetRaw) == 0 {
				return fmt.Errorf("at least one --target is required")
			}
			network, err := bkn.LoadNetwork(networkPath)
			if err != nil {
				return err
			}
			targets, err := input.ParseDeleteTargets(targetRaw)
			if err != nil {
				return err
			}
			plan := bkn.PlanDelete(network, targets, true)
			payload := deletePlanResult{OK: plan.OK(), Targets: plan.Targets, NotFound: plan.NotFound}
			lines := []string{
				fmt.Sprintf("Found: %d", len(plan.Targets)),
				fmt.Sprintf("Missing: %d", len(plan.NotFound)),
			}
			for _, target := range plan.Targets {
				lines = append(lines, fmt.Sprintf("target %s:%s", target.Type, target.ID))
			}
			if err := emit(cmd, opts.Format, strings.Join(lines, "\n"), payload); err != nil {
				return err
			}
			if !plan.OK() {
				return newSilentError("delete plan contains missing targets")
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&networkPath, "network", "", "Path to the root network file")
	cmd.Flags().StringArrayVar(&targetRaw, "target", nil, "Delete target in type:id form (repeatable)")
	return cmd
}

func newDeleteSimulateCommand(opts *Options) *cobra.Command {
	var networkPath string
	var targetRaw []string
	cmd := &cobra.Command{
		Use:   "simulate",
		Short: "Simulate deletion in memory",
		Example: "  bkn delete simulate --network examples/k8s-modular/index.bkn --target object:pod\n" +
			"  bkn delete simulate --network index.bkn --target relation:pod_on_node --format json",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(networkPath) == "" {
				return fmt.Errorf("--network is required")
			}
			if len(targetRaw) == 0 {
				return fmt.Errorf("at least one --target is required")
			}
			network, err := bkn.LoadNetwork(networkPath)
			if err != nil {
				return err
			}
			targets, err := input.ParseDeleteTargets(targetRaw)
			if err != nil {
				return err
			}
			plan := bkn.PlanDelete(network, targets, true)
			if !plan.OK() {
				payload := deleteSimulateResult{
					Plan: deletePlanResult{OK: plan.OK(), Targets: plan.Targets, NotFound: plan.NotFound},
				}
				if err := emit(cmd, opts.Format, "delete plan contains missing targets", payload); err != nil {
					return err
				}
				return newSilentError("delete plan contains missing targets")
			}
			next := bkn.NetworkWithout(network, plan.Targets)
			payload := deleteSimulateResult{
				Plan: deletePlanResult{OK: true, Targets: plan.Targets, NotFound: plan.NotFound},
				Before: map[string]int{
					"objects":   len(network.AllObjects()),
					"relations": len(network.AllRelations()),
					"actions":   len(network.AllActions()),
				},
				After: map[string]int{
					"objects":   len(next.AllObjects()),
					"relations": len(next.AllRelations()),
					"actions":   len(next.AllActions()),
				},
			}
			text := strings.Join([]string{
				fmt.Sprintf("Objects: %d -> %d", payload.Before["objects"], payload.After["objects"]),
				fmt.Sprintf("Relations: %d -> %d", payload.Before["relations"], payload.After["relations"]),
				fmt.Sprintf("Actions: %d -> %d", payload.Before["actions"], payload.After["actions"]),
			}, "\n")
			return emit(cmd, opts.Format, text, payload)
		},
	}
	cmd.Flags().StringVar(&networkPath, "network", "", "Path to the root network file")
	cmd.Flags().StringArrayVar(&targetRaw, "target", nil, "Delete target in type:id form (repeatable)")
	return cmd
}
