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

	noises := Noises{cell_values: NewDataMatrix[float64](10, 10)}

	light := Light{
		origin: Vec3Make(-1, 1, 0),
		dir:    Vec3Make(1, -0.25, 0).Normalized(),
		color:  Vec3Fill(1.0),
	}

	pprof.StartCPUProfile(f)
	for b.Loop() {
		ray_march(&image_target, &camera, &light, &noises, 1)
	}
	pprof.StopCPUProfile()
}
