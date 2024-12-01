package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"sort"
	"time"
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
		//log.Println(i, w[i], h[i])
	}
	log.Printf("n=%d, t=%d, sgm=%d\n", n, t, sgm)
	return Input{n, t, sgm, w, h}
}

type CmdWithScore struct {
	cmd   Cmd
	score int
}

const beamWidth = 10

func BeamSearch(in Input) {
	states := make([]State, 0)
	states = append(states, NewState(in))
	subStates := make([]State, 0)
	for t := 0; t < in.N; t++ {
		for w := 0; w < min(len(states), beamWidth); w++ {
			state := states[w]
			states = states[:w]
			cmds := cmdGenerate(t)
			log.Println(state.pos[0].t)
			for _, cmd := range cmds {
				now := state.Clone()
				now.do(in, cmd, t)
				subStates = append(subStates, now)
			}
		}
		sort.Slice(subStates, func(i, j int) bool {
			return subStates[i].score > subStates[j].score
		})
		break
	}
}

// queryを使わずに解く
func simSolver(in Input) (int, []Cmd) {
	best_score := math.MaxInt64
	best_cmds := make([]Cmd, in.N)
	for k := 0; k < 10000; k++ {
		state := NewState(in)
		cmds := make([]Cmd, in.N)
		for i := 0; i < in.N; i++ {
			cmd := Cmd{p: i, r: false, d: 'U', b: -1}
			if rand.Intn(2) == 1 {
				cmd.r = true
			}
			if rand.Intn(2) == 1 {
				cmd.d = 'L'
			}
			if i > 0 {
				cmd.b = rand.Intn(i) - 1
			}
			state.do(in, cmd, i)
			cmds[i] = cmd
		}
		if state.score < best_score {
			best_score = state.score
			copy(best_cmds, cmds)
			log.Println(k, "best_score", best_score)
		}
	}
	log.Printf("best_score=%d\n", best_score)
	return best_score, best_cmds
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

	bs, bc := simSolver(in)
	if bs < bestScore {
		bestScore = bs
		bestAnses = bc
	}
	fmt.Println(in.N)
	for i := 0; i < in.N; i++ {
		//log.Println(bestAnses[i].String())
		fmt.Println(bestAnses[i].String())
	}
	state := NewState(in)
	state.query(in, bestAnses)
	log.Printf("score=%d\n", state.score)
	BeamSearch(in)
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

// n: 追加する長方形
func cmdGenerate(n int) []Cmd {
	cmds := make([]Cmd, 0)
	for r := 0; r < 2; r++ {
		for d := 0; d < 2; d++ {
			for b := -1; b < n; b++ {
				cmds = append(cmds, Cmd{p: n, r: r == 1, d: "UL"[d], b: b})
			}
		}
	}
	for i, cmd := range cmds {
		log.Println(i, ":", cmd.String())
	}
	return cmds
}

type Pos struct {
	x1, x2, y1, y2 int
	r              bool
	t              int
}

func (p *Pos) reset() {
	p.x1 = -1
	p.x2 = -1
	p.y1 = -1
	p.y2 = -1
	p.r = false
	p.t = -1
}

type State struct {
	turn           int
	pos            [100]Pos
	W, H           int
	W2, H2         int // 更新前 undo用
	score_t, score int
}

func (s State) Clone() State {
	t := s
	return t
}

func NewState(in Input) State {
	s := State{}
	s.turn = 0
	for i := 0; i < 100; i++ {
		s.pos[i].reset()
	}
	s.W = 0
	s.H = 0
	s.W2 = 0
	s.H2 = 0
	s.score_t = 0
	s.score = 0
	return s
}

func (s *State) undo(c Cmd) {
	s.pos[c.p].reset()
	s.W = s.W2
	s.H = s.H2
	s.score = s.W + s.H
}

func (s *State) do(in Input, c Cmd, t int) {
	// cmdのチェック
	if s.pos[c.p].t >= 0 {
		log.Println("c.p:", c.p, s.pos[c.p].t)
		log.Println("c:", c, s.pos[c.p].t)
		panic("already used")
	} else if c.b >= 0 && s.pos[c.b].t < 0 {
		log.Println(c.String())
		log.Printf("b=%d, t=%d\n", c.b, s.pos[c.b].t)
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
	s.W2 = s.W
	s.H2 = s.H
	s.W = max(s.W, s.pos[c.p].x2)
	s.H = max(s.H, s.pos[c.p].y2)
	s.score = s.W + s.H
}

func (s *State) query(in Input, cmd []Cmd) {
	for t, c := range cmd {
		s.do(in, c, t)
	}
}

func main() {
	log.SetFlags(log.Lshortfile)
	startTIme := time.Now()
	in := input()
	solver(in)
	elap := time.Since(startTIme)
	log.Printf("time_ms=%d ms\n", elap.Milliseconds())
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
