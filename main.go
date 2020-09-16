package main

import (
	"fmt"
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
	filename       string
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

type SerialImageRenderer struct {
	row_renderer RowRenderer
}

func (r SerialImageRenderer) render_image(settings *settings) image.Image {
	row_renderer := r.row_renderer
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

	for py := 0; py < pixel_height; py++ {
		y := float64(py)*vscale + vintercept
		row_renderer.render_row(hintercept, hscale, y, py, settings, rendering)
	}

	return rendering
}

type ConcurrentImageRenderer struct {
	waitgroup    *sync.WaitGroup
	row_renderer RowRenderer
}

func (r ConcurrentImageRenderer) render_image(settings *settings) image.Image {
	row_renderer := r.row_renderer
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

	waitgroup := r.waitgroup

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

type AnimationRenderer interface {
	render(settings_receiver <-chan *settings)
}

type SerialAnimationRenderer struct {
	frame_renderer ImageRenderer
}

func (r SerialAnimationRenderer) render(settings_receiver <-chan *settings) {
	frame_renderer := r.frame_renderer
	settings := <-settings_receiver

	for settings != nil {
		log.Printf("Rendering %s", settings.filename)
		frame := frame_renderer.render_image(settings)
		file, _ := os.Create(settings.filename)
		png.Encode(file, frame)
		file.Close()

		settings = <-settings_receiver
	}
}

func main() {
	before := time.Now()

	var waitgroup sync.WaitGroup

	row_renderer := SerialRowRenderer{}
	frame_renderer := ConcurrentImageRenderer{
		waitgroup:    &waitgroup,
		row_renderer: row_renderer}
	renderer := SerialAnimationRenderer{frame_renderer: frame_renderer}
	settings_channel := make(chan *settings)
	go renderer.render(settings_channel)

	start_width := 1.0
	end_width := 0.005
	nframes := 30

	for i := 0; i < nframes; i += 1 {
		filename := fmt.Sprintf("frame%05d.png", i)
		width := start_width + float64(i)*(end_width-start_width)/float64(nframes)

		s := settings{
			image_width:    500,
			image_height:   500,
			center_x:       -0.7463,
			center_y:       0.1102,
			width:          width,
			abs_bound:      10000.0,
			max_iterations: 200,
			filename:       filename}

		settings_channel <- &s
	}

	settings_channel <- nil

	elapsed := time.Since(before)
	log.Printf("Done in %s", elapsed)
}
