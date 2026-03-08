package command

import (
	"fmt"
	"strings"

	bkn "github.com/kweaver-ai/bkn-specification/sdk/golang/bkn"
	"github.com/spf13/cobra"
)

type checksumGenerateResult struct {
	Root    string `json:"root"`
	Content string `json:"content"`
}

type checksumVerifyResult struct {
	Root   string   `json:"root"`
	OK     bool     `json:"ok"`
	Errors []string `json:"errors,omitempty"`
}

func newChecksumCommand(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "checksum",
		Short: "Generate or verify checksum.txt",
	}
	cmd.AddCommand(newChecksumGenerateCommand(opts), newChecksumVerifyCommand(opts))
	return cmd
}

func newChecksumGenerateCommand(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "generate <dir>",
		Short: "Generate checksum.txt for a business directory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			content, err := bkn.GenerateChecksumFile(args[0])
			if err != nil {
				return err
			}
			payload := checksumGenerateResult{Root: args[0], Content: content}
			text := fmt.Sprintf("Generated checksum.txt in %s", args[0])
			return emit(cmd, opts.Format, text, payload)
		},
	}
}

func newChecksumVerifyCommand(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "verify <dir>",
		Short: "Verify checksum.txt for a business directory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ok, errors := bkn.VerifyChecksumFile(args[0])
			payload := checksumVerifyResult{Root: args[0], OK: ok, Errors: errors}
			if ok {
				return emit(cmd, opts.Format, "Checksum OK", payload)
			}
			if err := emit(cmd, opts.Format, strings.Join(errors, "\n"), payload); err != nil {
				return err
			}
			return newSilentError("checksum verification failed")
		},
	}
}
