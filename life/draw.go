package life

import (
	"image"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

var (
	whiteImage = ebiten.NewImage(3, 3)

	// whiteSubImage is an internal sub image of whiteImage.
	// Use whiteSubImage at DrawTriangles instead of whiteImage in order to avoid bleeding edges.
	whiteSubImage = whiteImage.SubImage(image.Rect(1, 1, 2, 2)).(*ebiten.Image)

	black      = color.RGBA{0, 0, 0, 1}
	chartreuse = color.RGBA{128, 255, 0, 255}
)

func init() {
	whiteImage.Fill(color.White)
}

var (
	draws  = make([]time.Duration, n)
	iDraws int
)
