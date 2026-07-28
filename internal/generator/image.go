package generator

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"math/rand/v2"
	"os"
	"path/filepath"

	jpegstructure "github.com/dsoprea/go-jpeg-image-structure/v2"
	pngstructure "github.com/dsoprea/go-png-image-structure/v2"
	"github.com/ThothandSon/ghost-rizz/internal/exifutil"
)

// GenerateImages creates a specified number of procedural images injected with dummy EXIF metadata.
func GenerateImages(count int, outDir string) error {
	if count <= 0 {
		return nil
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}

	for i := 0; i < count; i++ {
		filename := filepath.Join(outDir, fmt.Sprintf("dummy_%04d", i))
		if i%2 == 0 {
			if err := createDummyJPEGWithEXIF(filename + ".jpg"); err != nil {
				return err
			}
		} else {
			if err := createDummyPNGWithEXIF(filename + ".png"); err != nil {
				return err
			}
		}
	}
	return nil
}

func createDummyJPEGWithEXIF(filename string) error {
	width, height := 800, 600
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	col := color.RGBA{uint8(rand.IntN(256)), uint8(rand.IntN(256)), uint8(rand.IntN(256)), 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{col}, image.Point{}, draw.Src)

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}); err != nil {
		return err
	}

	jmp := jpegstructure.NewJpegMediaParser()
	intfc, err := jmp.ParseBytes(buf.Bytes())
	if err != nil {
		return err
	}
	sl := intfc.(*jpegstructure.SegmentList)

	ib, err := exifutil.BuildBaseEXIF()
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
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{uint8(rand.IntN(256)), uint8(rand.IntN(256)), uint8(rand.IntN(256)), 255}}, image.Point{}, draw.Src)

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return err
	}

	pmp := pngstructure.NewPngMediaParser()
	intfc, err := pmp.ParseBytes(buf.Bytes())
	if err != nil {
		return err
	}
	cs := intfc.(*pngstructure.ChunkSlice)

	ib, err := exifutil.BuildBaseEXIF()
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
