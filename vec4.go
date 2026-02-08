package main

type Vec4 struct {
	X, Y, Z, W float64
}

func Vec4Fill(v float64) Vec4 {
	return Vec4{X: v, Y: v, Z: v, W: v}
}

func Vec4Make(v Vec3, w float64) Vec4 {
	return Vec4{X: v.X, Y: v.Y, Z: v.Z, W: w}
}

func (v Vec4) Scale(s float64) Vec4 {
	return Vec4{
		X: v.X * s,
		Y: v.Y * s,
		Z: v.Z * s,
		W: v.W * s,
	}
}

func (v Vec4) AddScalar(s float64) Vec4 {
	return Vec4{
		X: v.X + s,
		Y: v.Y + s,
		Z: v.Z + s,
		W: v.W + s,
	}
}

func (v Vec4) Add(v2 Vec4) Vec4 {
	return Vec4{
		X: v.X + v2.X,
		Y: v.Y + v2.Y,
		Z: v.Z + v2.Z,
		W: v.W + v2.W,
	}
}

func (v Vec4) Sub(v2 Vec4) Vec4 {
	return Vec4{
		X: v.X - v2.X,
		Y: v.Y - v2.Y,
		Z: v.Z - v2.Z,
		W: v.W - v2.W,
	}
}

func (v Vec4) Mul(v2 Vec4) Vec4 {
	return Vec4{
		X: v.X * v2.X,
		Y: v.Y * v2.Y,
		Z: v.Z * v2.Z,
		W: v.W * v2.W,
	}
}

func f4add(v1 [4]float64, v2 [4]float64) [4]float64 {
	return [4]float64{
		v1[0] + v2[0],
		v1[1] + v2[1],
		v1[2] + v2[2],
		v1[3] + v2[3],
	}
}
