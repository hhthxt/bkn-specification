package command

import (
	"github.com/kweaver-ai/bkn-specification/cli/internal/output"
	"github.com/spf13/cobra"
)

type Options struct {
	Format string
}

func Execute() error {
	return NewRootCommand().Execute()
}

func NewRootCommand() *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "bkn",
		Short: "Inspect, validate, and transform BKN files",
		Long:  "bkn provides focused command-line workflows for inspecting BKN files, validating data, evaluating risk, simulating deletes, and generating checksum or .bknd output.",
		Example: "  bkn inspect network examples/risk/index.bkn\n" +
			"  bkn validate network examples/risk/index.bkn\n" +
			"  bkn risk eval --network examples/risk/index.bkn --action restart_erp --context scenario_id=sec_t_01 --rules rules.json",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return output.ValidateFormat(opts.Format)
		},
	}
	cmd.PersistentFlags().StringVar(&opts.Format, "format", output.FormatText, "Output format: text or json")

	cmd.AddCommand(
		newInspectCommand(opts),
		newValidateCommand(opts),
		newDataCommand(opts),
		newRiskCommand(opts),
		newDeleteCommand(opts),
		newChecksumCommand(opts),
	)
	return cmd
}
