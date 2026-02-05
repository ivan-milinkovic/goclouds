package main

import (
	"math"
	"runtime"
	"sync"

	"github.com/aquilax/go-perlin"
)

var perlin_gen = perlin.NewPerlin(0.1, 1.0, 2, 1234) // contrast, zoom, iterations (details), seed
var max_jumps = 40
var cloud_color = Vec3Fill(0.5)

func ray_march(img *ImageTarget, camera *Camera, perlin_values *DataMatrix[float64], time float64) {
	sphere := Sphere{
		C: Vec3{0, 0, -1},
		R: 1,
	}

	light := DirectionalLight{
		dir:   Vec3Make(-1.5, 1.5, 0.75).Normalized(),
		color: Vec3Fill(1.0),
	}

	// Test ray at the center
	// ray := Ray{
	// 	origin: Vec3Fill(0),
	// 	dir:    Vec3Make(0, 0, -1),
	// }
	// march(&ray, &sphere, &light)
	// return

	// Single-threaded
	// for y := range img.H {
	// 	for x := range img.W {
	// 		ray := camera.MakeRay(x, y, img.W, img.H)
	// 		colorf := march(&ray, &sphere, &light)

	// 		mod := Vec3Fill(math.Sin(rl.GetTime() * 4)).Add(Vec3Fill(1))
	// 		colorf = colorf.Mul(mod)
	// 		p := pixel_from_fcolor(colorf)
	// 		img.Pixels[y*img.W+x] = p
	// 	}
	// }
	// fmt.Printf("Done all, frame %v\n", frame_id)

	// Multi-goroutine
	var wg sync.WaitGroup
	y_mark := 0 // run a single goroutine with data starting from from this index
	// var dH = 10 // increment on the y axis for each goroutine
	dH := img.H / runtime.NumCPU()

	for y_mark < img.H {
		wg.Add(1)
		go func(y_mark int, img *ImageTarget) {
			end := min(y_mark+dH, img.H)
			for y := y_mark; y < end; y++ {
				for x := range img.W {
					ray := camera.MakeRay(x, y, img.W, img.H)
					// colorf := march_solid(&ray, &sphere, &light)
					colorf := march_volume(&ray, &sphere, &light, perlin_values, time)

					p := pixel_from_float4(colorf)
					img.Pixels[y*img.W+x] = p
				}
			}
			wg.Done()
		}(y_mark, img)
		y_mark += dH
	}
	wg.Wait()
}

type Sphere struct {
	C Vec3
	R float64
}

type DirectionalLight struct {
	dir   Vec3
	color Vec3
}

func march_solid(starting_ray *Ray, sphere *Sphere, light *DirectionalLight) Vec3 {
	ray := *starting_ray
	background := Vec3Fill(0)
	count := 0
	for {
		ray_origin_in_sphere_space := ray.origin.Sub(sphere.C)
		sdf := sdfSphere(ray_origin_in_sphere_space, sphere.R)
		if sdf < 0.02 {
			// v := math.Abs(sdf / (sdf + 1))
			// return Vec3{X: v, Y: v, Z: v}

			// when orientations are introduced, the normals will have to be transformed
			// as long as there are only translations, directions are OK in any translated space (not rotated or scaled)
			sphere_normal := ray_origin_in_sphere_space.Sub(sphere.C).Normalized()
			light_amount := sphere_normal.Dot((*light).dir)
			return Vec3Fill(light_amount)
		}

		// advance ray
		dv := ray.dir.Scale(sdf)
		ray.origin = ray.origin.Add(dv)

		if sdf >= 10 || count >= 10 {
			return background
		}
		count++
	}
}

func march_volume(starting_ray *Ray, sphere *Sphere, light *DirectionalLight, noise_values *DataMatrix[float64], time float64) [4]float64 {
	ray := *starting_ray

	jump_count := 0
	var acc_color [4]float64

	for jump_count < max_jumps {
		jump_count++
		found := march_outside_volume(&ray, sphere, &jump_count)
		if !found {
			break
		}

		acc_color_v := march_through_volume(&ray, sphere, light, noise_values, time)
		acc_color = f4add(acc_color, acc_color_v)
	}
	return acc_color
}

