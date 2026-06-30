package gifbg

import (
	"fmt"
	"image"
	"image/color"
	"strings"
)

func renderFrame(f *Frame, cols, rows int) []string {
	grid := make([]string, rows)
	for row := 0; row < rows; row++ {
		var buf strings.Builder
		base := row * cols * 2
		for col := 0; col < cols; col++ {
			topPacked := f.Cells[base+col*2+cellTopIdx]
			botPacked := f.Cells[base+col*2+cellBotIdx]

			if topPacked == botPacked {
				r, g, b := unpackRGB(topPacked)
				buf.WriteString(fmt.Sprintf("\x1b[48;2;%d;%d;%dm ", r, g, b))
			} else {
				tr, tg, tb := unpackRGB(topPacked)
				br, bg, bb := unpackRGB(botPacked)
				buf.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm▀",
					tr, tg, tb, br, bg, bb))
			}
		}
		buf.WriteString("\x1b[0m")
		grid[row] = buf.String()
	}
	return grid
}

func renderFrameSubsample(f *Frame, cols, rows int) []string {
	srcCols := f.Cols
	srcRows := f.Rows
	if srcCols == 0 || srcRows == 0 {
		return nil
	}

	grid := make([]string, rows)
	for row := 0; row < rows; row++ {
		var buf strings.Builder
		sy := row * srcRows / rows
		if sy >= srcRows {
			sy = srcRows - 1
		}
		base := sy * srcCols * 2

		for col := 0; col < cols; col++ {
			sx := col * srcCols / cols
			if sx >= srcCols {
				sx = srcCols - 1
			}

			topPacked := f.Cells[base+sx*2+cellTopIdx]
			botPacked := f.Cells[base+sx*2+cellBotIdx]

			if topPacked == botPacked {
				r, g, b := unpackRGB(topPacked)
				buf.WriteString(fmt.Sprintf("\x1b[48;2;%d;%d;%dm ", r, g, b))
			} else {
				tr, tg, tb := unpackRGB(topPacked)
				br, bg, bb := unpackRGB(botPacked)
				buf.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm▀",
					tr, tg, tb, br, bg, bb))
			}
		}
		buf.WriteString("\x1b[0m")
		grid[row] = buf.String()
	}
	return grid
}

func (bg *Background) RowBGColor(row int) color.RGBA {
	if bg == nil || !bg.active || bg.numFrames == 0 {
		return color.RGBA{10, 10, 20, 255}
	}
	f := &bg.frames[bg.frameIdx]
	if f.Cells == nil || f.Cols == 0 {
		return color.RGBA{10, 10, 20, 255}
	}

	sy := row * f.Rows / bg.rows
	if sy >= f.Rows {
		sy = f.Rows - 1
	}
	if sy < 0 {
		sy = 0
	}

	r, g, b := f.TopAt(0, sy)
	return color.RGBA{R: r, G: g, B: b, A: 255}
}

func BGSequence(c color.RGBA) string {
	return fmt.Sprintf("\x1b[48;2;%d;%d;%dm", c.R, c.G, c.B)
}

func (bg *Background) AverageColor() color.RGBA {
	if bg == nil || !bg.active || bg.numFrames == 0 {
		return color.RGBA{10, 10, 20, 255}
	}
	f := &bg.frames[bg.frameIdx]
	if f.Cells == nil || len(f.Cells) == 0 {
		return color.RGBA{10, 10, 20, 255}
	}

	var rSum, gSum, bSum uint64
	count := len(f.Cells)
	for i := 0; i < count; i += 2 {
		tr, tg, tb := unpackRGB(f.Cells[i])
		rSum += uint64(tr)
		gSum += uint64(tg)
		bSum += uint64(tb)
	}
	n := uint64(count / 2)
	if n == 0 {
		return color.RGBA{10, 10, 20, 255}
	}
	return color.RGBA{
		R: uint8(rSum / n),
		G: uint8(gSum / n),
		B: uint8(bSum / n),
		A: 255,
	}
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
