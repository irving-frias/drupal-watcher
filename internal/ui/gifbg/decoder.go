package gifbg

import (
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"os"
	"path/filepath"
)

const (
	maxDecodeDim  = 1920
	maxGridCols   = 160
	maxGridRows   = 50
)

func decodeGIF(path string, cols, rows int) ([]Frame, error) {
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
	srcW := bounds.Dx()
	srcH := bounds.Dy()
	if srcW == 0 || srcH == 0 {
		srcW = g.Config.Width
		srcH = g.Config.Height
	}
	if srcW == 0 || srcH == 0 {
		return nil, nil
	}

	if cols > maxGridCols {
		cols = maxGridCols
	}
	if rows > maxGridRows {
		rows = maxGridRows
	}
	if cols < 1 {
		cols = 1
	}
	if rows < 1 {
		rows = 1
	}

	scale := 1.0
	if srcW > maxDecodeDim || srcH > maxDecodeDim {
		sx := float64(maxDecodeDim) / float64(srcW)
		sy := float64(maxDecodeDim) / float64(srcH)
		if sx < sy {
			scale = sx
		} else {
			scale = sy
		}
	}
	sw := int(float64(srcW) * scale)
	sh := int(float64(srcH) * scale)
	if sw < 1 {
		sw = 1
	}
	if sh < 1 {
		sh = 1
	}

	canvas := image.NewRGBA(image.Rect(0, 0, srcW, srcH))
	draw.Draw(canvas, canvas.Bounds(), image.Transparent, image.Point{}, draw.Src)

	var work *image.RGBA

	frames := make([]Frame, 0, len(g.Image))

	for i, src := range g.Image {
		srcBounds := src.Bounds()

		var prevSnap *image.RGBA
		if i > 0 && g.Disposal[i] == gif.DisposalPrevious {
			prevSnap = image.NewRGBA(canvas.Bounds())
			draw.Draw(prevSnap, prevSnap.Bounds(), canvas, image.Point{}, draw.Src)
		}

		draw.Draw(canvas, srcBounds, src, srcBounds.Min, draw.Over)

		if scale < 1.0 {
			if work == nil {
				work = image.NewRGBA(image.Rect(0, 0, sw, sh))
			}
			downscaleNearest(canvas, work)
		} else {
			work = canvas
		}

		cells := allocCells(cols * rows * 2)
		buildGrid(work, cells, cols, rows)

		delay := int(g.Delay[i]) * 10
		if delay < 20 {
			delay = 100
		}

		frames = append(frames, Frame{
			Cells: cells,
			Cols:  cols,
			Rows:  rows,
			Delay: delay,
		})

		g.Image[i] = nil

		switch g.Disposal[i] {
		case gif.DisposalBackground:
			bg := color.RGBA{0, 0, 0, 0}
			if int(g.BackgroundIndex) < len(src.Palette) {
				bg = color.RGBAModel.Convert(src.Palette[g.BackgroundIndex]).(color.RGBA)
			}
			for y := srcBounds.Min.Y; y < srcBounds.Max.Y; y++ {
				for x := srcBounds.Min.X; x < srcBounds.Max.X; x++ {
					canvas.Set(x, y, bg)
				}
			}
		case gif.DisposalPrevious:
			if prevSnap != nil {
				draw.Draw(canvas, canvas.Bounds(), prevSnap, image.Point{}, draw.Src)
			}
		}
	}

	return frames, nil
}

func buildGrid(src *image.RGBA, cells []uint32, cols, rows int) {
	b := src.Bounds()
	sw := b.Dx()
	sh := b.Dy()
	if sw == 0 || sh == 0 {
		return
	}

	wf := float64(sw - 1)
	hf := float64(sh - 1)

	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			idx := (row*cols + col) * 2
			nx := float64(col) / float64(cols)
			nyTop := float64(row) / float64(rows)
			nyBot := float64(row*2+1) / float64(rows*2)

			tr, tg, tb := sampleBilinearRGB(src, nx, nyTop, wf, hf)
			br, bg, bb := sampleBilinearRGB(src, nx, nyBot, wf, hf)

			cells[idx+cellTopIdx] = packRGB(tr, tg, tb)
			cells[idx+cellBotIdx] = packRGB(br, bg, bb)
		}
	}
}

func sampleBilinearRGB(img *image.RGBA, nx, ny, wf, hf float64) (uint8, uint8, uint8) {
	x := nx * wf
	y := ny * hf
	ix := int(x)
	iy := int(y)
	fx := x - float64(ix)
	fy := y - float64(iy)

	if ix < 0 {
		ix = 0
	}
	if iy < 0 {
		iy = 0
	}
	maxX := int(wf)
	maxY := int(hf)
	if ix >= maxX {
		ix = maxX
	}
	if iy >= maxY {
		iy = maxY
	}

	ix0 := ix
	ix1 := ix + 1
	iy0 := iy
	iy1 := iy + 1
	if ix1 > maxX {
		ix1 = maxX
	}
	if iy1 > maxY {
		iy1 = maxY
	}

	c00 := img.RGBAAt(ix0, iy0)
	c10 := img.RGBAAt(ix1, iy0)
	c01 := img.RGBAAt(ix0, iy1)
	c11 := img.RGBAAt(ix1, iy1)

	return uint8(blerp(float64(c00.R), float64(c10.R), float64(c01.R), float64(c11.R), fx, fy)),
		uint8(blerp(float64(c00.G), float64(c10.G), float64(c01.G), float64(c11.G), fx, fy)),
		uint8(blerp(float64(c00.B), float64(c10.B), float64(c01.B), float64(c11.B), fx, fy))
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
