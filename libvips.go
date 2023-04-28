package imageref

import (
	"github.com/davidbyttow/govips/v2/vips"
)

func init() {
	vips.LoggingSettings(nil, vips.LogLevelError)
	vips.Startup(&vips.Config{ReportLeaks: true})
}

func NewImageObject() ImageObject {
	return &ImageRef{}
}

func NewPixelRef() *PixelRef {
	return &PixelRef{}
}
