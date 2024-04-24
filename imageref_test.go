package imageref

import (
	"github.com/davidbyttow/govips/v2/vips"
	"testing"
)

func TestImageRef_ThumbnailMost(t *testing.T) {
	vipsImage, err := vips.Black(10, 5)
	if err != nil {
		t.Fatal(err)
	}

	vipsImageClone, err := vipsImage.Copy()
	if err != nil {
		t.Fatal(err)
	}

	imageRef := &ImageRef{ref: vipsImage}
	err = imageRef.ThumbnailMost(8)
	if err != nil {
		t.Fatal(err)
	}

	if max(imageRef.Width(), imageRef.Height()) != 8 {
		t.Fatalf("Expected thumbnail to be at most 8px, got %dx%d", imageRef.Width(), imageRef.Height())
	}

	err = vipsImageClone.Rotate(vips.Angle90)
	if err != nil {
		t.Fatal(err)
	}

	err = imageRef.ThumbnailMost(8)
	if err != nil {
		t.Fatal(err)
	}

	if max(imageRef.Width(), imageRef.Height()) != 8 {
		t.Fatalf("Expected thumbnail to be at most 8px, got %dx%d", imageRef.Width(), imageRef.Height())
	}
}
