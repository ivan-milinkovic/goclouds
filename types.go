package main

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type State struct {
	image_target *ImageTarget
	camera       *Camera
	light        *Light
	sphere       *Sphere
	geo          *Geometry
	noises       *Noises
	texture      *rl.Texture2D
}

type Pixel = color.RGBA

type ImageTarget struct {
	Pixels []Pixel
	W      int
	H      int
}

type Light struct { // point light
	origin Vec3
	color  Vec3
}

type ShadingType = int

const (
	ShadingType_NoLight         ShadingType = 0
	ShadingType_RayMarchedLight ShadingType = 1
)

type DensityType = int

const (
	DensityType_PerlinRuntime = 1
	DensityType_PerlinPreCalc = 2
	DensityType_Uniform       = 3
)

type RenderParameters struct {
	img    *ImageTarget
	camera *Camera
	light  *Light
	geo    *Geometry
	noises *Noises
	time   float64
}
