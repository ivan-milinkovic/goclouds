package main

import (
	"math"
	"runtime"
	"sync"

	"github.com/aquilax/go-perlin"
)

var perlin_gen = perlin.NewPerlin(1.0, 1.5, 2, 1234) // contrast, zoom, iterations (details), seed

var max_jumps = 40
var cloud_color = Vec3Fill(0.95)
var volume_resolution = 0.1

type ShadingType = int

const (
	ShadingType_NoLight         ShadingType = 0
	ShadingType_NaiveLight      ShadingType = 1
	ShadingType_RayMarchedLight ShadingType = 2
)

const shading_type = ShadingType_RayMarchedLight

type Sphere struct {
	C Vec3
	R float64
}

type Light struct { // point light
	origin Vec3
	dir    Vec3
	color  Vec3
}

func ray_march(img *ImageTarget, camera *Camera, light *Light, noises *Noises, time float64) {
	sphere := Sphere{
		C: Vec3{0, 0, -1},
		R: 1,
	}

	// Test ray at the center
	// ray := Ray{
	// 	origin: Vec3Fill(0),
	// 	dir:    Vec3Make(0, 0, -1),
	// }
	// march_volume(&ray, &sphere, &light, noises, time)
	// return

	// Test ray to the left of the sphere
	// ray := Ray{
	// 	origin: Vec3Fill(0),
	// 	dir:    Vec3Make(0, math.Atan(0.5), -1).Normalized(),
	// }
	// march_volume(&ray, &sphere, &light, noises, time)
	// return

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
					colorf := march_volume(&ray, &sphere, light, noises, time)

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

func march_solid(starting_ray *Ray, sphere *Sphere, light *Light) [4]float64 {
	ray := *starting_ray
	background := [4]float64{0, 0, 0, 0}
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
			dir_to_light := light.origin.Sub(ray.origin).Normalized()
			light_amount := sphere_normal.Dot(dir_to_light)
			return [4]float64{light_amount, light_amount, light_amount, 1.0}
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

func march_volume(starting_ray *Ray, sphere *Sphere, light *Light, noises *Noises, time float64) [4]float64 {
	ray := *starting_ray

	jump_count := 0
	var acc_color [4]float64

	for jump_count < max_jumps {
		jump_count++
		found := march_outside_volume(&ray, sphere, &jump_count)
		if !found {
			break
		}

		acc_color_v := march_through_volume(&ray, sphere, light, noises, time)
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

		// break out if moving away from near objects (won't work if there are both near and far objects)
		if sdf > prev_sdf {
			return false
		}

		dv := ray.dir.Scale(sdf) // advance ray; don't attempt to advance by zero
		ray.origin = ray.origin.Add(dv)
		prev_sdf = sdf
	}
	return false
}

func march_through_volume(ray *Ray, sphere *Sphere, light *Light, noises *Noises, time float64) [4]float64 {
	acc_density := 0.0
	acc_distance := 0.0      // accumulated distance inside the volume
	acc_color := Vec3Fill(0) // accumulated color
	acc_light_amount := 0.0

	// when orientations are introduced, the normals will have to be transformed
	// as long as there are only translations, directions are OK in any translated space (not rotated or scaled)

	for {
		point_in_sphere_space := ray.origin.Sub(sphere.C)
		sdf := sdfSphere(point_in_sphere_space, sphere.R)
		if sdf > 0 {
			break // went outside the volume
		}

		density := sample_density(ray.origin, noises, time)
		// density *= asymptote_to_one(math.Abs(sdf), 10.0) // make density closer to the surface softer
		acc_density += density

		// shade
		switch shading_type {
		case ShadingType_NoLight:
			acc_color = cloud_color

		case ShadingType_NaiveLight:
			sub_sphere_normal := point_in_sphere_space.Sub(sphere.C).Normalized()
			to_light_dir := (*light).dir.Scale(-1) // for directional light
			light_factor := sub_sphere_normal.Dot(to_light_dir)
			light_factor = max(0.2, light_factor)
			light_pass_through_amount := math.Exp(-acc_distance * acc_density) // Beer's law
			light_amount := light_pass_through_amount * light_factor
			point_light_color := light.color.Scale(light_amount)
			point_col := cloud_color.Mul(point_light_color)
			acc_color = acc_color.Add(point_col)

		// case ShadingType_RayMarchedLight:
		// 	distance_sampled_to_light, density_to_light := march_through_volume_to_light(ray.origin, sphere, light, noises, time)
		// 	light_amount := math.Exp(-distance_sampled_to_light * density_to_light) // Beer's law
		// 	light_color_at_point := light.color.Scale(light_amount)
		// 	point_color := cloud_color.Mul(light_color_at_point)
		// 	acc_color = acc_color.Add(point_color)

		case ShadingType_RayMarchedLight:
			distance_sampled_to_light, density_to_light := march_through_volume_to_light(ray.origin, sphere, light, noises, time)
			light_amount := math.Exp(-distance_sampled_to_light * density_to_light) // Beer's law
			acc_light_amount += light_amount
			acc_light_amount = asymptote_to_one_1(acc_light_amount)

		}

		// advance ray inside volume
		ds := volume_resolution
		dv := ray.dir.Scale(ds)
		ray.origin = ray.origin.Add(dv)
		acc_distance += ds
	}
	// diffuse := acc_color
	diffuse := cloud_color.Scale(acc_light_amount)
	// alpha := asymptote_to_one_3(acc_density)
	alpha := clamp01(0.8 * acc_density) // looks better
	return [4]float64{diffuse.X, diffuse.Y, diffuse.Z, alpha}
}

func march_through_volume_to_light(
	point Vec3,
	sphere *Sphere,
	light *Light,
	noises *Noises,
	time float64,
) (distance, density float64) {
	// directional light does not have an origin, just point towards where it's coming from (the oposite direction)
	dir_to_light := light.origin.Sub(point).Normalized()
	acc_distance := 0.0
	acc_density := 0.0
	for {
		point_in_sphere_space := point.Sub(sphere.C)
		sdf := sdfSphere(point_in_sphere_space, sphere.R)
		if sdf > 0 {
			acc_distance -= sdf // decrease by the over-shot distance outside the volume
			break               // went outside the volume
		}

		acc_density += sample_density(point, noises, time)

		// advance point towards light
		dv := dir_to_light.Scale(volume_resolution)
		point = point.Add(dv)
		acc_distance += volume_resolution
	}
	return acc_distance, acc_density
}

func sample_density(point Vec3, noises *Noises, time float64) float64 {
	// return 0.15

	noise_scale := 50.0
	noise_phase := time * 4
	noise_x := int(math.Abs(point.X*noise_scale + noise_phase*1))
	noise_y := int(math.Abs(point.Y*noise_scale + noise_phase*0))
	// noise_z := int(math.Abs(point.Z*noise_scale*2 + noise_phase*1))
	noise1 := noises.cell_values.get(noise_x, noise_y)
	// noise2 := noise_values.get(noise_y, noise_z)
	// noisef_0 := (noise1 + noise2) * 0.5
	noisef_0 := noise1

	perlin_scale_1 := 2.0
	perlin_phase_1 := time * 0.5
	perlin_1 := noises.perlin_gen.Noise3D(
		// perlin_1 := noises.perlin_values.getf(
		point.X*perlin_scale_1+perlin_phase_1*1,
		point.Y*perlin_scale_1+perlin_phase_1*0,
		point.Z*perlin_scale_1+perlin_phase_1*2,
	)
	perlin_1 = clamp01(perlin_1)

	perlin_scale_2 := 7.0
	perlin_phase_2 := time * 1
	perlin_2 := noises.perlin_gen.Noise3D(
		// perlin_2 := noises.perlin_values.getf(
		point.X*perlin_scale_2+perlin_phase_2*1,
		point.Y*perlin_scale_2+perlin_phase_2*0,
		math.Abs(point.Z)*perlin_scale_2+perlin_phase_2*1, // see Noise3D implementation, falls back to 3D if z < 0
	)
	perlin_2 = clamp01(perlin_2)

	// perlinf := perlin_2
	perlinf := (perlin_1 + perlin_2) * 0.5

	balance := 1.0                                   // cell texture is not good yet
	noisef := noisef_0*(1-balance) + perlinf*balance //+ noisef_1 + 0.1

	return noisef
}
