package processor

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/dsoprea/go-exif/v3"
)

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
	defer writer.Flush()

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

	for row := range rowCh {
		_ = writer.Write(row)
	}

	return nil
}
