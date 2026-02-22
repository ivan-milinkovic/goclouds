package main

import (
	"math"
	"runtime"
	"sync"
)

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
					colorf := march_volume(&ray, render_params)
					if RENDER_LIGHT_SOURCE {
						color_light_source := march_light(&ray, render_params)
						colorf = colorf.Add(color_light_source)
					}
					// color_solids := march_solid(&ray, render_params)
					// colorf := color_solids

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
	light := render_params.light
	background := Vec4{0, 0, 0, 0}
	count := 0
	for {
		sdf := render_params.geo.Sdf(ray.origin)

		// distort the shape
		// var x = int(16.0*ray.origin.X + render_params.time*8.0)
		// var y = int(16.0*ray.origin.Y + render_params.time*8.0)
		// sdf += render_params.noises.tex_values.getWrap(x, y)
		// sdf += 1.5 * render_params.noises.perlin_values.getFromVectorWrap(ray.origin.Scale(0.4).AddScalar(0.2*render_params.time))

		if sdf < 0.02 {
			// v := math.Abs(sdf / (sdf + 1))
			// return Vec3{X: v, Y: v, Z: v}

			// when orientations are introduced, the normals will have to be transformed
			// as long as there are only translations, directions are OK in any translated space (not rotated or scaled)
			normal := render_params.geo.NormalAt(ray.origin)
			dir_to_light := light.origin.Sub(ray.origin).Normalized()
			light_amount := normal.Dot(dir_to_light)
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

func march_light(starting_ray *Ray, render_params *RenderParameters) Vec4 {
	ray := *starting_ray
	light := render_params.light
	count := 0
	for {
		point_l := ray.origin.Sub(light.origin) // ray origin in light space
		sdf := sdfSphere(point_l, 0.1)          // render light as a sphere
		if sdf < 0.02 {
			return Vec4Make(light.color, 1)
		}

		// advance ray
		dv := ray.dir.Scale(sdf)
		ray.origin = ray.origin.Add(dv)

		if sdf >= 10 || count >= 10 {
			break
		}
		count++
	}
	return Vec4Fill(0.0)
}

func march_volume(starting_ray *Ray, render_params *RenderParameters) Vec4 {
	ray := *starting_ray

	jump_count := 0
	var acc_color Vec4

	for jump_count < MAX_JUMPS {
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
	prev_sdf := math.MaxFloat64
	for *jump_count < MAX_JUMPS {
		*jump_count++

		sdf := render_params.geo.Sdf(ray.origin)

		// distort the shape
		// var x = int(32.0*ray.origin.X + render_params.time*8.0)
		// var y = int(32.0*ray.origin.Y + render_params.time*8.0)
		// sdf += render_params.noises.tex_values.getWrap(x, y)
		// sdf += render_params.noises.perlin_values.getFromVectorWrap(ray.origin.Scale(0.2).AddScalar(0.2 * render_params.time))

		if sdf <= 0 {
			ray.origin = ray.origin.Add(ray.dir.Scale(sdf)) // move back to the beginning of the volume
			return true                                     // found a volume
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
	switch SHADING_TYPE {
	case ShadingType_NoLight:
		return march_through_volume_no_light(ray, render_params)
	case ShadingType_RayMarchedLight:
		// return march_through_volume_raymarched_light_1(ray, render_params)
		return march_through_volume_raymarched_light_2(ray, render_params)
	}
	return Vec4{0.2, 0, 0.1, 0}
}

func march_through_volume_no_light(ray *Ray, render_params *RenderParameters) Vec4 {
	acc_density := 0.0
	acc_distance := 0.0 // accumulated distance inside the volume
	count := 0.0

	var ds float64
	if SCALE_STEP_RES_TO_OBJECT {
		ds = render_params.geo.Size() / NUM_STEPS_OBJECT_SCALING
	} else {
		ds = VOLUME_RESOLUTION
	}

	// when orientations are introduced, the normals will have to be transformed
	// as long as there are only translations, directions are OK in any translated space (not rotated or scaled)

	for {
		sdf := render_params.geo.Sdf(ray.origin)
		if sdf > 0 {
			break // went outside the volume
		}

		density := sample_density(ray.origin, render_params.noises, render_params.time) * VOLUME_RESOLUTION

		// advance ray inside volume
		dv := ray.dir.Scale(ds)
		ray.origin = ray.origin.Add(dv)

		acc_density += density
		acc_distance += ds

		count += 1.0
		if count > float64(MAX_JUMPS) {
			break
		}
	}
	diffuse := cloud_color
	background_passthrough := beers_law(acc_distance, acc_density)
	alpha := 1 - background_passthrough
	return Vec4{diffuse.X, diffuse.Y, diffuse.Z, alpha}
}

// accumulating color
func march_through_volume_raymarched_light_1(ray *Ray, render_params *RenderParameters) Vec4 {
	light := render_params.light

	acc_density := 0.0
	acc_distance := 0.0      // accumulated distance inside the volume
	acc_color := Vec3Fill(0) // accumulated color
	acc_alpha := 0.0

	var ds float64
	if SCALE_STEP_RES_TO_OBJECT {
		ds = render_params.geo.Size() / NUM_STEPS_OBJECT_SCALING
	} else {
		ds = VOLUME_RESOLUTION
	}

	// when orientations are introduced, the normals will have to be transformed
	// as long as there are only translations, directions are OK in any translated space (not rotated or scaled)

	for {
		sdf := render_params.geo.Sdf(ray.origin)
		if sdf > 0 {
			break // went outside the volume
		}

		density := sample_density(ray.origin, render_params.noises, render_params.time) * VOLUME_RESOLUTION
		// density *= asymptote_to_one(math.Abs(sdf), 10.0) // make density closer to the surface softer
		acc_density += density

		distance_sampled_to_light, density_to_light := march_through_volume_to_light(ray.origin, render_params.geo, light, render_params.noises, render_params.time)
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

// accumulating light intensity
func march_through_volume_raymarched_light_2(ray *Ray, render_params *RenderParameters) Vec4 {
	light := render_params.light

	acc_mass := 0.0
	acc_distance := 0.0 // accumulated distance inside the volume
	acc_light_amount := 0.0
	acc_sdf := 0.0
	count := 0.0

	geo_size := render_params.geo.Size()
	var res_step float64
	if SCALE_STEP_RES_TO_OBJECT {
		res_step = geo_size / NUM_STEPS_OBJECT_SCALING
	} else {
		res_step = VOLUME_RESOLUTION
	}

	// when orientations are introduced, the normals will have to be transformed
	// as long as there are only translations, directions are OK in any translated space (not rotated or scaled)

	for {
		sdf := render_params.geo.Sdf(ray.origin)
		if sdf > 0 {
			acc_distance -= sdf // adjust for over-shooting beyond the volume
			break               // went outside the volume
		}
		abs_sdf := math.Abs(sdf)
		ds := min(abs_sdf, res_step)
		ds = max(res_step, 0) // must not be zero in order for ray to advance, initial sdf is zero as the ray is positioned precisely
		acc_sdf += abs_sdf

		mass := sample_density(ray.origin.Scale(1/geo_size), render_params.noises, render_params.time) * ds
		acc_mass += mass

		distance_sampled_to_light, mass_to_light := march_through_volume_to_light(ray.origin, render_params.geo, light, render_params.noises, render_params.time)
		light_amount := beers_law(distance_sampled_to_light, mass_to_light) // light transmittance from light to point
		light_amount *= beers_law(acc_distance, acc_mass)                   // light transmittance from point to camera
		// light_amount += MultipleOctaveScattering(density, 0.8)
		acc_light_amount += light_amount

		// advance ray inside volume
		rnd_offset := 0.0
		if RANDOMIZE_SAMPLING {
			// rnd_offset = ((hash(ray.origin.X) + hash(ray.origin.Y)) * 0.5) * 0.06
			rnd_offset = render_params.noises.perlin_values.getFromFloatsWrap(ray.origin.X, ray.origin.Y, ray.origin.Z) * 0.06
		}
		dv := ray.dir.Scale(ds + rnd_offset)
		ray.origin = ray.origin.Add(dv)
		acc_distance += ds

		count += 1.0
	}

	light_amount := acc_light_amount * 0.16
	light_color := light.color.Scale(light_amount)
	diffuse := cloud_color.Mul(light_color)

	alpha := 1 - beers_law(acc_distance, acc_mass)
	if EASE_IN_EDGES { // soften edges
		if EASE_IN_INSIDE_VOLUMES {
			// alpha *= ease_in(linear_step(0.0, 1.0, acc_mass))
			alpha *= ease_in(remap(acc_mass, 0, 1, 0, 1))
		}
		// alpha *= ease_in(linear_step(0.0, 6.0, acc_sdf)) // soften object outline; set by experimentation
		avg_sdf := acc_sdf / count
		alpha *= ease_in(clamp01(remap(avg_sdf/geo_size, 0, 0.2, 0, 1))) // distance of given fraction of R is faded out
		alpha = alpha * alpha
	}

	// col := light_color.Scale(1 - alpha).Add(cloud_color.Scale(alpha))
	// return Vec4{col.X, col.Y, col.Z, alpha}

	return Vec4{diffuse.X, diffuse.Y, diffuse.Z, alpha}
}

func march_through_volume_to_light(
	point Vec3,
	geo *Geometry,
	light *Light,
	noises *Noises,
	time float64,
) (distance, mass float64) {
	// As long as there are only translations, directions are OK in any translated space (not rotated or scaled)
	dir_to_light := light.origin.Sub(point).Normalized()

	acc_sdf := 0.0
	acc_distance := 0.0
	acc_mass := 0.0

	geo_size := geo.Size()
	var res_step float64
	if SCALE_STEP_RES_TO_OBJECT {
		res_step = geo_size / NUM_STEPS_OBJECT_SCALING
	} else {
		res_step = VOLUME_RESOLUTION
	}

	for {
		sdf := geo.Sdf(point)
		if sdf > 0 {
			acc_distance -= sdf // decrease by the over-shot distance outside the volume
			break               // went outside the volume
		}
		abs_sdf := math.Abs(sdf)
		ds := min(abs_sdf, res_step)
		ds = max(res_step, 0) // must not be zero
		acc_sdf += abs_sdf
		acc_mass += sample_density(point.Scale(1/geo_size), noises, time) * ds

		// advance point towards light
		dv := dir_to_light.Scale(res_step)
		point = point.Add(dv)
		acc_distance += VOLUME_RESOLUTION
	}

	// if EASE_IN_EDGES {
	// 	acc_mass *= ease_in(linear_step(0.0, 1.0, acc_sdf)) // soften towards surface
	// }

	return acc_distance, acc_mass
}
