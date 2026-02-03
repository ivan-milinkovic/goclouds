package main

import (
	"fmt"
	"math"
)

type Vec3 struct {
	X, Y, Z float64
}

func (v Vec3) String() string {
	return fmt.Sprintf("(%f, %f, %f)", v.X, v.Y, v.Z)
}

func (v Vec3) Len() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

func (v Vec3) LenSq() float64 {
	return v.X*v.X + v.Y*v.Y + v.Z*v.Z
}

func (v Vec3) Scale(s float64) Vec3 {
	return Vec3{
		X: v.X * s,
		Y: v.Y * s,
		Z: v.Z * s,
	}
}

func (v Vec3) Add(v2 Vec3) Vec3 {
	return Vec3{
		X: v.X + v2.X,
		Y: v.Y + v2.Y,
		Z: v.Z + v2.Z,
	}
}

func (v Vec3) Sub(v2 Vec3) Vec3 {
	return Vec3{
		X: v.X - v2.X,
		Y: v.Y - v2.Y,
		Z: v.Z - v2.Z,
	}
}

func (v *Vec3) Dot(v2 Vec3) float64 {
	return (*v).X*v2.X + (*v).Y*v2.Y + (*v).Z*v2.Z
}

func (v Vec3) Normalized() Vec3 {
	return v.Scale(1.0 / v.Len())
}

func (v *Vec3) Cross(v2 *Vec3) Vec3 {
	a := *v
	b := *v2
	return Vec3{
		X: a.Y*b.Z - a.Z*b.Y,
		Y: a.Z*b.X - a.X*b.Z,
		Z: a.X*b.Y - a.Y*b.X,
	}
}

// https://en.wikipedia.org/wiki/Rodrigues%27_rotation_formula
func VRotate(v *Vec3, axis *Vec3, rad float64) Vec3 {
	term1 := v.Scale(math.Cos(rad))

	v1 := axis.Cross(v)
	term2 := v1.Scale(math.Sin(rad))

	v2 := axis.Scale(axis.Dot(*v))
	term3 := v2.Scale(1 - math.Cos(rad))

	v3 := term1.Add(term2)
	return v3.Add(term3)
}

func Vec3Fill(v float64) Vec3 {
	return Vec3{X: v, Y: v, Z: v}
}

func Vec3Make(x, y, z float64) Vec3 {
	return Vec3{X: x, Y: y, Z: z}
}
