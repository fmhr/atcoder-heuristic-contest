package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
)

type Input struct {
	N, T, sgm int
	w, h      []int
}

func input() Input {
	var n, t, sgm int
	fmt.Scan(&n, &t, &sgm)
	w := make([]int, n)
	h := make([]int, n)
	for i := 0; i < n; i++ {
		fmt.Scan(&w[i], &h[i])
		log.Println(i, w[i], h[i])
	}
	return Input{n, t, sgm, w, h}
}

func solver(in Input) {
	var measured_w, measured_h int
	bestAnses := make([]Cmd, in.N)
	var bestScore int = math.MaxInt64
	for t := 0; t < in.T-1; t++ {
		fmt.Println(in.N)
		anses := make([]Cmd, in.N)
		for i := 0; i < in.N; i++ {
			var cmd Cmd
			cmd.p = i // 長方形の番号
			if 1 == rand.Intn(2) {
				cmd.r = true
			}
			cmd.d = 'U' // U：下から上に配置 L:右から左に配置
			if rand.Intn(2) == 0 {
				cmd.d = 'L'
			}
			cmd.b = -1
			if i > 0 {
				cmd.b = rand.Intn(i) - 1
			}
			anses[i] = cmd
			fmt.Println(cmd.String())
		}
		fmt.Scan(&measured_w, &measured_h)
		if measured_w+measured_h < bestScore {
			bestScore = measured_w + measured_h
			bestAnses = anses
			log.Println(t, measured_w+measured_h)
		}
	}
	log.Println("bestScore", bestScore)
	fmt.Println(in.N)
	for i := 0; i < in.N; i++ {
		log.Println(bestAnses[i].String())
		fmt.Println(bestAnses[i].String())
	}
	state := NewState(in)
	state.query(in, bestAnses)
}

type Cmd struct {
	p int  // 長方形の番号
	r bool // 1:90度回転
	d byte // U：下から上に配置 L:右から左に配置
	b int  // 基準となる長方形の番号
}

func (c Cmd) String() string {
	r := 0
	if c.r {
		r = 1
	}
	return fmt.Sprintf("%d %d %s %d", c.p, r, string(c.d), c.b)
}

type Pos struct {
	x1, x2, y1, y2 int
	r              bool
	t              int
}

type State struct {
	turn           int
	pos            []Pos
	W              int
	H              int
	W2             int
	H2             int
	score_t, score int
	comment        string
}

func NewState(in Input) State {
	s := State{}
	s.turn = 0
	s.pos = make([]Pos, in.N)
	for i := 0; i < in.N; i++ {
		s.pos[i] = Pos{-1, -1, -1, -1, false, -1}
	}
	s.W = 0
	s.H = 0
	s.W2 = 0
	s.H2 = 0
	s.score_t = 0
	s.score = 0
	s.comment = ""
	return s
}

func (s *State) query(in Input, cmd []Cmd) {
	for t, c := range cmd {
		// cmdのチェック
		if s.pos[c.p].t >= 0 {
			panic("already used")
		} else if c.b >= 0 && s.pos[c.b].t < 0 {
			panic("not used")
		}
		w, h := in.w[c.p], in.h[c.p]
		if c.r {
			w, h = h, w // 90度回転
		}
		if c.d == 'U' {
			x1 := 0 // 基準になるx座標
			if c.b >= 0 {
				x1 = s.pos[c.b].x2
			}
			x2 := x1 + w
			y1 := 0
			for _, q := range s.pos {
				if q.t >= 0 && max(x1, q.x1) < min(x2, q.x2) {
					y1 = max(y1, q.y2)
				}
			}
			y2 := y1 + h
			s.pos[c.p] = Pos{x1, x2, y1, y2, c.r, t}
		} else {
			y1 := 0 // 基準になるy座標
			if c.b >= 0 {
				y1 = s.pos[c.b].y2
			}
			y2 := y1 + h
			x1 := 0
			for _, q := range s.pos {
				if q.t >= 0 && max(y1, q.y1) < min(y2, q.y2) {
					x1 = max(x1, q.x2)
				}
			}
			x2 := x1 + w
			s.pos[c.p] = Pos{x1, x2, y1, y2, c.r, t}
		}
		s.W = max(s.W, s.pos[c.p].x2)
		s.H = max(s.H, s.pos[c.p].y2)
	}
	s.score = s.W + s.H
}

func main() {
	log.SetFlags(log.Lshortfile)
	in := input()
	solver(in)
}

// util
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
