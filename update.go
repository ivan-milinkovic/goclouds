package main

import (
	"math"
	"sync"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func update(img *ImageTarget, camera *Camera) {
	// for i := range img.W * img.H {
	// 	(*img).Pixels[i].B = byte(255 * math.Sin(rl.GetTime()*2))
	// }

	sphere := Sphere{
		C: Vec3{0, 0, -2},
		R: 1,
	}

	// ray := Ray{
	// 	origin: Vec3Fill(0),
	// 	dir:    Vec3Make(0, 0, -1),
	// }
	// march(&ray, &sphere)
	// panic(123)

	light := DirectionalLight{
		dir: Vec3Make(-1.5, 1.5, 0.75).Normalized(),
	}

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
	var dH = 10 // increment on the y axis for each goroutine
	// var dH := img.H / runtime.NumCPU()

	for y_mark < img.H {
		wg.Add(1)
		go func(y_mark int, img *ImageTarget) {
			end := min(y_mark+dH, img.H)
			for y := y_mark; y < end; y++ {
				for x := range img.W {
					ray := camera.MakeRay(x, y, img.W, img.H)
					colorf := march(&ray, &sphere, &light)

					mod := Vec3Fill(math.Sin(rl.GetTime() * 4)).Add(Vec3Fill(1))
					colorf = colorf.Mul(mod)
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

func march(starting_ray *Ray, sphere *Sphere, light *DirectionalLight) Vec3 {
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
