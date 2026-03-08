package command

import (
	"fmt"
	"os"
	"strings"

	"github.com/kweaver-ai/bkn-specification/cli/internal/output"
	"github.com/spf13/cobra"
)

func emit(cmd *cobra.Command, format string, text string, value any) error {
	if format == output.FormatJSON {
		return output.PrintJSON(cmd.OutOrStdout(), value)
	}
	return output.PrintText(cmd.OutOrStdout(), text)
}

func writeFileOrStdout(path string, content string) error {
	if path == "" {
		_, err := fmt.Fprintln(os.Stdout, content)
		return err
	}
	return os.WriteFile(path, []byte(content), 0644)
}

func joinIDs[T any](items []T, getID func(T) string) string {
	ids := make([]string, 0, len(items))
	for _, item := range items {
		ids = append(ids, getID(item))
	}
	return strings.Join(ids, ", ")
}
