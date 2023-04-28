package libvips

import (
	"github.com/davidbyttow/govips/v2/vips"
	"github.com/myposter-de/imageref/processor"
)

func init() {
	vips.LoggingSettings(nil, vips.LogLevelError)
	vips.Startup(&vips.Config{ReportLeaks: true})
}

func NewImageObject() processor.ImageObject {
	return &ImageRef{}
}

func NewPixelRef() *PixelRef {
	return &PixelRef{}
}
