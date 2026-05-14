package processor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/outis/ghost-rizz/internal/generator"
)

func TestGenerateRandomString(t *testing.T) {
	str := generateRandomString(10)
	if len(str) != 10 {
		t.Errorf("expected length 10, got %d", len(str))
	}
}

func TestProcessSingleImage(t *testing.T) {
	tmpDir := t.TempDir()

	if err := generator.GenerateImages(1, tmpDir); err != nil {
		t.Fatalf("failed to generate image: %v", err)
	}

	// Discover the generated file dynamically to avoid relying on name conventions.
	entries, err := os.ReadDir(tmpDir)
	if err != nil || len(entries) == 0 {
		t.Fatalf("no files generated in tmpDir")
	}
	inPath := filepath.Join(tmpDir, entries[0].Name())
	ext := filepath.Ext(entries[0].Name())

	outPathClean := filepath.Join(tmpDir, "out_clean"+ext)
	outPathFuzz := filepath.Join(tmpDir, "out_fuzz"+ext)

	if err := processSingleImage(inPath, outPathClean, "clean"); err != nil {
		t.Fatalf("processSingleImage clean failed: %v", err)
	}

	outPathClean2 := filepath.Join(tmpDir, "out_clean2"+ext)
	if err := processSingleImage(outPathClean, outPathClean2, "clean"); err != nil {
		t.Fatalf("processSingleImage clean on already clean image failed: %v", err)
	}

	if err := processSingleImage(inPath, outPathFuzz, "fuzz"); err != nil {
		t.Fatalf("processSingleImage fuzz failed: %v", err)
	}
}

func TestProcessSingleImage_ParseError(t *testing.T) {
	err := processSingleImage("/dev/null/does_not_exist.jpg", "out.jpg", "clean")
	if err == nil {
		t.Errorf("expected error when file does not exist")
	}
}

func TestProcessSingleImage_WriteError(t *testing.T) {
	tmpDir := t.TempDir()
	if err := generator.GenerateImages(1, tmpDir); err != nil {
		t.Fatalf("failed to generate image: %v", err)
	}
	entries, err := os.ReadDir(tmpDir)
	if err != nil || len(entries) == 0 {
		t.Fatalf("no files generated")
	}
	inPath := filepath.Join(tmpDir, entries[0].Name())

	if err := processSingleImage(inPath, "/dev/null/cannot_write"+filepath.Ext(entries[0].Name()), "clean"); err == nil {
		t.Errorf("expected error when output path is invalid")
	}
}

func TestCreateRandomEXIF(t *testing.T) {
	ib, err := createRandomEXIF()
	if err != nil {
		t.Fatalf("createRandomEXIF failed: %v", err)
	}
	if ib == nil {
		t.Fatal("expected non-nil IfdBuilder")
	}
}

func TestProcessSingleImage_UnknownMode(t *testing.T) {
	err := processSingleImage("/dev/null", "out.jpg", "invalid_mode")
	if err == nil {
		t.Errorf("expected error for unknown mode")
	}
}

func TestProcessSingleImage_OutFileError(t *testing.T) {
	tmpDir := t.TempDir()
	_ = generator.GenerateImages(1, tmpDir)
	entries, _ := os.ReadDir(tmpDir)
	inPath := filepath.Join(tmpDir, entries[0].Name())

	badOut := filepath.Join(tmpDir, "bad_out")
	_ = os.MkdirAll(badOut, 0755)

	err := processSingleImage(inPath, badOut, "clean")
	if err == nil {
		t.Errorf("expected error when output file is a directory")
	}
}
