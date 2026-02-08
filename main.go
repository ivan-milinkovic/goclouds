package main

// https://github.com/gen2brain/raylib-go/tree/master/examples

import (
	"fmt"
	"image/color"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	rl.InitWindow(int32(WINDOW_WIDTH), int32(WINDOW_HEIGHT), "goclouds") // must be at the top

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

		// Update
		time := rl.GetTime()
		render_parameters.time = time
		if rl.IsKeyReleased(rl.KeyUp) {
			perlin_preview_z += 1
		} else if rl.IsKeyReleased(rl.KeyDown) {
			perlin_preview_z -= 1
		}
		if rl.IsKeyReleased(rl.KeyOne) {
			density_type = DensityType_PerlinRuntime
		} else if rl.IsKeyReleased(rl.KeyTwo) {
			density_type = DensityType_PerlinPreCalc
		} else if rl.IsKeyReleased(rl.KeyThree) {
			density_type = DensityType_Uniform
		}

		if ANIMATE_LIGHT_POSITION {
			state.light.origin.X = 2 * math.Sin(time*0.4)
			// state.light.origin = VRotate(&state.light.origin, &Vec3{0, 1, 0}, 0.1)
		}

		// Render
		if PREVIEW_PERLIN {
			write_perlin_to_image(state, perlin_preview_z)
		} else {
			ray_march(&render_parameters)
		}
		rl.UpdateTexture(*state.texture, state.image_target.Pixels)

		rl.BeginDrawing()
		rl.ClearBackground(clear_color)
		// rl.DrawTexture(tex, int32(screen_w/2-vol_vport_w/2), int32(screen_h/2-vol_vport_h/2), rl.White)
		rl.DrawTexturePro(
			*state.texture,
			rl.NewRectangle(0, 0, float32(state.image_target.W), float32(state.image_target.H)),
			rl.NewRectangle(0, 0, float32(WINDOW_WIDTH), float32(WINDOW_HEIGHT)),
			rl.NewVector2(0, 0), 0,
			rl.White,
		)
		rl.DrawText(fmt.Sprintf("%v fps, dt: %.0fms", rl.GetFPS(), rl.GetFrameTime()*1000), 10, 10, 16, rl.White)
		rl.DrawText(fmt.Sprintf("noise: 1/2/3 keys, current: %d", density_type), 10, WINDOW_HEIGHT-20, 16, rl.White)
		rl.EndDrawing()
	}

	// rl.UnloadImage(img) // crashes
}

func initialize() *State {
	// prepare target image
	pixel_count := VOL_VIEWPORT_W * VOL_VIEWPORT_H
	image_target := ImageTarget{
		Pixels: make([]Pixel, pixel_count),
		W:      VOL_VIEWPORT_W,
		H:      VOL_VIEWPORT_H,
	}
	for i := range pixel_count {
		image_target.Pixels[i] = Pixel{R: 20, G: 20, B: 20, A: 255}
	}

	// prepare texture
	img_bytes := make([]byte, pixel_count*4) // used to copy to texture
	img := ImageFromRGBA(image_target.Pixels, &img_bytes, VOL_VIEWPORT_W, VOL_VIEWPORT_H)
	tex := rl.LoadTextureFromImage(img)

	// left-handed coordinate system
	near_plane_d := 1.0
	camera_origin := Vec3{0, 0, 0}
	camera := Camera{
		origin: camera_origin,
		p00:    Vec3{0, 0, camera_origin.Z + near_plane_d},
		aspect: float64(VOL_VIEWPORT_W) / float64(VOL_VIEWPORT_H),
	}

	// prepare perlin
	noises := NewNoises()

	light := Light{
		origin: Vec3Make(-2.5, 1.5, 2),
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

func write_perlin_to_image(state *State, z int) {
	for y := range state.image_target.H {
		for x := range state.image_target.W {
			val := state.noises.perlin_values.get(x, y, z)
			px := pixel_from_fvec3(Vec3Fill(val))
			state.image_target.Pixels[y*state.image_target.W+x] = px
		}
	}
}
