package main

import "math"

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
	if ix < 0 {
		ix = dm.W + ix
	}
	if iy < 0 {
		iy = dm.H + iy
	}
	if iz < 0 {
		iz = dm.D + iz
	}

	i := iy*dm.W*dm.D + ix*dm.D + iz
	return dm.values[i]
}

func (dm *Matrix3D[T]) getf(x, y, z float64) T {
	// get parts after decimal point
	x0 := math.Abs(math.Mod(x, 1))
	y0 := math.Abs(math.Mod(y, 1))
	z0 := math.Abs(math.Mod(z, 1))

	x1 := int(x0 * float64(dm.W))
	y1 := int(y0 * float64(dm.H))
	z1 := int(z0 * float64(dm.D))

	ix := x1 % dm.W
	iy := y1 % dm.H
	iz := z1 % dm.D

	i := iy*dm.W*dm.D + ix*dm.D + iz
	return dm.values[i]
}
