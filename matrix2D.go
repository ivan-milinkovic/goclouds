package main

type Matrix2D[T any] struct {
	values []T
	W, H   int
}

func NewMatrix2D[T any](w, h int) *Matrix2D[T] {
	var m Matrix2D[T]
	InitMatrix2D(&m, w, h)
	return &m
}

func InitMatrix2D[T any](m *Matrix2D[T], w, h int) {
	m.values = make([]T, w*h)
	m.W = w
	m.H = h
}

func (dm *Matrix2D[T]) set(value T, x int, y int) {
	dm.values[y*dm.W+x] = value
}

func (dm *Matrix2D[T]) get(x, y int) T {
	return dm.values[y*dm.W+x]
}

func (dm *Matrix2D[T]) setWrap(value T, x int, y int) {
	ix := x % dm.W
	iy := y % dm.H
	dm.values[iy*dm.W+ix] = value
}

func (dm *Matrix2D[T]) getWrap(x, y int) T {
	ix := x % dm.W
	iy := y % dm.H
	if ix < 0 {
		ix = dm.W - (-ix)
	}
	if iy < 0 {
		iy = dm.H - (-iy)
	}
	return dm.values[iy*dm.W+ix]
}
