package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"
)

func readInputFromFile(n int) (in Input) {
	f, ok := os.Open(fmt.Sprintf("tools/in/%04d.txt", n))
	if ok != nil {
		log.Fatal("File not found")
	}
	defer f.Close()

	fmt.Fscan(f, &in.N, &in.M, &in.K)

	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			fmt.Fscan(f, &in.board[i][j])
		}
	}

	for i := 0; i < 20; i++ {
		for j := 0; j < 9; j++ {
			fmt.Fscan(f, &in.stamps[i][j])
		}
	}

	return
}

func TestGenerateLoop(t *testing.T) {
	log.Println("TestGenerateLoop")
	log.Println(generateLoop())
	log.Println(len(generateLoop()))
}

func TestBeamSearchStates(t *testing.T) {
	maxScore := 0
	bss := NewBeamSearccArray[State]()
	for i := 0; i < beamWidth*5; i++ {
		s := State{score: rand.Intn(100)}
		bss.Insert(s)
		maxScore = max(maxScore, s.score)
	}
	bss2 := bss.List()
	if bss2[0].score != maxScore {
		t.Errorf("List()[0].score = %d; want %d", bss2[0].score, maxScore)
	}
}

func TestGenereteAction(t *testing.T) {
	log.Println("TestGenereteAction")
	in := readInputFromFile(0)
	startTime := time.Now()
	generateAction(in)
	log.Println("Time:", time.Since(startTime).Nanoseconds())
}
