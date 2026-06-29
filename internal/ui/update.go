package ui

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/irving-frias/drupal-watcher/internal/utils"
	"github.com/irving-frias/drupal-watcher/pkg/core"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		cw := msg.Width - 2
		statusStyle = statusStyle.Width(cw)
		eventsStyle = eventsStyle.Width(cw)
		cmdStyle = cmdStyle.Width(cw)
		helpStyle = helpStyle.Width(cw - 4)

		vpHeight := msg.Height - 9
		if vpHeight < 5 {
			vpHeight = 5
		}
		m.viewport.Width = cw - 4
		m.viewport.Height = vpHeight
		m.input.Width = cw - 6
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "ctrl+d":
			return m, tea.Quit

		case "?":
			m.showHelp = !m.showHelp
			return m, nil

		case "esc":
			if m.showHelp {
				m.showHelp = false
				return m, nil
			}

		case "home":
			m.viewport.GotoTop()
			return m, nil

		case "end":
			if !m.showHelp {
				m.autoScroll = !m.autoScroll
				if m.autoScroll {
					m.viewport.GotoBottom()
				}
			}
			return m, nil

		case "pgup", "pgdown":
			if !m.showHelp {
				m.autoScroll = false
				var cmd tea.Cmd
				m.viewport, cmd = m.viewport.Update(msg)
				return m, cmd
			}

		case "tab":
			m.completeInput()
			return m, nil

		case "up":
			if m.showHelp {
				return m, nil
			}
			if len(m.history) == 0 {
				return m, nil
			}
			if m.historyIdx == -1 {
				m.historyIdx = len(m.history) - 1
			} else if m.historyIdx > 0 {
				m.historyIdx--
			} else {
				return m, nil
			}
			m.input.SetValue(m.history[m.historyIdx])
			m.input.CursorEnd()
			return m, nil

		case "down":
			if m.showHelp {
				return m, nil
			}
			if m.historyIdx == -1 {
				return m, nil
			}
			if m.historyIdx < len(m.history)-1 {
				m.historyIdx++
				m.input.SetValue(m.history[m.historyIdx])
				m.input.CursorEnd()
			} else {
				m.historyIdx = -1
				m.input.SetValue("")
			}
			return m, nil

		case "enter":
			if m.showHelp {
				m.showHelp = false
				return m, nil
			}
			cmd := strings.TrimSpace(m.input.Value())
			m.input.SetValue("")
			m.historyIdx = -1
			m.completions = nil
			if cmd == "" {
				return m, nil
			}
			m.addToHistory(cmd)
			return m, m.executeCommand(cmd)

		default:
			if msg.String() != "" {
				m.completions = nil
			}
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}

	case tea.MouseMsg:
		if !m.showHelp {
			m.autoScroll = false
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}

	case tickMsg:
		m.updateStatus()
		m.viewport.SetContent(m.renderEvents())
		if m.autoScroll {
			m.viewport.GotoBottom()
		}
		return m, tickCmd()

	case engineEventMsg:
		evt := msg.Event
		ts := evt.Timestamp.Format("15:04:05")
		switch evt.Type {
		case core.EventChange:
			line := fmt.Sprintf("Change detected: %s", evt.File)
			if evt.Changes > 1 {
				line = fmt.Sprintf("%d changes detected (last: %s)", evt.Changes, evt.File)
			}
			m.pushEvent(eventLine{
				Timestamp: ts,
				Content:   line,
				Style:     infoStyle,
			})
		case core.EventCacheClear:
			icon := successStyle
			if evt.ExitCode != 0 {
				icon = errorStyle
			}
			tag := ""
			if evt.SiteName != "" {
				tag = " [" + evt.SiteName + "]"
				m.siteClears[evt.SiteName]++
			}
			line := fmt.Sprintf("drush %s%s (%v, exit %d)", evt.Commands, tag, evt.Duration.Round(time.Millisecond), evt.ExitCode)
			if evt.Stderr != "" {
				line += "\n" + dim.Render(evt.Stderr)
			}
			m.pushEvent(eventLine{
				Timestamp: ts,
				Content:   line,
				Style:     icon,
			})
		case core.EventError:
			content := fmt.Sprintf("Error: %v", evt.Error)
			if evt.File != "" {
				content = fmt.Sprintf("Error in %s: %v", evt.File, evt.Error)
			}
			m.pushEvent(eventLine{
				Timestamp: ts,
				Content:   content,
				Style:     errorStyle,
			})
		}
		m.viewport.SetContent(m.renderEvents())
		if m.autoScroll {
			m.viewport.GotoBottom()
		}
		return m, listenForEvents(m.eventChan)
	}

	return m, nil
}

