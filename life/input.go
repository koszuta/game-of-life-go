package life

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func isKeyStillPressed(k ebiten.Key) bool {
	return inpututil.KeyPressDuration(k) > 0
}
