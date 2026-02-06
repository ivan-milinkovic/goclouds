package main

import (
	"math"
	"runtime"
	"sync"
)

type RenderParameters struct {
	img    *ImageTarget
	camera *Camera
	light  *Light
	sphere *Sphere
	noises *Noises
	time   float64
}

type ShadingType = int

const (
	ShadingType_NoLight         ShadingType = 0
	ShadingType_NaiveLight      ShadingType = 1
	ShadingType_RayMarchedLight ShadingType = 2
)

type Sphere struct {
	C Vec3
	R float64
}

type Light struct { // point light
	origin Vec3
	dir    Vec3
	color  Vec3
}

var max_jumps = 40
var cloud_color = Vec3Fill(0.95)

const shading_type = ShadingType_RayMarchedLight
const scale_volume_res_per_object = true // scale ray advance step based on object size
const number_of_steps_for_object_scaling = 10
const volume_resolution = 0.1 // when not scaling
const ease_in_edges = false

func ray_march(render_params *RenderParameters) {

	// Test ray at the center
	// ray := Ray{
	// 	origin: Vec3Fill(0),
	// 	dir:    Vec3Make(0, 0, -1),
	// }
	// march_volume(&ray, &sphere, &light, noises, time)
	// return

	// Test ray to a tangent of the sphere
	// theta := math.Pi*0.5 - math.Asin(sphere.R/sphere.C.Sub(camera.origin).Len())
	// ray := Ray{
	// 	origin: Vec3Fill(0),
	// 	dir:    Vec3Make(0, math.Cos(theta), math.Sin(theta)).Normalized(),
	// }
	// march_volume(&ray, &sphere, light, noises, time)
	// return

	img := render_params.img
	camera := *render_params.camera

	// Multi-goroutine
	var wg sync.WaitGroup
	y_mark := 0 // run a single goroutine with data starting from from this index
	// var dH = 10 // increment on the y axis for each goroutine
	dH := img.H / runtime.NumCPU()

	for y_mark < img.H {
		wg.Add(1)
		go func(y_mark int, render_params *RenderParameters) {
			end := min(y_mark+dH, img.H)
			for y := y_mark; y < end; y++ {
				for x := range img.W {
					ray := camera.MakeRay(x, y, img.W, img.H)
					// colorf := march_solid(&ray, render_params)
					colorf := march_volume(&ray, render_params)

					p := pixel_from_fvec4(colorf)
					img.Pixels[y*img.W+x] = p
				}
			}
			wg.Done()
		}(y_mark, render_params)
		y_mark += dH
	}
	wg.Wait()
}

