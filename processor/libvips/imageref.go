package libvips

import (
	"github.com/davidbyttow/govips/v2/vips"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/myposter-de/imageref/constant/format"
	"github.com/myposter-de/imageref/processor"
	"github.com/myposter-de/imageref/processor/colorspace"
	"github.com/myposter-de/imageref/processor/composite"
	"math"
	"os"
)

type ImageRef struct {
	ref *vips.ImageRef
}

const LinearPrecision = 0.001

func (i *ImageRef) AdjustChroma(m float64) error {
	if math.Abs(m-1) < LinearPrecision {
		return nil
	}
	return i.ref.Linear([]float64{1, m, 1, 1}, []float64{0, 0, 0, 0})
}

func (i *ImageRef) Write(path string) error {
	params := vips.NewPngExportParams()
	params.Compression = 0
	params.StripMetadata = true
	params.Interlace = false
	data, _, err := i.ref.ExportPng(params)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, os.ModePerm)
}

func (i *ImageRef) Compare(reference processor.ImageObject) (float64, error) {
	c, err := i.ref.Copy()
	if err != nil {
		return 0, err
	}
	ref := reference.(*ImageRef).ref
	cref, err := ref.Copy()
	if err != nil {
		return 0, err
	}
	err = c.Composite(cref, vips.BlendModeDifference, 0, 0)
	if err != nil {
		return 0, err
	}

	err = c.Composite(c, vips.BlendModeMultiply, 0, 0)
	if err != nil {
		return 0, err
	}
	//remove alpha
	//this is needed because alpha is always added on composite and by default alpha layer will contain values of 255,
	//which means 100% opacity
	err = c.ExtractBand(0, 3)
	if err != nil {
		return 0, err
	}

	return c.Average()
}

func (i *ImageRef) Crop(width uint, height uint, x int, y int) error {
	return i.ref.ExtractArea(x, y, int(width), int(height))
}

func (i *ImageRef) ImportFile(path string) error {
	img, err := vips.NewImageFromFile(path)
	if err != nil {
		return err
	}
	i.ref = img
	return nil
}

