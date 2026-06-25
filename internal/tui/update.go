package tui

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/irving-frias/drupal-watcher/internal/watcher"
)

func fmtDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		vpHeight := msg.Height - 4 - 3 // status(2) + input(1) + borders & padding
		if vpHeight < 3 {
			vpHeight = 3
		}
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = vpHeight
		m.input.Width = msg.Width - 6
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "ctrl+d":
			return m, tea.Quit
		case "enter":
			cmd := strings.TrimSpace(m.input.Value())
			m.input.SetValue("")
			if cmd == "" {
				return m, nil
			}
			return m, m.executeCommand(cmd)
		default:
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}

	case tickMsg:
		uptime := time.Since(m.Watcher.Stats.StartTime)
		var allocMB float64
		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)
		allocMB = float64(mem.Alloc) / 1024 / 1024

		m.status = statusLine{
			PID:        os.Getpid(),
			Uptime:     fmtDuration(uptime),
			Changes:    m.Watcher.Stats.Changes.Load(),
			Clears:     m.Watcher.Stats.Clears.Load(),
			WatchCount: m.Watcher.WatchCount.Load(),
			AllocMB:    allocMB,
			Running:    true,
		}
		m.viewport.SetContent(m.renderEvents())
		return m, tickCmd()

	case watcherEventMsg:
		evt := msg.Event
		ts := evt.Timestamp.Format("15:04:05")
		switch evt.Type {
		case watcher.EventChange:
			line := fmt.Sprintf("Change detected: %s", evt.File)
			if evt.Changes > 1 {
				line = fmt.Sprintf("%d changes detected (last: %s)", evt.Changes, evt.File)
			}
			m.pushEvent(eventLine{
				Timestamp: ts,
				Content:   line,
				Style:     infoStyle,
			})
		case watcher.EventDrush:
			icon := successStyle
			if evt.ExitCode != 0 {
				icon = errorStyle
			}
			line := fmt.Sprintf("drush %s (%v, exit %d)", evt.Commands, evt.Duration.Round(time.Millisecond), evt.ExitCode)
			m.pushEvent(eventLine{
				Timestamp: ts,
				Content:   line,
				Style:     icon,
			})
		case watcher.EventError:
			m.pushEvent(eventLine{
				Timestamp: ts,
				Content:   fmt.Sprintf("Error: %v", evt.Error),
				Style:     errorStyle,
			})
		}
		m.viewport.SetContent(m.renderEvents())
		m.viewport.GotoBottom()
		return m, listenForEvents(m.Watcher)

	case errMsg:
		m.pushEvent(eventLine{
			Timestamp: time.Now().Format("15:04:05"),
			Content:   fmt.Sprintf("Error: %v", msg.Err),
			Style:     errorStyle,
		})
		return m, nil
	}

	return m, nil
}

func (m *Model) executeCommand(cmd string) tea.Cmd {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return nil
	}

	switch parts[0] {
	case "status":
		uptime := time.Since(m.Watcher.Stats.StartTime)
		m.pushEvent(eventLine{
			Timestamp: time.Now().Format("15:04:05"),
			Content: fmt.Sprintf("PID %d | Changes: %d | Clears: %d | Uptime: %s",
				os.Getpid(), m.Watcher.Stats.Changes.Load(), m.Watcher.Stats.Clears.Load(), fmtDuration(uptime)),
			Style: infoStyle,
		})
	case "help":
		m.pushEvent(eventLine{
			Timestamp: time.Now().Format("15:04:05"),
			Content: "Commands: status, stats, help, stop, quit",
			Style:   infoStyle,
		})
	case "stop", "quit", "exit":
		return tea.Quit
	default:
		m.pushEvent(eventLine{
			Timestamp: time.Now().Format("15:04:05"),
			Content:   fmt.Sprintf("Unknown: %s. Type help.", cmd),
			Style:     warnStyle,
		})
	}
	return nil
}
