package rules

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanFilesEmpty(t *testing.T) {
	dir := t.TempDir()
	files, err := ScanFiles(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected no files, got %d", len(files))
	}
}

func TestScanFilesRootLevel(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "shared.md"), []byte("# shared"), 0644)

	files, err := ScanFiles(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if files[0].Tools != nil {
		t.Errorf("expected no pre-assigned tools for root file")
	}
}

func TestScanFilesClaudeSubdir(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".claude"), 0755)
	os.WriteFile(filepath.Join(dir, ".claude", "auth.md"), []byte("# auth"), 0644)

	files, err := ScanFiles(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if len(files[0].Tools) != 1 || files[0].Tools[0] != "claude" {
		t.Errorf("expected claude tool assignment, got %v", files[0].Tools)
	}
}

func TestScanFilesEmptyFileError(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "empty.md"), []byte{}, 0644)

	_, err := ScanFiles(dir)
	if err == nil {
		t.Error("expected error for empty file")
	}
}

func TestScanFilesMissingDir(t *testing.T) {
	_, err := ScanFiles("/nonexistent/path")
	if err == nil {
		t.Error("expected error for missing directory")
	}
}
