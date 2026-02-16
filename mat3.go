package main

type Mat3 struct {
	M [9]float64
}

func (mat *Mat3) set(val float64, x, y int) {
	mat.M[y*3+x] = val
}

func (mat *Mat3) get(x, y int) float64 {
	return mat.M[y*3+x]
}

func (mat *Mat3) Mult(v Vec3) Vec3 {
	i := Vec3FromSlice(mat.M[0:3])
	j := Vec3FromSlice(mat.M[3:6])
	k := Vec3FromSlice(mat.M[6:9])
	return Vec3{
		i.Dot(v),
		j.Dot(v),
		k.Dot(v),
	}
}
