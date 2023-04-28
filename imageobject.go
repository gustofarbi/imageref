package imageref

import (
	"github.com/lucasb-eyer/go-colorful"
)

type ImageObject interface {
	Import(bytes []byte) error
	Width() uint
	Resize(width uint, height uint) error
	DistortPerspective(distortion []float64) error
	Composite(node ImageObject, mode Composite) error
	Height() uint
	HasImage() bool
	Negate() error
	Clone() (ImageObject, error)
	TransformColorspace(t ColorspaceType) error
	AdjustLightness(modifier float64) error
	Contrast(modifier float64) error
	Tint(tint PixelObject) error
	AddAlpha() error
	Export(format string) ([]byte, error)
	ImportFile(path string) error
	Crop(width uint, height uint, x int, y int) error
	Compare(reference ImageObject) (float64, error)
	Write(path string) error
	CopyTransparency(node ImageObject) error
	AdjustChroma(m float64) error
	Color(color colorful.Color) error
	Thumbnail(width int) error
	Close()
}
