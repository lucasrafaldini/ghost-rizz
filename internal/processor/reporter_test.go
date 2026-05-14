package processor

import (
	"encoding/csv"
	"os"
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
	defer f.Close()

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
	err := GenerateReport("/dev/null/invalid", t.TempDir()+"/report.csv")
	if err == nil {
		t.Errorf("expected error for invalid input dir")
	}
}

func TestGenerateReport_InvalidOutCSV(t *testing.T) {
	err := GenerateReport(t.TempDir(), "/dev/null/invalid/report.csv")
	if err == nil {
		t.Errorf("expected error for invalid output csv")
	}
}

func TestGenerateReport_NonJpeg(t *testing.T) {
	inDir := t.TempDir()
	badFile := filepath.Join(inDir, "bad.txt")
	os.WriteFile(badFile, []byte("not a real jpeg"), 0644)

	err := GenerateReport(inDir, filepath.Join(inDir, "report.csv"))
	if err != nil {
		t.Errorf("GenerateReport failed: %v", err)
	}
}

func TestGenerateReport_ParseError(t *testing.T) {
	inDir := t.TempDir()
	badFile := filepath.Join(inDir, "bad.jpg")
	os.WriteFile(badFile, []byte("not a real jpeg"), 0644)

	err := GenerateReport(inDir, filepath.Join(inDir, "report.csv"))
	if err != nil {
		t.Errorf("GenerateReport failed: %v", err)
	}
}
