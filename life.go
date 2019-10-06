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
	turns, draws, animations, resizes           int64
	turnTime, drawTime, animateTime, resizeTime time.Duration

	turnRate            int
	window              *pixelgl.Window
	winWidth, winHeight float64

	rows, cols                int
	grid, buff                []bool
	diff                      []int
	vertsPerCell, vertsPerRow int
	triangles                 pixel.TrianglesData
	batch                     *pixel.Batch

	aliveColor pixel.RGBA = chartreuse
	deadColor  pixel.RGBA = black
)

func main() {
	rand.Seed(time.Now().UnixNano())
	flag.IntVar(&rows, "rows", 100, "number of rows")
	flag.IntVar(&cols, "cols", 100, "number of columns")
	flag.IntVar(&turnRate, "rate", 12, "turns per second")
	flag.Parse()
	pixelgl.Run(run)
}

func run() {
	monitorWidth, monitorHeight := pixelgl.PrimaryMonitor().Size()
	refreshRate := pixelgl.PrimaryMonitor().RefreshRate()
	win, err := pixelgl.NewWindow(pixelgl.WindowConfig{
		Title:     "Life (" + strconv.FormatFloat(refreshRate, 'f', 2, 64) + " FPS)",
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
		Y: 31,
	})

	if window.VSync() && turnRate > int(refreshRate) {
		turnRate = int(refreshRate)
	}

	grid = make([]bool, rows*cols)
	buff = make([]bool, rows*cols)
	diff = make([]int, 0, rows*cols)
	initGrid()

	vertsPerCell = 2 * 3
	vertsPerRow = vertsPerCell * cols
	triangles = make(pixel.TrianglesData, vertsPerRow*rows)
	batch = pixel.NewBatch(&triangles, nil)

	fmt.Printf("\nmonitor:\t%dx%dpx\n", int(monitorWidth), int(monitorHeight))
	fmt.Printf("refresh rate:\t%dHz\n", int(refreshRate))
	fmt.Printf("window:\t\t%dx%dpx\n", int(window.Bounds().W()), int(window.Bounds().H()))
	fmt.Printf("grid size:\t%dx%d\n", rows, cols)
	fmt.Printf("turn rate:\t%d\n", turnRate)
	fmt.Printf("polygons: \t%d\n\n", len(triangles)/3)

	frame := 0
	lastFrame := frame
	var t int64 = 0
	var turnAccumulator int64 = 0
	var turnDt int64 = time.Second.Nanoseconds() / int64(turnRate)
	var titleAccumulator int64 = 0
	var titleDt int64 = time.Second.Nanoseconds()
	last := time.Now()
	for !window.Closed() {
		now := time.Now()
		frameTime := now.Sub(last).Nanoseconds()
		turnAccumulator += frameTime
		titleAccumulator += frameTime
		last = now

		if window.JustPressed(pixelgl.KeyEscape) {
			window.Destroy()
			break
		}

		if window.JustPressed(pixelgl.KeySpace) {
			clearScreen()
			initGrid()
			turnAccumulator = 0
		}

		winWidthNow := window.Bounds().W()
		winHeightNow := window.Bounds().H()
		if winWidth != winWidthNow || winHeight != winHeightNow {
			fmt.Printf("window changed to %dx%d\n", int(winWidthNow), int(winHeightNow))
			winWidth = winWidthNow
			winHeight = winHeightNow
			updateVertexPositions()
		}

		if titleAccumulator >= titleDt {
			currentFPS := float64(frame-lastFrame) / time.Duration(titleAccumulator).Seconds()
			window.SetTitle("Life (" + strconv.FormatFloat(currentFPS, 'f', 2, 64) + " FPS)")
			lastFrame = frame
			titleAccumulator -= titleDt
		}

		if turnAccumulator >= turnDt {
			setCellColor()
			turn()
			turnAccumulator -= turnDt
			t += turnDt
		}

		multiplier := float64(turnAccumulator) / float64(turnDt)
		animate(multiplier)
		render()
		window.Update()
		frame++
	}

	fmt.Printf("\n%d frames drawn\n", frame)
	fmt.Printf("Avg resize time %s\n", time.Duration(resizeTime.Nanoseconds()/resizes))
	fmt.Printf("Avg turn time %s\n", time.Duration(turnTime.Nanoseconds()/turns))
	fmt.Printf("Avg animate time %s\n", time.Duration(animateTime.Nanoseconds()/animations))
	fmt.Printf("Avg draw time %s\n", time.Duration(drawTime.Nanoseconds()/draws))
}

func initGrid() {
	diff = diff[:0]
	for cell := 0; cell < rows*cols; cell++ {
		grid[cell] = rand.Intn(3) == 0
		if grid[cell] {
			diff = append(diff, cell)
		}
	}
}

func clearScreen() {
	for cell := 0; cell < rows*cols; cell++ {
		for vert := 0; vert < vertsPerCell; vert++ {
			triangles[cell*vertsPerCell+vert].Color = deadColor
		}
	}
}

