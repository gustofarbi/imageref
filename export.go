package imageref

import "github.com/davidbyttow/govips/v2/vips"

type ExportParams struct {
	// common params
	StripMetadata      bool
	Quality            int
	Interlace          bool

	// jpeg
	OptimizeCoding     bool
	SubsampleMode      vips.SubsampleMode
	TrellisQuant       bool
	OvershootDeringing bool
	OptimizeScans      bool
	QuantTable         int

	// png
	Compression   int
	Filter        vips.PngFilter
	Palette       bool
	Dither        float64
	Bitdepth      int
	Profile       string

	// webp
	Lossless        bool
	NearLossless    bool
	ReductionEffort int
	IccProfile      string
}
