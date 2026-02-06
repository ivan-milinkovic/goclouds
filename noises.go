package main

import (
	"github.com/aquilax/go-perlin"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Noises struct {
	tex_values    *DataMatrix[float64]
	perlin_values *Matrix3D[float64]
	perlin_gen    *perlin.Perlin
}

func NewNoises() *Noises {
	noise_img := rl.LoadImage("tex/cells.png")
	// noise_img := rl.LoadImage("tex/perlin 10 - 256x256.png")
	noise_pixels := rl.LoadImageColors(noise_img)
	noise_values := NewDataMatrix[float64](int(noise_img.Width), int(noise_img.Height))
	for i := range noise_img.Width * noise_img.Height {
		noise_values.values[i] = float64(noise_pixels[i].R) / 255.0
	}

	var perlin_gen = perlin.NewPerlin(0.2, 1.0, 1, 1234) // contrast, zoom, iterations (details), seed

	// does not tile, blocky appearance
	dim := 64
	w, h, d := dim, dim, dim
	perlin_values := NewMatrix3D[float64](w, h, d)
	for y := range h {
		for x := range w {
			for z := range d {
				fx := float64(x) / float64(w)
				fy := float64(y) / float64(h)
				fz := float64(z) / float64(d)
				val := clamp01(perlin_gen.Noise3D(fx, fy, fz))
				perlin_values.set(val, x, y, z)
			}
		}
	}

	return &Noises{
		tex_values:    noise_values,
		perlin_values: perlin_values,
		perlin_gen:    perlin_gen,
	}
}

/*
// https://www.shadertoy.com/view/WdXGRj?__cf_chl_tk=9luEUdOzm.bxXjKfn8FkQQsZhrXf1A1g1eark8FC7YU-1770320491-1.0.1.1-rpvaq3.HnaxLY0F0kDek8EIrPjd.LC4KhjupDAtPrnA

func Vec3Floor(v Vec3) Vec3 {
	return Vec3{
		math.Floor(v.X),
		math.Floor(v.Y),
		math.Floor(v.Z),
	}
}

func Vec3Fract(v Vec3) Vec3 {
	return Vec3{
		math.Mod(v.X, 1.0),
		math.Mod(v.Y, 1.0),
		math.Mod(v.Z, 1.0),
	}
}

func hash(n float64) float64 {
	return math.Mod(math.Sin(n)*43758.5453, 1)
}

func mix(a, b, f float64) float64 {
	return a*(1-f) + b*f
}

func noise(x Vec3) float64 {
	p := Vec3Floor(x)
	f := Vec3Fract(x)

	f = f.Mul(f).Mul(f.Scale(2).AddScalar(-3).Scale(-1))

	n := p.X + p.Y*57.0 + 113.0*p.Z

	res := mix(mix(mix(hash(n+0.0), hash(n+1.0), f.X),
		mix(hash(n+57.0), hash(n+58.0), f.X), f.Y),
		mix(mix(hash(n+113.0), hash(n+114.0), f.X),
			mix(hash(n+170.0), hash(n+171.0), f.X), f.Y), f.Z)
	return res
}

func fbm(p Vec3) float64 {
	var f float64
	f = 0.5000 * noise(p)
	// p = m * p * 2.02
	f += 0.2500 * noise(p)
	// p = m * p * 2.03
	f += 0.12500 * noise(p)
	// p = m * p * 2.01
	f += 0.06250 * noise(p)
	return f
}
*/
