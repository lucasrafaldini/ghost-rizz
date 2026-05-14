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

	// Basic check that root tags exist
	_, err = ib.FindTagWithName("Make")
	if err != nil {
		t.Errorf("BuildBaseEXIF: missing 'Make' tag: %v", err)
	}
}
