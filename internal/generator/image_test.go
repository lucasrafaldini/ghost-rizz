package generator

import (
	"os"
	"path/filepath"
	"runtime"
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
	badPath := filepath.Join(t.TempDir(), "file.txt")
	_ = os.WriteFile(badPath, []byte("x"), 0644)
	err := GenerateImages(1, filepath.Join(badPath, "invalid_dir"))
	if err == nil {
		t.Errorf("expected error when output dir is invalid")
	}
}

func TestGenerateImages_ZeroCount(t *testing.T) {
	err := GenerateImages(0, t.TempDir())
	if err != nil {
		t.Errorf("expected no error for zero count, got %v", err)
	}
}

func TestCreateDummyJPEGWithEXIF_FileError(t *testing.T) {
	// Trying to write a file to an unwritable path
	err := createDummyJPEGWithEXIF(filepath.Join(t.TempDir(), "nonexistent_dir", "invalid.jpg"))
	if err == nil {
		t.Errorf("expected error when file path is invalid")
	}
}

func TestCreateDummyPNGWithEXIF_FileError(t *testing.T) {
	// Trying to write a PNG to an unwritable path
	err := createDummyPNGWithEXIF(filepath.Join(t.TempDir(), "nonexistent_dir", "invalid.png"))
	if err == nil {
		t.Errorf("expected error when file path is invalid")
	}
}

func TestCreateDummyJPEGWithEXIF_OutFileError(t *testing.T) {
	tmpDir := t.TempDir()
	badFile := filepath.Join(tmpDir, "bad.jpg")
	_ = os.MkdirAll(badFile, 0755)
	err := createDummyJPEGWithEXIF(badFile)
	if err == nil {
		t.Errorf("expected error when output file is a directory")
	}
}

func TestCreateDummyPNGWithEXIF_OutFileError(t *testing.T) {
	tmpDir := t.TempDir()
	badFile := filepath.Join(tmpDir, "bad.png")
	_ = os.MkdirAll(badFile, 0755)
	err := createDummyPNGWithEXIF(badFile)
	if err == nil {
		t.Errorf("expected error when output file is a directory")
	}
}





func TestCreateDummyPNGWithEXIF_WriteError(t *testing.T) {
	badPath := "/proc/invalid/path/png"
	if runtime.GOOS == "darwin" {
		badPath = "/System/nonexistent.png"
	}

	err := createDummyPNGWithEXIF(badPath)
	if err == nil {
		t.Errorf("expected error for unwritable path")
	}
}
