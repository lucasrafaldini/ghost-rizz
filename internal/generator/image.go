package generator

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/dsoprea/go-exif/v3"
	exifcommon "github.com/dsoprea/go-exif/v3/common"
	jpegstructure "github.com/dsoprea/go-jpeg-image-structure/v2"
)

func GenerateImages(count int, outDir string) error {
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}
	rand.Seed(time.Now().UnixNano())

	var wg sync.WaitGroup
	errCh := make(chan error, count)

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			filename := filepath.Join(outDir, fmt.Sprintf("dummy_%04d.jpg", index))
			if err := createDummyJPEGWithEXIF(filename); err != nil {
				errCh <- fmt.Errorf("failed to create image %d: %w", index, err)
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		fmt.Println(err)
	}
	return nil
}

func createDummyJPEGWithEXIF(filename string) error {
	width, height := 800, 600
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	col := color.RGBA{uint8(rand.Intn(255)), uint8(rand.Intn(255)), uint8(rand.Intn(255)), 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{col}, image.Point{}, draw.Src)

	tempFile := filename + ".tmp"
	f, err := os.Create(tempFile)
	if err != nil {
		return err
	}
	if err := jpeg.Encode(f, img, &jpeg.Options{Quality: 80}); err != nil {
		f.Close()
		return err
	}
	f.Close()
	defer os.Remove(tempFile)

	jmp := jpegstructure.NewJpegMediaParser()
	intfc, err := jmp.ParseFile(tempFile)
	if err != nil {
		return err
	}
	sl := intfc.(*jpegstructure.SegmentList)

	ib, err := buildBaseEXIF()
	if err != nil {
		return err
	}

	err = sl.SetExif(ib)
	if err != nil {
		return err
	}

	outF, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer outF.Close()

	return sl.Write(outF)
}

func buildBaseEXIF() (*exif.IfdBuilder, error) {
	im, err := exifcommon.NewIfdMappingWithStandard()
	if err != nil {
		return nil, err
	}
	ti := exif.NewTagIndex()

	ib := exif.NewIfdBuilder(im, ti, exifcommon.IfdStandardIfdIdentity, exifcommon.EncodeDefaultByteOrder)

	_ = ib.AddStandardWithName("ImageWidth", []uint32{800})
	_ = ib.AddStandardWithName("ImageLength", []uint32{600})
	_ = ib.AddStandardWithName("Make", "Ghost-Rizz Generator")
	_ = ib.AddStandardWithName("Model", "Dummy 1.0")
	_ = ib.AddStandardWithName("Software", "Ghost-Rizz OS 1.0")
	_ = ib.AddStandardWithName("DateTime", "2026:05:13 12:00:00")

	exifIb, _ := exif.GetOrCreateIbFromRootIb(ib, "IFD/Exif")
	_ = exifIb.AddStandardWithName("DateTimeOriginal", "2026:05:13 12:00:00")
	_ = exifIb.AddStandardWithName("ExposureTime", []exifcommon.Rational{{Numerator: 1, Denominator: 100}})

	gpsIb, _ := exif.GetOrCreateIbFromRootIb(ib, "IFD/GPSInfo")
	_ = gpsIb.AddStandardWithName("GPSLatitudeRef", "N")
	_ = gpsIb.AddStandardWithName("GPSLatitude", []exifcommon.Rational{
		{Numerator: 40, Denominator: 1}, {Numerator: 45, Denominator: 1}, {Numerator: 0, Denominator: 1},
	})

	return ib, nil
}
