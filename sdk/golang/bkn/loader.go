// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

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
	".bkn": true, ".md": true,
}

// Root file discovery: network.bkn
const rootCandidateName = "network.bkn"

// LoadNetwork loads a network file and recursively resolves its includes.
// Supported extensions: .bkn, .md. Root file should be type: network.
// When rootPath is a directory, the root file is discovered automatically (network.bkn).
// If the root has no includes, same-directory BKN files are loaded implicitly.
// Otherwise only files listed in frontmatter includes are loaded.
func LoadNetwork(rootPath string) (*BknNetwork, error) {
	fsys := NewOSFileSystem()
	return LoadNetworkWithFS(fsys, rootPath)
}

// LoadNetworkWithFS 使用指定的文件系统加载网络
func LoadNetworkWithFS(fsys FileSystem, rootPath string) (*BknNetwork, error) {
	absRoot := fsys.Abs(rootPath)

	// 检查是否是目录
	if fsys.IsDir(absRoot) {
		var err error
		absRoot, err = DiscoverRootFileWithFS(fsys, absRoot)
		if err != nil {
			return nil, err
		}
	}

	rootDoc, err := LoadWithFS(fsys, absRoot)
	if err != nil {
		return nil, err
	}

	loadedPaths := map[string]bool{absRoot: true}
	recursionStack := map[string]bool{}
	var includes []BknDocument
	baseDir := fsys.Dir(absRoot)

	if len(rootDoc.Frontmatter.Includes) > 0 {
		if err := resolveIncludesWithFS(fsys, rootDoc, baseDir, loadedPaths, recursionStack, &includes); err != nil {
			return nil, err
		}
	} else {
		// No includes: for type: network only, implicitly load same-dir files
		docType := strings.ToLower(strings.TrimSpace(rootDoc.Frontmatter.Type))
		if docType == "network" {
			implicitPaths, err := collectSameDirBknFilesWithFS(fsys, baseDir, absRoot)
			if err != nil {
				return nil, err
			}
			for _, incPath := range implicitPaths {
				incAbs := fsys.Abs(incPath)
				if loadedPaths[incAbs] {
					continue
				}
				if recursionStack[incAbs] {
					return nil, fmt.Errorf("circular include detected: %s (resolved to %s)", fsys.Base(incPath), incAbs)
				}
				loadedPaths[incAbs] = true
				incDoc, err := LoadWithFS(fsys, incAbs)
				if err != nil {
					return nil, err
				}
				includes = append(includes, *incDoc)
				recursionStack[incAbs] = true
				if err := resolveIncludesWithFS(fsys, incDoc, fsys.Dir(incAbs), loadedPaths, recursionStack, &includes); err != nil {
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

// DiscoverRootFile discovers the root network file in a directory.
// Order: network.bkn > network.md > index.bkn > index.md.
// If none exist, and exactly one file in the directory has type: network,
// use that file. Otherwise return error.
func DiscoverRootFile(directory string) (string, error) {
	fsys := NewOSFileSystem()
	return DiscoverRootFileWithFS(fsys, directory)
}

// DiscoverRootFileWithFS 使用指定的文件系统发现根文件
func DiscoverRootFileWithFS(fsys FileSystem, directory string) (string, error) {
	abs := fsys.Abs(directory)

	// 1. Check named candidate
	candidate := fsys.Join(abs, rootCandidateName)
	if _, err := fsys.Stat(candidate); err == nil {
		ext := fsys.Ext(candidate)
		if supportedExtensions[ext] {
			return candidate, nil
		}
	}

	// 2. Scan same directory for type: network files
	var networkFiles []string
	entries, err := fsys.ReadDir(abs)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := fsys.Ext(e.Name())
		if !supportedExtensions[ext] {
			continue
		}
		p := fsys.Join(abs, e.Name())
		data, err := fsys.ReadFile(p)
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
			names[i] = fsys.Base(p)
		}
		return "", fmt.Errorf("multiple network roots in %s: %v; use %s as the single root", abs, names, rootCandidateName)
	}
	return "", fmt.Errorf("no root network file found in %s; expected %s or a single type: network file", abs, rootCandidateName)
}

func collectSameDirBknFiles(directory, rootPath string) ([]string, error) {
	fsys := NewOSFileSystem()
	return collectSameDirBknFilesWithFS(fsys, directory, rootPath)
}

// bknSubdirs lists the standard subdirectories for BKN definitions per DESIGN.md §3.1.
var bknSubdirs = []string{"object_types", "relation_types", "action_types", "risk_types"}

func collectSameDirBknFilesWithFS(fsys FileSystem, directory, rootPath string) ([]string, error) {
	abs := fsys.Abs(directory)
	absRoot := fsys.Abs(rootPath)
	rootName := fsys.Base(absRoot)

	excludeNames := map[string]bool{rootName: true}
	candidate := fsys.Join(abs, rootCandidateName)
	if _, err := fsys.Stat(candidate); err == nil {
		excludeNames[rootCandidateName] = true
	}

	var result []string

	// Scan same-directory files
	entries, err := fsys.ReadDir(abs)
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
		ext := fsys.Ext(e.Name())
		if !supportedExtensions[ext] {
			continue
		}
		p := fsys.Join(abs, e.Name())
		data, err := fsys.ReadFile(p)
		if err != nil {
			continue
		}
		if _, err := Parse(string(data), p); err != nil {
			continue
		}
		result = append(result, p)
	}

	// Scan standard subdirectories (object_types/, relation_types/, action_types/, risk_types/)
	for _, subdir := range bknSubdirs {
		subdirPath := fsys.Join(abs, subdir)
		if !fsys.IsDir(subdirPath) {
			continue
		}
		subEntries, err := fsys.ReadDir(subdirPath)
		if err != nil {
			continue
		}
		for _, e := range subEntries {
			if e.IsDir() {
				continue
			}
			ext := fsys.Ext(e.Name())
			if !supportedExtensions[ext] {
				continue
			}
			p := fsys.Join(subdirPath, e.Name())
			data, err := fsys.ReadFile(p)
			if err != nil {
				continue
			}
			if _, err := Parse(string(data), p); err != nil {
				continue
			}
			result = append(result, p)
		}
	}

	sort.Strings(result)
	return result, nil
}

func checkExtension(path string) error {
	ext := strings.ToLower(filepath.Ext(path))
	if !supportedExtensions[ext] {
		return fmt.Errorf("unsupported file extension: %q; BKN supports: .bkn, .md", ext)
	}
	return nil
}

// Load loads and parses a single .bkn/.md file.
// Supported extensions: .bkn, .md. Content must satisfy BKN
// frontmatter, type, and structure requirements regardless of extension.
func Load(path string) (*BknDocument, error) {
	fsys := NewOSFileSystem()
	return LoadWithFS(fsys, path)
}

// LoadWithFS 使用指定的文件系统加载文件
func LoadWithFS(fsys FileSystem, path string) (*BknDocument, error) {
	if err := checkExtension(path); err != nil {
		return nil, err
	}
	data, err := fsys.ReadFile(path)
	if err != nil {
		return nil, err
	}
	abs := fsys.Abs(path)
	return Parse(string(data), abs)
}

// resolveIncludes recursively resolves includes from a document's frontmatter.
// Deduplication: paths in loadedPaths are skipped (already loaded).
// Circular: only when path is in recursionStack (back to self in chain).
func resolveIncludes(doc *BknDocument, baseDir string, loadedPaths, recursionStack map[string]bool, result *[]BknDocument) error {
	fsys := NewOSFileSystem()
	return resolveIncludesWithFS(fsys, doc, baseDir, loadedPaths, recursionStack, result)
}

func resolveIncludesWithFS(fsys FileSystem, doc *BknDocument, baseDir string, loadedPaths, recursionStack map[string]bool, result *[]BknDocument) error {
	// Only network type documents can have includes
	docType := strings.ToLower(strings.TrimSpace(doc.Frontmatter.Type))
	if docType != "network" {
		return nil
	}

	for _, includeRel := range doc.Frontmatter.Includes {
		includePath := fsys.Join(baseDir, includeRel)
		absPath := fsys.Abs(includePath)

		// Check for circular include first (before deduplication)
		if recursionStack[absPath] {
			return fmt.Errorf("circular include detected: %s (resolved to %s)", includeRel, absPath)
		}

		// Deduplication: skip if already loaded
		if loadedPaths[absPath] {
			continue
		}

		if _, err := fsys.Stat(absPath); err != nil {
			return fmt.Errorf("include file not found: %s (resolved to %s)", includeRel, absPath)
		}

		loadedPaths[absPath] = true
		incDoc, err := LoadWithFS(fsys, absPath)
		if err != nil {
			return err
		}
		*result = append(*result, *incDoc)

		// Add to recursion stack before recursing
		recursionStack[absPath] = true
		err = resolveIncludesWithFS(fsys, incDoc, fsys.Dir(absPath), loadedPaths, recursionStack, result)
		delete(recursionStack, absPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func validateNetworkReferences(network *BknNetwork) error {
	// Connection validation removed as Connection type is no longer supported
	return nil
}

// Deprecated: Use FileSystem.ReadFile instead.
func LoadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// Deprecated: Use FileSystem.Stat instead.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Deprecated: Use FileSystem.IsDir instead.
func IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
