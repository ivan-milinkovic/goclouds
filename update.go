package main

import "math"

func update(img *ImageTarget, camera *Camera) {
	// for i := range img.W * img.H {
	// 	(*img).Pixels[i].B = byte(255 * math.Sin(rl.GetTime()*2))
	// }

	sphere := Sphere{
		C: Vec3{0, 0, -2},
		R: 1,
	}

	// ray := Ray{origin: Vec3{0, 0, 1}, dir: Vec3{0, 0, -1}}
	// ray_in_sphere_space := ray.origin.Sub(sphere.C)
	// sdf := sdfSphere(ray_in_sphere_space, sphere.R)
	// println(sdf)
	// panic(123)

	for y := range img.H {
		for x := range img.W {
			ray := camera.MakeRay(x, y, img.W, img.H)
			colorf := march(&ray, &sphere)

			p := pixel_from_fcolor(colorf)
			img.Pixels[y*img.W+x] = p
		}
	}
}

type Sphere struct {
	C Vec3
	R float64
}

func march(starting_ray *Ray, sphere *Sphere) Vec3 {
	ray := *starting_ray
	count := 0
	for {
		ray_in_sphere_space := ray.origin.Sub(sphere.C)
		sdf := sdfSphere(ray_in_sphere_space, sphere.R)
		if sdf < 0.2 {
			v := math.Abs(sdf / (sdf + 1))
			return Vec3{X: v, Y: v, Z: v}
		}

		// advance ray
		dv := ray.dir.Scale(sdf)
		ray.origin = ray.origin.Add(dv)

		if count >= 10 {
			return Vec3{X: 0, Y: 0, Z: 0}
		}
		count++
	}
}
