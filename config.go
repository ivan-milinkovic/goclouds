package main

const window_width = 640
const window_height = 480

// viewport size for rendering volumetrics
const vol_viewport_w = window_width / 2
const vol_viewport_h = window_height / 2

const shading_type = ShadingType_RayMarchedLight
const max_jumps = 40
const scale_volume_res_per_object = true // scale ray advance step based on object size
const number_of_steps_for_object_scaling = 10
const volume_resolution = 0.1 // when not scaling
const ease_in_edges = true

const PREVIEW_PERLIN = false

var cloud_color = Vec3{0.95, 0.95, 0.95}
