package main

import (
	"flag"
	"fmt"
	"image/color"
	"math/rand"
	"strconv"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

var totalLifeTime, totalDrawTime time.Duration
var lifeCount, drawCount int64

var fps, turn_rate int
var frames_per_turn float64
var rows, cols int
var grid, buff []bool
var diff, all []uint32
var window *pixelgl.Window
var picture *pixel.PictureData
var width, height int

func main() {
	rand.Seed(time.Now().UnixNano())
	flag.IntVar(&fps, "fps", 60, "frames per second")
	flag.IntVar(&rows, "rows", 36, "number of rows")
	flag.IntVar(&cols, "cols", 64, "number of columns")
	flag.IntVar(&turn_rate, "rate", 12, "turn rate")
	flag.Parse()
	pixelgl.Run(run)
}

func run() {
	width, height := pixelgl.PrimaryMonitor().Size()
	win, err := pixelgl.NewWindow(pixelgl.WindowConfig{
		Title:     "Life (" + strconv.Itoa(fps) + " FPS)",
		Bounds:    pixel.R(0, 0, 1000, 1000),
		Resizable: true,
	})
	if err != nil {
		panic(err)
	}
	window = win
	window.SetPos(pixel.Vec{1, 31})
	picture = pixel.MakePictureData(window.Bounds())

	fmt.Printf("monitor=%dx%d\n", int(width), int(height))
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

	diff = make([]uint32, 0, rows*cols)
	all = make([]uint32, rows*cols)
	for i := range all {
		all[i] = uint32(i)
	}

	init_grid()

	frame := 0
	last := time.Now()
	for now := range time.Tick(time.Second / time.Duration(fps)) {
		if frame % fps == fps - 1 {
			curr_fps := float64(fps) * float64(time.Second) / float64(now.Sub(last))
			last = now
			window.SetTitle("Life (" + strconv.FormatFloat(curr_fps, 'f', 2, 64) + " FPS)")
		}

		if window.Closed() {
			break
		}
		if window.JustPressed(pixelgl.KeySpace) {
			init_grid()
		}

		window.Update()
		if frame%fps%turn_rate == 0 {
			// fmt.Printf("Turning\n")
			turn()
		}
		// fmt.Printf("Frame %d\n", frame)
		draw(frame % fps)

		frame++
	}
	// fmt.Printf("Avg life time %s\n", time.Duration(totalLifeTime.Nanoseconds()/lifeCount))
	// fmt.Printf("Avg draw time %s\n", time.Duration(totalDrawTime.Nanoseconds()/drawCount))
}

func init_grid() {
	diff = diff[:0]
	for i := range grid {
		// grid[i] = byte(rand.Intn(3)) % 2
		grid[i] = rand.Intn(3)%2 == 0
		if grid[i] {
			diff = append(diff, uint32(i))
		}
	}
}

func turn() {
	// defer func(start time.Time) {
	// 	lifeCount++
	// 	elapsed := time.Since(start)
	// 	totalLifeTime += elapsed
	// 	// fmt.Printf("life time %s\n", elapsed)
	// }(time.Now())

	diff = diff[:0]
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			var r_up, r_down, c_right, c_left int
			if r == 0 {
				r_down = rows - 1
			} else {
				r_down = r - 1
			}
			if c == 0 {
				c_left = cols - 1
			} else {
				c_left = c - 1
			}
			if r == rows-1 {
				r_up = 0
			} else {
				r_up = r + 1
			}
			if c == cols-1 {
				c_right = 0
			} else {
				c_right = c + 1
			}

			r_up = r_up * cols
			r_down = r_down * cols
			rr := r * cols

			var living_neighbors byte
			if grid[r_up+c] {
				living_neighbors++
			}
			if grid[r_up+c_right] {
				living_neighbors++
			}
			if grid[rr+c_right] {
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
			if grid[rr+c_left] {
				living_neighbors++
			}
			if grid[r_up+c_left] {
				living_neighbors++
			}

			i := rr + c
			// fmt.Printf("%d changed\n", i)
			if grid[i] {
				if living_neighbors < 2 {
					buff[i] = false
					diff = append(diff, uint32(i))
				} else if living_neighbors < 4 {
					buff[i] = true
					// if grid[i] != 7 {
					// 	buff[i] = grid[i] + 1
					// } else {
					// 	buff[i] = 7
					// }
				} else {
					buff[i] = false
					diff = append(diff, uint32(i))
				}
			} else {
				if living_neighbors == 3 {
					buff[i] = true
					diff = append(diff, uint32(i))
				} else {
					buff[i] = false
				}
			}
		}
	}
	grid, buff = buff, grid
}

func draw(frame int) {
	// defer func(start time.Time) {
	// 	drawCount++
	// 	elapsed := time.Since(start)
	// 	totalDrawTime += elapsed
	// 	// fmt.Printf("draw time %s\n", elapsed)
	// }(time.Now())

	cells := diff

	win_width := int(window.Bounds().Max.X)
	win_height := int(window.Bounds().Max.Y)
	if width != win_width || height != win_height {
		picture = pixel.MakePictureData(window.Bounds())
		width = win_width
		height = win_height
		cells = all
		fmt.Printf("changed to %dx%d\n", width, height)
	}

	col_w := width / cols
	row_h := height / rows
	// fmt.Printf("\nrow_h=%d col_w=%d\n", row_h, col_w)

	w_rem := width % cols
	h_rem := height % rows
	// fmt.Printf("h_rem=%d w_rem=%d\n", h_rem, w_rem)
	// fmt.Printf("%dpx\n", (int(width) - w_rem) * (int(height) - h_rem))

	for _, i := range cells {
		r := int(i) / cols
		c := int(i) % cols
		// fmt.Printf("diff[%d] = (%d, %d)\n", i, r, c)

		var r_start_extra, r_end_extra, c_start_extra, c_end_extra int
		if r < h_rem {
			r_start_extra = r
			r_end_extra = 1
		} else {
			r_start_extra = h_rem
			r_end_extra = 0
		}
		if c < w_rem {
			c_start_extra = c
			c_end_extra = 1
		} else {
			c_start_extra = w_rem
			c_end_extra = 0
		}
		r_start := r*row_h + r_start_extra
		r_end := r_start + row_h + r_end_extra
		c_start := c*col_w + c_start_extra
		c_end := c_start + col_w + c_end_extra
		// fmt.Printf("row(%d) %d-%d\t\t", r, r_start, r_end)
		// fmt.Printf("col(%d) %d-%d\n", c, c_start, c_end)

		var color color.RGBA
		if grid[i] {
			color = alive_color
		} else {
			color = dead_color
		}

		for y := r_start; y < r_end; y++ {
			for x := c_start; x < c_end; x++ {
				picture.Pix[y*picture.Stride+x] = color
			}
		}
	}
	sprite := pixel.NewSprite(picture, picture.Rect)
	sprite.Draw(window, pixel.IM.Moved(window.Bounds().Center()))
}

var alive_color color.RGBA = color.RGBA{
	R: 128,
	G: 255,
	B: 0,
	A: 255,
}

var dead_color color.RGBA = color.RGBA{
	R: 0,
	G: 0,
	B: 0,
	A: 255,
}
