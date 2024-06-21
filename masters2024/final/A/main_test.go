package main

import (
	"log"
	"testing"
)

func TestA(t *testing.T) {
	p := Point{1, 1}
	v := Vector{0, 0}
	target := Point{-1, -1}
	power := 100
	ay, ax := CalculateAcceleration(p, v, target, power)
	log.Println("A", int(ay), int(ax))
}
