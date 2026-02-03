package main

type Camera struct {
	origin     Vec3 //
	p00        Vec3 // image plane center
	aspect     float64
	near_plane float64
}

type Ray struct {
	origin Vec3
	dir    Vec3
}

func (c *Camera) MakeRay(x int, y int, img_w int, img_h int) Ray {
	cam := *c

	canvas_x := float64(x) / float64(img_w)
	canvas_x = canvas_x*2 - 1 // make [-1, 1], center around zero
	canvas_x *= cam.aspect

	canvas_y := float64(y) / float64(img_h)
	canvas_y = canvas_y*2 - 1 // make [-1, 1], center around zero
	canvas_y *= -1            // flip y (to go up), as the image is interpreted from top to bottom (y goes down)

	canvas_h := 2.0
	canvas_w := canvas_h * cam.aspect
	dx := float64(canvas_w) / float64(img_w)
	dy := float64(canvas_h) / float64(img_h)

	p_canvas := Vec3{X: canvas_x, Y: canvas_y, Z: 0.0}
	p_canvas.X += dx * 0.5
	p_canvas.Y -= dy * 0.5
	p_canvas = p_canvas.Add(cam.p00)

	ray_dir := p_canvas.Sub(cam.origin).Normalized()

	ray := Ray{
		origin: cam.origin,
		dir:    ray_dir,
	}

	return ray
}
