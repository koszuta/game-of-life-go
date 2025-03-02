package main

import (
	"flag"
	"fmt"
	"math"
	"math/rand/v2"
	"strconv"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

const lerpMag = 13.37

var (
	alive = pixel.RGBA{R: 1.0, G: 1.0, B: 1.0, A: 1.0} // white
	// alive = pixel.RGBA{R: 0.5, G: 1.0, B: 0.0, A: 1.0} // chartreuse
	// alive = pixel.RGBA{R: 0.5, G: 0.5, B: 1.0, A: 1.0} // purple
	dead = pixel.RGBA{R: 0.0, G: 0.0, B: 0.0, A: 1.0} // black

	turns, draws, animations, resizes           int64
	turnTime, drawTime, animateTime, resizeTime time.Duration

	turnRate            int
	winWidth, winHeight float64

	nRows, nCols              int
	grid, buff                []bool
	diff                      []int
	vertsPerCell, vertsPerRow int
	triangles                 pixel.TrianglesData
	batch                     *pixel.Batch
)

func main() {
	flag.IntVar(&nRows, "rows", 100, "number of rows")
	flag.IntVar(&nCols, "cols", 100, "number of columns")
	flag.IntVar(&turnRate, "rate", 12, "turns per second")
	flag.Parse()
	pixelgl.Run(run)
}

func run() {
	monitorWidth, monitorHeight := pixelgl.PrimaryMonitor().Size()
	refreshRate := pixelgl.PrimaryMonitor().RefreshRate()

	window, err := pixelgl.NewWindow(pixelgl.WindowConfig{
		Title:     "Life (" + strconv.FormatFloat(refreshRate, 'f', 2, 64) + " FPS)",
		Bounds:    pixel.R(0, 0, 1000, 1000),
		Resizable: true,
		VSync:     true,
	})
	if err != nil {
		panic(err)
	}

	if window.VSync() && turnRate > int(refreshRate) {
		turnRate = int(refreshRate)
	}

	grid = make([]bool, nRows*nCols)
	buff = make([]bool, nRows*nCols)
	diff = make([]int, 0, nRows*nCols)
	initGrid()

	vertsPerCell = 2 * 3
	vertsPerRow = vertsPerCell * nCols
	triangles = make(pixel.TrianglesData, vertsPerRow*nRows)
	batch = pixel.NewBatch(&triangles, nil)

	fmt.Println()
	fmt.Printf("monitor:      %dx%dp\n", int(monitorWidth), int(monitorHeight))
	fmt.Printf("refresh rate: %d Hz\n", int(refreshRate))
	fmt.Printf("window:       %dx%dp\n", int(window.Bounds().W()), int(window.Bounds().H()))
	fmt.Printf("grid size:    %dx%d\n", nRows, nCols)
	fmt.Printf("turn rate:    %d\n", turnRate)
	fmt.Printf("polygons:     %d\n\n", len(triangles)/3)

	var (
		frame            int
		turnAccumulator  int64
		titleAccumulator int64
	)
	dtTurn := time.Second.Nanoseconds() / int64(turnRate)
	dtTitle := time.Second.Nanoseconds()
	lastFrame := frame
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

		if titleAccumulator >= dtTitle {
			currentFPS := float64(frame-lastFrame) / time.Duration(titleAccumulator).Seconds()
			window.SetTitle("Life (" + strconv.FormatFloat(currentFPS, 'f', 2, 64) + " FPS) ")
			lastFrame = frame
			titleAccumulator -= dtTitle
		}

		if turnAccumulator >= dtTurn {
			setCellColor()
			turn()
			turnAccumulator -= dtTurn
		}

		multiplier := float64(turnAccumulator) / float64(dtTurn)
		animate(multiplier)

		start := time.Now()
		batch.Draw(window)
		draws++
		drawTime += time.Since(start)

		window.Update()
		frame++
	}

	fmt.Println()
	fmt.Printf("%d frames drawn\n", frame)
	fmt.Printf("Avg resize time %s\n", time.Duration(resizeTime.Nanoseconds()/resizes))
	fmt.Printf("Avg turn time %s\n", time.Duration(turnTime.Nanoseconds()/turns))
	fmt.Printf("Avg animate time %s\n", time.Duration(animateTime.Nanoseconds()/animations))
	fmt.Printf("Avg draw time %s\n", time.Duration(drawTime.Nanoseconds()/draws))
}

func initGrid() {
	diff = diff[:0]
	for cell := 0; cell < nRows*nCols; cell++ {
		grid[cell] = rand.IntN(3) == 0
		if grid[cell] {
			diff = append(diff, cell)
		}
	}
}

func clearScreen() {
	for cell := 0; cell < nRows*nCols; cell++ {
		for vert := 0; vert < vertsPerCell; vert++ {
			triangles[cell*vertsPerCell+vert].Color = dead
		}
	}
}

func setCellColor() {
	for _, cell := range diff {
		var color pixel.RGBA
		if grid[cell] {
			color = alive
		} else {
			color = dead
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

	diff = diff[:0]
	defer func() {
		grid, buff = buff, grid
	}()

	for row := 0; row < nRows; row++ {
		for col := 0; col < nCols; col++ {
			rowUp := row + 1
			rowDown := row - 1
			colLeft := col - 1
			colRight := col + 1

			if row == 0 {
				rowDown = nRows - 1
			}
			if col == 0 {
				colLeft = nCols - 1
			}
			if row == nRows-1 {
				rowUp = 0
			}
			if col == nCols-1 {
				colRight = 0
			}

			rowUp = rowUp * nCols
			rowDown = rowDown * nCols
			rowSame := row * nCols

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

			cell := row*nCols + col
			buff[cell] = livingNeighbors == 3 || (grid[cell] && livingNeighbors == 2)
			if grid[cell] != buff[cell] {
				diff = append(diff, cell)
			}
		}
	}
}

func animate(m float64) {
	defer func(start time.Time) {
		animations++
		animateTime += time.Since(start)
	}(time.Now())

	defer func() {
		batch.Dirty()
	}()

	aliveFade := lerp(dead, alive, m)
	deadFade := lerp(alive, dead, m)

	for _, cell := range diff {
		var color pixel.RGBA
		if grid[cell] {
			color = aliveFade
		} else {
			color = deadFade
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

	rowHeight := winHeight / float64(nRows)
	colWidth := winWidth / float64(nCols)
	for i := 0; i < nRows; i++ {
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

		for j := 0; j < nCols-1; j++ {
			d := rowStart + j*vertsPerCell

			lowerVertex := pixel.Vec{
				X: float64(j+1) * colWidth,
				Y: float64(i) * rowHeight,
			}
			triangles[d+2].Position = lowerVertex
			triangles[d+4].Position = lowerVertex
			triangles[d+6].Position = lowerVertex

			upperVertex := pixel.Vec{
				X: float64(j+1) * colWidth,
				Y: float64(i+1) * rowHeight,
			}
			triangles[d+5].Position = upperVertex
			triangles[d+7].Position = upperVertex
			triangles[d+9].Position = upperVertex
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

func lerp(a, b pixel.RGBA, t float64) pixel.RGBA {
	// 1/1+e^-2*mag*t+mag clamped to [0,1]
	t = math.Max(0.0, math.Min(1.0, 1.0/(1.0+math.Exp(-2.0*lerpMag*t+lerpMag))))
	return pixel.RGBA{
		R: a.R + (b.R-a.R)*t,
		G: a.G + (b.G-a.G)*t,
		B: a.B + (b.B-a.B)*t,
		A: a.A + (b.A-a.A)*t,
	}
}
