package bkn

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Supported file extensions for BKN content. .md is allowed as a carrier;
// content must still satisfy BKN frontmatter/type/structure requirements.
var supportedExtensions = map[string]bool{
	".bkn": true, ".bknd": true, ".md": true,
}

func checkExtension(path string) error {
	ext := strings.ToLower(filepath.Ext(path))
	if !supportedExtensions[ext] {
		return fmt.Errorf("unsupported file extension: %q; BKN supports: .bkn, .bknd, .md", ext)
	}
	return nil
}

// Load loads and parses a single .bkn/.bknd/.md file.
// Supported extensions: .bkn, .bknd, .md. Content must satisfy BKN
// frontmatter, type, and structure requirements regardless of extension.
func Load(path string) (*BknDocument, error) {
	if err := checkExtension(path); err != nil {
		return nil, err
	}
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

// LoadNetwork loads a network file and recursively resolves its includes.
// Supported extensions: .bkn, .bknd, .md. Root file should be type: network.
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