func march_outside_volume(ray *Ray, sphere *Sphere, jump_count *int) bool {
	prev_sdf := math.MaxFloat64
	for *jump_count < max_jumps {
		*jump_count++

		ray_origin_in_sphere_space := ray.origin.Sub(sphere.C)
		sdf := sdfSphere(ray_origin_in_sphere_space, sphere.R)

		if sdf <= 0 {
			return true // found a volume
		}
		if sdf > prev_sdf { // break out if moving away from all objects
			return false
		}

		prev_sdf = sdf
		// advance ray
		dv := ray.dir.Scale(sdf) // don't attempt to advance by zero
		ray.origin = ray.origin.Add(dv)
	}
	return false
}

func march_through_volume(ray *Ray, sphere *Sphere, light *DirectionalLight, noise_values *DataMatrix[float64], time float64) [4]float64 {
	acc_density := 0.0
	volume_acc_dist := 0.0   // accumulated distance inside a volume
	acc_color := Vec3Fill(0) // accumulated color

	// when orientations are introduced, the normals will have to be transformed
	// as long as there are only translations, directions are OK in any translated space (not rotated or scaled)

	// sample perlin
	if time > 1000000 {
		time = 0.0
	}

	for {
		point_in_sphere_space := ray.origin.Sub(sphere.C)
		sdf := sdfSphere(point_in_sphere_space, sphere.R)
		if sdf > 0 {
			break // went outside the volume
		}

		density := sample_density(ray.origin, noise_values, perlin_gen, time)
		acc_density += density
		absorbed := math.Exp(-volume_acc_dist * density) // Beer's law
		absorbed *= 0.25

		// shade

		// no light
		// acc_color = acc_color.AddScalar(1 - absorbed)
		// acc_color = acc_color.Add(cloud_color.Scale(0.5 * (1 - absorbed)))

		// with light

		sub_sphere_normal := point_in_sphere_space.Sub(sphere.C).Normalized()
		light_factor := sub_sphere_normal.Dot((*light).dir)
		light_amount := (absorbed) * light_factor * 0.2
		point_light_color := light.color.Scale(light_amount)
		// point_cloud_col := cloud_color.Scale(1.0 * (1 - absorbed))
		point_cloud_col := cloud_color
		point_col := point_cloud_col.Mul(point_light_color)
		acc_color = acc_color.Add(point_col)

		// advance ray inside volume
		ds := sphere.R / 16.0
		dv := ray.dir.Scale(ds)
		ray.origin = ray.origin.Add(dv)
		volume_acc_dist += ds
	}
	return [4]float64{acc_color.X, acc_color.Y, acc_color.Z, 1.0}
}

func sample_density(point Vec3, noise_values *DataMatrix[float64], perlin *perlin.Perlin, time float64) float64 {
	// noise_scale := 25.0
	// noise_phase := time * 10
	// noise_x := int(math.Abs(point.X*noise_scale + noise_phase*1))
	// noise_y := int(math.Abs(point.Y*noise_scale + noise_phase*0))
	// // noise_z := int(math.Abs(point.Z*noise_scale*2 + noise_phase*1))
	// noise1 := noise_values.get(noise_x, noise_y)
	// // noise2 := noise_values.get(noise_x, noise_z)
	// // noisef_0 := (noise1 + noise2) * 0.5
	// noisef_0 := noise1

	// perlin_scale_1 := 2.0
	// perlin_phase_1 := time * 1
	// perlin_1 := perlin_gen.Noise3D(
	// 	point.X*perlin_scale_1+perlin_phase_1*1,
	// 	point.Y*perlin_scale_1+perlin_phase_1*0,
	// 	point.Z*perlin_scale_1+perlin_phase_1*2,
	// )

	perlin_scale_2 := 8.0
	perlin_phase_2 := time * 1
	perlin_2 := perlin_gen.Noise3D(
		point.X*perlin_scale_2+perlin_phase_2*1,
		point.Y*perlin_scale_2+perlin_phase_2*0,
		math.Abs(point.Z)*perlin_scale_2+perlin_phase_2*1, // see Noise3D implementation, falls back to 3D if z < 0
	)
	// perlin_2 = 1 - perlin_2

	perlinf := perlin_2
	// perlinf := (perlin_1 + perlin_2) * 0.5

	// balance := 1.0
	// noisef := noisef_0*(1-balance) + perlinf*balance //+ noisef_1 + 0.1
	noisef := perlinf
	return noisef
}
