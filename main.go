package main

import (
	"log"

	"github.com/koszuta/game-of-life-go/life"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	game := life.NewOldLife(life.Conf{
		// game := life.NewMapLife(life.Conf{
		WindowHeight: 1000,
		WindowWidth:  1000,
		NRows:        1000,
		NCols:        1000,
		TPS:          12,
	})
	if err := ebiten.RunGame(game); err != nil {
		log.Fatalln(err)
	}
}