func march_solid(starting_ray *Ray, render_params *RenderParameters) Vec4 {
	ray := *starting_ray
	sphere := render_params.sphere
	light := render_params.light
	background := Vec4{0, 0, 0, 0}
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
			return Vec4{light_amount, light_amount, light_amount, 1.0}
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

func march_volume(starting_ray *Ray, render_params *RenderParameters) Vec4 {
	ray := *starting_ray

	jump_count := 0
	var acc_color Vec4

	for jump_count < max_jumps {
		jump_count++
		found := march_outside_volume(&ray, render_params, &jump_count)
		if !found {
			break
		}

		acc_color_v := march_through_volume(&ray, render_params)
		acc_color = acc_color.Add(acc_color_v)
	}
	return acc_color
}

func march_outside_volume(ray *Ray, render_params *RenderParameters, jump_count *int) bool {
	sphere := render_params.sphere
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

func march_through_volume(ray *Ray, render_params *RenderParameters) Vec4 {
	switch shading_type {
	case ShadingType_NoLight:
		return march_through_volume_no_light(ray, render_params)
	case ShadingType_NaiveLight:
		return march_through_volume_naive_light(ray, render_params)
	case ShadingType_RayMarchedLight:
		// return march_through_volume_raymarched_light_1(ray, render_params)
		return march_through_volume_raymarched_light_2(ray, render_params)
	}
	return Vec4{0.2, 0, 0.1, 0}
}

func march_through_volume_no_light(ray *Ray, render_params *RenderParameters) Vec4 {
	sphere := render_params.sphere

	acc_density := 0.0
	acc_distance := 0.0 // accumulated distance inside the volume
	count := 0.0

	var ds float64
	if scale_volume_res_per_object {
		ds = sphere.R / number_of_steps_for_object_scaling
	} else {
		ds = volume_resolution
	}

	// when orientations are introduced, the normals will have to be transformed
	// as long as there are only translations, directions are OK in any translated space (not rotated or scaled)

	for {
		point_in_sphere_space := ray.origin.Sub(sphere.C)
		sdf := sdfSphere(point_in_sphere_space, sphere.R)
		if sdf > 0 {
			break // went outside the volume
		}

		density := sample_density(ray.origin, render_params.noises, render_params.time) * volume_resolution

		// advance ray inside volume
		dv := ray.dir.Scale(ds)
		ray.origin = ray.origin.Add(dv)

		acc_density += density
		acc_distance += ds

		count += 1.0
		if count > float64(max_jumps) {
			break
		}
	}
	diffuse := cloud_color
	background_passthrough := beers_law(acc_distance, acc_density)
	alpha := 1 - background_passthrough
	return Vec4{diffuse.X, diffuse.Y, diffuse.Z, alpha}
}

func march_through_volume_naive_light(ray *Ray, render_params *RenderParameters) Vec4 {
	sphere := render_params.sphere
	light := render_params.light

	acc_density := 0.0
	acc_distance := 0.0      // accumulated distance inside the volume
	acc_color := Vec3Fill(0) // accumulated color
	count := 0.0

	var ds float64
	if scale_volume_res_per_object {
		ds = sphere.R / number_of_steps_for_object_scaling
	} else {
		ds = volume_resolution
	}

	// when orientations are introduced, the normals will have to be transformed
	// as long as there are only translations, directions are OK in any translated space (not rotated or scaled)

	for {
		point_in_sphere_space := ray.origin.Sub(sphere.C)
		sdf := sdfSphere(point_in_sphere_space, sphere.R)
		if sdf > 0 {
			break // went outside the volume
		}

		density := sample_density(ray.origin, render_params.noises, render_params.time) * volume_resolution
		// density *= asymptote_to_one(math.Abs(sdf), 10.0) // make density closer to the surface softer
		acc_density += density

		sub_sphere_normal := point_in_sphere_space.Sub(sphere.C).Normalized()
		dir_to_light := (*light).origin.Sub(ray.origin).Normalized()
		light_factor := sub_sphere_normal.Dot(dir_to_light)
		// light_factor = max(0.05, light_factor)
		point_light_color := light.color.Scale(light_factor)
		point_col := cloud_color.Mul(point_light_color)
		acc_color = acc_color.Add(point_col)

		// advance ray inside volume
		dv := ray.dir.Scale(ds)
		ray.origin = ray.origin.Add(dv)
		acc_distance += ds

		count += 1.0
	}
	diffuse := acc_color
	alpha := 1 - beers_law(acc_distance, acc_density)
	return Vec4{diffuse.X, diffuse.Y, diffuse.Z, alpha}
}

func march_through_volume_raymarched_light_1(ray *Ray, render_params *RenderParameters) Vec4 {
	sphere := render_params.sphere
	light := render_params.light

	acc_density := 0.0
	acc_distance := 0.0      // accumulated distance inside the volume
	acc_color := Vec3Fill(0) // accumulated color
	acc_alpha := 0.0

	var ds float64
	if scale_volume_res_per_object {
		ds = sphere.R / number_of_steps_for_object_scaling
	} else {
		ds = volume_resolution
	}

	// when orientations are introduced, the normals will have to be transformed
	// as long as there are only translations, directions are OK in any translated space (not rotated or scaled)

	for {
		point_in_sphere_space := ray.origin.Sub(sphere.C)
		sdf := sdfSphere(point_in_sphere_space, sphere.R)
		if sdf > 0 {
			break // went outside the volume
		}

		density := sample_density(ray.origin, render_params.noises, render_params.time) * volume_resolution
		// density *= asymptote_to_one(math.Abs(sdf), 10.0) // make density closer to the surface softer
		acc_density += density

		distance_sampled_to_light, density_to_light := march_through_volume_to_light(ray.origin, sphere, light, render_params.noises, render_params.time)
		light_amount := beers_law(distance_sampled_to_light, density_to_light)
		light_color_at_point := light.color.Scale(light_amount)
		point_color := cloud_color.Mul(light_color_at_point)
		acc_color = acc_color.Add(point_color)
		acc_alpha += 1 - beers_law(acc_distance, acc_density)

		// advance ray inside volume
		dv := ray.dir.Scale(ds)
		ray.origin = ray.origin.Add(dv)
		acc_distance += ds
	}
	diffuse := acc_color
	alpha := 1 - beers_law(acc_distance, acc_density)
	return Vec4{diffuse.X, diffuse.Y, diffuse.Z, alpha}
}

func march_through_volume_raymarched_light_2(ray *Ray, render_params *RenderParameters) Vec4 {
	sphere := render_params.sphere
	light := render_params.light

	acc_density := 0.0
	acc_distance := 0.0 // accumulated distance inside the volume
	acc_light_amount := 0.0
	avg_sdf := 0.0
	count := 0.0

	var ds float64
	if scale_volume_res_per_object {
		ds = sphere.R / number_of_steps_for_object_scaling
	} else {
		ds = volume_resolution
	}

	// when orientations are introduced, the normals will have to be transformed
	// as long as there are only translations, directions are OK in any translated space (not rotated or scaled)

	for {
		point_in_sphere_space := ray.origin.Sub(sphere.C)
		sdf := sdfSphere(point_in_sphere_space, sphere.R)
		if sdf > 0 {
			break // went outside the volume
		}
		avg_sdf += math.Abs(sdf)

		density := sample_density(ray.origin, render_params.noises, render_params.time) //* volume_resolution
		acc_density += density

		distance_sampled_to_light, density_to_light := march_through_volume_to_light(ray.origin, sphere, light, render_params.noises, render_params.time)
		light_amount := beers_law(distance_sampled_to_light, density_to_light)
		acc_light_amount += light_amount

		// advance ray inside volume
		dv := ray.dir.Scale(ds)
		ray.origin = ray.origin.Add(dv)
		acc_distance += ds

		count += 1.0
	}
	avg_sdf /= count
	light_amount := acc_light_amount / count // average
	diffuse := cloud_color.Scale(light_amount)
	alpha := 1 - beers_law(acc_distance, acc_density)
	if ease_in_edges {
		alpha *= ease_in(avg_sdf) // soften edges, if avg. sdf is small, then only near the surface density was sampled
		// alpha *= clamp01(avg_sdf)
	}
	return Vec4{diffuse.X, diffuse.Y, diffuse.Z, alpha}
}

func march_through_volume_to_light(
	point Vec3,
	sphere *Sphere,
	light *Light,
	noises *Noises,
	time float64,
) (distance, density float64) {
	// As long as there are only translations, directions are OK in any translated space (not rotated or scaled)
	light_origin_s := light.origin.Sub(sphere.C) // light origin in sphere space
	point_s := point.Sub(sphere.C)               // point in sphere space
	dir_to_light := light_origin_s.Sub(point_s).Normalized()

	acc_distance := 0.0
	acc_density := 0.0

	var ds float64
	if scale_volume_res_per_object {
		ds = sphere.R / number_of_steps_for_object_scaling
	} else {
		ds = volume_resolution
	}

	for {
		sdf := sdfSphere(point_s, sphere.R)
		if sdf > 0 {
			acc_distance -= sdf // decrease by the over-shot distance outside the volume
			break               // went outside the volume
		}

		acc_density += sample_density(point, noises, time) //* volume_resolution

		// advance point towards light
		dv := dir_to_light.Scale(ds)
		point_s = point_s.Add(dv)
		acc_distance += volume_resolution
	}
	return acc_distance, acc_density
}

func sample_density(point Vec3, noises *Noises, time float64) float64 {
	// scale by resolution so it looks the same regardless of resolution value
	// return 0.05

	noise_scale := 20.0
	noise_phase := time * 4
	noise_x := int(math.Abs(point.X*noise_scale + noise_phase*1))
	noise_y := int(math.Abs(point.Y*noise_scale + noise_phase*1))
	// noise_z := int(math.Abs(point.Z*noise_scale*2 + noise_phase*1))
	noise1 := noises.tex_values.get(noise_x, noise_y)
	// noise2 := noises.tex_values.get(noise_y, noise_z)
	// noisef_0 := (noise1 + noise2) * 0.5
	noisef_0 := noise1
	// return noisef_0

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

	perlin_balance := 0.5
	perlinf := ((1-perlin_balance)*perlin_1 + perlin_balance*perlin_2)

	balance := 1.0                                   // cell texture is not good yet
	noisef := noisef_0*(1-balance) + perlinf*balance //+ noisef_1 + 0.1

	return noisef
}
