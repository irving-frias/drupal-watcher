package metrics

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Stats struct {
	mu sync.Mutex

	StartTime      time.Time         `json:"startTime"`
	TotalChanges   int64             `json:"totalChanges"`
	TotalClears    int64             `json:"totalClears"`
	ClearsPerSite  map[string]int64  `json:"clearsPerSite"`
	ChangesPerMin  []int64           `json:"changesPerMin"`
	Errors         int64             `json:"errors"`
	lastMinute     int
	currentMinute  int64
}

var global Stats

func Init() {
	global = Stats{
		StartTime:     time.Now(),
		ClearsPerSite: make(map[string]int64),
		ChangesPerMin: make([]int64, 0, 60),
	}
}

func RecordChange() {
	global.mu.Lock()
	defer global.mu.Unlock()
	global.TotalChanges++
	now := time.Now()
	minute := now.Minute()
	if minute == global.lastMinute {
		global.currentMinute++
	} else {
		if len(global.ChangesPerMin) >= 60 {
			global.ChangesPerMin = global.ChangesPerMin[1:]
		}
		global.ChangesPerMin = append(global.ChangesPerMin, global.currentMinute)
		global.currentMinute = 1
		global.lastMinute = minute
	}
}

func RecordClear(siteName string) {
	global.mu.Lock()
	defer global.mu.Unlock()
	global.TotalClears++
	if siteName != "" {
		global.ClearsPerSite[siteName]++
	}
}

func RecordError() {
	global.mu.Lock()
	defer global.mu.Unlock()
	global.Errors++
}

func Snapshot() SnapshotData {
	global.mu.Lock()
	defer global.mu.Unlock()
	cp := make([]int64, len(global.ChangesPerMin))
	copy(cp, global.ChangesPerMin)
	sites := make(map[string]int64, len(global.ClearsPerSite))
	for k, v := range global.ClearsPerSite {
		sites[k] = v
	}
	return SnapshotData{
		Uptime:       time.Since(global.StartTime),
		TotalChanges: global.TotalChanges,
		TotalClears:  global.TotalClears,
		ClearsPerSite: sites,
		ChangesPerMin: cp,
		CurrentMinute: global.currentMinute,
		Errors:        global.Errors,
	}
}

type SnapshotData struct {
	Uptime        time.Duration    `json:"uptime"`
	TotalChanges  int64            `json:"totalChanges"`
	TotalClears   int64            `json:"totalClears"`
	ClearsPerSite map[string]int64 `json:"clearsPerSite"`
	ChangesPerMin []int64          `json:"changesPerMin"`
	CurrentMinute int64            `json:"currentMinute"`
	Errors        int64            `json:"errors"`
}

func Save(path string) error {
	global.mu.Lock()
	snap := SnapshotData{
		Uptime:        time.Since(global.StartTime),
		TotalChanges:  global.TotalChanges,
		TotalClears:   global.TotalClears,
		ClearsPerSite: global.ClearsPerSite,
		ChangesPerMin: global.ChangesPerMin,
		CurrentMinute: global.currentMinute,
		Errors:        global.Errors,
	}
	global.mu.Unlock()

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(snap)
}
