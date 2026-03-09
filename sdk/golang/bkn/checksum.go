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

const checksumFilename = "checksum.txt"

// GenerateChecksumFile validates BKN inputs, then generates checksum.txt in
// the given business directory. Covers .bkn, .bknd, and SKILL.md. Returns the
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

		var line string
		if name == "SKILL.md" {
			line = computeSkillChecksum(path, rel)
		} else if ext == ".bkn" {
			line = computeBknChecksum(path, rel)
		} else if ext == ".bknd" {
			line = computeBkndChecksum(path, rel)
		}
		if line != "" {
			entries = append(entries, line)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(entries)

	concat := strings.Join(entries, "\n") + "\n"
	if len(entries) == 0 {
		concat = ""
	}
	aggHash := sha256.Sum256([]byte(concat))
	aggLine := "sha256:" + hex.EncodeToString(aggHash[:]) + "  *"

	now := time.Now().UTC().Format("2006-01-02T15:04:05Z07:00")
	lines := []string{
		"# BKN Directory Checksum",
		"# generated: " + now,
		aggLine,
		"",
	}
	lines = append(lines, entries...)
	content := strings.Join(lines, "\n")

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
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".bkn" && ext != ".bknd" && ext != ".md" {
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

	for _, rootPath := range unique {
		rel, _ := filepath.Rel(root, rootPath)
		rel = filepath.ToSlash(rel)
		network, loadErr := LoadNetwork(rootPath)
		if loadErr != nil {
			return fmt.Errorf("checksum validation failed for network %s: %w", rel, loadErr)
		}
		result := ValidateNetworkData(network)
		if result.OK() {
			continue
		}
		return fmt.Errorf(
			"checksum validation failed for network %s: %s",
			rel,
			result.Errors[0].String(),
		)
	}
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
		parts := strings.SplitN(line, "  ", 2)
		if len(parts) == 2 {
			declared[strings.TrimSpace(parts[1])] = strings.TrimSpace(parts[0])
		}
	}

	var errors []string
	// Collect and verify each file
	var actualEntries []string
	filepath.Walk(abs, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || filepath.Base(path) == checksumFilename {
			return nil
		}
		rel, _ := filepath.Rel(abs, path)
		rel = filepath.ToSlash(rel)
		name := filepath.Base(path)
		ext := strings.ToLower(filepath.Ext(path))

		var line string
		if name == "SKILL.md" {
			line = computeSkillChecksum(path, rel)
		} else if ext == ".bkn" {
			line = computeBknChecksum(path, rel)
		} else if ext == ".bknd" {
			line = computeBkndChecksum(path, rel)
		}
		if line != "" {
			actualEntries = append(actualEntries, line)
			actualHash := strings.Split(line, "  ")[0]
			if decl, ok := declared[rel]; ok {
				if decl != actualHash {
					errors = append(errors, "Mismatch: "+rel)
				}
				delete(declared, rel)
			} else {
				errors = append(errors, "Unexpected file: "+rel)
			}
		}
		return nil
	})

	for rel := range declared {
		if rel != "*" {
			errors = append(errors, "Missing file: "+rel)
		}
	}

	// Verify aggregate
	sort.Strings(actualEntries)
	concat := strings.Join(actualEntries, "\n") + "\n"
	if len(actualEntries) == 0 {
		concat = ""
	}
	aggHash := sha256.Sum256([]byte(concat))
	expectedAgg := "sha256:" + hex.EncodeToString(aggHash[:])
	if decl, ok := declared["*"]; ok && decl != expectedAgg {
		errors = append(errors, "Aggregate checksum mismatch")
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
	return "sha256:" + hex.EncodeToString(h[:]) + "  " + rel
}

func computeBknChecksum(path, rel string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	_, body := splitFrontmatter(string(data))
	norm := normalizeForChecksum(body)
	h := sha256.Sum256([]byte(norm))
	return "sha256:" + hex.EncodeToString(h[:]) + "  " + rel
}

func computeBkndChecksum(path, rel string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	_, body := splitFrontmatter(string(data))
	// Parse table, sort rows, re-serialize for order-insensitive hash
	canonical := canonicalizeBkndTable(strings.TrimSpace(body))
	h := sha256.Sum256([]byte(normalizeForChecksum(canonical)))
	return "sha256:" + hex.EncodeToString(h[:]) + "  " + rel
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

func canonicalizeBkndTable(body string) string {
	lines := strings.Split(body, "\n")
	var tableLines []string
	for _, line := range lines {
		s := strings.TrimSpace(line)
		if strings.HasPrefix(s, "|") {
			tableLines = append(tableLines, s)
		} else if len(tableLines) > 0 {
			break
		}
	}
	if len(tableLines) < 2 {
		return body
	}
	headers := splitTableRow(tableLines[0])
	sortedHeaders := make([]string, len(headers))
	copy(sortedHeaders, headers)
	sort.Strings(sortedHeaders)

	sep := strings.TrimSpace(tableLines[1])
	dataStart := 2
	if len(sep) < 2 || !strings.Contains(sep, "-") {
		dataStart = 1
	}
	var rows [][]string
	for _, line := range tableLines[dataStart:] {
		cells := splitTableRow(line)
		cellMap := make(map[string]string)
		for i, h := range headers {
			if i < len(cells) {
				cellMap[h] = cells[i]
			} else {
				cellMap[h] = ""
			}
		}
		row := make([]string, len(sortedHeaders))
		for i, h := range sortedHeaders {
			row[i] = cellMap[h]
		}
		rows = append(rows, row)
	}
	// Sort rows by all columns (order-insensitive)
	sort.Slice(rows, func(i, j int) bool {
		for k := 0; k < len(sortedHeaders); k++ {
			if rows[i][k] != rows[j][k] {
				return rows[i][k] < rows[j][k]
			}
		}
		return false
	})
	var out []string
	// Use sorted header order and canonical separator for output
	headerLine := "| " + strings.Join(sortedHeaders, " | ") + " |"
	sepLine := strings.Repeat("|---", len(sortedHeaders)) + "|"
	out = append(out, headerLine, sepLine)
	for _, row := range rows {
		out = append(out, "| "+strings.Join(row, " | ")+" |")
	}
	return strings.Join(out, "\n")
}

func splitTableRow(row string) []string {
	row = strings.TrimSpace(row)
	row = strings.TrimPrefix(row, "|")
	row = strings.TrimSuffix(row, "|")
	parts := strings.Split(row, "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}
