package processor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/outis/ghost-rizz/internal/generator"
)

func TestProcessImages(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	err := generator.GenerateImages(2, inDir)
	if err != nil {
		t.Fatalf("failed to generate images: %v", err)
	}

	err = ProcessImages(inDir, outDir, "clean")
	if err != nil {
		t.Fatalf("ProcessImages (clean) failed: %v", err)
	}

	err = ProcessImages(inDir, outDir, "fuzz")
	if err != nil {
		t.Fatalf("ProcessImages (fuzz) failed: %v", err)
	}

	files, err := os.ReadDir(outDir)
	if err != nil {
		t.Fatalf("failed to read outDir: %v", err)
	}

	if len(files) != 4 {
		t.Errorf("expected 4 files in output, got %d", len(files))
	}
}

func TestProcessImages_InvalidInDir(t *testing.T) {
	err := ProcessImages(filepath.Join(t.TempDir(), "nonexistent"), t.TempDir(), "clean")
	if err == nil {
		t.Errorf("expected error for invalid input dir")
	}
}

func TestProcessImages_InvalidOutDir(t *testing.T) {
	badPath := filepath.Join(t.TempDir(), "file.txt")
	_ = os.WriteFile(badPath, []byte("x"), 0644)
	err := ProcessImages(t.TempDir(), filepath.Join(badPath, "invalid_dir"), "clean")
	if err == nil {
		t.Errorf("expected error for invalid output dir")
	}
}

func TestProcessImages_ProcessSingleImageError(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	badFile := filepath.Join(inDir, "bad.jpg")
	_ = os.WriteFile(badFile, []byte("not a real jpeg"), 0644)

	err := ProcessImages(inDir, outDir, "clean")
	if err == nil {
		t.Errorf("ProcessImages expected error but got nil")
	}
}
