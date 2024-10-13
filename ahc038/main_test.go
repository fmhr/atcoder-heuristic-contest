package main

import (
	"testing"
)

func TestRotate(t *testing.T) {
	p := Point{Y: 0, X: 1}
	center := Point{0, 0}

	// 時計回り
	p = p.Rotate(center, CW)
	if p.Y != 1 || p.X != 0 {
		t.Fatalf("rotete error, expected(1, 0), got(%d, %d)", p.Y, p.X)
	}
	// 反時計回り
	p = p.Rotate(center, CCW)
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
		now    int
		target int
		expend int
	}{
		{1, 1, None},
		{1, 4, CCW},
		{2, 3, CW},
		{1, 3, CW},
		{3, 0, CW},
		{2, 1, CCW},
		{2, 2, None},
		{0, 2, CW},
		{1, 3, CW},
		{3, 4, CW},
	}
	for i, tc := range testCases {
		result := chooseRotation(tc.now, tc.target)
		if result != tc.expend {
			t.Fatalf("case: %d ChooseRotation error, expected %d, got %d", i, tc.expend, result)
		}
	}
}

func TestRotateRobot(t *testing.T) {
	r := Node{
		Point: Point{Y: 0, X: 0},
	}
	n := Node{
		Point:     Point{Y: 0, X: 1},
		direction: 2,
		parent:    &r,
		length:    1,
	}
	r.children = append(r.children, &n)
	RotateRobot(CW, &n, r.Point)
	RotateRobot(CW, &n, r.Point)
	RotateRobot(CW, &n, r.Point)
	RotateRobot(CW, &n, r.Point)
	if n.direction != 2 {
		t.Fatalf("RotateRobot error, expected 2, got %d", n.direction)
	}
	RotateRobot(CCW, &n, r.Point)
	RotateRobot(CCW, &n, r.Point)
	RotateRobot(CCW, &n, r.Point)
	RotateRobot(CCW, &n, r.Point)
	if n.direction != 2 {
		t.Fatalf("RotateRobot error, expected 2, got %d", n.direction)
	}
}

func TestCalcRelatevePosition(t *testing.T) {
	var s State
	r := Node{
		Point: Point{Y: 0, X: 0},
	}
	n := Node{
		Point:     Point{Y: 0, X: 1},
		direction: 2,
		parent:    &r,
		length:    1,
	}
	s.nodes[0] = r
	s.nodes[1] = n
	r.children = append(r.children, &n)
	s.calcRelatevePosition()
}

func TestFindMthCombinatind(t *testing.T) {
	op := []int{1, 2, 3, 4}
	length := 3
	for i := 0; i < 20; i++ {
		result := findMthCombinatin(op, length, i)
		t.Logf("result: %v", result)
	}
}

func TestPathToRoot(t *testing.T) {
	var s State
	s.nodes[0].Point = Point{Y: 1, X: 1}
	s.nodes[0].index = 0
	// node0 -> node1
	s.nodes[1].Point = Point{Y: 1, X: 2}
	s.nodes[1].index = 1
	s.nodes[1].parent = &s.nodes[0]
	s.nodes[0].children = append(s.nodes[0].children, &s.nodes[1])
	// node1 -> node2
	s.nodes[2].Point = Point{Y: 2, X: 5}
	s.nodes[2].index = 2
	s.nodes[2].parent = &s.nodes[1]
	s.nodes[1].children = append(s.nodes[1].children, &s.nodes[2])
	// node0 -> node3
	s.nodes[3].Point = Point{Y: 3, X: 5}
	s.nodes[3].index = 3
	s.nodes[3].parent = &s.nodes[0]
	s.nodes[0].children = append(s.nodes[0].children, &s.nodes[3])

	path := pathToRoot(&s.nodes[2])
	expected := []int{2, 1, 0}
	for i, v := range path {
		if v.index != expected[i] {
			t.Fatalf("PathToRoot error, expected %v, got %v", expected, path)
		}
	}
}
