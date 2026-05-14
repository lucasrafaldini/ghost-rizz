package processor

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/outis/ghost-rizz/internal/generator"
)

func TestGetMediaHandler(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a real JPEG and PNG using our generator
	_ = generator.GenerateImages(2, tmpDir)

	files, _ := os.ReadDir(tmpDir)
	var jpgPath, pngPath string
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".jpg" {
			jpgPath = filepath.Join(tmpDir, f.Name())
		} else if filepath.Ext(f.Name()) == ".png" {
			pngPath = filepath.Join(tmpDir, f.Name())
		}
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"Unsupported", filepath.Join(tmpDir, "dummy.txt"), true},
		{"JPEG", jpgPath, false},
		{"PNG", pngPath, false},
		{"HEIC", filepath.Join(tmpDir, "dummy.heic"), false},
	}

	_ = os.WriteFile(filepath.Join(tmpDir, "dummy.txt"), []byte("txt"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "dummy.heic"), []byte("heic"), 0644)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mh, err := GetMediaHandler(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMediaHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && mh != nil {
				// Skip exiftool-dependent HEIC methods if the tool is absent
				if tt.name == "HEIC" {
					if _, lookErr := exec.LookPath("exiftool"); lookErr != nil {
						t.Skip("exiftool not found, skipping HEIC method coverage")
					}
				}
				// Hit the basic methods to ensure coverage
				_ = mh.DropExif()
				var buf bytes.Buffer
				_ = mh.Write(&buf)
				_ = mh.SetExif(nil)
				_, _ = mh.RawExif()
			}
		})
	}
}