func setCellColor() {
	for _, cell := range diff {
		var color pixel.RGBA
		if grid[cell] {
			color = aliveColor
		} else {
			color = deadColor
		}
		for vert := 0; vert < vertsPerCell; vert++ {
			triangles[cell*vertsPerCell+vert].Color = color
		}
	}
}

func turn() {
	defer func(start time.Time) {
		turns++
		turnTime += time.Since(start)
	}(time.Now())

	defer func() {
		grid, buff = buff, grid
	}()

	diff = diff[:0]
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			rowUp := row + 1
			rowDown := row - 1
			colLeft := col - 1
			colRight := col + 1

			if row == 0 {
				rowDown = rows - 1
			}
			if col == 0 {
				colLeft = cols - 1
			}
			if row == rows-1 {
				rowUp = 0
			}
			if col == cols-1 {
				colRight = 0
			}

			rowUp = rowUp * cols
			rowDown = rowDown * cols
			rowSame := row * cols

			var livingNeighbors byte
			if grid[rowUp+col] {
				livingNeighbors++
			}
			if grid[rowUp+colRight] {
				livingNeighbors++
			}
			if grid[rowSame+colRight] {
				livingNeighbors++
			}
			if grid[rowDown+colRight] {
				livingNeighbors++
			}
			if grid[rowDown+col] {
				livingNeighbors++
			}
			if grid[rowDown+colLeft] {
				livingNeighbors++
			}
			if grid[rowSame+colLeft] {
				livingNeighbors++
			}
			if grid[rowUp+colLeft] {
				livingNeighbors++
			}

			cell := row*cols + col
			buff[cell] = livingNeighbors == 3 || (grid[cell] && livingNeighbors == 2)
			if grid[cell] != buff[cell] {
				diff = append(diff, cell)
			}
		}
	}
}

func render() {
	defer func(start time.Time) {
		draws++
		drawTime += time.Since(start)
	}(time.Now())

	batch.Draw(window)
}

func animate(multiplier float64) {
	defer func(start time.Time) {
		animations++
		animateTime += time.Since(start)
	}(time.Now())

	defer func() {
		batch.Dirty()
	}()

	multiplier = math.Max(0.0, math.Min(1.0, 1.0/(1.0+math.Exp(-12.0*multiplier+6.0))))
	aliveColorChanging := pixel.RGBA{
		R: aliveColor.R * multiplier,
		G: aliveColor.G * multiplier,
		B: aliveColor.B * multiplier,
		A: 1.0,
	}
	deadColorChanging := pixel.RGBA{
		R: aliveColor.R * (1.0 - multiplier),
		G: aliveColor.G * (1.0 - multiplier),
		B: aliveColor.B * (1.0 - multiplier),
		A: 1.0,
	}

	for _, cell := range diff {
		var color pixel.RGBA
		if grid[cell] {
			color = aliveColorChanging
		} else {
			color = deadColorChanging
		}
		for v := 0; v < 6; v++ {
			triangles[cell*vertsPerCell+v].Color = color
		}
	}
}

func updateVertexPositions() {
	defer func(start time.Time) {
		resizes++
		resizeTime += time.Since(start)
	}(time.Now())

	defer func() {
		batch.Dirty()
	}()

	rowHeight := winHeight / float64(rows)
	colWidth := winWidth / float64(cols)
	for i := 0; i < rows; i++ {
		rowStart := i * vertsPerRow
		rowEnd := (i + 1) * vertsPerRow
		triangles[rowStart].Position = pixel.Vec{
			X: 0,
			Y: float64(i) * rowHeight,
		}
		triangles[rowStart+1].Position = pixel.Vec{
			X: 0,
			Y: float64(i+1) * rowHeight,
		}
		triangles[rowStart+3].Position = triangles[rowStart+1].Position

		for j := 0; j < cols-1; j++ {
			lowerVertex := pixel.Vec{
				X: float64(j+1) * colWidth,
				Y: float64(i) * rowHeight,
			}
			triangles[rowStart+j*vertsPerCell+2].Position = lowerVertex
			triangles[rowStart+j*vertsPerCell+4].Position = lowerVertex
			triangles[rowStart+j*vertsPerCell+6].Position = lowerVertex

			upperVertex := pixel.Vec{
				X: float64(j+1) * colWidth,
				Y: float64(i+1) * rowHeight,
			}
			triangles[rowStart+j*vertsPerCell+5].Position = upperVertex
			triangles[rowStart+j*vertsPerCell+7].Position = upperVertex
			triangles[rowStart+j*vertsPerCell+9].Position = upperVertex
		}

		triangles[rowEnd-4].Position = pixel.Vec{
			X: winWidth,
			Y: float64(i) * rowHeight,
		}
		triangles[rowEnd-2].Position = triangles[rowEnd-4].Position
		triangles[rowEnd-1].Position = pixel.Vec{
			X: winWidth,
			Y: float64(i+1) * rowHeight,
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
