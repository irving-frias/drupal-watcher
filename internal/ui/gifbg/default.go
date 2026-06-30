package gifbg

import (
	"math"
)

const (
	defCount = 30
	defDelay = 100
)

func defaultFrames(cols, rows int) []Frame {
	if cols < 1 {
		cols = 80
	}
	if rows < 1 {
		rows = 24
	}

	numFrames := defCount
	frames := make([]Frame, 0, numFrames)

	for i := 0; i < numFrames; i++ {
		cells := allocCells(cols * rows * 2)
		phase := float64(i) / float64(numFrames) * 2 * math.Pi

		for y := 0; y < rows; y++ {
			for x := 0; x < cols; x++ {
				nx := float64(x) / float64(cols)
				ny := float64(y) / float64(rows)

				topV := math.Sin(phase+nx*3.0+ny*2.0)*0.5 + 0.5
				topV *= 0.3

				botV := math.Sin(phase*1.3+nx*2.0+ny*3.0)*0.5 + 0.5
				botV *= 0.2

				tr := uint8(topV * 25)
				tg := uint8(topV*15 + botV*10)
				tb := uint8(topV*80 + botV*40 + 15)

				nx2 := float64(x*2+1) / float64(cols*2)
				ny2 := float64(y) / float64(rows)
				bv := math.Sin(phase+nx2*3.0+ny2*2.0)*0.5 + 0.5
				bv *= 0.3
				bu := math.Sin(phase*1.3+nx2*2.0+ny2*3.0)*0.5 + 0.5
				bu *= 0.2

				br := uint8(bv * 25)
				bg := uint8(bv*15 + bu*10)
				bb := uint8(bv*80 + bu*40 + 15)

				idx := (y*cols + x) * 2
				cells[idx+cellTopIdx] = packRGB(tr, tg, tb)
				cells[idx+cellBotIdx] = packRGB(br, bg, bb)
			}
		}

		frames = append(frames, Frame{
			Cells: cells,
			Cols:  cols,
			Rows:  rows,
			Delay: defDelay,
		})
	}

	return frames
}