func (i *ImageRef) Export(outputFormat string) ([]byte, error) {
	var result []byte
	var err error

	switch outputFormat {
	case format.Jpg:
		result, _, err = i.ref.ExportJpeg(vips.NewJpegExportParams())
	case format.WebP:
		result, _, err = i.ref.ExportWebp(&vips.WebpExportParams{
			StripMetadata:   false,
			Quality:         75,
			Lossless:        false,
			ReductionEffort: 0,
		})
	case format.Png:
		fallthrough
	default:
		result, _, err = i.ref.ExportPng(vips.NewPngExportParams())
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (i *ImageRef) AddAlpha() error {
	return i.ref.AddAlpha()
}

func (i *ImageRef) Color(color colorful.Color) error {
	r, g, b := color.RGB255()
	return i.ref.Linear([]float64{0, 0, 0, 0}, []float64{float64(r), float64(g), float64(b), math.MaxUint8})
}

func (i *ImageRef) Tint(tint processor.PixelObject) error {
	var err error
	err = i.ref.ToColorSpace(vips.InterpretationRGB16)
	if err != nil {
		return err
	}
	err = i.ref.UnpremultiplyAlpha()
	if err != nil {
		return err
	}
	alphaLayer, err := i.ref.Copy()
	if err != nil {
		return err
	}
	err = alphaLayer.ExtractBand(3, 1)
	if err != nil {
		return err
	}
	err = i.ref.ExtractBand(0, 3)
	if err != nil {
		return err
	}
	tintRef, err := i.ref.Copy()
	if err != nil {
		return err
	}
	t := tint.(*PixelRef).Color
	r, g, b, _ := t.RGBA()
	err = tintRef.Linear([]float64{0, 0, 0}, []float64{float64(r), float64(g), float64(b)})
	if err != nil {
		return err
	}
	err = i.ref.Composite(tintRef, vips.BlendModeSoftLight, 0, 0)
	if err != nil {
		return err
	}
	err = i.ref.Composite(tintRef, vips.BlendModeSoftLight, 0, 0)
	if err != nil {
		return err
	}
	err = i.ref.Composite(tintRef, vips.BlendModeSoftLight, 0, 0)
	if err != nil {
		return err
	}
	err = i.ref.ExtractBand(0, 3)
	if err != nil {
		return err
	}
	err = i.ref.BandJoin(alphaLayer)
	if err != nil {
		return err
	}
	return i.ref.ToColorSpace(vips.InterpretationLCH)
}

func (i *ImageRef) Contrast(factor float64) error {
	if math.Abs(factor-float64(1)) < LinearPrecision {
		return nil
	}
	var err error
	r := i.ref
	err = i.ref.ToColorSpace(vips.InterpretationSRGB)
	if err != nil {
		return err
	}
	a := factor
	b := -(128*factor - 128)

	return r.Linear([]float64{a, a, a, 1}, []float64{b, b, b, 0})
}

func (i *ImageRef) TransformColorspace(t colorspace.Type) error {
	return i.ref.ToColorSpace(parseVipsColorspace(t))
}

func parseVipsColorspace(c colorspace.Type) vips.Interpretation {
	switch c {
	case colorspace.SRGB:
		return vips.InterpretationSRGB
	case colorspace.Gray:
		return vips.InterpretationGrey16
	default:
		return 0
	}
}

func (i *ImageRef) Clone() (processor.ImageObject, error) {
	c, err := i.ref.Copy()
	if err != nil {
		return nil, err
	}
	return &ImageRef{ref: c}, nil
}

func (i *ImageRef) Negate() error {
	err := i.ref.Invert()
	if err != nil {
		return err
	}
	if i.ref.Bands() == 2 {
		return i.ref.Linear([]float64{0, 1}, []float64{0, 0})
	}
	if i.ref.Bands() == 4 {
		return i.ref.Linear([]float64{0, 0, 0, 1}, []float64{0, 0, 0, 0})
	}
	return nil
}

func (i *ImageRef) HasImage() bool {
	return i.ref.Width() != 0
}

func (i *ImageRef) Height() uint {
	return uint(i.ref.Height())
}

func (i *ImageRef) Composite(overlay processor.ImageObject, mode composite.Mode) error {
	return i.ref.Composite(overlay.(*ImageRef).ref, blendMode(mode), 0, 0)
}

func blendMode(mode composite.Mode) vips.BlendMode {
	switch mode {
	case composite.DestIn:
		return vips.BlendModeDestIn
	case composite.Over:
		return vips.BlendModeOver
	case composite.DestOver:
		return vips.BlendModeDestOver
	default:
		return 0
	}
}

func (i *ImageRef) DistortPerspective(distortion []float64) error {
	return DistortPerspective(i.ref, distortion)
}

func (i *ImageRef) Resize(width uint, height uint) error {
	var err error = nil

	xscale := float64(width) / float64(i.ref.Width())
	yscale := float64(height) / float64(i.ref.Height())

	err = i.ref.ResizeWithVScale(xscale, yscale, vips.KernelLanczos3)
	if err != nil {
		return err
	}

	return nil
}

func (i *ImageRef) Width() uint {
	return uint(i.ref.Width())
}

func (i *ImageRef) Import(bytes []byte) error {
	ref, err := vips.NewImageFromBuffer(bytes)
	if err != nil {
		return err
	}
	i.ref = ref
	return nil
}

func (i *ImageRef) CopyTransparency(overlay processor.ImageObject) error {
	overlayVips := overlay.(*ImageRef).ref
	baseTransparency, err := i.ref.Copy()
	if err != nil {
		return err
	}

	err = overlayVips.ExtractBand(overlayVips.Bands()-1, 1)
	if err != nil {
		return err
	}
	err = baseTransparency.ExtractBand(baseTransparency.Bands()-1, 1)
	if err != nil {
		return err
	}

	err = baseTransparency.Composite(overlayVips, vips.BlendModeMultiply, 0, 0)

	if err != nil {
		return err
	}

	err = baseTransparency.ExtractBand(0, 1)

	if err != nil {
		return err
	}

	err = i.ref.ExtractBand(0, i.ref.Bands()-1)
	if err != nil {
		return err
	}
	return i.ref.BandJoin(baseTransparency)
}

func (i *ImageRef) AdjustLightness(multiplier float64) error {
	if math.Abs(multiplier-1) < LinearPrecision {
		return nil
	}
	return i.ref.Linear([]float64{multiplier, 1, 1, 1}, []float64{0, 0, 0, 0})
}

func (i *ImageRef) Thumbnail(width int) error {
	return i.ref.Thumbnail(width, 0, vips.InterestingAll)
}

func (i *ImageRef) Close() {
	i.ref.Close()
}
