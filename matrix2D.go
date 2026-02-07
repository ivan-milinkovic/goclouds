package main

type Matrix2D[T any] struct {
	values []T
	W, H   int
}

func NewDataMatrix[T any](w, h int) *Matrix2D[T] {
	return &Matrix2D[T]{
		values: make([]T, w*h),
		W:      w,
		H:      h,
	}
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
		ix = dm.W - ix
	}
	if iy < 0 {
		iy = dm.H - iy
	}
	return dm.values[iy*dm.W+ix]
}
