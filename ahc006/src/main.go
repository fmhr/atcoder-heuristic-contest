package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof"
	"sort"
	"strconv"
	"time"
)

var sc = bufio.NewScanner(os.Stdin)
var buff []byte

func nextString() string {
	sc.Scan()
	return sc.Text()
}

func nextFloat64() float64 {
	sc.Scan()
	f, err := strconv.ParseFloat(sc.Text(), 64)
	if err != nil {
		panic(err)
	}
	return f
}

func nextInt() int {
	sc.Scan()
	i, err := strconv.Atoi(sc.Text())
	if err != nil {
		panic(err)
	}
	return i
}

var MAX = math.MaxInt64

func maxInt(a ...int) int {
	r := a[0]
	for i := 0; i < len(a); i++ {
		if r < a[i] {
			r = a[i]
		}
	}
	return r
}
func minInt(a ...int) int {
	r := a[0]
	for i := 0; i < len(a); i++ {
		if r > a[i] {
			r = a[i]
		}
	}
	return r
}
func sum(a []int) (r int) {
	for i := range a {
		r += a[i]
	}
	return r
}
func absInt(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

func init() {
	sc.Split(bufio.ScanWords)
	sc.Buffer(buff, bufio.MaxScanTokenSize*1024)
	log.SetFlags(log.Lshortfile)
}

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

func main() {
	log.SetFlags(log.Lshortfile)
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
	input()
	// ... rest of the program ...

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

func input() {
	orders := make([]Order, 0, 1000)
	for i := 0; i < 1000; i++ {
		var o Order
		o.num = i + 1
		o.u.x = nextInt()
		o.u.y = nextInt()
		o.v.x = nextInt()
		o.v.y = nextInt()
		orders = append(orders, o)
	}
	solver(orders)
}

var timeout bool
var TimeLimit time.Duration = 1900

func solver(orders []Order) {
	//startTime := time.Now()
	go func() {
		time.Sleep(TimeLimit * time.Millisecond)
		timeout = true
	}()
	selectedList := make([]int, 50)
	var usedList [1001]bool
	for i := 1; i <= 50; i++ {
		selectedList[i-1] = i
		usedList[i] = true
	}
	bestTime := 10000000
	var bestSelect []int
	//var bestUsedList [1001]bool
	nowTime := sumTime(greedyRoot(makeOrder(orders, selectedList)))
	var aIndex, b int
	loop := 0
	//logstring := ""
	for {
		loop++
		if timeout {
			break
		}
		// if i%500 == 0 && i > 0 && len(bestSelect) == 50 {
		// 	nowTime = bestTime
		// 	copy(selectedList, bestSelect)
		// 	usedList = bestUsedList
		// }
		aIndex = rand.Intn(50)
		a := selectedList[aIndex]
		b = a
		for b == a || usedList[b] {
			b = rand.Intn(1000)
		}
		selectedList = delete(selectedList, a)
		usedList[a] = false
		selectedList = append(selectedList, b)
		usedList[b] = true
		t := sumTime(greedyRoot(makeOrder(orders, selectedList)))
		if t < nowTime {
			nowTime = t
			if nowTime < bestTime {
				bestTime = nowTime
				bestSelect = selectedList
				//bestUsedList = usedList
			}
		} else {
			selectedList = delete(selectedList, b)
			usedList[b] = false
			selectedList = append(selectedList, a)
			usedList[a] = true
		}
	}
	output(makeOrder(orders, bestSelect))
	output2(greedyRoot(makeOrder(orders, bestSelect)))
	//log.Println(loop)
	//log.Println(logstring)
}

func makeOrder(orders []Order, selectedList []int) []Order {
	if len(selectedList) != 50 {
		panic("Noo")
	}
	os := make([]Order, 0, len(selectedList))
	for i := 0; i < len(selectedList); i++ {
		os = append(os, orders[selectedList[i]])
	}
	return os
}

func greedyRoot(os []Order) (root []Point) {
	root = make([]Point, 0, 102)
	root = append(root, Point{400, 400})
	nowPoint := Point{400, 400}
	for {
		var update bool
		var nextPoint Point
		var closePoint Point
		var closeTime int
		closeTime = 800 + 800
		var selectedStatus int
		var selectedOrder int
		var newStatus int
		for i := 0; i < 50; i++ {
			if os[i].status == noPick {
				nextPoint = os[i].u
				newStatus = picked
				update = true
			} else if os[i].status == picked {
				nextPoint = os[i].v
				newStatus = done
				update = true
			} else if os[i].status == done {
				continue
			}
			t := calTime(nowPoint, nextPoint)
			if closeTime > t {
				closeTime = t
				closePoint = nextPoint
				selectedStatus = newStatus
				selectedOrder = i
			}
		}
		if !update {
			break
		}
		nowPoint = closePoint
		root = append(root, closePoint)
		os[selectedOrder].status = selectedStatus
	}
	root = append(root, Point{400, 400})
	return
}

func sumTime(root []Point) int {
	var t int
	for i := 0; i < len(root)-1; i++ {
		t += calTime(root[i], root[i+1])
	}
	return t
}

func output(orders []Order) {
	fmt.Print(len(orders))
	for i := 0; i < len(orders); i++ {
		fmt.Print(" ", orders[i].num)
	}
	fmt.Println("")
}

func output2(root []Point) {
	fmt.Print(len(root))
	for i := 0; i < len(root); i++ {
		fmt.Print(" " + itoa(root[i].x) + " " + itoa(root[i].y))
	}
	fmt.Println("")
}

type Order struct {
	num    int
	u      Point
	v      Point
	status int
}

const (
	noPick = 0
	picked = 1
	done   = 2
)

type Point struct {
	x, y int
}

func calTime(a, b Point) int {
	return absInt(a.x-b.x) + absInt(a.y-b.y)
}

func itoa(i int) string {
	return strconv.Itoa(i)
}

func delete(s []int, v int) []int {
	sort.Ints(s)
	index := sort.Search(len(s), func(i int) bool { return s[i] >= v })
	s[index], s[len(s)-1] = s[len(s)-1], s[index]
	return s[:len(s)-1]
}