func openURL(url string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("open", url)
		case "linux":
			cmd = exec.Command("xdg-open", url)
		case "windows":
			cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
		default:
			cmd = exec.Command("open", url)
		}
		_ = cmd.Start()
		return nil
	}
}

func (m *Model) completeInput() {
	input := strings.TrimSpace(m.input.Value())
	parts := strings.Fields(input)

	if m.completions != nil {
		m.completionIdx = (m.completionIdx + 1) % len(m.completions)
		newVal := m.completions[m.completionIdx]
		if len(parts) >= 2 && parts[0] == "filter" {
			newVal = "filter " + newVal
		} else {
			newVal += " "
		}
		m.input.SetValue(newVal)
		m.input.CursorEnd()
		return
	}

	if input == "" {
		m.completions = commands
		m.completionIdx = 0
		m.input.SetValue(m.completions[m.completionIdx] + " ")
		m.input.CursorEnd()
		return
	}

	if len(parts) == 1 && !strings.HasSuffix(input, " ") {
		prefix := parts[0]
		var matches []string
		for _, cmd := range commands {
			if strings.HasPrefix(cmd, prefix) && cmd != prefix {
				matches = append(matches, cmd)
			}
		}
		if len(matches) > 0 {
			m.completions = matches
			m.completionIdx = 0
			m.input.SetValue(m.completions[m.completionIdx] + " ")
			m.input.CursorEnd()
		}
		return
	}

	if len(parts) >= 2 && parts[0] == "filter" && !strings.HasSuffix(input, " ") {
		prefix := parts[1]
		var matches []string
		for name := range m.siteClears {
			if strings.HasPrefix(name, prefix) {
				matches = append(matches, name)
			}
		}
		if len(matches) > 0 {
			m.completions = matches
			m.completionIdx = 0
			m.input.SetValue("filter " + m.completions[m.completionIdx])
			m.input.CursorEnd()
		}
		return
	}
}

func (m *Model) executeCommand(cmd string) tea.Cmd {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return nil
	}

	switch parts[0] {
	case "status":
		changes, clears := m.engineInfo.Stats()
		uptime := time.Since(m.engineInfo.StartTime())
		m.pushEvent(eventLine{
			Timestamp: time.Now().Format("15:04:05"),
			Content: fmt.Sprintf("PID %d | Changes: %d | Clears: %d | Uptime: %s",
				m.status.PID, changes, clears, utils.FormatDuration(uptime)),
			Style: infoStyle,
		})
	case "help":
		m.showHelp = true
	case "filter":
		if len(parts) > 1 {
			m.siteFilter = parts[1]
			m.pushEvent(eventLine{
				Timestamp: time.Now().Format("15:04:05"),
				Content:   fmt.Sprintf("Filtering events by site: %s", cyan.Render(m.siteFilter)),
				Style:     infoStyle,
			})
		} else {
			m.siteFilter = ""
			m.pushEvent(eventLine{
				Timestamp: time.Now().Format("15:04:05"),
				Content:   "Filter cleared",
				Style:     infoStyle,
			})
		}
	case "stats":
		if len(m.siteClears) == 0 {
			m.pushEvent(eventLine{
				Timestamp: time.Now().Format("15:04:05"),
				Content:   "No site-specific clears yet",
				Style:     infoStyle,
			})
		} else {
			var sb strings.Builder
			sb.WriteString("Clears per site: ")
			first := true
			for name, count := range m.siteClears {
				if !first {
					sb.WriteString(", ")
				}
				sb.WriteString(fmt.Sprintf("%s: %d", cyan.Render(name), count))
				first = false
			}
			m.pushEvent(eventLine{
				Timestamp: time.Now().Format("15:04:05"),
				Content:   sb.String(),
				Style:     infoStyle,
			})
		}
	case "star":
		m.pushEvent(eventLine{
			Timestamp: time.Now().Format("15:04:05"),
			Content:   "Opening " + cyan.Render("github.com/irving-frias/drupal-watcher") + " in your browser...",
			Style:     infoStyle,
		})
		return openURL("https://github.com/irving-frias/drupal-watcher")

	case "dismiss":
		m.showStar = false
		writeStarDismissed(m.root)
		m.pushEvent(eventLine{
			Timestamp: time.Now().Format("15:04:05"),
			Content:   "Star banner dismissed permanently. Type " + cyan.Render("star") + " to reopen.",
			Style:     infoStyle,
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
