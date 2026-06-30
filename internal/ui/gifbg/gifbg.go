package gifbg

type Background struct {
	path      string
	frames    []frameData
	numFrames int
	frameIdx  int
	cols      int
	rows      int

	active  bool
	enabled bool
	lastErr string
	dirty   bool
}

func New(path string) *Background {
	bg := &Background{active: false, enabled: false}
	bg.load(path)
	return bg
}

func (bg *Background) load(path string) {
	bg.active = false
	bg.enabled = false
	bg.lastErr = ""

	if path == "" || path == "default" {
		frames := defaultFrames()
		if frames == nil {
			bg.lastErr = "failed to generate default frames"
			return
		}
		bg.frames = frames
		bg.numFrames = len(frames)
		bg.path = ""
		bg.frameIdx = 0
		bg.cols = 80
		bg.rows = 24
		bg.active = true
		bg.enabled = true
		return
	}

	frames, err := decodeGIF(path)
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
	bg.cols = 80
	bg.rows = 24
	bg.active = true
	bg.enabled = true
}

func (bg *Background) Load(path string) {
	if bg == nil {
		return
	}
	bg.load(path)
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
	bg.frameIdx = (bg.frameIdx + 1) % bg.numFrames
	bg.dirty = true
}

func (bg *Background) Resize(cols, rows int) {
	if bg == nil || !bg.active || bg.numFrames == 0 {
		return
	}
	bg.cols = cols
	bg.rows = rows
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
	rgba := bg.frames[bg.frameIdx].RGBA
	if rgba == nil {
		return nil
	}
	return renderHalfBlock(rgba, cols, rows)
}
