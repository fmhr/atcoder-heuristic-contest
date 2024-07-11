package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

var sc = bufio.NewScanner(os.Stdin)
var buff []byte

func init() {
	sc.Split(bufio.ScanWords)
	sc.Buffer(buff, bufio.MaxScanTokenSize*1024)
	log.SetFlags(log.Lshortfile)
}

// https://golang.org/pkg/runtime/pprof/
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	// ... rest of the program ...
	solver()

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		runtime.GC()    // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
}

var N int
var targetScore float64

func solver() {
	n, reqs := input()
	N = n
	targetScore = calcTargetScore(reqs)
	annealing(reqs)
	// ----
	// ads := make([]Ad, n)
	// minimam(&reqs, &ads)
	// s := score(reqs, ads)
	// log.Printf("score=%f", s)
	// output(&ads)
	// ----
}

func minimam(re *[]Req, ad *[]Ad) {
	for i := 0; i < N; i++ {
		(*ad)[i].a = (*re)[i].x
		(*ad)[i].b = (*re)[i].y
		(*ad)[i].dx = 1
		(*ad)[i].dy = 1
	}
}

type Point struct {
	x, y int
}

// Req is Request from input
type Req struct {
	x, y int
	r    int // 広告の面積
}

// Ad is Advertising space
type Ad struct {
	a, b   int // 左上 出力のa,bに相当
	dx, dy int // 辺の長さ
}

func (a Ad) size() int {
	return a.dx * a.dy
}

func (a Ad) x1() int {
	return a.a
}
func (a Ad) x2() int {
	return a.a + a.dx
}
func (a Ad) y1() int {
	return a.b
}
func (a Ad) y2() int {
	return a.b + a.dy
}

func (a Ad) rect() (r Rect) {
	r.x1 = a.x1()
	r.y1 = a.y1()
	r.x2 = a.x2()
	r.y2 = a.y2()
	return
}

type Rect struct {
	x1, x2, y1, y2 int
}

