package imageref

import (
	"github.com/lucasb-eyer/go-colorful"
	"image/color"
)

type PixelRef struct {
	Color color.Color
}

func NewPixel() *PixelRef {
	return &PixelRef{}
}

func (p *PixelRef) SetColor(c color.Color) {
	p.Color = c
}

func (p *PixelRef) AdjustLightness(factor float64) {
	c, _ := colorful.MakeColor(p.Color)
	h, s, l := c.Hsl()
	l *= factor
	p.Color = colorful.Hsl(h, s, l)
}

func (p *PixelRef) AdjustSaturation(factor float64) {
	c, _ := colorful.MakeColor(p.Color)
	h, s, l := c.Hsl()
	s *= factor
	p.Color = colorful.Hsl(h, s, l)
}

func (p *PixelRef) RGBA() (r, g, b, a uint32) {
	return p.Color.RGBA()
}
