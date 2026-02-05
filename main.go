package main

// https://github.com/gen2brain/raylib-go/tree/master/examples

import (
	"fmt"
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Pixel = color.RGBA

type ImageTarget struct {
	Pixels []Pixel
	W      int
	H      int
}

func main() {
	screen_w := 640
	screen_h := 480
	rl.InitWindow(int32(screen_w), int32(screen_h), "goclouds") // must be at the top

	// prepare target image
	pixel_count := screen_w * screen_h
	image_target := ImageTarget{
		Pixels: make([]Pixel, pixel_count),
		W:      screen_w,
		H:      screen_h,
	}
	for i := range pixel_count {
		image_target.Pixels[i] = Pixel{R: 20, G: 20, B: 20, A: 255}
	}

	// prepare texture
	img_bytes := make([]byte, pixel_count*4) // used to copy to texture
	img := ImageFromRGBA(image_target.Pixels, &img_bytes, screen_w, screen_h)
	tex := rl.LoadTextureFromImage(img)

	// right-handed coordinate system
	camera := Camera{
		origin: Vec3{0, 0, 1},
		p00:    Vec3{0, 0, 0},
		aspect: float64(screen_w) / float64(screen_h),
	}

	// prepare perlin
	noise_img := rl.LoadImage("cells.png")
	noise_pixels := rl.LoadImageColors(noise_img)
	noise_values := NewDataMatrix[float64](int(noise_img.Width), int(noise_img.Height))
	for i := range noise_img.Width * noise_img.Height {
		noise_values.values[i] = float64(noise_pixels[i].R) / 255.0
	}

	// clear_color := rl.Black
	// clear_color := color.RGBA{30, 50, 70, 255}
	clear_color := color.RGBA{40, 40, 40, 255}

	// rl.SetTargetFPS(60)
	for !rl.WindowShouldClose() {

		t := rl.GetTime()
		ray_march(&image_target, &camera, noise_values, t)

		rl.BeginDrawing()
		rl.ClearBackground(clear_color)
		rl.UpdateTexture(tex, image_target.Pixels)
		rl.DrawTexture(tex, 0, 0, rl.White)
		rl.DrawText(fmt.Sprintf("%v fps, dt: %.0fms", rl.GetFPS(), rl.GetFrameTime()*1000), 10, 10, 16, rl.White)
		rl.EndDrawing()
	}

	// rl.UnloadImage(img) // crashes
}
