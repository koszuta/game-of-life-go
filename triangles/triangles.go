package main

import (
	"fmt"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

type Vec2 struct {
	x, y int
}

var (
	width, height int
	window        *pixelgl.Window

	rows, cols int
	grid       []bool
	cells      []Vec2
)

func main() {
	pixelgl.Run(run)
}

func run() {
	monitor_width, monitor_height := pixelgl.PrimaryMonitor().Size()
	fmt.Printf("%f x %f\n", monitor_width, monitor_height)
	win, err := pixelgl.NewWindow(pixelgl.WindowConfig{
		Title:  "Triangles",
		Bounds: pixel.R(0, 0, 1000, 1000),
	})
	if err != nil {
		panic(err)
	}
	win.SetPos(pixel.Vec{
		X: 1,
		Y: 31,
	})

	rows = 7
	cols = 7
	grid = make([]bool, rows*cols)
	for i := range grid {
		grid[i] = i%2 == 0
	}

	width := win.Bounds().Max.X
	height := win.Bounds().Max.Y
	row_height := height / float64(rows)
	col_width := width / float64(cols)
	fmt.Printf("row_height=%f col_width=%f\n\n", row_height, col_width)

	rowTriangles := make([]*pixel.TrianglesData, rows)
	for i := range rowTriangles {
		rowTriangles[i] = pixel.MakeTrianglesData(2 * 3 * cols)
		triangles := *rowTriangles[i]

		triangles[0].Position = pixel.Vec{
			X: 0,
			Y: float64(i) * row_height,
		}
		triangles[1].Position = pixel.Vec{
			X: 0,
			Y: float64(i+1) * row_height,
		}
		triangles[3].Position = triangles[1].Position

		for j := 0; j < cols-1; j++ {
			lower_vertex := pixel.Vec{
				X: float64(j+1) * col_width,
				Y: float64(i) * row_height,
			}
			triangles[6*j+2].Position = lower_vertex
			triangles[6*j+4].Position = lower_vertex
			triangles[6*j+6].Position = lower_vertex

			upper_vertex := pixel.Vec{
				X: float64(j+1) * col_width,
				Y: float64(i+1) * row_height,
			}
			triangles[6*j+5].Position = upper_vertex
			triangles[6*j+7].Position = upper_vertex
			triangles[6*j+9].Position = upper_vertex
		}

		triangles[len(triangles)-4].Position = pixel.Vec{
			X: width,
			Y: float64(i) * row_height,
		}
		triangles[len(triangles)-2].Position = triangles[len(triangles)-4].Position
		triangles[len(triangles)-1].Position = pixel.Vec{
			X: width,
			Y: float64(i+1) * row_height,
		}
	}

	for row := 0; row < rows; row++ {
		triangles := *rowTriangles[row]
		for col := 0; col < cols; col++ {
			var color pixel.RGBA
			if grid[row*cols+col] {
				color = chartreuse
			} else {
				color = black
			}
			for i := 0; i < 6; i++ {
				triangles[6*col+i].Color = color
			}
		}
	}

	for _, triangles := range rowTriangles {
		win.MakeTriangles(triangles).Draw()
	}

	for !win.Closed() {
		win.Update()
	}
}

var chartreuse pixel.RGBA = pixel.RGBA{
	R: 0.5,
	G: 1.0,
	B: 0.0,
	A: 1.0,
}

var black pixel.RGBA = pixel.RGBA{
	R: 0.0,
	G: 0.0,
	B: 0.0,
	A: 1.0,
}
