package main

type Sphere struct {
	C Vec3
	R float64
}

func (sphere *Sphere) Sdf(point Vec3) float64 {
	point_s := point.Sub(sphere.C) // point in sphere space
	return sdfSphere(point_s, sphere.R)
}

func (sphere *Sphere) Size() float64 {
	return sphere.R
}

func (sphere *Sphere) NormalAt(point Vec3) Vec3 {
	point_s := point.Sub(sphere.C) // point in sphere space
	return point_s.Sub(sphere.C).Normalized()
}
