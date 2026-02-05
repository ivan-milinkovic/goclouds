package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func ImageFromRGBA(pixels []Pixel, img_bytes *[]byte, w, h int) *rl.Image {
	for i, pixel := range pixels {
		(*img_bytes)[i*4+0] = pixel.R
		(*img_bytes)[i*4+1] = pixel.G
		(*img_bytes)[i*4+2] = pixel.B
		(*img_bytes)[i*4+3] = pixel.A
	}
	img := rl.NewImage(*img_bytes, int32(w), int32(h), 1, rl.UncompressedR8g8b8a8)
	return img
}

func pixel_from_fcolor(fcol Vec3) Pixel {
	p := Pixel{
		R: byte_color_value_from_float(fcol.X),
		G: byte_color_value_from_float(fcol.Y),
		B: byte_color_value_from_float(fcol.Z),
		A: 255,
	}
	return p
}

func pixel_from_float4(fcol [4]float64) Pixel {
	p := Pixel{
		R: byte_color_value_from_float(fcol[0]),
		G: byte_color_value_from_float(fcol[1]),
		B: byte_color_value_from_float(fcol[2]),
		A: byte_color_value_from_float(fcol[3]),
	}
	return p
}

func byte_color_value_from_float(f float64) byte {
	f_clamped := clamp01(f)
	vb := byte(f_clamped * 255)
	return vb
}

func clamp(val, minv, maxv float64) float64 {
	return max(min(val, maxv), minv)
}

func clamp01(val float64) float64 {
	return clamp(val, 0, 1)
}

func f4add(v1 [4]float64, v2 [4]float64) [4]float64 {
	return [4]float64{
		v1[0] + v2[0],
		v1[1] + v2[1],
		v1[2] + v2[2],
		v1[3] + v2[3],
	}
}

func asymptote_to_one_1(x float64) float64 {
	// desmos code: y\ =\ \frac{x}{\left(x+1\right)}
	return x / (x + 1)
}

func asymptote_to_one_2(x float64, compress float64) float64 {
	// desmos code
	// y=\left(\frac{1}{\left(1+e^{\left(-\left|x\cdot10\right|\right)}\right)}-0.5\right)\cdot2
	x = math.Abs(x)
	x *= compress
	sig := 1 / (1 + math.Exp(x)) // sigmoid
	sig -= 0.5
	sig *= 2
	return sig
}

func asymptote_to_one_3(x float64) float64 {
	// desmos code: y=\log\left(x+1\right)
	return math.Log(x + 1)
}

func beers_law(distance, absorption float64) float64 {
	// desmos code: y=\exp\left(-x\cdot d\right)
	return math.Exp(-distance * absorption)
}
