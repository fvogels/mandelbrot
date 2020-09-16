package main

import (
	"image"
	"image/color"
	"image/png"
	_ "image/png"
	"log"
	"os"
	"sync"
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

func compute_iterations(x, y float64, settings *settings) uint32 {
	re := x
	im := y
	abs := re*re + im*im
	abs_bound := settings.abs_bound
	max_iterations := settings.max_iterations
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

func render_pixel(x, y float64, settings *settings) color.RGBA {
	max_iterations := settings.max_iterations
	iterations := compute_iterations(x, y, settings)
	intensity := float64(iterations) / float64(max_iterations)
	color_component := uint8(intensity * 255.0)
	color := color.RGBA{color_component, color_component, color_component, 255}

	return color
}

type RowRenderer interface {
	render_row(x_start, x_step, y float64, py int, settings *settings, image *image.RGBA)
}

type SerialRowRenderer struct{}

func (_ SerialRowRenderer) render_row(x_start, x_step, y float64, py int, settings *settings, image *image.RGBA) {
	pixel_width := settings.image_width

	for i := 0; i < pixel_width; i++ {
		x := x_start + x_step*float64(i)
		color := render_pixel(x, y, settings)
		image.Set(i, py, color)
	}
}

type ConcurrentRowRenderer struct {
	waitgroup *sync.WaitGroup
}

func (r ConcurrentRowRenderer) render_row(x_start, x_step, y float64, py int, settings *settings, image *image.RGBA) {
	pixel_width := settings.image_width

	for i := 0; i < pixel_width; i++ {
		r.waitgroup.Add(1)

		go func(i int) {
			defer r.waitgroup.Done()

			x := x_start + x_step*float64(i)
			color := render_pixel(x, y, settings)
			image.Set(i, py, color)
		}(i)
	}
}

type ImageRenderer interface {
	render_image(settings *settings) image.Image
}

type SerialImageRenderer struct{}

func (_ SerialImageRenderer) render_image(settings *settings) image.Image {
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
	rendering := image.NewRGBA(size)
	row_renderer := SerialRowRenderer{}

	for py := 0; py < pixel_height; py++ {
		y := float64(py)*vscale + vintercept
		row_renderer.render_row(hintercept, hscale, y, py, settings, rendering)
	}

	return rendering
}

func render_image_concurrent_rows(settings *settings) image.Image {
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
	rendering := image.NewRGBA(size)

	var waitgroup sync.WaitGroup
	row_renderer := SerialRowRenderer{}
	// row_renderer := ConcurrentRowRenderer{waitgroup: &waitgroup}

	for py := 0; py < pixel_height; py++ {
		waitgroup.Add(1)

		go func(py int) {
			defer waitgroup.Done()

			y := float64(py)*vscale + vintercept
			row_renderer.render_row(hintercept, hscale, y, py, settings, rendering)
		}(py)
	}

	waitgroup.Wait()

	return rendering
}

func main() {
	before := time.Now()

	s := settings{
		image_width:    3440,
		image_height:   1440,
		center_x:       -1.25,
		center_y:       0,
		width:          0.25,
		abs_bound:      10000.0,
		max_iterations: 200}

	image := render_image_concurrent_rows(&s)

	file, _ := os.Create("result.png")
	defer file.Close()
	png.Encode(file, image)

	elapsed := time.Since(before)
	log.Printf("Done in %s", elapsed)
}
