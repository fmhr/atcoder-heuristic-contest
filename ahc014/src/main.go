package main

import (
	"fmt"
	"log"
	"math"
	"time"
)

func main() {
	log.SetFlags(log.Lshortfile)
	solver()
}

var N, M int32

func solver() {
	startTime := time.Now()
	n, m, points := read()
	N, M = n, m
	s := NewState(N, M, points)
	////////////////////////////////
	s.greedy()
	output(s.squares)
	timeDistance := time.Since(startTime)
	log.Printf("time=%d", timeDistance.Milliseconds())
	s.computeScore()
}

// スコア計算に必要
var inputDots []Point
var inputDotsMap map[Point]struct{}

func read() (N, M int32, points []Point) {
	fmt.Scan(&N, &M)
	points = make([]Point, M)
	for i := 0; i < int(M); i++ {
		fmt.Scan(&points[i].x, &points[i].y)
	}
	inputDots = make([]Point, len(points))
	copy(inputDots, points)
	inputDotsMap = make(map[Point]struct{})
	for _, p := range inputDots {
		inputDotsMap[p] = struct{}{}
	}
	return N, M, points
}

func output(ans []Square) {
	fmt.Println(len(ans))
	for i := 0; i < len(ans); i++ {
		fmt.Println(ans[i].String())
	}
}

var DXY = [8]Point{{1, 0}, {1, 1}, {0, 1}, {-1, 1}, {-1, 0}, {-1, -1}, {0, -1}, {1, -1}}
var rectangleDXY = [8][3]int{{0, 1, 2}, {4, 3, 2}, {4, 5, 6}, {0, 7, 6},
	{1, 2, 3}, {7, 0, 1}, {5, 6, 7}, {3, 4, 5}}

func dx(dir, i int) int32 {
	return DXY[rectangleDXY[dir][i]].x * diagonal(dir, i)
}

func dy(dir, i int) int32 {
	return DXY[rectangleDXY[dir][i]].y * diagonal(dir, i)
}

type Point struct {
	x, y int32
}

func (p Point) equal(q Point) bool {
	return p.x == q.x && p.y == q.y
}

func (p *Point) Add(q Point) {
	p.x += q.x
	p.y += q.y
}

func (p Point) String() string {
	return fmt.Sprint(p.x) + " " + fmt.Sprint(p.y)
}

type Square [4]Point

func (s Square) String() string {
	return s[0].String() + " " + s[1].String() + " " + s[2].String() + " " + s[3].String()
}

type State struct {
	hasPoint [61][61]bool
	used     [61][61][8]bool
	squares  []Square
}

func NewState(N, M int32, points []Point) (s State) {
	for i := 0; i < int(M); i++ {
		s.hasPoint[points[i].x][points[i].y] = true
	}
	return s
}

func (s State) HasPoint(p Point) bool {
	if outField(p.x) || outField(p.y) {
		return false
	}
	return s.hasPoint[p.x][p.y]
}

func (s State) checkMove(rect Square) error {
	if s.hasPoint[rect[0].x][rect[0].y] {
		return fmt.Errorf(rect[0].String(), " already contains a dot.")
	}
	for i := 1; i < 4; i++ {
		if !s.hasPoint[rect[i].x][rect[i].y] {
			return fmt.Errorf(rect[i].String(), " does not contain a dot.")
		}
	}
	dx01 := rect[1].x - rect[0].x
	dy01 := rect[1].y - rect[0].y
	dx03 := rect[3].x - rect[0].x
	dy03 := rect[3].y - rect[0].y
	if dx01*dx03+dy01*dy03 != 0 {
		// 傾いてない時、両方が０、傾いている時、x,yを逆にして等しい
		return fmt.Errorf("illegal rectangle")
	} else if dx01 != 0 && dy01 != 0 && absI8(dx01) != absI8(dy01) {
		// 斜めの時、dx,dyは同じ距離
		return fmt.Errorf("illegal rectangle")
	} else if !(rect[1].x+dx03 == rect[2].x && rect[1].y+dy03 == rect[2].y) {
		// P1+(dx,dy)==P2
		return fmt.Errorf("illegal rectangle")
	} else {
		// 点と点を繋いでいるラインが全て使われていないか調べる
		for i := 0; i < 4; i++ {
			xy := rect[i]
			txy := rect[(i+1)%4]
			dx := signum(txy.x - xy.x)
			dy := signum(txy.y - xy.y)
			dir := selectDir(dx, dy)
			for !xy.equal(txy) {
				if !xy.equal(rect[i]) && s.hasPoint[xy.x][xy.y] {
					return fmt.Errorf("there is an obstacle at %s", xy.String())
				}
				if s.used[xy.x][xy.y][dir] {
					return fmt.Errorf("overlapped rectangles")
				}
				xy.x += dx
				xy.y += dy
				if s.used[xy.x][xy.y][dir^4] {
					return fmt.Errorf("overlapped rectangles")
				}
			}
		}
	}
	return nil
}

