package main

import (
	"image"
	"image/color"
	"image/png"
	_ "image/png"
	"os"
)

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

func main() {
	pixel_width := 500
	pixel_height := 500
	abs_bound := 1000.0
	max_iterations := uint32(200)

	var centerx float64 = -1.25
	var centery float64 = 0
	var width float64 = 0.25
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
	file, _ := os.Create("result.png")
	defer file.Close()

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

	png.Encode(file, image)
}
