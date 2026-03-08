package bkn

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateChecksumFile_ValidationFails(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "objects"), 0755); err != nil {
		t.Fatalf("mkdir objects: %v", err)
	}
	if err := os.Mkdir(filepath.Join(root, "connections"), 0755); err != nil {
		t.Fatalf("mkdir connections: %v", err)
	}
	if err := os.Mkdir(filepath.Join(root, "data"), 0755); err != nil {
		t.Fatalf("mkdir data: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "index.bkn"), []byte(`---
type: network
id: demo
name: Demo
includes:
  - connections/erp.bkn
  - objects/pod.bkn
  - data/pod.bknd
---

# Demo
`), 0644); err != nil {
		t.Fatalf("write index: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "objects", "pod.bkn"), []byte(`---
type: object
id: pod
network: demo
---

## Object: pod

### Data Source

| Type | ID | Name |
|------|----|------|
| connection | erp | ERP |

### Data Properties

| Property | Primary Key |
|----------|-------------|
| id | YES |
`), 0644); err != nil {
		t.Fatalf("write object: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "connections", "erp.bkn"), []byte(`---
type: connection
id: erp
network: demo
---

## Connection: erp

**ERP**
`), 0644); err != nil {
		t.Fatalf("write connection: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "data", "pod.bknd"), []byte(`---
type: data
object: pod
---

## Data

| id |
|----|
| pod-1 |
`), 0644); err != nil {
		t.Fatalf("write data: %v", err)
	}

	network, err := LoadNetwork(filepath.Join(root, "index.bkn"))
	if err != nil {
		t.Fatalf("load network: %v", err)
	}
	validation := ValidateNetworkData(network)
	if validation.OK() {
		t.Fatal("expected network validation to fail before checksum generation")
	}

	_, err = GenerateChecksumFile(root)
	if err == nil {
		t.Fatal("expected checksum generation to fail")
	}
	if !strings.Contains(err.Error(), "checksum validation failed") {
		t.Fatalf("expected validation failure, got %q", err.Error())
	}
	if _, statErr := os.Stat(filepath.Join(root, checksumFilename)); !os.IsNotExist(statErr) {
		t.Fatalf("expected checksum.txt not to be written, got %v", statErr)
	}
}
