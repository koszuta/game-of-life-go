package life

import (
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

const n = 42

type Lifer interface {
	Reset()
	Step()
}

type Conf struct {
	WindowWidth  int
	WindowHeight int
	NRows        int
	NCols        int
	TPS          int
}

func newVert(x, y int, xScale, yScale float32, c color.RGBA) ebiten.Vertex {
	return ebiten.Vertex{
		DstX:   float32(x) * xScale,
		DstY:   float32(y) * yScale,
		ColorR: float32(c.R) / 255.0,
		ColorG: float32(c.G) / 255.0,
		ColorB: float32(c.B) / 255.0,
		ColorA: float32(c.A) / 255.0,
	}
}

var (
	debug = false

	updates = make([]time.Duration, n)
	iUpdate int
)
