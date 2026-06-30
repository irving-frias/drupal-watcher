package gifbg

import "sync"

var cellSlicePool = sync.Pool{
	New: func() any {
		s := make([]uint32, 0, 80*24*2)
		return &s
	},
}

func allocCells(n int) []uint32 {
	ptr := cellSlicePool.Get().(*[]uint32)
	if n > cap(*ptr) {
		cellSlicePool.Put(ptr)
		s := make([]uint32, n)
		return s
	}
	s := (*ptr)[:n]
	clear(s)
	return s
}

func freeCells(s []uint32) {
	s = s[:0]
	cellSlicePool.Put(&s)
}

func packRGB(r, g, b uint8) uint32 {
	return uint32(r)<<16 | uint32(g)<<8 | uint32(b)
}

func unpackRGB(packed uint32) (uint8, uint8, uint8) {
	return uint8(packed >> 16), uint8(packed >> 8), uint8(packed)
}

const (
	cellTopIdx = 0
	cellBotIdx = 1
)

type Frame struct {
	Cells []uint32
	Cols  int
	Rows  int
	Delay int
}

func (f *Frame) TopAt(col, row int) (uint8, uint8, uint8) {
	idx := (row*f.Cols + col) * 2
	return unpackRGB(f.Cells[idx+cellTopIdx])
}

func (f *Frame) BottomAt(col, row int) (uint8, uint8, uint8) {
	idx := (row*f.Cols + col) * 2
	return unpackRGB(f.Cells[idx+cellBotIdx])
}

func (f *Frame) setCell(col, row int, tr, tg, tb, br, bg, bb uint8) {
	idx := (row*f.Cols + col) * 2
	f.Cells[idx+cellTopIdx] = packRGB(tr, tg, tb)
	f.Cells[idx+cellBotIdx] = packRGB(br, bg, bb)
}

func (f *Frame) Release() {
	if f.Cells != nil {
		freeCells(f.Cells)
		f.Cells = nil
	}
}
