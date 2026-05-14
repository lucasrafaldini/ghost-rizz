package processor

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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

	csvFile, err := os.Create(outCSV)
	if err != nil {
		return err
	}
	defer func() { _ = csvFile.Close() }()

	writer := csv.NewWriter(csvFile)

	header := []string{"Filename", "HasEXIF", "Make", "Model", "Software", "DateTime", "DateTimeOriginal", "ExposureTime", "GPSLatitude", "Error"}
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
			row := []string{name, "false", "", "", "", "", "", "", "", ""}

			if strings.HasSuffix(strings.ToLower(name), ".heic") || strings.HasSuffix(strings.ToLower(name), ".heif") {
				if err := CheckExifTool(); err != nil {
					row[1] = "error"
					row[9] = err.Error()
					rowCh <- row
					return
				}
			}

			mh, err := GetMediaHandler(path)
			if err != nil {
				row[1] = "error"
				row[9] = err.Error()
				rowCh <- row
				return
			}

			rawExif, err := mh.RawExif()
			if err != nil {
				if !strings.Contains(err.Error(), "no exif data") {
					row[1] = "error"
					row[9] = err.Error()
				}
				rowCh <- row
				return
			}

			row[1] = "true"

			flatTags, _, err := exif.GetFlatExifData(rawExif, nil)
			if err == nil {
				tagMap := make(map[string]string)
				for _, t := range flatTags {
					// Key by IfdPath and TagName to avoid collisions (e.g. IFD0 vs ExifIFD)
					key := fmt.Sprintf("%s:%s", t.IfdPath, t.TagName)
					tagMap[key] = t.FormattedFirst
				}
				row[2] = tagMap["IFD0:Make"]
				row[3] = tagMap["IFD0:Model"]
				row[4] = tagMap["IFD0:Software"]
				row[5] = tagMap["IFD0:DateTime"]
				row[6] = tagMap["IFD/Exif:DateTimeOriginal"]
				row[7] = tagMap["IFD/Exif:ExposureTime"]
				row[8] = tagMap["IFD/GPSInfo:GPSLatitude"]
			}

			rowCh <- row
		}(inPath, fileName)
	}

	go func() {
		wg.Wait()
		close(rowCh)
	}()

	var rows [][]string
	for row := range rowCh {
		rows = append(rows, row)
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i][0] < rows[j][0]
	})

	var writeErr error
	for _, row := range rows {
		if writeErr == nil {
			if err := writer.Write(row); err != nil {
				writeErr = err
			}
		}
	}

	// Flush even on writeErr to ensure partial results are written if needed,
	// or just return the error. Copilot suggested being more deliberate.
	writer.Flush()
	if flushErr := writer.Error(); flushErr != nil && writeErr == nil {
		writeErr = flushErr
	}

	if writeErr != nil {
		_ = csvFile.Close()
		_ = os.Remove(outCSV) // Remove partial/broken report
		return writeErr
	}

	return nil
}
