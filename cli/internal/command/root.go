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
		Use:           "bkn",
		Short:         "BKN command-line tools",
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
