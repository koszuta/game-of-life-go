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
	grid  []bool
	diff  []int
}

func NewOldLife(conf Conf) *OldLife {
	ebiten.SetTPS(conf.TPS)

	w, h := conf.WindowWidth, conf.WindowHeight
	ebiten.SetWindowTitle("The Game of Life (Old)")
	ebiten.SetWindowSize(w, h)

	nRows, nCols := conf.NRows, conf.NCols
	l := &OldLife{
		nRows: nRows,
		nCols: nCols,
		grid:  make([]bool, nRows*nCols),
	}
	l.Reset()
	return l
}

func (life *OldLife) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		life.Reset()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		debug = !debug
	}

	life.Step()

	return nil
}

func (life *OldLife) Draw(screen *ebiten.Image) {
	var nDrawn time.Duration
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
			perCell := tDraws / nDraws / nDrawn
			ebitenutil.DebugPrint(screen,
				fmt.Sprintf("TPS: %.1f\tFPS: %.1f\tUpdate took %v\tDraw took %v\tPer cell took %v",
					ebiten.ActualTPS(), ebiten.ActualFPS(), avgUpdate, avgDraw, perCell))
		}()
	}
	defer func(start time.Time) {
		draws[iDraws] = time.Since(start)
		iDraws = (iDraws + 1) % n
	}(time.Now())

	screen.Fill(black)

	vertices := make([]ebiten.Vertex, 0, int(math.Pow(2, 10)))
	indices := make([]uint16, 0, cap(vertices)/4*6) // 6 indices per 4 vertices

	w, h := ebiten.WindowSize()
	xScale, yScale := float32(w)/float32(life.nCols), float32(h)/float32(life.nRows)
	for i, alive := range life.grid {
		if !alive {
			continue
		}
		nDrawn++

		x, y := life.deindex(i)
		vertices = append(vertices,
			newVert(x, y, xScale, yScale, chartreuse),     // upper left
			newVert(x+1, y, xScale, yScale, chartreuse),   // upper right
			newVert(x, y+1, xScale, yScale, chartreuse),   // lower left
			newVert(x+1, y+1, xScale, yScale, chartreuse), // lower right
		)
		j := uint16(len(vertices) - 1)
		indices = append(indices, j-3, j-1, j-2, j-2, j-1, j)

		if len(vertices)+4 >= cap(vertices) {
			screen.DrawTriangles(vertices, indices, whiteSubImage, nil)
			vertices, indices = vertices[:0], indices[:0]
		}
	}
	screen.DrawTriangles(vertices, indices, whiteSubImage, nil)
}

func (life *OldLife) deindex(i int) (x, y int) {
	return i % life.nCols, i / life.nCols
}

func (life *OldLife) Layout(w, h int) (int, int) {
	return w, h
}

func (life *OldLife) Reset() {
	clear(updates)
	clear(draws)

	life.diff = life.diff[:0]
	for i := range life.nRows * life.nCols {
		alive := rand.IntN(100) < n
		life.grid[i] = alive
		if alive {
			life.diff = append(life.diff, i)
		}
	}
}

func (life *OldLife) Step() {
	defer func(start time.Time) {
		updates[iUpdate] = time.Since(start)
		iUpdate = (iUpdate + 1) % n
	}(time.Now())

	l := len(life.grid)
	buff := make([]bool, l)
	defer func() {
		life.grid = buff
	}()

	life.diff = life.diff[:0]
	for i := range life.nRows * life.nCols {
		var neighbors int
		{ // upper left
			j := i - life.nCols - 1
			if j < 0 {
				j += l
			}
			if life.grid[j] {
				neighbors++
			}
		}
		{ // upper center
			j := i - life.nCols
			if j < 0 {
				j += l
			}
			if life.grid[j] {
				neighbors++
			}
		}
		{ // upper right
			j := i - life.nCols + 1
			if j < 0 {
				j += l
			}
			if life.grid[j] {
				neighbors++
			}
		}
		{ // left
			j := i - 1
			if j < 0 {
				j += l
			}
			if life.grid[j] {
				neighbors++
			}
		}
		{ // right
			j := i + 1
			if j >= l {
				j -= l
			}
			if life.grid[j] {
				neighbors++
			}
		}
		{ // lower left
			j := i + life.nCols - 1
			if j >= l {
				j -= l
			}
			if life.grid[j] {
				neighbors++
			}
		}
		{ // lower center
			j := i + life.nCols
			if j >= l {
				j -= l
			}
			if life.grid[j] {
				neighbors++
			}
		}
		{ // lower right
			j := i + life.nCols + 1
			if j >= l {
				j -= l
			}
			if life.grid[j] {
				neighbors++
			}
		}

		wasAlive := life.grid[i]
		nowAlive := neighbors == 3 || (wasAlive && neighbors == 2)

		buff[i] = nowAlive
		if wasAlive != nowAlive {
			life.diff = append(life.diff, i)
		}
	}
}
