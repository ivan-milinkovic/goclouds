package main

import "math"

func sample_density(point Vec3, noises *Noises, time float64) float64 {
	// return 0.05
	// return sample_density_runtime_perlin(point, noises, time)
	// return sample_density_pre_calc_perlin_1(point, noises, time)
	return sample_density_pre_calc_perlin_2(point, noises, time)
	// return sample_density_2D_texture(point, noises, time)
}

func sample_density_runtime_perlin(point Vec3, noises *Noises, time float64) float64 {

	var p1 float64 = 123
	{
		scale := 2.0
		phase := time * 0.5
		p1 = noises.perlin_gen.Noise3D(
			point.X*scale+phase*1,
			point.Y*scale+phase*0,
			point.Z*scale+phase*2,
		)
		p1 = clamp01(p1)
	}

	var p2 float64
	{
		scale := 7.0
		phase := time * 1
		p2 = noises.perlin_gen.Noise3D(
			point.X*scale+phase*1,
			point.Y*scale+phase*0,
			math.Abs(point.Z)*scale+phase*1, // see Noise3D implementation, falls back to 3D if z < 0
		)
		p2 = clamp01(p2)
	}

	p := mix(p1, p2, 0.0)
	return p
}

func sample_density_pre_calc_perlin_1(point Vec3, noises *Noises, time float64) float64 {
	scale := 0.8
	phase := time * 0.08
	coords := point.Scale(scale).Add(Vec3{phase * 1, phase * 0, phase * 2})
	coords = coords.Add(Vec3{0.2, 0.2, 0.2}) // avoid tiling edges where it's all zeros
	perlin_1 := noises.perlin_values.getFromVectorWrap(coords)
	perlin_1 = clamp01(perlin_1)
	return perlin_1
}

func sample_density_pre_calc_perlin_2(point Vec3, noises *Noises, time float64) float64 {
	var p1 float64
	{
		scale := 0.4
		phase := time * 0.08
		coords := point.Scale(scale).Add(Vec3{phase * 1, phase * 0, phase * 2})
		coords = coords.Add(Vec3{0.05, 0.1, 0.15}) // avoid tiling edges where it's all zeros
		p1 = noises.perlin_values.getFromVectorWrap(coords)
		p1 = clamp01(p1)
	}
	var p2 float64
	{
		scale := 1.0
		phase := time * 0.08
		coords := point.Scale(scale).Add(Vec3{phase * 1, phase * 0, phase * 2})
		coords = coords.Add(Vec3{0.05, 0.1, 0.15}) // avoid tiling edges where it's all zeros
		p2 = noises.perlin_values.getFromVectorWrap(coords)
		p2 = clamp01(p2)
	}
	p := mix(p1, p2, 0.5)
	return p
}

func sample_density_2D_texture(point Vec3, noises *Noises, time float64) float64 {
	noise_scale := 20.0
	noise_phase := time * 4
	noise_x := int(math.Abs(point.X*noise_scale + noise_phase*1))
	noise_y := int(math.Abs(point.Y*noise_scale + noise_phase*1))
	// noise_z := int(math.Abs(point.Z*noise_scale*2 + noise_phase*1))
	noise1 := noises.tex_values.getWrap(noise_x, noise_y)
	// noise2 := noises.tex_values.get(noise_y, noise_z)
	// noisef_0 := (noise1 + noise2) * 0.5
	noisef := noise1
	return noisef
}
