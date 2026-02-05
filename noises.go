package main

import rl "github.com/gen2brain/raylib-go/raylib"

type Noises struct {
	cell_values   *DataMatrix[float64]
	perlin_values *Matrix3D[float64]
}

func NewNoises() *Noises {
	noise_img := rl.LoadImage("cells.png")
	noise_pixels := rl.LoadImageColors(noise_img)
	noise_values := NewDataMatrix[float64](int(noise_img.Width), int(noise_img.Height))
	for i := range noise_img.Width * noise_img.Height {
		noise_values.values[i] = float64(noise_pixels[i].R) / 255.0
	}

	return &Noises{
		cell_values: noise_values,
	}
}
