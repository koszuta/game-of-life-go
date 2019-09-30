package main

import (
	"flag"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

var (
	turns, draws, resizes                               int64
	total_turn_time, total_draw_time, total_resize_time time.Duration

	turn_rate             int
	window                *pixelgl.Window
	win_width, win_height float64

	rows, cols                    int
	grid, buff                    []bool
	diff                          []int
	verts_per_cell, verts_per_row int
	triangles                     pixel.TrianglesData
	batch                         *pixel.Batch

	alive_color pixel.RGBA = chartreuse
	dead_color  pixel.RGBA = black
)

func main() {
	rand.Seed(time.Now().UnixNano())
	flag.IntVar(&rows, "rows", 100, "number of rows")
	flag.IntVar(&cols, "cols", 100, "number of columns")
	flag.IntVar(&turn_rate, "rate", 12, "turns per second")
	flag.Parse()
	pixelgl.Run(run)
}

func run() {
	monitor_width, monitor_height := pixelgl.PrimaryMonitor().Size()
	refresh_rate := pixelgl.PrimaryMonitor().RefreshRate()
	win, err := pixelgl.NewWindow(pixelgl.WindowConfig{
		Title:     "Life (" + strconv.FormatFloat(refresh_rate, 'f', 2, 64) + " FPS)",
		Bounds:    pixel.R(0, 0, 1000, 1000),
		Resizable: true,
		VSync:     true,
	})
	if err != nil {
		panic(err)
	}
	window = win
	window.SetPos(pixel.Vec{
		X: 1,
		Y: 26,
	})

	if window.VSync() && turn_rate > int(refresh_rate) {
		turn_rate = int(refresh_rate)
	}

	grid = make([]bool, rows*cols)
	buff = make([]bool, rows*cols)
	diff = make([]int, 0, rows*cols)
	init_grid()

	verts_per_cell = 2 * 3
	verts_per_row = verts_per_cell * cols
	triangles = make(pixel.TrianglesData, verts_per_row*rows)
	batch = pixel.NewBatch(&triangles, nil)

	fmt.Printf("\nmonitor:\t%dx%dpx\n", int(monitor_width), int(monitor_height))
	fmt.Printf("refresh rate:\t%dHz\n", int(refresh_rate))
	fmt.Printf("window:\t\t%dx%dpx\n", int(window.Bounds().W()), int(window.Bounds().H()))
	fmt.Printf("grid size:\t%dx%d\n", rows, cols)
	fmt.Printf("turn rate:\t%d\n\n", turn_rate)

	frame := 0
	last_frame := frame
	var t int64 = 0
	var turn_accumulator int64 = 0
	var turn_dt int64 = time.Second.Nanoseconds() / int64(turn_rate)
	var title_accumulator int64 = 0
	var title_dt int64 = time.Second.Nanoseconds()
	last := time.Now()
	for !window.Closed() {
		now := time.Now()
		frame_time := now.Sub(last).Nanoseconds()
		turn_accumulator += frame_time
		title_accumulator += frame_time
		last = now

		if window.JustPressed(pixelgl.KeyEscape) {
			window.Destroy()
			break
		}

		if window.JustPressed(pixelgl.KeySpace) {
			clear_screen()
			init_grid()
			turn_accumulator = 0
		}

		win_width_now := window.Bounds().W()
		win_height_now := window.Bounds().H()
		if win_width != win_width_now || win_height != win_height_now {
			fmt.Printf("window changed to %dx%d\n", int(win_width_now), int(win_height_now))
			win_width = win_width_now
			win_height = win_height_now
			update_vertex_positions()
		}

		if title_accumulator >= title_dt {
			curr_fps := float64(frame-last_frame) / time.Duration(title_accumulator).Seconds()
			window.SetTitle("Life (" + strconv.FormatFloat(curr_fps, 'f', 2, 64) + " FPS)")
			last_frame = frame
			title_accumulator -= title_dt
		}

		if turn_accumulator >= turn_dt {
			set_cell_color()
			turn()
			turn_accumulator -= turn_dt
			t += turn_dt
		}

		multiplier := float64(turn_accumulator) / float64(turn_dt)
		render(multiplier)
		window.Update()
		frame++
	}

	fmt.Printf("\n%d frames drawn\n", frame)
	fmt.Printf("Avg draw time %s\n", time.Duration(total_draw_time.Nanoseconds()/draws))
	fmt.Printf("Avg turn time %s\n", time.Duration(total_turn_time.Nanoseconds()/turns))
	fmt.Printf("Avg resize time %s\n", time.Duration(total_resize_time.Nanoseconds()/resizes))
}

func init_grid() {
	diff = diff[:0]
	for cell := 0; cell < rows*cols; cell++ {
		grid[cell] = rand.Intn(3) == 0
		if grid[cell] {
			diff = append(diff, cell)
		}
	}
}

func clear_screen() {
	for cell := 0; cell < rows*cols; cell++ {
		for vert := 0; vert < verts_per_cell; vert++ {
			triangles[cell*verts_per_cell+vert].Color = dead_color
		}
	}
}

func set_cell_color() {
	for _, cell := range diff {
		var color pixel.RGBA
		if grid[cell] {
			color = alive_color
		} else {
			color = dead_color
		}
		for vert := 0; vert < verts_per_cell; vert++ {
			triangles[cell*verts_per_cell+vert].Color = color
		}
	}
}

func turn() {
	defer func(start time.Time) {
		turns++
		total_turn_time += time.Since(start)
	}(time.Now())

	defer func() {
		grid, buff = buff, grid
	}()

	diff = diff[:0]
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			r_up := row + 1
			r_down := row - 1
			c_left := col - 1
			c_right := col + 1

			if row == 0 {
				r_down = rows - 1
			}
			if col == 0 {
				c_left = cols - 1
			}
			if row == rows-1 {
				r_up = 0
			}
			if col == cols-1 {
				c_right = 0
			}

			r_up = r_up * cols
			r_down = r_down * cols
			r_same := row * cols

			var living_neighbors byte
			if grid[r_up+col] {
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
			if grid[r_down+col] {
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

			cell := row*cols + col
			buff[cell] = living_neighbors == 3 || (grid[cell] && living_neighbors == 2)
			if grid[cell] != buff[cell] {
				diff = append(diff, cell)
			}
		}
	}
}

func render(multiplier float64) {
	defer func(start time.Time) {
		draws++
		total_draw_time += time.Since(start)
	}(time.Now())

	defer func() {
		batch.Dirty()
		batch.Draw(window)
	}()

	multiplier = math.Max(0.0, math.Min(1.0, 1.0/(1.0+math.Exp(-12.0*multiplier+6.0))))
	// fmt.Printf("multiplier %f\n", multiplier)
	alive_color_changing := pixel.RGBA{
		R: alive_color.R * multiplier,
		G: alive_color.G * multiplier,
		B: alive_color.B * multiplier,
		A: 1.0,
	}
	dead_color_changing := pixel.RGBA{
		R: alive_color.R * (1.0 - multiplier),
		G: alive_color.G * (1.0 - multiplier),
		B: alive_color.B * (1.0 - multiplier),
		A: 1.0,
	}

	for _, cell := range diff {
		var color pixel.RGBA
		if grid[cell] {
			color = alive_color_changing
		} else {
			color = dead_color_changing
		}
		for v := 0; v < 6; v++ {
			triangles[cell*verts_per_cell+v].Color = color
		}
	}
}

func update_vertex_positions() {
	defer func(start time.Time) {
		resizes++
		total_resize_time += time.Since(start)
	}(time.Now())

	defer func() {
		batch.Dirty()
	}()

	row_height := win_height / float64(rows)
	col_width := win_width / float64(cols)
	for i := 0; i < rows; i++ {
		row_start := i * verts_per_row
		row_end := (i + 1) * verts_per_row
		triangles[row_start].Position = pixel.Vec{
			X: 0,
			Y: float64(i) * row_height,
		}
		triangles[row_start+1].Position = pixel.Vec{
			X: 0,
			Y: float64(i+1) * row_height,
		}
		triangles[row_start+3].Position = triangles[row_start+1].Position

		for j := 0; j < cols-1; j++ {
			lower_vertex := pixel.Vec{
				X: float64(j+1) * col_width,
				Y: float64(i) * row_height,
			}
			triangles[row_start+j*verts_per_cell+2].Position = lower_vertex
			triangles[row_start+j*verts_per_cell+4].Position = lower_vertex
			triangles[row_start+j*verts_per_cell+6].Position = lower_vertex

			upper_vertex := pixel.Vec{
				X: float64(j+1) * col_width,
				Y: float64(i+1) * row_height,
			}
			triangles[row_start+j*verts_per_cell+5].Position = upper_vertex
			triangles[row_start+j*verts_per_cell+7].Position = upper_vertex
			triangles[row_start+j*verts_per_cell+9].Position = upper_vertex
		}

		triangles[row_end-4].Position = pixel.Vec{
			X: win_width,
			Y: float64(i) * row_height,
		}
		triangles[row_end-2].Position = triangles[row_end-4].Position
		triangles[row_end-1].Position = pixel.Vec{
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
