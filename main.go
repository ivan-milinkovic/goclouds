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
	rl.InitWindow(int32(screen_w), int32(screen_h), "goclouds")

	pixel_count := screen_w * screen_h
	image_target := ImageTarget{
		Pixels: make([]Pixel, pixel_count),
		W:      screen_w,
		H:      screen_h,
	}
	for i := range pixel_count {
		image_target.Pixels[i] = Pixel{R: 100, G: 100, B: 100, A: 255}
	}

	img_bytes := make([]byte, pixel_count*4) // used to copy to texture
	img := ImageFromRGBA(image_target.Pixels, &img_bytes, screen_w, screen_h)
	tex := rl.LoadTextureFromImage(img)

	// right-handed coordinate system
	camera := Camera{
		origin: Vec3{0, 0, 1},
		p00:    Vec3{0, 0, 0},
		aspect: float64(screen_w) / float64(screen_h),
	}

	// rl.SetTargetFPS(60)

	for !rl.WindowShouldClose() {

		ray_march(&image_target, &camera)

		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)
		rl.UpdateTexture(tex, image_target.Pixels)
		rl.DrawTexture(tex, 0, 0, rl.White)
		rl.DrawText(fmt.Sprintf("%v fps, dt: %.0fms", rl.GetFPS(), rl.GetFrameTime()*1000), 10, 10, 16, rl.White)
		rl.EndDrawing()
	}

	// rl.UnloadImage(img) // crashes
}
