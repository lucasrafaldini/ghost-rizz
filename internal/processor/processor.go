package processor

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func ProcessImages(inDir, outDir, mode string) error {
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}

	files, err := ioutil.ReadDir(inDir)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(files))

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(strings.ToLower(file.Name()), ".jpg") {
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

	for err := range errCh {
		fmt.Println(err)
	}
	return nil
}
