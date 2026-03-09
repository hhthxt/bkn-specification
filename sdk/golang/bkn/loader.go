package bkn

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Supported file extensions for BKN content. .md is allowed as a carrier;
// content must still satisfy BKN frontmatter/type/structure requirements.
var supportedExtensions = map[string]bool{
	".bkn": true, ".bknd": true, ".md": true,
}

// Root file discovery order: network.bkn > network.md > index.bkn > index.md
var rootCandidateNames = []string{"network.bkn", "network.md", "index.bkn", "index.md"}

// DiscoverRootFile discovers the root network file in a directory.
// Order: network.bkn > network.md > index.bkn > index.md.
// If none exist, and exactly one file in the directory has type: network,
// use that file. Otherwise return error.
func DiscoverRootFile(directory string) (string, error) {
	abs, err := filepath.Abs(directory)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("not a directory: %s", abs)
	}

	// 1. Check named candidates in order
	for _, name := range rootCandidateNames {
		candidate := filepath.Join(abs, name)
		if st, err := os.Stat(candidate); err == nil && !st.IsDir() {
			ext := strings.ToLower(filepath.Ext(candidate))
			if supportedExtensions[ext] {
				return candidate, nil
			}
		}
	}

	// 2. Scan same directory for type: network files
	var networkFiles []string
	entries, err := os.ReadDir(abs)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if !supportedExtensions[ext] {
			continue
		}
		p := filepath.Join(abs, e.Name())
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		doc, err := Parse(string(data), p)
		if err != nil {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(doc.Frontmatter.Type), "network") {
			networkFiles = append(networkFiles, p)
		}
	}

	if len(networkFiles) == 1 {
		return networkFiles[0], nil
	}
	if len(networkFiles) > 1 {
		names := make([]string, len(networkFiles))
		for i, p := range networkFiles {
			names[i] = filepath.Base(p)
		}
		return "", fmt.Errorf("multiple network roots in %s: %v; use network.bkn or index.bkn as the single root", abs, names)
	}
	return "", fmt.Errorf("no root network file found in %s; expected one of: %s or a single type: network file", abs, strings.Join(rootCandidateNames, ", "))
}

func collectSameDirBknFiles(directory, rootPath string) ([]string, error) {
	abs, err := filepath.Abs(directory)
	if err != nil {
		return nil, err
	}
	absRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, err
	}
	rootName := filepath.Base(absRoot)

	excludeNames := map[string]bool{rootName: true}
	for _, name := range rootCandidateNames {
		candidate := filepath.Join(abs, name)
		if _, err := os.Stat(candidate); err == nil {
			excludeNames[name] = true
		}
	}

	var result []string
	entries, err := os.ReadDir(abs)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if excludeNames[e.Name()] {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if !supportedExtensions[ext] {
			continue
		}
		p := filepath.Join(abs, e.Name())
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		if _, err := Parse(string(data), p); err != nil {
			continue
		}
		result = append(result, p)
	}
	sort.Strings(result)
	return result, nil
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
// When rootPath is a directory, the root file is discovered automatically
// (network.bkn > network.md > index.bkn > index.md).
// If the root has no includes, same-directory BKN files are loaded implicitly.
// Otherwise only files listed in frontmatter includes are loaded.
func LoadNetwork(rootPath string) (*BknNetwork, error) {
	absRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(absRoot)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		absRoot, err = DiscoverRootFile(absRoot)
		if err != nil {
			return nil, err
		}
	}

	rootDoc, err := Load(absRoot)
	if err != nil {
		return nil, err
	}
	loadedPaths := map[string]bool{absRoot: true}
	recursionStack := map[string]bool{}
	var includes []BknDocument
	baseDir := filepath.Dir(absRoot)

	if len(rootDoc.Frontmatter.Includes) > 0 {
		if err := resolveIncludes(rootDoc, baseDir, loadedPaths, recursionStack, &includes); err != nil {
			return nil, err
		}
	} else {
		// No includes: for type: network only, implicitly load same-dir files
		docType := strings.ToLower(strings.TrimSpace(rootDoc.Frontmatter.Type))
		if docType == "network" {
			implicitPaths, err := collectSameDirBknFiles(baseDir, absRoot)
			if err != nil {
				return nil, err
			}
			for _, incPath := range implicitPaths {
				incAbs, _ := filepath.Abs(incPath)
				if loadedPaths[incAbs] {
					continue
				}
				if recursionStack[incAbs] {
					return nil, fmt.Errorf("circular include detected: %s (resolved to %s)", filepath.Base(incPath), incAbs)
				}
				loadedPaths[incAbs] = true
				incDoc, err := Load(incAbs)
				if err != nil {
					return nil, err
				}
				includes = append(includes, *incDoc)
				recursionStack[incAbs] = true
				if err := resolveIncludes(incDoc, filepath.Dir(incAbs), loadedPaths, recursionStack, &includes); err != nil {
					return nil, err
				}
				delete(recursionStack, incAbs)
			}
		}
	}

	network := &BknNetwork{
		Root:     *rootDoc,
		Includes: includes,
	}
	if err := validateNetworkReferences(network); err != nil {
		return nil, err
	}
	return network, nil
}

// resolveIncludes recursively resolves includes from a document's frontmatter.
// Deduplication: paths in loadedPaths are skipped (already loaded).
// Circular: only when path is in recursionStack (back to self in chain).
func resolveIncludes(doc *BknDocument, baseDir string, loadedPaths, recursionStack map[string]bool, result *[]BknDocument) error {
	for _, includeRel := range doc.Frontmatter.Includes {
		includePath := filepath.Join(baseDir, includeRel)
		absPath, err := filepath.Abs(includePath)
		if err != nil {
			absPath = includePath
		}
		if loadedPaths[absPath] {
			continue // Deduplication: already loaded via another path
		}
		if recursionStack[absPath] {
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
		recursionStack[absPath] = true
		err = resolveIncludes(incDoc, filepath.Dir(absPath), loadedPaths, recursionStack, result)
		delete(recursionStack, absPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func validateNetworkReferences(network *BknNetwork) error {
	for _, obj := range network.AllObjects() {
		if obj.DataSource == nil {
			continue
		}
		if strings.ToLower(strings.TrimSpace(obj.DataSource.Type)) != "connection" {
			continue
		}
		connectionID := strings.TrimSpace(obj.DataSource.ID)
		if connectionID == "" || network.GetConnection(connectionID) == nil {
			return fmt.Errorf("object %q references missing connection %q", obj.ID, connectionID)
		}
	}
	return nil
}
