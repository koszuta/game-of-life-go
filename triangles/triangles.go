package main

import (
	// "flag"
	"fmt"
	// "image/color"
	// "math"
	// "math/rand"
	// "strconv"
	// "time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

func main() {
	pixelgl.Run(run2)
}

func run2() {
	monitor_width, monitor_height := pixelgl.PrimaryMonitor().Size()
	fmt.Printf("%fx%f\n", monitor_width, monitor_height)
	win, err := pixelgl.NewWindow(pixelgl.WindowConfig{
		Title:     "Triangles",
		Bounds:    pixel.R(0, 0, 1000, 1000),
	})
	if err != nil {
		panic(err)
	}
	win.SetPos(pixel.Vec{
		X: 1,
		Y: 31,
	})

	rows := 5
	cols := 5
	grid := make([]bool, rows*cols)
	for i := range grid {
		grid[i] = i % 2 == 0
	}

	rowTriangles := make([]*pixel.TrianglesData, rows)
	for i := range rowTriangles {
		rowTriangles[i] = pixel.MakeTrianglesData(2*3*cols)
	}

	for row := 0; row < rows; row++ {
		
	}



	for !win.Closed() {
		win.Update()
	}
}