func intersect(r1, r2 Rect) bool {
	return minInt(r1.x2, r2.x2) > maxInt(r1.x1, r2.x1) && minInt(r1.y2, r2.y2) > maxInt(r1.y1, r2.y1)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// 広告を設置するスペースは10000 * 10000 の正方形
// 企業は点(xi+0.5, yi+0.5)を含む面積riの広告スペース
// 満足度
//	点(x,y)を含まない場合 pi = 0
//	点(x,y)を含む場合　pi = (1 - (1 - min(ri, si)/ max(ri, si))^2 si:広告面積

// 広告は面積のみ決められた長方形　辺の長さは任意に決められる

var W = 10000

var timeout bool
var TimeLimit time.Duration = 4500

func annealing(req []Req) {
	startTime := time.Now()
	log.Println(startTime)
	go func() {
		time.Sleep(TimeLimit * time.Millisecond)
		timeout = true
	}()
	var s State
	s.init(req)
	loop := 0
	for {
		loop++
		if timeout {
			break
		}
		var next info
		var now info
		var d int
		// --------------------------------------------------------
		// ４方向に成長させる
		i := randn(N)
		if s.ad[i].size >= s.ad[i].reqsize {
			goto END
		}
		d = randn(4)
		next = s.ad[i]
		now = s.ad[i]
		switch d {
		case 0: // -y方向に伸ばす
			next.y1--
			for x := next.x1; x < next.x2; x++ {
				if s.used[next.y1][x] {
					goto END
				}
			}
		case 1: // x方向に伸ばす
			next.x2++
			for y := next.y1; y < next.y2; y++ {
				if s.used[y][next.x2-1] {
					goto END
				}
			}
		case 2: // y方向に伸ばす
			next.y2++
			for x := next.x1; x < next.x2; x++ {
				if s.used[next.y2-1][x] {
					goto END
				}
			}
		case 3: // -x方向に伸ばす
			next.x1--
			for y := next.y1; y < next.y2; y++ {
				if s.used[y][next.x1] {
					goto END
				}
			}
		}
		next.calscore()
		if next.score > now.score {
			s.ad[i] = next
			s.score = s.score - now.score + next.score
			switch d {
			case 0:
				for x := next.x1; x < next.x2; x++ {
					s.used[next.y1][x] = true
				}
			case 1:
				for y := next.y1; y < next.y2; y++ {
					s.used[y][next.x2-1] = true
				}
			case 2:
				for x := next.x1; x < next.x2; x++ {
					s.used[next.y2-1][x] = true
				}
			case 3:
				for y := next.y1; y < next.y2; y++ {
					s.used[y][next.x1] = true
				}
				continue
			}
		}
		if s.score < 890000000 {
			continue
		}
		// --------------------------------------------------
		// ４方向に移動させる
		i = randn(N)
		d = randn(4)
		next = s.ad[i]
		now = s.ad[i]
		switch d {
		case 0: // y方向に動かす(下)
			for x := next.x1; x < next.x2; x++ {
				if s.used[next.y2+1][x] {
					goto END
				}
			}
			next.y1++
			next.y2++
		case 1: // x方向に動かす
			for y := next.y1; y < next.y2; y++ {
				if s.used[y][next.x2+1] {
					goto END
				}
			}
			next.x2++
			next.x1++
		case 2: // -y方向に動かす
			for x := next.x1; x < next.x2; x++ {
				if s.used[next.y2+1][x] {
					goto END
				}
			}
			next.y1++
			next.y2++
		case 3: // -x方向に動かす
			for y := next.y1; y < next.y2; y++ {
				if s.used[y][next.x1-1] {
					goto END
				}
			}
			next.x2--
			next.x1--
		}
		if next.fulfill() {
			s.ad[i] = next
			switch d {
			case 0:
				for x := next.x1; x < next.x2; x++ {
					s.used[next.y2][x] = true
					s.used[next.y1-1][x] = false
				}
			case 1:
				for y := next.y1; y < next.y2; y++ {
					s.used[y][next.x2] = true
					s.used[y][next.x1-1] = false
				}
			case 2:
				for x := next.x1; x < next.x2; x++ {
					s.used[next.y2][x] = true
					s.used[next.y1-1][x] = false
				}
			case 3:
				for y := next.y1; y < next.y2; y++ {
					s.used[y][next.x1] = true
					s.used[y][next.x2+1] = false
				}
			}
			continue
		}
		// ---------------------------------
		// 4方向のいずれかに縮める
		// if time.Since(startTime) > TimeLimit-100 {
		// 	continue
		// }
		i = randn(N)
		next = s.ad[i]
		now = s.ad[i]
		d = randn(4)
		switch d {
		case 0: // y方向に縮める
			if next.y2-next.y1 > 1 {
				next.y1++
			}
		case 1: // x方向に縮める
			if next.x2-next.x1 > 1 {
				next.x1++
			}
		case 2: // -y方向に縮める
			if next.y2-next.y1 > 1 {
				next.y2--
			}
		case 3: // -x方向に縮める
			if next.x2-next.x1 > 1 {
				next.x2--
			}
		}
		if next.fulfill() {
			next.calscore()
			if next.score < now.score {
				continue
			}
			s.ad[i] = next
			s.score = s.score - now.score + next.score
			switch d {
			case 0:
				for x := next.x1; x < next.x2; x++ {
					s.used[next.y1-1][x] = false
				}
			case 1:
				for y := next.y1; y < next.y2; y++ {
					s.used[y][next.x1-1] = false
				}
			case 2:
				for x := next.x1; x < next.x2; x++ {
					s.used[next.y2+1][x] = false
				}
			case 3:
				for y := next.y1; y < next.y2; y++ {
					s.used[y][next.x2+1] = false
				}
			}
		}
		continue
	END:
	}
	log.Printf("score=%v loop=%d", int(math.Round(s.score)), loop)
	s.output()
}

type info struct {
	x1, y1, x2, y2 int
	size           int
	reqsize        int
	rx, ry         int
	score          float64
}

func (i *info) calscore() {
	// if i.x1 >= i.x2 || i.y1 >= i.y2 {
	// 	i.score = -1
	// }
	i.size = (i.x2 - i.x1) * (i.y2 - i.y1)
	s := float64(minInt(i.size, i.reqsize)) / float64(maxInt(i.size, i.reqsize))
	tmp := 1.0 - (1.0-s)*(1.0-s)
	i.score = 1000000000 * tmp / float64(N)
}

func (i info) fulfill() (b bool) {
	b = i.rx >= i.x1 && i.rx < i.x2 && i.ry < i.y2 && i.ry >= i.y1
	return
}

type State struct {
	ad    []info
	used  [10003][10003]bool
	score float64
}

// 番兵を置きたいので、inputのx,y(0~99999)をx,y(1~100000)にする
func (s *State) init(req []Req) {
	s.ad = make([]info, len(req))
	for i := range req {
		if s.used[req[i].y][req[i].x] {
			log.Println("EEEEEEERRRRRRR")
		}
		req[i].x++
		req[i].y++
		s.ad[i].x1 = req[i].x
		s.ad[i].y1 = req[i].y
		s.ad[i].x2 = req[i].x + 1
		s.ad[i].y2 = req[i].y + 1
		s.ad[i].rx = req[i].x
		s.ad[i].ry = req[i].y
		s.ad[i].size = 1
		s.ad[i].reqsize = req[i].r
		s.ad[i].calscore()
		s.used[req[i].y][req[i].x] = true
		s.score += s.ad[i].score
	}
	for i := 0; i < 10003; i++ {
		s.used[i][10001] = true
		s.used[i][10002] = true
	}
	for i := 0; i < 10003; i++ {
		s.used[10001][i] = true
		s.used[10002][i] = true
	}
	for i := 0; i < 10003; i++ {
		s.used[i][0] = true
	}
	for i := 0; i < 10003; i++ {
		s.used[0][i] = true
	}

}

func (s State) output() {
	for _, ad := range s.ad {
		fmt.Printf("%d %d %d %d\n", ad.x1-1, ad.y1-1, ad.x2-1, ad.y2-1)
	}
}

func input() (int, []Req) {
	var n int
	fmt.Scanln(&n)
	rqs := make([]Req, n)
	for i := 0; i < n; i++ {
		fmt.Scan(&rqs[i].x, &rqs[i].y, &rqs[i].r)
		// rqs[i].x = nextInt() // 0 <= x,y <= 9999
		// rqs[i].y = nextInt()
		// rqs[i].r = nextInt() // 1 <= r <= 10000 * 10000
	}
	return n, rqs
}

func output(ads *[]Ad) {
	// 長方形の対角となる２頂点の座標(a, b),(c, d)を出力せよ
	for _, ad := range *ads {
		fmt.Println(ad.a, ad.b, ad.a+ad.dx, ad.b+ad.dy)
	}
}

// スコア確認用なので高速化は不要
func score(input []Req, out []Ad) (score float64) {
	for i := 0; i < N; i++ {
		if out[i].a < 0 || out[i].x2() > W || out[i].b < 0 || out[i].y2() > W {
			log.Printf("rectangle %d is out of range\n", i)
			return 0
		}
		if out[i].x1() >= out[i].x2() || out[i].y1() >= out[i].y2() {
			log.Printf("rectangle %d does not have positive area\n", i)
			return 0
		}
		if !(out[i].x1() <= input[i].x && input[i].x < out[i].x2() && out[i].y1() <= input[i].y && input[i].y < out[i].y2()) {
			log.Printf("rectangle %d does not contain point %v\n", i, input[i])
			continue
		}
		for j := 0; j < i; j++ {
			if intersect(out[i].rect(), out[j].rect()) {
				log.Printf("retangles %d and %d overlap\n", i, j)
				return 0
			}
		}
		s := float64(minInt(out[i].size(), input[i].r)) / float64(maxInt(out[i].size(), input[i].r))
		score += 1.0 - (1.0-s)*(1.0-s)
	}
	score = math.Round(1000000000 * score / float64(N))
	return
}

func calcTargetScore(input []Req) (score float64) {
	for i := 0; i < N; i++ {
		s := float64(input[i].r) / float64(input[i].r)
		score += 1.0 - (1.0-s)*(1.0-s)
	}
	score = math.Round(1000000000 * score / float64(N))
	return
}

var (
	x uint32 = 123456789
	y uint32 = 362436069
	z uint32 = 521288629
	w uint32 = 88675123
	t uint32
)

func xorshift() uint {
	t = x ^ (x << 11)
	x = y
	y = z
	z = w
	w = w ^ (w >> 19) ^ (t ^ (t >> 8))
	return uint(w)
}

func randn(n int) int {
	return int(xorshift() % uint(n))
}
