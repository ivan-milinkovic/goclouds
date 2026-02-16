package main

// https://iquilezles.org/articles/distfunctions/

func sdfSphere(p Vec3, r float64) float64 {
	return p.Len() - r
}

func sdfBox(p, b Vec3) float64 {
	q := p.Abs().Sub(b)
	return q.Max(0).Len() + min(max(q.X, max(q.Y, q.Z)), 0)
}
