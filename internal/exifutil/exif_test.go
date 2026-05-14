package exifutil

import (
	"testing"

	"github.com/dsoprea/go-exif/v3"
	exifcommon "github.com/dsoprea/go-exif/v3/common"
)

func TestAddTag(t *testing.T) {
	im, _ := exifcommon.NewIfdMappingWithStandard()
	ti := exif.NewTagIndex()
	ib := exif.NewIfdBuilder(im, ti, exifcommon.IfdStandardIfdIdentity, exifcommon.EncodeDefaultByteOrder)

	// Test valid tag
	if err := AddTag(ib, "Make", "TestMake"); err != nil {
		t.Errorf("AddTag(Make) failed: %v", err)
	}

	// Test invalid tag
	if err := AddTag(ib, "InvalidTagXYZ", "value"); err == nil {
		t.Errorf("AddTag(InvalidTagXYZ) expected error, got nil")
	}
}

func TestBuildBaseEXIF(t *testing.T) {
	ib, err := BuildBaseEXIF()
	if err != nil {
		t.Fatalf("BuildBaseEXIF failed: %v", err)
	}
	if ib == nil {
		t.Fatal("expected non-nil IfdBuilder")
	}

	// Check root tags
	for _, tag := range []string{"ImageWidth", "ImageLength", "Make", "Model", "Software", "DateTime"} {
		_, err = ib.FindTagWithName(tag)
		if err != nil {
			t.Errorf("BuildBaseEXIF: missing %q tag: %v", tag, err)
		}
	}

	// Check Exif IFD
	exifIb, err := exif.GetOrCreateIbFromRootIb(ib, "IFD/Exif")
	if err != nil {
		t.Fatalf("failed to get Exif IFD: %v", err)
	}
	for _, tag := range []string{"DateTimeOriginal", "ExposureTime"} {
		_, err = exifIb.FindTagWithName(tag)
		if err != nil {
			t.Errorf("BuildBaseEXIF: missing %q tag: %v", tag, err)
		}
	}

	// Check GPS IFD
	gpsIb, err := exif.GetOrCreateIbFromRootIb(ib, "IFD/GPSInfo")
	if err != nil {
		t.Fatalf("failed to get GPS IFD: %v", err)
	}
	for _, tag := range []string{"GPSLatitudeRef", "GPSLatitude"} {
		_, err = gpsIb.FindTagWithName(tag)
		if err != nil {
			t.Errorf("BuildBaseEXIF: missing %q tag: %v", tag, err)
		}
	}
}

