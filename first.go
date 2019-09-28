package main

import (
	"flag"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

type Vec2 struct {
	x int
	y int
}

var (
	turnCount, drawCount, resizeCount             int64
	totalTurnTime, totalDrawTime, totalResizeTime time.Duration

	fps, turn_rate int
	window                *pixelgl.Window
	win_width, win_height float64

	rows, cols      int
	grid, buff      []bool
	rowTriangles    []*pixel.TrianglesData
)

func main() {
	rand.Seed(time.Now().UnixNano())
	flag.IntVar(&fps, "fps", 60, "frames per second")
	flag.IntVar(&rows, "rows", 100, "number of rows")
	flag.IntVar(&cols, "cols", 100, "number of columns")
	flag.IntVar(&turn_rate, "rate", 12, "turn rate")
	flag.Parse()
	pixelgl.Run(run)
}

func run() {
	monitor_width, monitor_height := pixelgl.PrimaryMonitor().Size()
	win, err := pixelgl.NewWindow(pixelgl.WindowConfig{
		Title:     "Life (" + strconv.Itoa(fps) + " FPS)",
		Bounds:    pixel.R(0, 0, 1000, 1000),
		Resizable: true,
	})
	if err != nil {
		panic(err)
	}
	window = win
	window.SetPos(pixel.Vec{
		X: 1,
		Y: 31,
	})
	// picture = pixel.MakePictureData(window.Bounds())

	fmt.Printf("monitor=%dx%d\n", int(monitor_width), int(monitor_height))
	fmt.Printf("fps=%d\n", fps)
	fmt.Printf("rate=%d\n", turn_rate)
	fmt.Printf("grid=%dx%d\n", rows, cols)

	if turn_rate > fps {
		turn_rate = 1
	} else {
		turn_rate = int(fps / turn_rate)
	}

	grid = make([]bool, rows*cols)
	buff = make([]bool, rows*cols)

	init_grid()

	rowTriangles = make([]*pixel.TrianglesData, rows)
	for i := range rowTriangles {
		rowTriangles[i] = pixel.MakeTrianglesData(2 * 3 * cols)
	}

	win_width = window.Bounds().Max.X
	win_height = window.Bounds().Max.Y
	update_vertex_positions()

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

	frame := 0
	last := time.Now()
	for now := range time.Tick(time.Second / time.Duration(fps)) {
		// if frame == 10*fps {
		// 	window.Destroy()
		// 	break
		// }

		if window.Closed() {
			break
		}

		win_width_now := window.Bounds().Max.X
		win_height_now := window.Bounds().Max.Y
		if win_width != win_width_now || win_height != win_height_now {
			win_width = win_width_now
			win_height = win_height_now
			update_vertex_positions()
			fmt.Printf("window changed to %s x %s\n", strconv.FormatFloat(win_width, 'f', 1, 64), strconv.FormatFloat(win_height, 'f', 1, 64))
		}

		if window.JustPressed(pixelgl.KeySpace) {
			init_grid()
		}

		if frame%fps == fps-1 {
			curr_fps := float64(fps) * float64(time.Second) / float64(now.Sub(last))
			last = now
			window.SetTitle("Life (" + strconv.FormatFloat(curr_fps, 'f', 2, 64) + " FPS)")
		}

		window.Update()

		if frame%fps%turn_rate == 0 {
			turn()
		}

		draw(frame)
		frame++
	}
	fmt.Printf("\nAvg turn time %s\n", time.Duration(totalTurnTime.Nanoseconds()/turnCount))
	fmt.Printf("\nAvg draw time %s\n", time.Duration(totalDrawTime.Nanoseconds()/drawCount))
}

func init_grid() {
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			grid[r*cols+c] = rand.Intn(3) == 0
		}
	}
}

func turn() {
	defer func(start time.Time) {
		turnCount++
		elapsed := time.Since(start)
		totalTurnTime += elapsed
		// fmt.Printf("life time %s\n", elapsed)
	}(time.Now())

	for r := 0; r < rows; r++ {
		triangles := *rowTriangles[r]
		for c := 0; c < cols; c++ {

			r_up := r + 1
			r_down := r - 1
			c_left := c - 1
			c_right := c + 1

			if r == 0 {
				r_down = rows - 1
			}
			if c == 0 {
				c_left = cols - 1
			}
			if r == rows-1 {
				r_up = 0
			}
			if c == cols-1 {
				c_right = 0
			}

			r_up = r_up * cols
			r_down = r_down * cols
			r_same := r * cols

			var living_neighbors byte
			if grid[r_up+c] {
				living_neighbors++
			}
			if grid[r_up+c_right] {
				living_neighbors++
			}
			if grid[r_same+c_right] {
				living_neighbors++
			}
			if grid[r_down+c_right] {
				living_neighbors++
			}
			if grid[r_down+c] {
				living_neighbors++
			}
			if grid[r_down+c_left] {
				living_neighbors++
			}
			if grid[r_same+c_left] {
				living_neighbors++
			}
			if grid[r_up+c_left] {
				living_neighbors++
			}

			i := r*cols + c
			buff[i] = living_neighbors == 3 || (grid[i] && living_neighbors == 2)

			var color pixel.RGBA
			if buff[i] {
				color = chartreuse
			} else {
				color = black
			}
			for i := 0; i < 6; i++ {
				triangles[6*c+i].Color = color
			}
		}
	}

	grid, buff = buff, grid
}

func draw(frame int) {
	defer func(start time.Time) {
		drawCount++
		elapsed := time.Since(start)
		totalDrawTime += elapsed
		// fmt.Printf("draw time %s\n", elapsed)
	}(time.Now())

	for _, triangles := range rowTriangles {
		window.MakeTriangles(triangles).Draw()
	}
}

func update_vertex_positions() {
	defer func(start time.Time) {
		resizeCount++
		elapsed := time.Since(start)
		totalResizeTime += elapsed
	}(time.Now())

	row_height := win_height / float64(rows)
	col_width := win_width / float64(cols)
	fmt.Printf("row_height=%f col_width=%f\n\n", row_height, col_width)
	for i, it := range rowTriangles {
		triangles := *it
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
			X: win_width,
			Y: float64(i) * row_height,
		}
		triangles[len(triangles)-2].Position = triangles[len(triangles)-4].Position
		triangles[len(triangles)-1].Position = pixel.Vec{
			X: win_width,
			Y: float64(i+1) * row_height,
		}
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
