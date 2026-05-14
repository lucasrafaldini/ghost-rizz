package generator

import (
	"os"
	"testing"
)

func TestGenerateImages(t *testing.T) {
	tmpDir := t.TempDir()

	err := GenerateImages(2, tmpDir)
	if err != nil {
		t.Fatalf("GenerateImages failed: %v", err)
	}

	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to read tmpdir: %v", err)
	}

	if len(files) != 2 {
		t.Errorf("expected 2 files, got %d", len(files))
	}

	for _, f := range files {
		info, _ := f.Info()
		if info.Size() == 0 {
			t.Errorf("expected non-empty file %s", f.Name())
		}
	}
}

func TestGenerateImages_DirCreationError(t *testing.T) {
	err := GenerateImages(1, "/dev/null/invalid_dir")
	if err == nil {
		t.Errorf("expected error when output dir is invalid")
	}
}

func TestCreateDummyJPEGWithEXIF_FileError(t *testing.T) {
	// Trying to write a file to an unwritable path
	err := createDummyJPEGWithEXIF("/dev/null/invalid.jpg")
	if err == nil {
		t.Errorf("expected error when file path is invalid")
	}
}

func TestBuildBaseEXIF(t *testing.T) {
	ib, err := buildBaseEXIF()
	if err != nil {
		t.Fatalf("buildBaseEXIF failed: %v", err)
	}
	if ib == nil {
		t.Fatal("expected non-nil IfdBuilder")
	}
}
