package main

const WINDOW_WIDTH = 640
const WINDOW_HEIGHT = 480

// viewport size for rendering volumetrics
const VOL_VIEWPORT_W = WINDOW_WIDTH / 2
const VOL_VIEWPORT_H = WINDOW_HEIGHT / 2

const SHADING_TYPE = ShadingType_RayMarchedLight
const MAX_JUMPS = 40                  // max jumps for a single ray
const SCALE_STEP_RES_TO_OBJECT = true // scale ray advance step based on object size
const NUM_STEPS_OBJECT_SCALING = 10
const VOLUME_RESOLUTION = 0.1 // when not scaling
const EASE_IN_EDGES = true
const EASE_IN_INSIDE_VOLUMES = true

var cloud_color = Vec3{0.95, 0.95, 0.95}
var density_type = DensityType_PerlinPreCalc // updated by key shortcuts 1,2,3

const RENDER_LIGHT_SOURCE = false
const ANIMATE_LIGHT_POSITION = false

const PREVIEW_PERLIN = false
