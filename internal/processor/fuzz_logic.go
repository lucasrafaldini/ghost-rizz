package processor

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"

	"github.com/dsoprea/go-exif/v3"
	exifcommon "github.com/dsoprea/go-exif/v3/common"
	"github.com/outis/ghost-rizz/internal/exifutil"
)

func processSingleImage(inPath, outPath, mode string) error {
	mh, err := GetMediaHandler(inPath)
	if err != nil {
		return err
	}

	switch mode {
	case "clean":
		err = mh.DropExif()
		if err != nil && !errors.Is(err, exif.ErrNoExif) && !strings.Contains(strings.ToLower(err.Error()), "no exif") {
			return err
		}
	case "fuzz":
		ib, err := createRandomEXIF()
		if err != nil {
			return err
		}
		err = mh.SetExif(ib)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown mode: %q", mode)
	}

	outF, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer func() { _ = outF.Close() }()

	return mh.Write(outF)
}

func createRandomEXIF() (*exif.IfdBuilder, error) {
	im, err := exifcommon.NewIfdMappingWithStandard()
	if err != nil {
		return nil, err
	}
	ti := exif.NewTagIndex()
	ib := exif.NewIfdBuilder(im, ti, exifcommon.IfdStandardIfdIdentity, exifcommon.EncodeDefaultByteOrder)

	randomDate := func() string {
		return fmt.Sprintf("%04d:%02d:%02d %02d:%02d:%02d",
			rand.Intn(130)+1970, rand.Intn(12)+1, rand.Intn(28)+1,
			rand.Intn(24), rand.Intn(60), rand.Intn(60))
	}

	for _, call := range []func() error{
		func() error { return exifutil.AddTag(ib, "Make", generateRandomString(10)) },
		func() error { return exifutil.AddTag(ib, "Model", generateRandomString(12)) },
		func() error { return exifutil.AddTag(ib, "Software", generateRandomString(15)) },
		func() error { return exifutil.AddTag(ib, "DateTime", randomDate()) },
	} {
		if err := call(); err != nil {
			return nil, err
		}
	}

	exifIb, err := exif.GetOrCreateIbFromRootIb(ib, "IFD/Exif")
	if err != nil {
		return nil, err
	}
	for _, call := range []func() error{
		func() error { return exifutil.AddTag(exifIb, "DateTimeOriginal", randomDate()) },
		func() error {
			return exifutil.AddTag(exifIb, "ExposureTime", []exifcommon.Rational{
				{Numerator: 1, Denominator: uint32(rand.Intn(1000) + 1)},
			})
		},
	} {
		if err := call(); err != nil {
			return nil, err
		}
	}

	gpsIb, err := exif.GetOrCreateIbFromRootIb(ib, "IFD/GPSInfo")
	if err != nil {
		return nil, err
	}

	latRefs := []string{"N", "S"}
	lonRefs := []string{"E", "W"}
	for _, call := range []func() error{
		func() error { return exifutil.AddTag(gpsIb, "GPSLatitudeRef", latRefs[rand.Intn(len(latRefs))]) },
		func() error {
			return exifutil.AddTag(gpsIb, "GPSLatitude", []exifcommon.Rational{
				{Numerator: uint32(rand.Intn(90)), Denominator: 1},
				{Numerator: uint32(rand.Intn(60)), Denominator: 1},
				{Numerator: uint32(rand.Intn(6000)), Denominator: 100},
			})
		},
		func() error { return exifutil.AddTag(gpsIb, "GPSLongitudeRef", lonRefs[rand.Intn(len(lonRefs))]) },
		func() error {
			return exifutil.AddTag(gpsIb, "GPSLongitude", []exifcommon.Rational{
				{Numerator: uint32(rand.Intn(180)), Denominator: 1},
				{Numerator: uint32(rand.Intn(60)), Denominator: 1},
				{Numerator: uint32(rand.Intn(6000)), Denominator: 100},
			})
		},
	} {
		if err := call(); err != nil {
			return nil, err
		}
	}

	return ib, nil
}

// generateRandomString creates a random alphanumeric string of length n.
func generateRandomString(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
