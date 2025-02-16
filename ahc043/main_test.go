package main

import (
	"log"
	"math/rand"
	"testing"
)

func TestShortestPaht(t *testing.T) {
	f := NewField(50)
	for i := 0; i < 50; i++ {
		for j := 0; j < 50; j++ {
			f.cell[i][j] = EMPTY
			if rand.Intn(100) < 10 {
				f.cell[i][j] = STATION
			}
		}
	}
	a := Pos{Y: 0, X: 0}
	b := Pos{Y: 49, X: 0}
	f.cell[a.Y][a.X] = EMPTY
	f.cell[b.Y][b.X] = EMPTY
	path := f.shortestPath(a, b)
	//log.Println(path)
	if path == nil {
		log.Println("no path")
		return
	}
	for _, p := range path {
		f.cell[p.Y][p.X] = 7
	}
	for i := 0; i < 50; i++ {
		str := ""
		for j := 0; j < 50; j++ {
			str += railMap[f.cell[i][j]] + " "
		}
		//log.Printf("%02d %s\n", i, str)
	}
	rtn := f.selectRails(path)
	//log.Println(rtn)
	for i := 0; i < len(rtn); i++ {
		f.cell[path[i].Y][path[i].X] = rtn[i]
	}
	for i := 0; i < 50; i++ {
		str := ""
		for j := 0; j < 50; j++ {
			str += railMap[f.cell[i][j]] + " "
		}
		log.Printf("%02d %s\n", i, str)
	}
}

func TestCountSrcDst(t *testing.T) {

}
