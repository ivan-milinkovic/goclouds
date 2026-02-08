package main

// go test -bench .
// go tool pprof cpu.prof
// go tool pprof -http=:8080 cpu.prof

import (
	"os"
	"runtime/pprof"
	"testing"
)

func BenchmarkProf(b *testing.B) {
	f, err := os.Create("cpu.prof")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	screen_w := 640
	screen_h := 480
	pixel_count := screen_w * screen_h
	camera := Camera{
		origin: Vec3{0, 0, 1},
		p00:    Vec3{0, 0, 0},
		aspect: float64(screen_w) / float64(screen_h),
	}
	image_target := ImageTarget{
		Pixels: make([]Pixel, pixel_count),
		W:      screen_w,
		H:      screen_h,
	}

	noises := NewNoises()
	noises.tex_values = NewDataMatrix[float64](10, 10) // noises.tex_values.W becomes 0 for some reason if run without debugging

	light := Light{
		origin: Vec3Make(-1, 1, 0),
		color:  Vec3Fill(1.0),
	}

	sphere := Sphere{
		C: Vec3{0, 0, -1},
		R: 1,
	}

	render_parameters := RenderParameters{
		img:    &image_target,
		camera: &camera,
		light:  &light,
		sphere: &sphere,
		noises: noises,
		time:   0.0,
	}

	pprof.StartCPUProfile(f)
	for b.Loop() {
		ray_march(&render_parameters)
	}
	pprof.StopCPUProfile()
}
