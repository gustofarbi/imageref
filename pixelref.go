package imageref

import (
	"github.com/lucasb-eyer/go-colorful"
	"image/color"
)

type HslColor interface {
	Hsl() (float64, float64, float64)
}
type PixelRef struct {
	Color color.Color
}

func (p *PixelRef) SetColor(_ string) bool {
	return true
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

func (p *PixelRef) SetHSL(color HslColor) {
	p.Color = colorful.Hsl(color.Hsl())
}
