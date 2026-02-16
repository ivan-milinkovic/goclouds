package main

type Geometry struct {
	sphere *Sphere
	box    *Box
}

func (geo *Geometry) Sdf(point Vec3) float64 {
	if USE_BOX_GEOMETRY {
		return geo.box.Sdf(point)
	} else {
		return geo.sphere.Sdf(point)
	}
}

func (geo *Geometry) Size() float64 {
	if USE_BOX_GEOMETRY {
		return geo.box.Size()
	} else {
		return geo.sphere.Size()
	}
}

func (geo *Geometry) NormalAt(point Vec3) Vec3 {
	if USE_BOX_GEOMETRY {
		return geo.sphere.NormalAt(point) // todo
	} else {
		return geo.sphere.NormalAt(point)
	}
}
