package processor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/outis/ghost-rizz/internal/generator"
)

func TestIsSupportedFormat(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{"jpg", "test.jpg", true},
		{"jpeg", "test.jpeg", true},
		{"png", "test.png", true},
		{"heic", "test.heic", true},
		{"heif", "test.heif", true},
		{"txt", "test.txt", false},
		{"noext", "test", false},
		{"upper", "TEST.JPG", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSupportedFormat(tt.filename); got != tt.want {
				t.Errorf("isSupportedFormat(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestProcessImages(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	err := generator.GenerateImages(2, inDir)
	if err != nil {
		t.Fatalf("failed to generate images: %v", err)
	}

	// Add a non-supported file to test the skip branch
	_ = os.WriteFile(filepath.Join(inDir, "skip.txt"), []byte("skip"), 0644)

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

func TestProcessImages_HEIC(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	// Create a dummy HEIC file
	heicFile := filepath.Join(inDir, "test.heic")
	_ = os.WriteFile(heicFile, []byte("fake heic"), 0644)

	// ProcessImages should call CheckExifTool.
	// If exiftool is absent, it returns an error. If present, it continues.
	// Either way, it hits the branches.
	_ = ProcessImages(inDir, outDir, "clean")
}

func TestProcessImages_UnreadableFile(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	file := filepath.Join(inDir, "test.jpg")
	_ = os.WriteFile(file, []byte("fake jpg"), 0000)
	defer func() { _ = os.Chmod(file, 0644) }()

	err := ProcessImages(inDir, outDir, "clean")
	if err == nil {
		t.Errorf("expected error for unreadable file")
	}
}

func TestProcessImages_InvalidMode(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	err := generator.GenerateImages(1, inDir)
	if err != nil {
		t.Fatalf("failed to generate image: %v", err)
	}

	err = ProcessImages(inDir, outDir, "invalid_mode")
	if err == nil {
		t.Errorf("expected error for invalid mode in ProcessImages")
	}
}
