package bkn

import (
	"fmt"
	"os"
	"path/filepath"
)

// Load loads and parses a single .bkn/.bknd file.
func Load(path string) (*BknDocument, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	return Parse(string(data), abs)
}

// LoadNetwork loads a network .bkn file and recursively resolves its includes.
// Only files listed in frontmatter includes are loaded.
func LoadNetwork(rootPath string) (*BknNetwork, error) {
	absRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, err
	}
	rootDoc, err := Load(absRoot)
	if err != nil {
		return nil, err
	}
	loadedPaths := map[string]bool{absRoot: true}
	var includes []BknDocument
	baseDir := filepath.Dir(absRoot)
	if err := resolveIncludes(rootDoc, baseDir, loadedPaths, &includes); err != nil {
		return nil, err
	}
	return &BknNetwork{
		Root:     *rootDoc,
		Includes: includes,
	}, nil
}

func resolveIncludes(doc *BknDocument, baseDir string, loadedPaths map[string]bool, result *[]BknDocument) error {
	for _, includeRel := range doc.Frontmatter.Includes {
		includePath := filepath.Join(baseDir, includeRel)
		absPath, err := filepath.Abs(includePath)
		if err != nil {
			absPath = includePath
		}
		if loadedPaths[absPath] {
			return fmt.Errorf("circular include detected: %s (resolved to %s)", includeRel, absPath)
		}
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			return fmt.Errorf("include file not found: %s (resolved to %s)", includeRel, absPath)
		}
		loadedPaths[absPath] = true
		incDoc, err := Load(absPath)
		if err != nil {
			return err
		}
		*result = append(*result, *incDoc)
		if err := resolveIncludes(incDoc, filepath.Dir(absPath), loadedPaths, result); err != nil {
			return err
		}
	}
	return nil
}
