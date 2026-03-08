package command

import (
	"github.com/kweaver-ai/bkn-specification/cli/internal/output"
	"github.com/spf13/cobra"
)

func emit(cmd *cobra.Command, format string, text string, value any) error {
	if format == output.FormatJSON {
		return output.PrintJSON(cmd.OutOrStdout(), value)
	}
	return output.PrintText(cmd.OutOrStdout(), text)
}
