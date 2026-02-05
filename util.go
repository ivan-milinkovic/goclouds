package main

import rl "github.com/gen2brain/raylib-go/raylib"

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
	vf := min(f*255, 255)
	vf = max(vf, 0)
	vb := byte(vf)
	return vb
}
