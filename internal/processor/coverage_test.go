package processor

import (
	"os"
	"path/filepath"
	"testing"
)

// mockExifTool creates a dummy exiftool script that either succeeds or fails.
func mockExifTool(t *testing.T, shouldFail bool) string {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "exiftool")
	content := "#!/bin/sh\n"
	if shouldFail {
		content += "echo 'Error happened' >&2\nexit 1\n"
	} else {
		content += "echo 'Fake Exiftool Output'\nexit 0\n"
	}
	_ = os.WriteFile(scriptPath, []byte(content), 0755)
	return tmpDir
}

func TestHeicHandler_Coverage_Full(t *testing.T) {
	// Mock exiftool success
	mockPath := mockExifTool(t, false)
	oldPath := os.Getenv("PATH")
	_ = os.Setenv("PATH", mockPath+string(os.PathListSeparator)+oldPath)
	defer func() { _ = os.Setenv("PATH", oldPath) }()

	h := &heicHandler{filepath: "test.heic"}

	// Test RawExif success path
	_, _ = h.RawExif()

	// Test Write success path
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "out.heic")
	outF, _ := os.Create(outPath)
	_ = h.Write(outF)
	_ = outF.Close()

	// Test SetExif nil
	_ = h.SetExif(nil)

	// Mock exiftool failure
	mockPathFail := mockExifTool(t, true)
	_ = os.Setenv("PATH", mockPathFail+string(os.PathListSeparator)+oldPath)
	_, _ = h.RawExif()

	f2, _ := os.Create(filepath.Join(tmpDir, "out2.heic"))
	_ = h.Write(f2)
	_ = f2.Close()
}

func TestCheckExifTool_Coverage_Full(t *testing.T) {
	// Success path
	mockPath := mockExifTool(t, false)
	oldPath := os.Getenv("PATH")
	_ = os.Setenv("PATH", mockPath+string(os.PathListSeparator)+oldPath)
	defer func() { _ = os.Setenv("PATH", oldPath) }()
	_ = CheckExifTool()

	// Fail path
	_ = os.Setenv("PATH", "/non/existent")
	_ = CheckExifTool()
}

func TestProcessSingleImage_Additional(t *testing.T) {
	tmpDir := t.TempDir()
	badPng := filepath.Join(tmpDir, "bad.png")
	_ = os.WriteFile(badPng, []byte("not a png"), 0644)

	// Error in GetMediaHandler
	_ = processSingleImage(badPng, filepath.Join(tmpDir, "out.png"), "clean")

	// Unknown mode
	_ = processSingleImage(badPng, filepath.Join(tmpDir, "out.png"), "unknown")
}
