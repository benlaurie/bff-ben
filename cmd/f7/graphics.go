//go:build graphics
// +build graphics

package main

import (
	"image"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"github.com/crazy3lf/colorconv"
)

func graphics(universe *[65536]uint8) {
	myApp := app.New()
	w := myApp.NewWindow("Raster")

	raster := canvas.NewRasterWithPixels(
		func(x, y, w, h int) color.Color {

			x = x * SQRT_ULEN / w
			y = y * SQRT_ULEN / h
			n := x + y*SQRT_ULEN
			if n >= ULEN {
				hsl, _ := colorconv.HSLToColor(0.0, 0.0, 0.0)
				return hsl
			}

			op := universe[n]

			if op > MAX_OP {
				l := juice(n)/(512/100.) + .5
				hsl, _ := colorconv.HSLToColor(0.0, 0.0, l)
				return hsl
			}

			//fmt.Println(s)

			hsl, err := colorconv.HSLToColor(float64(op)/float64(MAX_OP+1)*360.0, 0.9, 0.5)

			if err != nil {
				panic(err)
			}

			return hsl

		})

	w.SetContent(raster)
	w.Resize(fyne.NewSize(SQRT_ULEN*2, SQRT_ULEN*2))

	i_w := myApp.NewWindow("Instructions")
	i_raster := canvas.NewRaster(
		func(w, h int) image.Image {
			var ops [256]uint64
			max := uint64(0)
			for i := 0; i < ULEN; i++ {
				ops[universe[i]]++
				if ops[universe[i]] > max {
					max = ops[universe[i]]
				}
			}
			image := image.NewRGBA(image.Rect(0, 0, w, h))
			for y := 0; y < h; y++ {
				op := y * 256 / h
				if op > 255 {
					op = 255
				}
				hue := float64(op) / float64(MAX_OP+1) * 360.0
				l := float64(ops[op]) / float64(max)
				s := 1.0
				if op > MAX_OP {
					hue = 0.0
					s = 0.0
				}

				hsl, err := colorconv.HSLToColor(hue, s, l)
				if err != nil {
					panic(err)
				}

				for x := 0; x < w; x++ {
					image.Set(x, y, hsl)
				}
			}
			return image
		})
	i_w.SetContent(i_raster)
	i_w.Resize(fyne.NewSize(128, 512))

	go func() {
		for {
			raster.Refresh()
			i_raster.Refresh()
			time.Sleep(20 * time.Millisecond)
		}
	}()

	i_w.Show()
	w.ShowAndRun()
}
