package processor

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

func isSupportedFormat(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".heic" || ext == ".heif"
}

// ProcessImages reads supported image files from inDir, processes them using
// the specified mode, and writes the results to outDir.
func ProcessImages(inDir, outDir, mode string) error {
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}

	files, err := os.ReadDir(inDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() && isSupportedFormat(file.Name()) {
			if strings.HasSuffix(strings.ToLower(file.Name()), ".heic") || strings.HasSuffix(strings.ToLower(file.Name()), ".heif") {
				if _, err := exec.LookPath("exiftool"); err != nil {
					return fmt.Errorf("exiftool is required for HEIC processing but was not found in PATH")
				}
				break
			}
		}
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(files))

	for _, file := range files {
		if file.IsDir() || !isSupportedFormat(file.Name()) {
			continue
		}

		inPath := filepath.Join(inDir, file.Name())
		ext := filepath.Ext(file.Name())
		base := strings.TrimSuffix(file.Name(), ext)
		outPath := filepath.Join(outDir, fmt.Sprintf("%s_%s%s", base, mode, ext))

		wg.Add(1)
		go func(in, out string) {
			defer wg.Done()
			if err := processSingleImage(in, out, mode); err != nil {
				errCh <- fmt.Errorf("failed to process %s: %w", in, err)
			}
		}(inPath, outPath)
	}

	wg.Wait()
	close(errCh)

	var errs []error
	for err := range errCh {
		fmt.Println(err)
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}
