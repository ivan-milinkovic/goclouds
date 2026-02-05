package main

type Matrix3D[T any] struct {
	values  []T
	W, H, D int
}

func NewMatrix3D[T any](w, h, d int) *Matrix3D[T] {
	return &Matrix3D[T]{
		values: make([]T, w*h*d),
		W:      w,
		H:      h,
		D:      d,
	}
}

func (dm *Matrix3D[T]) set(value T, x, y, z int) {
	ix := x % dm.W
	iy := y % dm.H
	iz := z % dm.D

	i := iy*dm.W*dm.D + ix*dm.D + iz

	dm.values[i] = value
}

func (dm *Matrix3D[T]) get(x, y, z int) T {
	ix := x % dm.W
	iy := y % dm.H
	iz := z % dm.D

	i := iy*dm.W*dm.D + ix*dm.D + iz
	return dm.values[i]
}
