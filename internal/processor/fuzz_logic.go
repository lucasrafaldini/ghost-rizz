package processor

import (
	"fmt"
	"math/rand"
	"os"
	"strings"

	"github.com/dsoprea/go-exif/v3"
	exifcommon "github.com/dsoprea/go-exif/v3/common"
	jpegstructure "github.com/dsoprea/go-jpeg-image-structure/v2"
)

func processSingleImage(inPath, outPath, mode string) error {
	jmp := jpegstructure.NewJpegMediaParser()
	intfc, err := jmp.ParseFile(inPath)
	if err != nil {
		return err
	}
	sl := intfc.(*jpegstructure.SegmentList)

	if mode == "clean" {
		_, err = sl.DropExif()
		if err != nil && !strings.Contains(err.Error(), "no exif data") {
			return err
		}
	} else if mode == "fuzz" {
		ib, err := createRandomEXIF()
		if err != nil {
			return err
		}
		err = sl.SetExif(ib)
		if err != nil {
			return err
		}
	}

	outF, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer outF.Close()

	return sl.Write(outF)
}

func createRandomEXIF() (*exif.IfdBuilder, error) {
	im, err := exifcommon.NewIfdMappingWithStandard()
	if err != nil {
		return nil, err
	}
	ti := exif.NewTagIndex()
	ib := exif.NewIfdBuilder(im, ti, exifcommon.IfdStandardIfdIdentity, exifcommon.EncodeDefaultByteOrder)

	_ = ib.AddStandardWithName("Make", generateRandomString(10))
	_ = ib.AddStandardWithName("Model", generateRandomString(12))
	_ = ib.AddStandardWithName("Software", generateRandomString(15))
	_ = ib.AddStandardWithName("DateTime", fmt.Sprintf("20%02d:%02d:%02d %02d:%02d:%02d", rand.Intn(30), rand.Intn(12)+1, rand.Intn(28)+1, rand.Intn(24), rand.Intn(60), rand.Intn(60)))

	exifIb, _ := exif.GetOrCreateIbFromRootIb(ib, "IFD/Exif")
	_ = exifIb.AddStandardWithName("DateTimeOriginal", fmt.Sprintf("20%02d:%02d:%02d %02d:%02d:%02d", rand.Intn(30), rand.Intn(12)+1, rand.Intn(28)+1, rand.Intn(24), rand.Intn(60), rand.Intn(60)))
	_ = exifIb.AddStandardWithName("ExposureTime", []exifcommon.Rational{{Numerator: 1, Denominator: uint32(rand.Intn(1000) + 1)}})

	gpsIb, _ := exif.GetOrCreateIbFromRootIb(ib, "IFD/GPSInfo")

	refs := []string{"N", "S"}

	_ = gpsIb.AddStandardWithName("GPSLatitudeRef", refs[rand.Intn(len(refs))])
	_ = gpsIb.AddStandardWithName("GPSLatitude", []exifcommon.Rational{
		{Numerator: uint32(rand.Intn(90)), Denominator: 1},
		{Numerator: uint32(rand.Intn(60)), Denominator: 1},
		{Numerator: uint32(rand.Intn(6000)), Denominator: 100},
	})

	return ib, nil
}

func generateRandomString(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
