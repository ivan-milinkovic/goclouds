package main

import (
	"math"
	"runtime"
	"sync"
)

func ray_march(img *ImageTarget, camera *Camera, perlin_values *DataMatrix[float64], time float64) {
	sphere := Sphere{
		C: Vec3{0, 0, -2},
		R: 1,
	}

	light := DirectionalLight{
		dir: Vec3Make(-1.5, 1.5, 0.75).Normalized(),
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

					// GetTime is extremelly slow duw to system calls
					// mod := Vec3Fill(math.Sin(rl.GetTime() * 4)).Add(Vec3Fill(1))
					// colorf = colorf.Mul(mod)
					p := pixel_from_fcolor(colorf)
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

func march_volume(starting_ray *Ray, sphere *Sphere, light *DirectionalLight, perlin_values *DataMatrix[float64], time float64) Vec3 {
	ray := *starting_ray
	acc_color := Vec3Fill(0) // accumulated color
	acc_density := 0.0
	max_jumps := 40
	jump_count := 0
	prev_sdf := math.MaxFloat64
	for jump_count < max_jumps {
		jump_count++

		ray_origin_in_sphere_space := ray.origin.Sub(sphere.C)
		sdf := sdfSphere(ray_origin_in_sphere_space, sphere.R)

		if sdf > prev_sdf { // break out early if moving away from all objects
			break
		}
		prev_sdf = sdf

		if sdf > 0 {
			// advance ray outside of volumes
			dv := ray.dir.Scale(sdf) // don't attempt to advance by zero
			ray.origin = ray.origin.Add(dv)
			continue
		}

		// inside volume

		// when orientations are introduced, the normals will have to be transformed
		// as long as there are only translations, directions are OK in any translated space (not rotated or scaled)

		// sample perlin
		// perlin_scale := 50.0 * math.Sin(time)
		perlin_scale := 40.0
		if time > 1000000 {
			time = 0.0
		}
		perlin_phase := time * 10
		perlin_x := int(math.Abs(ray.origin.X*perlin_scale + perlin_phase))
		perlin_y := int(math.Abs(ray.origin.Y*perlin_scale + perlin_phase))
		perlin_z := int(math.Abs(ray.origin.Z*perlin_scale + perlin_phase))
		perlin1 := perlin_values.get(perlin_x, perlin_y)
		perlin2 := perlin_values.get(perlin_y, perlin_z)
		perlin3 := perlin_values.get(perlin_x, perlin_z)
		perlin := (perlin1 + perlin2 + perlin3) * 0.33

		// density := 0.025
		density := perlin
		acc_density += density

		// shade

		// no light
		acc_color = acc_color.AddScalar(density)

		// with light
		// sub_sphere_normal := ray_origin_in_sphere_space.Sub(sphere.C).Normalized()
		// light_amount := sub_sphere_normal.Dot((*light).dir)
		// light_scale := acc_density // math.Abs(sdf)
		// acc_color = acc_color.AddScalar(light_scale * light_amount)

		// advance ray inside volume
		dv := ray.dir.Scale(0.025)
		ray.origin = ray.origin.Add(dv)
	}
	return acc_color
}
