package main

// https://github.com/gen2brain/raylib-go/tree/master/examples

import (
	"fmt"
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	rl.InitWindow(int32(window_width), int32(window_height), "goclouds") // must be at the top

	state := initialize()

	render_parameters := RenderParameters{
		img:    state.image_target,
		camera: state.camera,
		light:  state.light,
		sphere: state.sphere,
		noises: state.noises,
		time:   0.0,
	}

	// clear_color := rl.Black
	clear_color := color.RGBA{5, 10, 30, 255}
	perlin_preview_z := 10

	// rl.SetTargetFPS(60)
	for !rl.WindowShouldClose() {

		time := rl.GetTime()
		render_parameters.time = time
		if rl.IsKeyReleased(rl.KeyUp) {
			perlin_preview_z += 1
		} else if rl.IsKeyReleased(rl.KeyDown) {
			perlin_preview_z -= 1
		}
		// light.origin.X = 2 * math.Sin(time*0.5)
		// light.origin = VRotate(&light.origin, &Vec3{0, 1, 0}, 0.1)

		if PREVIEW_PERLIN {
			for y := range state.image_target.H {
				for x := range state.image_target.W {
					val := state.noises.perlin_values.get(x, y, perlin_preview_z)
					px := pixel_from_fvec3(Vec3Fill(val))
					state.image_target.Pixels[y*state.image_target.W+x] = px
				}
			}
		} else {
			ray_march(&render_parameters)
		}

		rl.BeginDrawing()
		rl.ClearBackground(clear_color)
		rl.UpdateTexture(*state.texture, state.image_target.Pixels)
		// rl.DrawTexture(tex, int32(screen_w/2-vol_vport_w/2), int32(screen_h/2-vol_vport_h/2), rl.White)
		rl.DrawTexturePro(
			*state.texture,
			rl.NewRectangle(0, 0, float32(state.image_target.W), float32(state.image_target.H)),
			rl.NewRectangle(0, 0, float32(window_width), float32(window_height)),
			rl.NewVector2(0, 0), 0,
			rl.White,
		)
		rl.DrawText(fmt.Sprintf("%v fps, dt: %.0fms", rl.GetFPS(), rl.GetFrameTime()*1000), 10, 10, 16, rl.White)
		rl.EndDrawing()
	}

	// rl.UnloadImage(img) // crashes
}

func initialize() *State {
	// prepare target image
	pixel_count := vol_viewport_w * vol_viewport_h
	image_target := ImageTarget{
		Pixels: make([]Pixel, pixel_count),
		W:      vol_viewport_w,
		H:      vol_viewport_h,
	}
	for i := range pixel_count {
		image_target.Pixels[i] = Pixel{R: 20, G: 20, B: 20, A: 255}
	}

	// prepare texture
	img_bytes := make([]byte, pixel_count*4) // used to copy to texture
	img := ImageFromRGBA(image_target.Pixels, &img_bytes, vol_viewport_w, vol_viewport_h)
	tex := rl.LoadTextureFromImage(img)

	// right-handed coordinate system
	near_plane_d := 1.0
	camera_origin := Vec3{0, 0, 0}
	camera := Camera{
		origin: camera_origin,
		p00:    Vec3{0, 0, camera_origin.Z + near_plane_d},
		aspect: float64(vol_viewport_w) / float64(vol_viewport_h),
	}

	// prepare perlin
	noises := NewNoises()

	light := Light{
		origin: Vec3Make(-4, 4, 0),
		dir:    Vec3Make(1, -0.25, 0).Normalized(),
		color:  Vec3Fill(1.0),
	}

	sphere := Sphere{
		C: Vec3{0, 0, 2},
		R: 1,
	}

	state := State{
		image_target: &image_target,
		camera:       &camera,
		light:        &light,
		sphere:       &sphere,
		noises:       noises,
		texture:      &tex,
	}
	return &state
}
