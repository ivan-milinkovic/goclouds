package main

type Geometry struct {
	sphere *Sphere
	box    *Box
}

type Sphere struct {
	C Vec3
	R float64
}

type Box struct {
	whd    Vec3 // width, height, depth
	origin Vec3
}

func (sphere *Sphere) Sdf(point Vec3) float64 {
	point_s := point.Sub(sphere.C) // point in sphere space
	return sdfSphere(point_s, sphere.R)
}

func (sphere *Sphere) NormalAt(point Vec3) Vec3 {
	point_s := point.Sub(sphere.C) // point in sphere space
	return point_s.Sub(sphere.C).Normalized()
}

func (geo *Geometry) Sdf(point Vec3) float64 {
	return geo.sphere.Sdf(point)
}

func (geo *Geometry) Size() float64 {
	return geo.sphere.R
}

func (geo *Geometry) NormalAt(point Vec3) Vec3 {
	return geo.sphere.NormalAt(point)
}
