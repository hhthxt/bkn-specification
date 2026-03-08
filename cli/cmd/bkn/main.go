package main

import (
	"fmt"
	"os"

	"github.com/kweaver-ai/bkn-specification/cli/internal/command"
)

func main() {
	if err := command.Execute(); err != nil {
		if !command.IsSilentError(err) {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}
