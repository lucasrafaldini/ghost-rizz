package processor

import (
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
	
	err := generator.GenerateImages(1, tmpDir)
	if err != nil {
		t.Fatalf("failed to generate image: %v", err)
	}

	inPath := filepath.Join(tmpDir, "dummy_0000.jpg")
	outPathClean := filepath.Join(tmpDir, "dummy_0000_clean.jpg")
	outPathFuzz := filepath.Join(tmpDir, "dummy_0000_fuzz.jpg")

	err = processSingleImage(inPath, outPathClean, "clean")
	if err != nil {
		t.Fatalf("processSingleImage clean failed: %v", err)
	}

	outPathClean2 := filepath.Join(tmpDir, "dummy_0000_clean2.jpg")
	err = processSingleImage(outPathClean, outPathClean2, "clean")
	if err != nil {
		t.Fatalf("processSingleImage clean on already clean image failed: %v", err)
	}

	err = processSingleImage(inPath, outPathFuzz, "fuzz")
	if err != nil {
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
	generator.GenerateImages(1, tmpDir)
	inPath := filepath.Join(tmpDir, "dummy_0000.jpg")
	
	err := processSingleImage(inPath, "/dev/null/cannot_write.jpg", "clean")
	if err == nil {
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
