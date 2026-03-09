package bkn

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const checksumFilename = "CHECKSUM"

// GenerateChecksumFile validates BKN inputs, then generates CHECKSUM in
// the given business directory. Covers .bkn and SKILL.md. Returns the
// content written.
func GenerateChecksumFile(root string) (string, error) {
	abs, err := filepath.Abs(root)
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
	if err := validateChecksumInputs(abs); err != nil {
		return "", err
	}

	var entries []string
	err = filepath.Walk(abs, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || filepath.Base(path) == checksumFilename {
			return nil
		}
		rel, _ := filepath.Rel(abs, path)
		rel = filepath.ToSlash(rel)
		name := filepath.Base(path)
		ext := strings.ToLower(filepath.Ext(path))

		if name == "SKILL.md" {
			line := computeSkillChecksum(path, rel)
			if line != "" {
				entries = append(entries, line)
			}
		} else if ext == ".bkn" {
			lines := computeBknChecksum(path)
			entries = append(entries, lines...)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(entries)

	now := time.Now().UTC().Format("2006-01-02T15:04:05Z07:00")
	lines := []string{
		"# BKN Directory Checksum",
		"# generated: " + now,
	}
	lines = append(lines, entries...)
	content := strings.Join(lines, "\n") + "\n"

	outPath := filepath.Join(abs, checksumFilename)
	if err := os.WriteFile(outPath, []byte(content), 0644); err != nil {
		return "", err
	}
	return content, nil
}

func validateChecksumInputs(root string) error {
	var networkPaths []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		// Only validate .bkn files, not .md files (SKILL.md is not a BKN file)
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".bkn" {
			return nil
		}

		doc, loadErr := Load(path)
		if loadErr != nil {
			rel, _ := filepath.Rel(root, path)
			return fmt.Errorf("checksum validation failed for %s: %w", filepath.ToSlash(rel), loadErr)
		}
		if strings.EqualFold(strings.TrimSpace(doc.Frontmatter.Type), "network") {
			networkPaths = append(networkPaths, path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Use root discovery per directory to avoid validating both network.bkn and
	// index.bkn when both exist (network.bkn takes priority).
	dirsWithNetworks := make(map[string]bool)
	for _, p := range networkPaths {
		dirsWithNetworks[filepath.Dir(p)] = true
	}
	var rootsToValidate []string
	for d := range dirsWithNetworks {
		rootPath, discoverErr := DiscoverRootFile(d)
		if discoverErr != nil {
			return fmt.Errorf("checksum validation failed: %w", discoverErr)
		}
		rootsToValidate = append(rootsToValidate, rootPath)
	}
	// Deduplicate
	seen := make(map[string]bool)
	var unique []string
	for _, r := range rootsToValidate {
		if !seen[r] {
			seen[r] = true
			unique = append(unique, r)
		}
	}

	// Network data validation removed - .bknd format no longer supported
	_ = unique
	return nil
}

// VerifyChecksumFile verifies checksum.txt against actual files.
// Returns (ok, errorMessages).
func VerifyChecksumFile(root string) (bool, []string) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return false, []string{err.Error()}
	}
	ckPath := filepath.Join(abs, checksumFilename)
	data, err := os.ReadFile(ckPath)
	if err != nil {
		return false, []string{checksumFilename + " not found"}
	}

	declared := make(map[string]string)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Parse format: definition_type:id  sha256:hash
		parts := strings.SplitN(line, "  ", 2)
		if len(parts) == 2 {
			declared[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	var errors []string
	// Collect and verify each definition
	var actualEntries []string
	filepath.Walk(abs, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || filepath.Base(path) == checksumFilename {
			return nil
		}
		name := filepath.Base(path)
		ext := strings.ToLower(filepath.Ext(path))

		if name == "SKILL.md" {
			// SKILL.md is not a definition, skip in verification
			return nil
		} else if ext == ".bkn" {
			lines := computeBknChecksum(path)
			for _, line := range lines {
				actualEntries = append(actualEntries, line)
				parts := strings.SplitN(line, "  ", 2)
				if len(parts) == 2 {
					defKey := strings.TrimSpace(parts[0])
					actualHash := strings.TrimSpace(parts[1])
					if decl, ok := declared[defKey]; ok {
						if decl != actualHash {
							errors = append(errors, "Mismatch: "+defKey)
						}
						delete(declared, defKey)
					} else {
						errors = append(errors, "Unexpected definition: "+defKey)
					}
				}
			}
		}
		return nil
	})

	for defKey := range declared {
		if defKey != "*" {
			errors = append(errors, "Missing definition: "+defKey)
		}
	}

	return len(errors) == 0, errors
}

func computeSkillChecksum(path, rel string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	norm := normalizeForChecksum(string(data))
	h := sha256.Sum256([]byte(norm))
	return rel + "  sha256:" + hex.EncodeToString(h[:16])
}

// computeBknChecksum computes checksums for all definitions in a .bkn file.
// Returns multiple lines for files with multiple definitions.
// Format: definition_type:definition_id  sha256:hash
func computeBknChecksum(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	content := string(data)
	fm, err := ParseFrontmatter(content)
	if err != nil {
		return nil
	}

	var results []string
	typeVal := strings.TrimSpace(fm.Type)

	// For network type, use network:id format
	if typeVal == "network" {
		_, body := splitFrontmatter(content)
		norm := normalizeForChecksum(body)
		h := sha256.Sum256([]byte(norm))
		id := fm.ID
		if id == "" {
			id = "network"
		}
		results = append(results, "network:"+id+"  sha256:"+hex.EncodeToString(h[:16]))
		return results
	}

	// For definition types, parse body and compute per definition
	objects, relations, actions := ParseBody(content)

	for _, obj := range objects {
		_, body := splitFrontmatter(content)
		// Extract just this object's definition section
		objSection := extractDefinitionSection(body, "Object", obj.ID)
		norm := normalizeForChecksum(objSection)
		h := sha256.Sum256([]byte(norm))
		results = append(results, "object_type:"+obj.ID+"  sha256:"+hex.EncodeToString(h[:16]))
	}

	for _, rel := range relations {
		_, body := splitFrontmatter(content)
		relSection := extractDefinitionSection(body, "Relation", rel.ID)
		norm := normalizeForChecksum(relSection)
		h := sha256.Sum256([]byte(norm))
		results = append(results, "relation_type:"+rel.ID+"  sha256:"+hex.EncodeToString(h[:16]))
	}

	for _, act := range actions {
		_, body := splitFrontmatter(content)
		actSection := extractDefinitionSection(body, "Action", act.ID)
		norm := normalizeForChecksum(actSection)
		h := sha256.Sum256([]byte(norm))
		results = append(results, "action_type:"+act.ID+"  sha256:"+hex.EncodeToString(h[:16]))
	}

	return results
}

// extractDefinitionSection extracts a specific definition section from body
func extractDefinitionSection(body, defType, id string) string {
	lines := strings.Split(body, "\n")
	var result []string
	inSection := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for section start: ## Object: id or ## Relation: id, etc.
		if strings.HasPrefix(trimmed, "## ") {
			if inSection {
				// We were in a section, now found next section
				break
			}
			// Check if this is our target section
			prefix := "## " + defType + ": " + id
			if strings.HasPrefix(trimmed, prefix) {
				inSection = true
				result = append(result, line)
				continue
			}
		}

		if inSection {
			// Check for subsection (### or deeper)
			if strings.HasPrefix(trimmed, "###") {
				result = append(result, line)
			} else if strings.HasPrefix(trimmed, "## ") {
				// Next definition at same level
				break
			} else {
				result = append(result, line)
			}
		}
	}

	return strings.Join(result, "\n")
}

// normalizeForChecksum normalizes text before hashing so that blank lines,
// CRLF/LF differences, trailing whitespace, and table-cell padding do not
// affect the checksum. Semantic content changes still change the checksum.
func normalizeForChecksum(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	lines := strings.Split(s, "\n")
	var out []string
	for _, line := range lines {
		trimmed := strings.TrimRight(line, " \t")
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return strings.Join(out, "\n")
}
