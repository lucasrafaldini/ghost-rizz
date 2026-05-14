package processor

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/dsoprea/go-exif/v3"
)

// GenerateReport scans a directory for supported images and outputs a CSV report of their EXIF metadata.
func GenerateReport(inDir, outCSV string) error {
	files, err := os.ReadDir(inDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() && isSupportedFormat(file.Name()) {
			if strings.HasSuffix(strings.ToLower(file.Name()), ".heic") || strings.HasSuffix(strings.ToLower(file.Name()), ".heif") {
				if err := CheckExifTool(); err != nil {
					return err
				}
				break
			}
		}
	}

	csvFile, err := os.Create(outCSV)
	if err != nil {
		return err
	}
	defer func() { _ = csvFile.Close() }()

	writer := csv.NewWriter(csvFile)

	header := []string{"Filename", "HasEXIF", "Make", "Model", "Software", "DateTime", "DateTimeOriginal", "ExposureTime", "GPSLatitude"}
	if err := writer.Write(header); err != nil {
		return err
	}

	rowCh := make(chan []string, len(files))
	var wg sync.WaitGroup

	for _, file := range files {
		if file.IsDir() || !isSupportedFormat(file.Name()) {
			continue
		}

		inPath := filepath.Join(inDir, file.Name())
		fileName := file.Name()

		wg.Add(1)
		go func(path, name string) {
			defer wg.Done()
			row := []string{name, "false", "", "", "", "", "", "", ""}

			mh, err := GetMediaHandler(path)
			if err != nil {
				row[1] = "error"
				rowCh <- row
				return
			}

			rawExif, err := mh.RawExif()
			if err != nil {
				if !strings.Contains(err.Error(), "no exif data") {
					row[1] = "error"
				}
				rowCh <- row
				return
			}

			row[1] = "true"

			flatTags, _, err := exif.GetFlatExifData(rawExif, nil)
			if err == nil {
				tagMap := make(map[string]string)
				for _, t := range flatTags {
					tagMap[t.TagName] = t.FormattedFirst
				}
				row[2] = tagMap["Make"]
				row[3] = tagMap["Model"]
				row[4] = tagMap["Software"]
				row[5] = tagMap["DateTime"]
				row[6] = tagMap["DateTimeOriginal"]
				row[7] = tagMap["ExposureTime"]
				row[8] = tagMap["GPSLatitude"]
			}

			rowCh <- row
		}(inPath, fileName)
	}

	go func() {
		wg.Wait()
		close(rowCh)
	}()

	var writeErr error
	for row := range rowCh {
		if writeErr == nil {
			if err := writer.Write(row); err != nil {
				writeErr = err
				// drain remaining rows so the closer goroutine can finish
			}
		}
	}
	if writeErr != nil {
		return writeErr
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return err
	}

	return nil
}
