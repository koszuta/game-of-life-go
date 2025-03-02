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

type MapLife struct {
	nRows int
	nCols int

	cells map[Vec2D]MapCell
	diff  []Vec2D

	vertices []ebiten.Vertex
}

type MapCell struct {
	created time.Time
}

type Vec2D struct {
	x, y int
}

func NewMapLife(conf Conf) *MapLife {
	ebiten.SetTPS(conf.TPS)

	w, h := conf.WindowWidth, conf.WindowHeight
	ebiten.SetWindowTitle("The Game of Life (Map)")
	ebiten.SetWindowSize(w, h)

	nRows, nCols := conf.NRows, conf.NCols
	l := &MapLife{
		nRows:    nRows,
		nCols:    nCols,
		cells:    make(map[Vec2D]MapCell),
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

func (l *MapLife) cellScale() (x, y float32) {
	w, h := ebiten.WindowSize()
	return float32(w) / float32(l.nCols), float32(h) / float32(l.nRows)
}

func (l *MapLife) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyPageUp) {
		tps := ebiten.TPS() + 1
		ebiten.SetTPS(min(tps, 60))
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyPageDown) {
		tps := ebiten.TPS() - 1
		ebiten.SetTPS(max(tps, 1))
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		l.Reset()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		debug = !debug
	}

	l.Step()
	return nil
}

func (l *MapLife) Draw(screen *ebiten.Image) {
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
			if nUpdates == 0 {
				nUpdates = 1
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
	for p := range l.cells {
		if len(vertices)+4 >= math.MaxUint16 {
			screen.DrawTriangles(vertices, indices, whiteSubImage, nil)
			vertices, indices = vertices[:0], indices[:0]
		}
		vertices = append(vertices,
			newVert(p.x, p.y, xScale, yScale, chartreuse),     // upper left
			newVert(p.x+1, p.y, xScale, yScale, chartreuse),   // upper right
			newVert(p.x, p.y+1, xScale, yScale, chartreuse),   // lower left
			newVert(p.x+1, p.y+1, xScale, yScale, chartreuse), // lower right
		)
		j := uint16(len(vertices) - 1)
		indices = append(indices, j-3, j-1, j-2, j-2, j-1, j)
	}
	screen.DrawTriangles(vertices, indices, whiteSubImage, nil)
}

func (l *MapLife) Layout(h, w int) (int, int) {
	return h, w
}

func (l *MapLife) Reset() {
	clear(updates)
	clear(draws)

	clear(l.cells)
	l.diff = l.diff[:0]
	for y := range l.nRows {
		for x := range l.nCols {
			if rand.IntN(3) != 0 {
				continue
			}
			p := Vec2D{x, y}
			l.cells[p] = MapCell{created: time.Now()}
			l.diff = append(l.diff, p)
		}
	}
}

func (l *MapLife) Step() {
	defer func(start time.Time) {
		updates[iUpdate] = time.Since(start)
		iUpdate = (iUpdate + 1) % n
	}(time.Now())

	next := make(map[Vec2D]MapCell, len(l.cells))
	defer func() {
		l.cells = next
	}()

	// TODO: Performance...
	// Can't use a map; way too slow
	// Try a 2D array + a list of alive cells?
	// How to track alive neighbors?
	// Linked list?
	l.diff = l.diff[:0]
	neighbors := make(map[Vec2D]int, 0)
	for p, c := range l.cells {
		var n int
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				if dx == 0 && dy == 0 {
					continue
				}
				p.x, p.y = p.x+dx, p.y+dy
				if p.x < 0 || p.y < 0 || p.x >= l.nCols || p.y >= l.nRows {
					continue
				}
				if _, ok := l.cells[p]; ok {
					n++
				} else {
					neighbors[p]++
				}
				p.x, p.y = p.x-dx, p.y-dy
			}
		}

		if n == 2 || n == 3 {
			next[p] = c // lives
		} else {
			l.diff = append(l.diff, p) // dies
		}
	}
	for p, n := range neighbors {
		if n == 3 {
			next[p] = MapCell{created: time.Now()} // reborn
			l.diff = append(l.diff, p)
		}
	}
}
