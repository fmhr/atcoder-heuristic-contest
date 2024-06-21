package main

import (
	"fmt"
	"log"
	"strings"
)

type Position struct {
	x, y int
}

const (
	zero = iota
	Cow
	Pig
	Rabbit
	Dog
	Cat
)

var AnimalCount [6]int

type Pet struct {
	Position
	typ int
}

type Human struct {
	Position
	index int
}

var N int // ペットの数 10<=N<=20
var M int // 人間の数 5<=M<=10

func main() {
	log.SetFlags(log.Lshortfile)
	fmt.Scan(&N)
	pets := make([]Pet, 0, N)
	for i := 0; i < N; i++ {
		var p Pet
		fmt.Scan(&p.x, &p.y, &p.typ)
		pets = append(pets, p)
		AnimalCount[p.typ]++
	}
	fmt.Scan(&M)
	AnimalCount[0] = M
	humans := make([]Human, 0, M)
	for i := 0; i < M; i++ {
		var h Human
		fmt.Scan(&h.x, &h.y)
		h.index = i + 1
		humans = append(humans, h)
	}
	for t := 0; t < 300; t++ {
		ans := strings.Repeat(".", M)
		fmt.Println(ans)
	}
	log.Printf("human=%d cow=%d pig=%d rabbit=%d dog=%d cat=%d\n", AnimalCount[0], AnimalCount[1], AnimalCount[2], AnimalCount[3], AnimalCount[4], AnimalCount[5])
}
