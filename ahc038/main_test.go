package main

import (
	"testing"
)

func TestRotate(t *testing.T) {
	p := Point{Y: 0, X: 1}
	center := Point{0, 0}

	// 時計回り
	p.Rotate(center, CW)
	if p.Y != 1 || p.X != 0 {
		t.Fatalf("rotete error, expected(1, 0), got(%d, %d)", p.Y, p.X)
	}
	// 反時計回り
	p.Rotate(center, CCW)
	if p.Y != 0 || p.X != 1 {
		t.Fatalf("rotete error, expected(0, 1), got(%d, %d)", p.Y, p.X)
	}
}

func TestStateMove(t *testing.T) {
	var s State
	s.nodes[0].Point = Point{Y: 1, X: 1}
	// node0 -> node1
	s.nodes[1].Point = Point{Y: 1, X: 2}
	s.nodes[1].parent = &s.nodes[0]
	s.nodes[0].children = append(s.nodes[0].children, &s.nodes[1])
	// node1 -> node2
	s.nodes[2].Point = Point{Y: 2, X: 5}
	s.nodes[2].parent = &s.nodes[1]
	s.nodes[1].children = append(s.nodes[1].children, &s.nodes[2])
	// node0 -> node3
	s.nodes[3].Point = Point{Y: 3, X: 5}
	s.nodes[3].parent = &s.nodes[0]
	s.nodes[0].children = append(s.nodes[0].children, &s.nodes[3])

	expected := [4]Point{
		{Y: 0, X: 2},
		{Y: 0, X: 3},
		{Y: 1, X: 6},
		{Y: 2, X: 6},
	}

	// move
	s.MoveRobot(Up, &s.nodes[0])
	s.MoveRobot(Right, &s.nodes[0])
	for i := 0; i < 4; i++ {
		if s.nodes[i].Y != expected[i].Y || s.nodes[i].X != expected[i].X {
			t.Fatalf("move error, expected(%d, %d), got(%d, %d)", expected[i].Y, expected[i].X, s.nodes[i].Y, s.nodes[i].X)
		}
	}
}

func TestChooseRotation(t *testing.T) {
	testCases := []struct {
		n      int
		x      int
		expend int
	}{
		{0, 1, 1},
		{1, 0, -1},
		{2, 3, 1},
		{0, 3, -1},
		{3, 0, 1},
		{1, 1, 0},
		{2, 2, 0},
		{0, 2, 2},
		{1, 3, 2},
	}
	for _, tc := range testCases {
		result := chooseRotation(tc.n, tc.x)
		if result != tc.expend {
			t.Fatalf("ChooseRotation error, expected %d, got %d", tc.expend, result)
		}
	}
}
