package life

import (
	"fmt"
	"math"
	"math/rand/v2"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type OldLife struct {
	nRows int
	nCols int

	grid []bool
	diff []int

	vertices []ebiten.Vertex
}

func NewOldLife(conf Conf) *OldLife {
	ebiten.SetTPS(conf.TPS)

	w, h := conf.WindowWidth, conf.WindowHeight
	ebiten.SetWindowTitle("The Game of Life (Old)")
	ebiten.SetWindowSize(w, h)

	nRows, nCols := conf.NRows, conf.NCols
	l := &OldLife{
		nRows:    nRows,
		nCols:    nCols,
		grid:     make([]bool, nRows*nCols),
		vertices: make([]ebiten.Vertex, 0, (nRows+1)*(nCols+1)),
	}
	xScale, yScale := l.cellScale()
	for y := range nRows + 1 {
		for x := range nCols + 1 {
			l.vertices = append(l.vertices, newVert(x, y, xScale, yScale, chartreuse))
		}
	}
	l.Reset()
	return l
}

func (l *OldLife) cellScale() (x, y float32) {
	w, h := ebiten.WindowSize()
	return float32(w) / float32(l.nCols), float32(h) / float32(l.nRows)
}

func (l *OldLife) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		l.Reset()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyD) {
		debug = !debug
	}

	l.Step()

	return nil
}

func (l *OldLife) Draw(screen *ebiten.Image) {
	if debug {
		defer func() {
			var tDraws, nDraws time.Duration
			for _, d := range draws {
				if d != 0 {
					tDraws += d
					nDraws++
				}
			}
			var tUpdates, nUpdates time.Duration
			for _, u := range updates {
				if u != 0 {
					tUpdates += u
					nUpdates++
				}
			}
			avgUpdate := (tUpdates / nUpdates).Round(10 * time.Microsecond)
			avgDraw := (tDraws / nDraws).Round(10 * time.Microsecond)
			ebitenutil.DebugPrint(screen,
				fmt.Sprintf("TPS: %.1f\tFPS: %.1f\tUpdate took %v\tDraw took %v",
					ebiten.ActualTPS(), ebiten.ActualFPS(), avgUpdate, avgDraw))
		}()
	}
	defer func(start time.Time) {
		draws[iDraws] = time.Since(start)
		iDraws = (iDraws + 1) % n
	}(time.Now())

	screen.Fill(black)

	vertices := make([]ebiten.Vertex, 0, math.MaxUint16)
	indices := make([]uint16, 0)

	w, h := ebiten.WindowSize()
	xScale, yScale := float32(w)/float32(l.nCols), float32(h)/float32(l.nRows)
	for i, alive := range l.grid {
		if !alive {
			continue
		}
		if len(vertices)+4 >= math.MaxUint16 {
			screen.DrawTriangles(vertices, indices, whiteSubImage, nil)
			vertices, indices = vertices[:0], indices[:0]
		}
		x, y := l.deindex(i)
		vertices = append(vertices,
			newVert(x, y, xScale, yScale, chartreuse),     // upper left
			newVert(x+1, y, xScale, yScale, chartreuse),   // upper right
			newVert(x, y+1, xScale, yScale, chartreuse),   // lower left
			newVert(x+1, y+1, xScale, yScale, chartreuse), // lower right
		)
		j := uint16(len(vertices) - 1)
		indices = append(indices, j-3, j-1, j-2, j-2, j-1, j)
	}
	screen.DrawTriangles(vertices, indices, whiteSubImage, nil)
}

func (l *OldLife) deindex(i int) (x, y int) {
	return i % l.nCols, i / l.nCols
}

func (l *OldLife) Layout(h, w int) (int, int) {
	return h, w
}

func (l *OldLife) Reset() {
	clear(updates)
	clear(draws)

	l.diff = l.diff[:0]
	for cell := range l.nRows * l.nCols {
		l.grid[cell] = rand.IntN(3) == 0
		if l.grid[cell] {
			l.diff = append(l.diff, cell)
		}
	}
}

func (l *OldLife) Step() {
	defer func(start time.Time) {
		updates[iUpdate] = time.Since(start)
		iUpdate = (iUpdate + 1) % n
	}(time.Now())

	next := make([]bool, len(l.grid))
	defer func() {
		l.grid = next
	}()

	l.diff = l.diff[:0]
	for row := range l.nRows {
		for col := range l.nCols {
			rowUp := row + 1
			rowDown := row - 1
			colLeft := col - 1
			colRight := col + 1

			if row == 0 {
				rowDown = l.nRows - 1
			}
			if col == 0 {
				colLeft = l.nCols - 1
			}
			if row == l.nRows-1 {
				rowUp = 0
			}
			if col == l.nCols-1 {
				colRight = 0
			}

			rowUp = rowUp * l.nCols
			rowDown = rowDown * l.nCols
			rowSame := row * l.nCols

			var livingNeighbors int
			if l.grid[rowUp+col] {
				livingNeighbors++
			}
			if l.grid[rowUp+colRight] {
				livingNeighbors++
			}
			if l.grid[rowSame+colRight] {
				livingNeighbors++
			}
			if l.grid[rowDown+colRight] {
				livingNeighbors++
			}
			if l.grid[rowDown+col] {
				livingNeighbors++
			}
			if l.grid[rowDown+colLeft] {
				livingNeighbors++
			}
			if l.grid[rowSame+colLeft] {
				livingNeighbors++
			}
			if l.grid[rowUp+colLeft] {
				livingNeighbors++
			}

			cell := row*l.nCols + col
			alive := livingNeighbors == 3 || (l.grid[cell] && livingNeighbors == 2)
			next[cell] = alive
			if l.grid[cell] != alive {
				l.diff = append(l.diff, cell)
			}
		}
	}
}
