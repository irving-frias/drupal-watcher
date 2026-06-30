package gifbg

import "fmt"

const (
	defaultCols = 80
	defaultRows = 24
	frameWindow = 15
)

type Background struct {
	path      string
	frames    []Frame
	numFrames int
	frameIdx  int
	cols      int
	rows      int

	active       bool
	enabled      bool
	lastErr      string
	dirty        bool

	streaming    bool
	windowStart  int
}

func New(path string) *Background {
	bg := &Background{active: false, enabled: false}
	bg.load(path, defaultCols, defaultRows)
	return bg
}

func (bg *Background) load(path string, cols, rows int) {
	bg.active = false
	bg.enabled = false
	bg.lastErr = ""

	if cols < 1 {
		cols = defaultCols
	}
	if rows < 1 {
		rows = defaultRows
	}

	if path == "" || path == "default" {
		frames := defaultFrames(cols, rows)
		if frames == nil {
			bg.lastErr = "failed to generate default frames"
			return
		}
		bg.frames = frames
		bg.numFrames = len(frames)
		bg.path = ""
		bg.frameIdx = 0
		bg.cols = cols
		bg.rows = rows
		bg.active = true
		bg.enabled = true
		bg.streaming = false
		return
	}

	frames, err := decodeGIF(path, cols, rows)
	if err != nil {
		bg.lastErr = err.Error()
		return
	}
	if len(frames) == 0 {
		bg.lastErr = "GIF has no frames"
		return
	}

	bg.frames = frames
	bg.numFrames = len(frames)
	bg.path = path
	bg.frameIdx = 0
	bg.cols = cols
	bg.rows = rows
	bg.active = true
	bg.enabled = true
	bg.streaming = false
}

func (bg *Background) Load(path string) {
	if bg == nil {
		return
	}
	bg.load(path, bg.cols, bg.rows)
}

func (bg *Background) Active() bool {
	return bg != nil && bg.active
}

func (bg *Background) Enabled() bool {
	return bg != nil && bg.active && bg.enabled
}

func (bg *Background) SetEnabled(on bool) {
	if bg == nil || !bg.active {
		return
	}
	if bg.enabled == on {
		return
	}
	bg.enabled = on
	if on {
		bg.dirty = true
	}
}

func (bg *Background) LastError() string {
	if bg == nil {
		return ""
	}
	return bg.lastErr
}

func (bg *Background) FrameIndex() int {
	if bg == nil {
		return 0
	}
	return bg.frameIdx
}

func (bg *Background) NumFrames() int {
	if bg == nil {
		return 0
	}
	return bg.numFrames
}

func (bg *Background) Path() string {
	if bg == nil {
		return ""
	}
	return bg.path
}

func (bg *Background) FrameDuration(frameIdx int) int {
	if bg == nil || frameIdx < 0 || frameIdx >= bg.numFrames {
		return 100
	}
	return bg.frames[frameIdx].Delay
}

func (bg *Background) NextFrame() {
	if bg == nil || !bg.active || bg.numFrames == 0 || !bg.enabled {
		return
	}

	prevIdx := bg.frameIdx
	bg.frameIdx = (bg.frameIdx + 1) % bg.numFrames

	if bg.numFrames > frameWindow {
		if prevIdx >= 0 {
			bg.frames[prevIdx].Release()
		}
	}

	bg.dirty = true
}

func (bg *Background) Resize(cols, rows int) {
	if bg == nil || !bg.active || bg.numFrames == 0 {
		return
	}
	if cols == bg.cols && rows == bg.rows {
		return
	}

	bg.cols = cols
	bg.rows = rows

	if bg.path == "" || bg.path == "default" {
		bg.frames = defaultFrames(cols, rows)
	} else {
		bg.load(bg.path, cols, rows)
	}

	if bg.enabled {
		bg.dirty = true
	}
}

func (bg *Background) FrameRows(cols, rows int) []string {
	if bg == nil || !bg.active || !bg.enabled || bg.numFrames == 0 {
		return nil
	}
	if cols != bg.cols || rows != bg.rows {
		bg.cols = cols
		bg.rows = rows
	}
	bg.dirty = false
	return bg.renderGrid(cols, rows)
}

func (bg *Background) renderGrid(cols, rows int) []string {
	f := &bg.frames[bg.frameIdx]
	if f.Cells == nil {
		return nil
	}

	if cols == f.Cols && rows == f.Rows {
		return renderFrame(f, cols, rows)
	}

	return renderFrameSubsample(f, cols, rows)
}

func (bg *Background) ReleaseAll() {
	if bg == nil || bg.frames == nil {
		return
	}
	for i := range bg.frames {
		bg.frames[i].Release()
	}
	bg.frames = nil
	bg.numFrames = 0
}

func (bg *Background) ResetAnimation() {
	bg.frameIdx = 0
	bg.dirty = true
}

func (bg *Background) String() string {
	if bg == nil || !bg.active {
		return "gif: inactive"
	}
	return fmt.Sprintf("gif: %d frames @ %dx%d [frame %d]",
		bg.numFrames, bg.cols, bg.rows, bg.frameIdx)
}
