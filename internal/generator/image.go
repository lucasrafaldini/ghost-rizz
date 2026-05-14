package generator

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"math/rand"
	"os"
	"path/filepath"
	"sync"

	"github.com/dsoprea/go-exif/v3"
	exifcommon "github.com/dsoprea/go-exif/v3/common"
	jpegstructure "github.com/dsoprea/go-jpeg-image-structure/v2"
	pngstructure "github.com/dsoprea/go-png-image-structure/v2"
)

// GenerateImages creates a specified number of procedural images injected with dummy EXIF metadata.
func GenerateImages(count int, outDir string) error {
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}
	var wg sync.WaitGroup
	errCh := make(chan error, count)

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			var err error
			filename := filepath.Join(outDir, fmt.Sprintf("dummy_%04d", index))
			if index%2 == 0 {
				err = createDummyJPEGWithEXIF(filename + ".jpg")
			} else {
				err = createDummyPNGWithEXIF(filename + ".png")
			}
			if err != nil {
				errCh <- fmt.Errorf("failed to create image %d: %w", index, err)
			}
		}(i)
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
		_ = f.Close()
		return err
	}
	_ = f.Close()
	defer func() { _ = os.Remove(tempFile) }()

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
	defer func() { _ = outF.Close() }()

	return sl.Write(outF)
}

func createDummyPNGWithEXIF(filename string) error {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{uint8(rand.Intn(255)), uint8(rand.Intn(255)), uint8(rand.Intn(255)), 255}}, image.Point{}, draw.Src)

	tempFile := filename + ".tmp.png"
	f, err := os.Create(tempFile)
	if err != nil {
		return err
	}
	if err := png.Encode(f, img); err != nil {
		_ = f.Close()
		return err
	}
	_ = f.Close()
	defer func() { _ = os.Remove(tempFile) }()

	pmp := pngstructure.NewPngMediaParser()
	intfc, err := pmp.ParseFile(tempFile)
	if err != nil {
		return err
	}
	cs := intfc.(*pngstructure.ChunkSlice)

	ib, err := buildBaseEXIF()
	if err != nil {
		return err
	}

	err = cs.SetExif(ib)
	if err != nil {
		return err
	}

	outF, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() { _ = outF.Close() }()

	return cs.WriteTo(outF)
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

	exifIb, err := exif.GetOrCreateIbFromRootIb(ib, "IFD/Exif")
	if err != nil {
		return nil, err
	}
	_ = exifIb.AddStandardWithName("DateTimeOriginal", "2026:05:13 12:00:00")
	_ = exifIb.AddStandardWithName("ExposureTime", []exifcommon.Rational{{Numerator: 1, Denominator: 100}})

	gpsIb, err := exif.GetOrCreateIbFromRootIb(ib, "IFD/GPSInfo")
	if err != nil {
		return nil, err
	}
	_ = gpsIb.AddStandardWithName("GPSLatitudeRef", "N")
	_ = gpsIb.AddStandardWithName("GPSLatitude", []exifcommon.Rational{
		{Numerator: 40, Denominator: 1}, {Numerator: 45, Denominator: 1}, {Numerator: 0, Denominator: 1},
	})

	return ib, nil
}
