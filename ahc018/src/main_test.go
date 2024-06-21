package main

import (
	"log"
	"testing"
)

func TestRouteSearch(t *testing.T) {
	a := Point{3, 10}
	b := Point{10, 30}
	if len(routeSearch(a, b)) != abs(b.y-a.y)+abs(b.x-a.x)+1 {
		t.Error("got =", routeSearch(a, b))
	}
	if len(routeSearch(b, a)) != abs(b.y-a.y)+abs(b.x-a.x)+1 {
		t.Error("got =", routeSearch(a, b))
	}
}

func TestAdjacentCell(t *testing.T) {
	a := Point{5, 0}
	r := adjacentCells(a)
	log.Println(r)
}

func TestTrialDig(t *testing.T) {
	trialDig()
}
