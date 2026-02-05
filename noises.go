package main

import (
	"github.com/aquilax/go-perlin"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Noises struct {
	cell_values   *DataMatrix[float64]
	perlin_values *Matrix3D[float64]
	perlin_gen    *perlin.Perlin
}

func NewNoises() *Noises {
	noise_img := rl.LoadImage("cells.png")
	noise_pixels := rl.LoadImageColors(noise_img)
	noise_values := NewDataMatrix[float64](int(noise_img.Width), int(noise_img.Height))
	for i := range noise_img.Width * noise_img.Height {
		noise_values.values[i] = float64(noise_pixels[i].R) / 255.0
	}

	var perlin_gen = perlin.NewPerlin(1.0, 1.5, 2, 1234) // contrast, zoom, iterations (details), seed

	dim := 64
	w, h, d := dim, dim, dim
	perlin_values := NewMatrix3D[float64](w, h, d)
	for y := range h {
		for x := range w {
			for z := range d {
				fx := float64(x) / float64(w)
				fy := float64(y) / float64(h)
				fz := float64(z) / float64(d)
				val := perlin_gen.Noise3D(fx, fy, fz)
				perlin_values.set(val, x, y, z)
			}
		}
	}

	return &Noises{
		cell_values:   noise_values,
		perlin_values: perlin_values,
		perlin_gen:    perlin_gen,
	}
}
