package main

type DataMatrix[T any] struct {
	values []T
	W, H   int
}

func NewDataMatrix[T any](w, h int) *DataMatrix[T] {
	return &DataMatrix[T]{
		values: make([]T, w*h),
		W:      w,
		H:      h,
	}
}

func (dm *DataMatrix[T]) set(value T, x int, y int) {
	ix := x % dm.W
	iy := y % dm.H
	dm.values[iy*dm.W+ix] = value
}

func (dm *DataMatrix[T]) get(x, y int) T {
	ix := x % dm.W
	iy := y % dm.H
	return dm.values[iy*dm.W+ix]
}
