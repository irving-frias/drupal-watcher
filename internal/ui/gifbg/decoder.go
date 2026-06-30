package gifbg

import (
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"os"
	"path/filepath"
)

const maxDecodeDim = 320

type frameData struct {
	RGBA  *image.RGBA
	Delay int
}

func decodeGIF(path string) ([]frameData, error) {
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	g, err := gif.DecodeAll(f)
	if err != nil {
		return nil, err
	}

	if len(g.Image) == 0 {
		return nil, nil
	}

	bounds := g.Image[0].Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	if w == 0 || h == 0 {
		w = g.Config.Width
		h = g.Config.Height
	}
	if w == 0 || h == 0 {
		return nil, nil
	}

	scale := 1.0
	if w > maxDecodeDim || h > maxDecodeDim {
		sx := float64(maxDecodeDim) / float64(w)
		sy := float64(maxDecodeDim) / float64(h)
		if sx < sy {
			scale = sx
		} else {
			scale = sy
		}
	}
	sw := int(float64(w) * scale)
	sh := int(float64(h) * scale)
	if sw < 1 {
		sw = 1
	}
	if sh < 1 {
		sh = 1
	}

	canvas := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(canvas, canvas.Bounds(), image.Transparent, image.Point{}, draw.Src)

	var frames []frameData

	for i, src := range g.Image {
		bounds := src.Bounds()

		var prevSnap *image.RGBA
		if i > 0 && g.Disposal[i] == gif.DisposalPrevious {
			prevSnap = image.NewRGBA(canvas.Bounds())
			draw.Draw(prevSnap, prevSnap.Bounds(), canvas, image.Point{}, draw.Src)
		}

		draw.Draw(canvas, bounds, src, bounds.Min, draw.Over)

		var frame *image.RGBA
		if scale < 1.0 {
			frame = image.NewRGBA(image.Rect(0, 0, sw, sh))
			downscaleNearest(canvas, frame)
		} else {
			frame = image.NewRGBA(image.Rect(0, 0, w, h))
			draw.Draw(frame, frame.Bounds(), canvas, image.Point{}, draw.Src)
		}

		switch g.Disposal[i] {
		case gif.DisposalBackground:
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					c := color.RGBA{0, 0, 0, 0}
					if int(g.BackgroundIndex) < len(src.Palette) {
						c = color.RGBAModel.Convert(src.Palette[g.BackgroundIndex]).(color.RGBA)
					}
					canvas.Set(x, y, c)
				}
			}
		case gif.DisposalPrevious:
			if prevSnap != nil {
				draw.Draw(canvas, canvas.Bounds(), prevSnap, image.Point{}, draw.Src)
			}
		}

		rgba := image.NewRGBA(frame.Bounds())
		draw.Draw(rgba, rgba.Bounds(), frame, image.Point{}, draw.Src)

		delay := int(g.Delay[i]) * 10
		if delay < 20 {
			delay = 100
		}

		frames = append(frames, frameData{
			RGBA:  rgba,
			Delay: delay,
		})
	}

	return frames, nil
}

func downscaleNearest(src *image.RGBA, dst *image.RGBA) {
	sb := src.Bounds()
	sw := sb.Dx()
	sh := sb.Dy()
	dw := dst.Bounds().Dx()
	dh := dst.Bounds().Dy()

	for dy := 0; dy < dh; dy++ {
		for dx := 0; dx < dw; dx++ {
			sx := dx * sw / dw
			sy := dy * sh / dh
			c := src.At(sx+sb.Min.X, sy+sb.Min.Y)
			dst.Set(dx, dy, c)
		}
	}
}
