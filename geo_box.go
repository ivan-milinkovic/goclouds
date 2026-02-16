package main

type Box struct {
	whd    Vec3 // width, height, depth
	origin Vec3
}

func (box *Box) Sdf(point Vec3) float64 {
	point_b := point.Sub(box.origin) // point in box space
	return sdfBox(point_b, box.whd)
}

func (box *Box) Size() float64 {
	return max(max(box.whd.X, box.whd.Y), box.whd.Z)
}
