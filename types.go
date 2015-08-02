package main

type Direction int

const (
	DOWN Direction = iota
	UP
	LEFT
	RIGHT
	BACK
	FORWARD
	UNKNOWN
)

type Named interface {
	Name() string
}

type Position struct {
	x int
	y int
	z int
}
