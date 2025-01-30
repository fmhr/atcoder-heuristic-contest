package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime/pprof"
	"time"
)

type Node struct {
	id       uint64
	parent   *Node
	children []*Node
	h        *int64
}

func (n *Node) Root() *Node {
	if n.parent == nil {
		return n
	}
	return n.parent.Root()
}

func (n *Node) isIsorat() bool {
	return n.parent == nil && len(n.children) == 0
}

func (n *Node) Connect(m *Node, in Input) bool {
	if n.isIsorat() && m.isIsorat() {
		// どちらも孤立しているとき, nを初期化して,mをnにつなぐ
		//if in.A[n.id] < in.A[m.id] {
		//n, m = m, n
		//}
		var h int64 = 1
		n.h = &h
		m.parent = n
		m.h = n.h
		n.children = append(n.children, m)
	} else if !n.isIsorat() && !m.isIsorat() {
		// どちらもつながっているとき
		if n.Root() == m.Root() {
			// nとmが同じ木に属しているとき
			return false
		}
		nRoot, mRoot := n.Root(), m.Root()
		if *nRoot.h+*mRoot.h > in.H {
			// 高さがHを超えるとき
			return false
		}
		// mをnにつなぐ
		*nRoot.h += *mRoot.h
		m.parent = n
		n.children = append(n.children, m)
	} else {
		// 片方が孤立しているとき
		// mを孤立している方にする
		if n.isIsorat() {
			// nが孤立しているので入れ替える
			n, m = m, n
		}
		nRoot := n.Root()
		if *nRoot.h == in.H {
			return false
		}
		// mをnにつなぐ
		*nRoot.h++
		m.parent = n
		n.children = append(n.children, m)
	}
	return true
}

type Ans [1000]int

func solver(in Input) {
	//log.Printf("%+v", in)
	graphs := make([][]uint64, in.N)
	for i := 0; i < int(in.M); i++ {
		a, b := in.edges[i][0], in.edges[i][1]
		graphs[a] = append(graphs[a], b)
		graphs[b] = append(graphs[b], a)
	}

	for i := 0; i < int(in.N); i++ {
		log.Printf("%d %+v", i, graphs[i])
	}
	var nodes [1000]Node
	for i := 0; i < 1000; i++ {
		nodes[i].id = uint64(i)
	}
	for i := 0; i < int(in.N); i++ {
		for _, j := range graphs[i] {
			if nodes[i].Connect(&nodes[j], in) {
				log.Println("connect", i, j)
				break
			}
		}
	}

	for i := 0; i < int(in.N); i++ {
		if nodes[i].parent == nil {
			fmt.Println("-1 ")
		} else {
			fmt.Print(nodes[i].parent.id, " ")
		}
	}
	fmt.Println("")
	log.Printf("parent %v", nodes[0].parent.id)
}

type Input struct {
	N, M   uint64
	H      int64
	A      [1000]int64
	edges  [3000][2]uint64
	points [1000][2]int64
}

func input() (in Input) {
	fmt.Scan(&in.N, &in.M, &in.H)
	for i := 0; i < int(in.N); i++ {
		fmt.Scan(&in.A[i])
	}
	for i := 0; i < int(in.M); i++ {
		fmt.Scan(&in.edges[i][0], &in.edges[i][1])
	}
	for i := 0; i < int(in.N); i++ {
		fmt.Scan(&in.points[i][0], &in.points[i][1])
	}
	return
}

var ATCODER int
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")

func main() {
	log.SetFlags(log.Lshortfile)
	if os.Getenv("ATCODER") == "1" {
		ATCODER = 1
		log.SetOutput(io.Discard)
	}
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

	startTIme := time.Now()
	in := input()
	solver(in)
	log.Printf("elapsed: %v", time.Since(startTIme))
}
