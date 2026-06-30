package gifbg

import (
	"image"
	"image/color"
	"image/draw"
	"math"
)

const (
	defWidth  = 80
	defHeight = 24
	defCount  = 30
	defDelay  = 100
)

func defaultFrames() []frameData {
	w, h := defWidth, defHeight
	numFrames := defCount
	delay := defDelay

	var frames []frameData

	for i := 0; i < numFrames; i++ {
		img := image.NewRGBA(image.Rect(0, 0, w, h))
		phase := float64(i) / float64(numFrames) * 2 * math.Pi

		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				nx := float64(x) / float64(w)
				ny := float64(y) / float64(h)

				v := math.Sin(phase+nx*3.0+ny*2.0)*0.5 + 0.5
				v *= 0.3

				u := math.Sin(phase*1.3+nx*2.0+ny*3.0)*0.5 + 0.5
				u *= 0.2

				r := uint8(v * 25)
				g := uint8(v*15 + u*10)
				b := uint8(v*80 + u*40 + 15)

				img.Set(x, y, color.RGBA{r, g, b, 255})
			}
		}

		rgba := image.NewRGBA(img.Bounds())
		draw.Draw(rgba, rgba.Bounds(), img, image.Point{}, draw.Src)

		frames = append(frames, frameData{
			RGBA:  rgba,
			Delay: delay,
		})
	}

	if len(frames) == 0 {
		return nil
	}
	return frames
}
