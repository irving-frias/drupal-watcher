package gifbg

import (
	"fmt"
	"image"
	"image/color"
	"strings"
)

func renderHalfBlock(rgba *image.RGBA, cols, rows int) []string {
	b := rgba.Bounds()
	sw := b.Dx()
	sh := b.Dy()
	if sw == 0 || sh == 0 {
		return nil
	}

	grid := make([]string, rows)
	for row := 0; row < rows; row++ {
		var buf strings.Builder
		for col := 0; col < cols; col++ {
			nx := float64(col) / float64(cols)
			nyTop := float64(row) / float64(rows)
			nyBot := float64(row*2+1) / float64(rows*2)

			top := sampleBilinear(rgba, nx, nyTop)
			bottom := sampleBilinear(rgba, nx, nyBot)

			if top == bottom {
				buf.WriteString(fmt.Sprintf("\x1b[48;2;%d;%d;%dm ", top.R, top.G, top.B))
			} else {
				buf.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm▀",
					top.R, top.G, top.B, bottom.R, bottom.G, bottom.B))
			}
		}
		buf.WriteString("\x1b[0m")
		grid[row] = buf.String()
	}
	return grid
}

func sampleBilinear(img *image.RGBA, nx, ny float64) color.RGBA {
	b := img.Bounds()
	w := float64(b.Dx())
	h := float64(b.Dy())
	if w == 0 || h == 0 {
		return color.RGBA{}
	}

	x := nx * (w - 1)
	y := ny * (h - 1)

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
	maxX := b.Dx() - 1
	maxY := b.Dy() - 1
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

	c00 := rgbaToFloat(img.RGBAAt(ix0, iy0))
	c10 := rgbaToFloat(img.RGBAAt(ix1, iy0))
	c01 := rgbaToFloat(img.RGBAAt(ix0, iy1))
	c11 := rgbaToFloat(img.RGBAAt(ix1, iy1))

	return color.RGBA{
		R: uint8(blerp(c00.R, c10.R, c01.R, c11.R, fx, fy)),
		G: uint8(blerp(c00.G, c10.G, c01.G, c11.G, fx, fy)),
		B: uint8(blerp(c00.B, c10.B, c01.B, c11.B, fx, fy)),
		A: 255,
	}
}

type floatColor struct {
	R, G, B float64
}

func rgbaToFloat(c color.RGBA) floatColor {
	return floatColor{float64(c.R), float64(c.G), float64(c.B)}
}

func blerp(c00, c10, c01, c11 float64, fx, fy float64) float64 {
	return c00*(1-fx)*(1-fy) + c10*fx*(1-fy) + c01*(1-fx)*fy + c11*fx*fy
}

func pixelAt(img *image.RGBA, x, y, w, h int) color.RGBA {
	if x < 0 {
		x = 0
	}
	if x >= w {
		x = w - 1
	}
	if y < 0 {
		y = 0
	}
	if y >= h {
		y = h - 1
	}
	r, g, b, _ := img.At(x, y).RGBA()
	return color.RGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: 255,
	}
}

func (bg *Background) RowBGColor(row int) color.RGBA {
	if bg == nil || !bg.active || bg.numFrames == 0 {
		return color.RGBA{10, 10, 20, 255}
	}
	rgba := bg.frames[bg.frameIdx].RGBA
	if rgba == nil {
		return color.RGBA{10, 10, 20, 255}
	}
	sw := rgba.Bounds().Dx()
	sh := rgba.Bounds().Dy()
	if sw == 0 || sh == 0 {
		return color.RGBA{10, 10, 20, 255}
	}

	ny := float64(row) / float64(bg.rows)
	return sampleBilinear(rgba, 0, ny)
}

func BGSequence(c color.RGBA) string {
	return fmt.Sprintf("\x1b[48;2;%d;%d;%dm", c.R, c.G, c.B)
}

func (bg *Background) AverageColor() color.RGBA {
	if bg == nil || !bg.active || bg.numFrames == 0 {
		return color.RGBA{10, 10, 20, 255}
	}
	rgba := bg.frames[bg.frameIdx].RGBA
	if rgba == nil {
		return color.RGBA{10, 10, 20, 255}
	}
	b := rgba.Bounds()
	total := uint64(b.Dx() * b.Dy())
	if total == 0 {
		return color.RGBA{10, 10, 20, 255}
	}
	var rSum, gSum, bSum uint64
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			px := rgba.RGBAAt(x, y)
			rSum += uint64(px.R)
			gSum += uint64(px.G)
			bSum += uint64(px.B)
		}
	}
	return color.RGBA{
		R: uint8(rSum / total),
		G: uint8(gSum / total),
		B: uint8(bSum / total),
		A: 255,
	}
}
