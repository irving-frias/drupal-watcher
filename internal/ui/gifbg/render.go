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
			topY := row * sh / rows
			bottomY := (row*2 + 1) * sh / (rows * 2)
			topX := col * sw / cols

			top := pixelAt(rgba, topX, topY, sw, sh)
			bottom := pixelAt(rgba, topX, bottomY, sw, sh)

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

	y := row * sh / bg.rows
	return pixelAt(rgba, 0, y, sw, sh)
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
