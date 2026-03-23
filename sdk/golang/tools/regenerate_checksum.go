package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kweaver-ai/bkn-specification/sdk/golang/bkn"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run regenerate_checksum.go <example-dir>")
		os.Exit(1)
	}

	examplesDir := os.Args[1]
	entries, err := os.ReadDir(examplesDir)
	if err != nil {
		fmt.Printf("Error reading dir: %v\n", err)
		os.Exit(1)
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(examplesDir, e.Name())
		if _, err := os.Stat(filepath.Join(dir, "network.bkn")); err != nil {
			continue
		}

		fmt.Printf("Regenerating CHECKSUM for %s...\n", e.Name())
		_, err := bkn.GenerateChecksumFile(dir)
		if err != nil {
			fmt.Printf("  Error: %v\n", err)
		} else {
			fmt.Printf("  Done\n")
		}
	}
}
