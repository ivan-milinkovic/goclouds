package main

// https://iquilezles.org/articles/distfunctions/

func sdfSphere(p Vec3, r float64) float64 {
	return p.Len() - r
}
