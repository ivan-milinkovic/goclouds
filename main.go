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
	vol_vport_w := screen_w / 2 // viewport for volumetrics
	vol_vport_h := screen_h / 2
	pixel_count := vol_vport_w * vol_vport_h
	image_target := ImageTarget{
		Pixels: make([]Pixel, pixel_count),
		W:      vol_vport_w,
		H:      vol_vport_h,
	}
	for i := range pixel_count {
		image_target.Pixels[i] = Pixel{R: 20, G: 20, B: 20, A: 255}
	}

	// prepare texture
	img_bytes := make([]byte, pixel_count*4) // used to copy to texture
	img := ImageFromRGBA(image_target.Pixels, &img_bytes, vol_vport_w, vol_vport_h)
	tex := rl.LoadTextureFromImage(img)

	// right-handed coordinate system
	camera := Camera{
		origin: Vec3{0, 0, 1},
		p00:    Vec3{0, 0, 0},
		aspect: float64(vol_vport_w) / float64(vol_vport_h),
	}

	// prepare perlin
	noises := NewNoises()

	light := Light{
		origin: Vec3Make(-2, 2, 0),
		dir:    Vec3Make(1, -0.25, 0).Normalized(),
		color:  Vec3Fill(1.0),
	}

	// clear_color := rl.Black
	clear_color := color.RGBA{5, 10, 30, 255}
	// clear_color := color.RGBA{40, 40, 40, 255}

	// rl.SetTargetFPS(60)
	for !rl.WindowShouldClose() {

		time := rl.GetTime()
		// light.origin.X = 2 * math.Sin(time*0.5)
		// light.origin = VRotate(&light.origin, &Vec3{0, 1, 0}, 0.1)
		ray_march(&image_target, &camera, &light, noises, time)

		rl.BeginDrawing()
		rl.ClearBackground(clear_color)
		rl.UpdateTexture(tex, image_target.Pixels)
		// rl.DrawTexture(tex, int32(screen_w/2-vol_vport_w/2), int32(screen_h/2-vol_vport_h/2), rl.White)
		rl.DrawTexturePro(
			tex,
			rl.NewRectangle(0, 0, float32(vol_vport_w), float32(vol_vport_h)),
			rl.NewRectangle(0, 0, float32(screen_w), float32(screen_h)),
			rl.NewVector2(0, 0), 0,
			rl.White,
		)
		rl.DrawText(fmt.Sprintf("%v fps, dt: %.0fms", rl.GetFPS(), rl.GetFrameTime()*1000), 10, 10, 16, rl.White)
		rl.EndDrawing()
	}

	// rl.UnloadImage(img) // crashes
}
