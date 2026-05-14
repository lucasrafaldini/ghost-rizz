package exifutil

import (
	"fmt"

	"github.com/dsoprea/go-exif/v3"
	exifcommon "github.com/dsoprea/go-exif/v3/common"
)

// AddTag is a helper that wraps AddStandardWithName and returns a descriptive error.
func AddTag(ib *exif.IfdBuilder, name string, value interface{}) error {
	if err := ib.AddStandardWithName(name, value); err != nil {
		return fmt.Errorf("failed to add EXIF tag %q: %w", name, err)
	}
	return nil
}

// BuildBaseEXIF creates a standard IfdBuilder with basic tags.
func BuildBaseEXIF() (*exif.IfdBuilder, error) {
	im, err := exifcommon.NewIfdMappingWithStandard()
	if err != nil {
		return nil, err
	}
	ti := exif.NewTagIndex()

	ib := exif.NewIfdBuilder(im, ti, exifcommon.IfdStandardIfdIdentity, exifcommon.EncodeDefaultByteOrder)

	for _, call := range []func() error{
		func() error { return AddTag(ib, "ImageWidth", []uint32{800}) },
		func() error { return AddTag(ib, "ImageLength", []uint32{600}) },
		func() error { return AddTag(ib, "Make", "Ghost-Rizz Generator") },
		func() error { return AddTag(ib, "Model", "Dummy 1.0") },
		func() error { return AddTag(ib, "Software", "Ghost-Rizz OS 1.0") },
		func() error { return AddTag(ib, "DateTime", "2026:05:13 12:00:00") },
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
		func() error { return AddTag(exifIb, "DateTimeOriginal", "2026:05:13 12:00:00") },
		func() error {
			return AddTag(exifIb, "ExposureTime", []exifcommon.Rational{{Numerator: 1, Denominator: 100}})
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
	for _, call := range []func() error{
		func() error { return AddTag(gpsIb, "GPSLatitudeRef", "N") },
		func() error {
			return AddTag(gpsIb, "GPSLatitude", []exifcommon.Rational{
				{Numerator: 40, Denominator: 1}, {Numerator: 45, Denominator: 1}, {Numerator: 0, Denominator: 1},
			})
		},
	} {
		if err := call(); err != nil {
			return nil, err
		}
	}

	return ib, nil
}