// add new Point and Square
func (s *State) applyMove(sq Square) {
	s.hasPoint[sq[0].x][sq[0].y] = true
	for i := 0; i < 4; i++ {
		xy := sq[i]
		txy := sq[(i+1)%4]
		dx := signum(txy.x - xy.x)
		dy := signum(txy.y - xy.y)
		var dir int
		for i := 0; i < 8; i++ {
			if DXY[i].x == dx && DXY[i].y == dy {
				dir = i
				break
			}
		}
		for !xy.equal(txy) {
			s.used[xy.x][xy.y][dir] = true
			xy.x += dx
			xy.y += dy
			s.used[xy.x][xy.y][dir^4] = true
		}
		s.hasPoint[xy.x][xy.y] = true
	}
	s.squares = append(s.squares, sq)
}

func weight(p Point) int32 {
	var dx, dy int32
	dx = int32(p.x - N/2)
	dy = int32(p.y - N/2)
	return dx*dx + dy*dy + 1
}

func (s State) computeScore() {
	var num int32
	var x, y int32
	for x = 0; x < N; x++ {
		for y = 0; y < N; y++ {
			if s.hasPoint[x][y] {
				num += weight(Point{x, y})
			}
		}
	}
	var den int32
	for x = 0; x < N; x++ {
		for y = 0; y < N; y++ {
			den += weight(Point{x, y})
		}
	}
	score := 1e6 * float64(N*N) / float64(M) * (float64(num) / float64(den))
	log.Printf("score=%d num=%d", int(math.Round(score)), len(s.squares))
}

func (s *State) greedy() {
	var i, j, k int32
	for k = 0; k < 10; k++ {
		for i = 0; i < N; i++ {
			for j = 0; j < N; j++ {
				if !s.hasPoint[i][j] {
					if !s.findRectangle(Point{i, j}, k+1) {
						s.findRhombus(Point{i, j}, k+1)
					}
				}
			}
		}
		log.Println(k, len(s.squares), k+1)
	}
}

// unusedDot -> Rectangle
// 残りの3点が使われていないかを調べる
func (s *State) findRectangle(start Point, max int32) bool {
	for dir := 0; dir < 4; dir++ {
		var r Square
		r[0] = start
		// 2点を決める
		var i, l int32
		for i = 1; i < max; i++ {
			x1 := start.x + dx(dir, 0)*i
			y1 := start.y
			if s.HasPoint(Point{x1, y1}) {
				for l = 1; l < max; l++ {
					x3 := start.x
					y3 := start.y + dy(dir, 2)*l
					if i == 1 && l == 1 && max < 4 {
						if (start.x%2+start.y)%2 == 1 && dir%2 == 1 {
							continue
						}
					}
					if s.HasPoint(Point{x3, y3}) {
						x2 := x1
						y2 := y3
						if s.HasPoint(Point{x3, y3}) {
							r[1] = Point{x1, y1}
							r[2] = Point{x2, y2}
							r[3] = Point{x3, y3}
							err := s.checkMove(r)
							if err == nil {
								s.applyMove(r)
								return true
							}
						}
					}
				}
			}
		}
	}
	return false
}

func (s *State) findRhombus(start Point, max int32) bool {
	var r Square
	r[0] = start
	var a, b int32
	for d := 4; d < 8; d++ {
		for a = 1; a < max; a++ {
			x1 := start.x + dx(d, 0)*a
			y1 := start.y + dy(d, 0)*a
			if s.HasPoint(Point{x1, y1}) {
				for b = 1; b < max; b++ {
					x3 := start.x + dx(d, 2)*b
					y3 := start.y + dy(d, 2)*b
					if a == 1 && b == 1 && max < 4 {
						if start.x%2 == 1 && d%2 == 0 {
							continue
						}
					}
					if s.HasPoint(Point{x3, y3}) {
						x2 := start.x + dx(d, 0)*a + dx(d, 2)*b
						y2 := start.y + dy(d, 0)*a + dy(d, 2)*b
						if s.HasPoint(Point{x2, y2}) {
							r[1] = Point{x1, y1}
							r[2] = Point{x2, y2}
							r[3] = Point{x3, y3}
							err := s.checkMove(r)
							if err == nil {
								s.applyMove(r)
								return true
							}
						}
					}
				}
			}
		}
	}
	return false
}

func inField(x int32) bool {
	return x >= 0 && x < N
}

func outField(x int32) bool {
	return !inField(x)
}

func diagonal(r, i int) int32 {
	if r >= 4 && i == 1 {
		return 2
	}
	return 1
}

func selectDir(dx, dy int32) int {
	for i := 0; i < 8; i++ {
		if DXY[i].x == dx && DXY[i].y == dy {
			return i
		}
	}
	return -1
}

func signum(a int32) int32 {
	if a == 0 {
		return 0
	} else if a > 0 {
		return 1
	} else {
		return -1
	}
}

func absI8(a int32) int32 {
	if a > 0 {
		return a
	} else {
		return -a
	}
}

func minI32(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}
