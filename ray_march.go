package main

import (
	"math"
	"runtime"
	"sync"

	"github.com/aquilax/go-perlin"
)

var perlin_gen = perlin.NewPerlin(2.0, 2.0, 1, 1234)

func ray_march(img *ImageTarget, camera *Camera, perlin_values *DataMatrix[float64], time float64) {
	sphere := Sphere{
		C: Vec3{0, 0, -2},
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

func march_volume(starting_ray *Ray, sphere *Sphere, light *DirectionalLight, noise_values *DataMatrix[float64], time float64) Vec3 {
	ray := *starting_ray
	cloud_color := Vec3Fill(0.5)
	acc_color := Vec3Fill(0) // accumulated color
	acc_density := 0.0
	volume_acc_dist := 0.0 // accumulated distance inside a volume
	max_jumps := 40
	jump_count := 0
	prev_sdf := math.MaxFloat64
	for jump_count < max_jumps {
		jump_count++

		ray_origin_in_sphere_space := ray.origin.Sub(sphere.C)
		sdf := sdfSphere(ray_origin_in_sphere_space, sphere.R)

		if sdf > 0 { // outside of any volume
			if sdf > prev_sdf { // break out early if moving away from all objects
				break
			}
			prev_sdf = sdf
			// advance ray outside of volumes
			dv := ray.dir.Scale(sdf) // don't attempt to advance by zero
			ray.origin = ray.origin.Add(dv)
			continue
		}

		// inside volume

		// when orientations are introduced, the normals will have to be transformed
		// as long as there are only translations, directions are OK in any translated space (not rotated or scaled)

		// sample perlin
		if time > 1000000 {
			time = 0.0
		}

		// noise_scale := 50.0
		// noise_phase := time * 10
		// noise_x := int(math.Abs(ray.origin.X*noise_scale + noise_phase*1))
		// noise_y := int(math.Abs(ray.origin.Y*noise_scale + noise_phase*0))
		// noise_z := int(math.Abs(ray.origin.Z*noise_scale*2 + noise_phase*1))
		// noise1 := noise_values.get(noise_x, noise_y)
		// noise2 := noise_values.get(noise_x, noise_z)
		// noisef := (noise1 + noise2) * 0.5

		noise_scale := 3.0
		noise_phase := time * 3
		noisef := perlin_gen.Noise3D(
			ray.origin.X*noise_scale+noise_phase*1,
			ray.origin.Y*noise_scale+noise_phase*0,
			ray.origin.Z*noise_scale+noise_phase*2,
		)

		// density := 0.025
		density := noisef
		acc_density += density
		absorbed := math.Exp(-volume_acc_dist * density)

		// shade

		// no light
		// acc_color = acc_color.AddScalar(1 - absorbed)
		// acc_color = acc_color.Add(cloud_color.Scale(0.5 * (1 - absorbed)))

		// with light
		sub_sphere_normal := ray_origin_in_sphere_space.Sub(sphere.C).Normalized()
		light_factor := sub_sphere_normal.Dot((*light).dir)
		light_amount := (1 - absorbed) * light_factor
		// light_amount := (absorbed) * light_factor
		point_light_color := light.color.Scale(light_amount)
		// point_cloud_col := cloud_color.Scale(1.0 * (1 - absorbed))
		point_cloud_col := cloud_color
		point_col := point_cloud_col.Mul(point_light_color)
		acc_color = acc_color.Add(point_col)

		// advance ray inside volume
		ds := sphere.R / 8.0
		dv := ray.dir.Scale(ds)
		ray.origin = ray.origin.Add(dv)
		volume_acc_dist += ds
	}
	return acc_color
}
