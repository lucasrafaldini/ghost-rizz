package processor

import (
	"encoding/csv"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/outis/ghost-rizz/internal/generator"
)

func TestGenerateReport(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	err := generator.GenerateImages(1, inDir)
	if err != nil {
		t.Fatalf("failed to generate image: %v", err)
	}

	// Add a non-supported file to test the skip branch
	_ = os.WriteFile(filepath.Join(inDir, "skip.txt"), []byte("skip"), 0644)

	// Add a directory to test the skip branch
	_ = os.MkdirAll(filepath.Join(inDir, "subdir"), 0755)

	err = ProcessImages(inDir, outDir, "clean")
	if err != nil {
		t.Fatalf("ProcessImages clean failed: %v", err)
	}

	err = ProcessImages(inDir, outDir, "fuzz")
	if err != nil {
		t.Fatalf("ProcessImages fuzz failed: %v", err)
	}

	outCSV := filepath.Join(inDir, "report.csv")
	err = GenerateReport(outDir, outCSV)
	if err != nil {
		t.Fatalf("GenerateReport failed: %v", err)
	}

	f, err := os.Open(outCSV)
	if err != nil {
		t.Fatalf("failed to open generated csv: %v", err)
	}
	defer func() { _ = f.Close() }()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("failed to read csv records: %v", err)
	}

	if len(records) != 3 {
		t.Errorf("expected 3 rows in CSV, got %d", len(records))
	}
}

func TestGenerateReport_InvalidInDir(t *testing.T) {
	err := GenerateReport(filepath.Join(t.TempDir(), "nonexistent"), t.TempDir()+"/report.csv")
	if err == nil {
		t.Errorf("expected error for invalid input dir")
	}
}

func TestGenerateReport_InvalidOutCSV(t *testing.T) {
	err := GenerateReport(t.TempDir(), filepath.Join(t.TempDir(), "nonexistent", "report.csv"))
	if err == nil {
		t.Errorf("expected error for invalid output csv")
	}
}

func TestGenerateReport_NonJpeg(t *testing.T) {
	inDir := t.TempDir()
	badFile := filepath.Join(inDir, "bad.txt")
	_ = os.WriteFile(badFile, []byte("not a real jpeg"), 0644)

	outCSV := filepath.Join(inDir, "report.csv")
	err := GenerateReport(inDir, outCSV)
	if err != nil {
		t.Errorf("GenerateReport failed: %v", err)
	}

	f, err := os.Open(outCSV)
	if err == nil {
		defer func() { _ = f.Close() }()
		records, _ := csv.NewReader(f).ReadAll()
		if len(records) != 1 {
			t.Errorf("expected 1 row (header only), got %d", len(records))
		}
	} else {
		t.Errorf("failed to open csv: %v", err)
	}
}

func TestGenerateReport_ParseError(t *testing.T) {
	inDir := t.TempDir()
	badFile := filepath.Join(inDir, "bad.jpg")
	_ = os.WriteFile(badFile, []byte("not a real jpeg"), 0644)

	outCSV := filepath.Join(inDir, "report.csv")
	err := GenerateReport(inDir, outCSV)
	if err != nil {
		t.Errorf("GenerateReport failed: %v", err)
	}

	f, err := os.Open(outCSV)
	if err == nil {
		defer func() { _ = f.Close() }()
		records, _ := csv.NewReader(f).ReadAll()
		if len(records) != 2 {
			t.Errorf("expected 2 rows, got %d", len(records))
		} else if records[1][1] != "error" {
			t.Errorf("expected HasEXIF to be 'error', got %s", records[1][1])
		}
	} else {
		t.Errorf("failed to open csv: %v", err)
	}
}

func TestGenerateReport_HEIC(t *testing.T) {
	inDir := t.TempDir()
	outCSV := filepath.Join(inDir, "report.csv")

	// Create a dummy HEIC file
	heicFile := filepath.Join(inDir, "test.heic")
	_ = os.WriteFile(heicFile, []byte("fake heic"), 0644)

	_, lookErr := exec.LookPath("exiftool")
	err := GenerateReport(inDir, outCSV)

	if lookErr != nil {
		// Since we process files concurrently, GenerateReport doesn't return error
		// from CheckExifTool immediately but records it in the CSV row.
		// Let's verify the CSV contains 'error'.
		f, _ := os.Open(outCSV)
		if f != nil {
			defer f.Close()
			records, _ := csv.NewReader(f).ReadAll()
			if len(records) > 1 && records[1][1] != "error" {
				t.Errorf("expected HEIC row to have error when exiftool is missing")
			}
		}
	} else if err != nil {
		t.Errorf("GenerateReport failed with HEIC even though exiftool is present: %v", err)
	}
}

func TestGenerateReport_MkdirError(t *testing.T) {
	// GenerateReport doesn't call MkdirAll, it assumes outCSV's parent exists.
	// But we can test ProcessImages's MkdirAll error here too or in processor_test.
	err := ProcessImages(t.TempDir(), "/dev/null/invalid", "clean")
	if err == nil {
		t.Errorf("expected error for invalid output dir in ProcessImages")
	}
}

func TestGenerateReport_OutCSVIsDir(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()
	err := GenerateReport(inDir, outDir)
	if err == nil {
		t.Errorf("expected error when outCSV is a directory")
	}
}

func TestGenerateReport_MkdirAllError(t *testing.T) {
	inDir := t.TempDir()

	parentFile := filepath.Join(t.TempDir(), "file")
	_ = os.WriteFile(parentFile, []byte("x"), 0644)
	badOutDir := filepath.Join(parentFile, "subdir")

	err := GenerateReport(inDir, badOutDir)
	if err == nil {
		t.Errorf("expected error when outDir creation fails due to parent being a file")
	}
}

func TestGenerateReport_EmptyDir(t *testing.T) {
	inDir := t.TempDir()
	outCSV := filepath.Join(t.TempDir(), "empty.csv")
	err := GenerateReport(inDir, outCSV)
	if err != nil {
		t.Errorf("expected no error for empty dir in GenerateReport, got %v", err)
	}
}

func TestGenerateReport_NoExt(t *testing.T) {
	inDir := t.TempDir()
	outCSV := filepath.Join(t.TempDir(), "noext.csv")
	_ = os.WriteFile(filepath.Join(inDir, "noext"), []byte("x"), 0644)
	err := GenerateReport(inDir, outCSV)
	if err != nil {
		t.Errorf("expected no error for file with no extension, got %v", err)
	}
}

func TestGenerateReport_MkdirAllError_Real(t *testing.T) {
	inDir := t.TempDir()
	parent := filepath.Join(t.TempDir(), "file")
	_ = os.WriteFile(parent, []byte("x"), 0644)
	outCSV := filepath.Join(parent, "report.csv")
	err := GenerateReport(inDir, outCSV)
	if err == nil {
		t.Errorf("expected error when parent dir is a file")
	}
}

func TestGenerateReport_AllSupported(t *testing.T) {
	inDir := t.TempDir()
	outCSV := filepath.Join(t.TempDir(), "all.csv")
	_ = os.WriteFile(filepath.Join(inDir, "f1.jpg"), []byte("x"), 0644)
	_ = os.WriteFile(filepath.Join(inDir, "f2.png"), []byte("x"), 0644)
	_ = GenerateReport(inDir, outCSV)
}
