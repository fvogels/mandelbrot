package main

import (
	"image"
	"image/color"
	"image/png"
	_ "image/png"
	"log"
	"os"
	"time"
)

type settings struct {
	image_width    int
	image_height   int
	center_x       float64
	center_y       float64
	width          float64
	abs_bound      float64
	max_iterations uint32
}

func mandelbrot(x, y, abs_bound float64, max_iterations uint32) uint32 {
	re := x
	im := y
	abs := re*re + im*im
	var iterations uint32 = 0

	for abs < abs_bound && iterations < max_iterations {
		new_re := re*re - im*im + x
		new_im := 2*re*im + y
		re = new_re
		im = new_im
		abs = re*re + im*im
		iterations += 1
	}

	return iterations
}

func render_image(settings *settings) image.Image {
	abs_bound := settings.abs_bound
	max_iterations := settings.max_iterations
	pixel_width := settings.image_width
	pixel_height := settings.image_height
	centerx := settings.center_x
	centery := settings.center_y
	width := settings.width

	var height float64 = (width * float64(pixel_height) / float64(pixel_width))
	var left float64 = centerx - width/2
	var top float64 = centery + height/2
	var right float64 = left + width
	var bottom float64 = top - height

	hscale := (right - left) / float64(pixel_width)
	vscale := (top - bottom) / float64(pixel_height)
	hintercept := left
	vintercept := bottom

	size := image.Rect(0, 0, pixel_width, pixel_height)
	image := image.NewRGBA(size)

	for py := 0; py < pixel_height; py++ {
		for px := 0; px < pixel_width; px++ {
			cx := float64(px)*hscale + hintercept
			cy := float64(py)*vscale + vintercept
			iterations := mandelbrot(cx, cy, abs_bound, max_iterations)
			intensity := float64(iterations) / float64(max_iterations)
			color_component := uint8(intensity * 255.0)
			color := color.RGBA{color_component, color_component, color_component, 255}
			image.Set(px, py, color)
		}
	}

	return image
}

func main() {
	before := time.Now()

	s := settings{
		image_width:    500,
		image_height:   500,
		center_x:       -1.25,
		center_y:       0,
		width:          0.25,
		abs_bound:      10000.0,
		max_iterations: 200}

	image := render_image(&s)

	file, _ := os.Create("result.png")
	defer file.Close()
	png.Encode(file, image)

	elapsed := time.Since(before)
	log.Printf("Done in %s", elapsed)
}
