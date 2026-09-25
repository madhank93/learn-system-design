package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/madhank93/learn-system-design/syslings/internal/exercise"
)

func (m Model) marker(t TopicEntry, ex exercise.Exercise) (string, lipgloss.Style) {
	switch {
	case t.St.Completed[ex.ID]:
		return "✓", solvedStyle
	case t.St.Attempted[ex.ID]:
		return "◐", startedStyle
	default:
		return "◌", noneStyle
	}
}

func badge(ex exercise.Exercise) string {
	if ex.Type == exercise.TypeScenario {
		return scenarioStyle.Render("◇scenario")
	}
	return fixStyle.Render("✎fix")
}

func badgeText(ex exercise.Exercise) string {
	if ex.Type == exercise.TypeScenario {
		return "◇scenario"
	}
	return "✎fix"
}

func (m Model) View() string {
	if m.Quitting {
		return "tearing down… run `syslings` to resume where you left off\n"
	}
	if !m.ready {
		return "loading…"
	}
	if m.Splash {
		return m.splashView()
	}

	var b strings.Builder
	b.WriteString(m.headerBar() + "\n")
	left := m.listView()
	right := paneStyle.Render(m.rightPane())
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, left, right) + "\n")
	b.WriteString(m.footer())
	return b.String()
}

func (m Model) headerBar() string {
	var dot string
	switch {
	case m.Booting:
		dot = warnStyle.Render(m.spin.View() + " starting " + m.pendingOrPlayingTopic() + "…")
	case m.ActiveTopic != "":
		dot = okStyle.Render("● up · " + m.ActiveTopic)
	default:
		dot = dimStyle.Render("○ down — ↵ boots it")
	}
	return titleStyle.Render("🧪 syslings") + "   stack: " + dot
}

func (m Model) pendingOrPlayingTopic() string {
	if t, _ := m.exAt(m.Playing); t.Name != "" {
		return t.Name
	}
	if r := m.selRef(); r.T >= 0 {
		return m.Topics[r.T].Name
	}
	return ""
}

func (m Model) listView() string {
	lw := leftWidth(m.w)
	inner := lw - 4
	bodyH := m.vp.Height
	var lines []string
	for i := range m.rows {
		if len(lines) >= bodyH {
			break
		}
		r := m.rows[i]
		if r.header != "" {
			done, total := m.topicProgress(r.header)
			lines = append(lines, moduleStyle.Render(truncate(r.header, inner-6))+dimStyle.Render(fmt.Sprintf(" %d/%d", done, total)))
			continue
		}
		t, ex := m.exAt(r.ref)
		mk, mkStyle := m.marker(t, ex)
		playMark := " "
		if r.ref == m.Playing {
			playMark = "▸"
		}
		bt := badgeText(ex)
		nameW := inner - 5 - len([]rune(bt))
		name := truncate(ex.Title, nameW)
		if m.isCursor(i) {
			lines = append(lines, cursorStyle.Render(padRight(" "+playMark+mk+" "+name+" "+bt, inner)))
		} else {
			lines = append(lines, " "+playMark+mkStyle.Render(mk)+" "+textStyle.Render(name)+" "+badge(ex))
		}
	}
	for len(lines) < bodyH {
		lines = append(lines, "")
	}
	return paneStyle.Width(lw - 2).Height(bodyH).Render(strings.Join(lines, "\n"))
}

func (m Model) topicProgress(name string) (int, int) {
	for _, t := range m.Topics {
		if t.Name == name {
			done := 0
			for _, ex := range t.Exs {
				if t.St.Completed[ex.ID] {
					done++
				}
			}
			return done, len(t.Exs)
		}
	}
	return 0, 0
}

func (m Model) isCursor(rowIdx int) bool {
	return len(m.sel) > 0 && m.sel[m.Cursor] == rowIdx
}

func (m Model) rightPane() string {
	m.vp.SetContent(m.rightContent())
	return m.vp.View()
}

func (m Model) rightContent() string {
	switch m.Mode {
	case modeHelp:
		return helpText()
	case modeOutput:
		return m.outputView()
	default:
		return m.detail()
	}
}

func (m Model) outputView() string {
	t, ex := m.exAt(m.Playing)
	title := ex.Title
	if title == "" {
		_, sel := m.exAt(m.selRef())
		title = sel.Title
	}
	var b strings.Builder
	b.WriteString(titleStyle.Render(title) + "\n\n")
	switch m.Phase {
	case PhaseSettingUp:
		b.WriteString(startedStyle.Render(m.spin.View()+" injecting the fault…") + "\n")
	case PhaseChecking:
		b.WriteString(startedStyle.Render(m.spin.View()+" checking against live traffic…") + "\n")
	case PhaseFailed:
		b.WriteString(colorizeCheck(m.Output) + "\n")
		if ex.Type == exercise.TypeFix {
			b.WriteString("\n" + dimStyle.Render("edit ") + linkStyle.Render("work/"+t.Name+"/"+ex.File) +
				dimStyle.Render(" and save — the check reruns on its own") + "\n")
		}
	case PhasePassed:
		b.WriteString(colorizeCheck(m.Output) + "\n\n")
		if m.allDone() {
			b.WriteString(okStyle.Render("All exercises complete! 🎉") + "\n")
		} else {
			b.WriteString(keybar([2]string{"↵", "start the next highlighted exercise"}) + "\n")
		}
	}
	if k := m.hintKey(m.playingOrSel()); m.Hints[k] > 0 {
		_, hx := m.exAt(m.playingOrSel())
		b.WriteString("\n")
		for i := 0; i < m.Hints[k] && i < len(hx.Hints); i++ {
			b.WriteString(warnStyle.Render("💡 ") + textStyle.Render(hx.Hints[i]) + "\n")
		}
	}
	return b.String()
}

func (m Model) playingOrSel() ref {
	if m.Playing.T >= 0 {
		return m.Playing
	}
	return m.selRef()
}

// colorizeCheck paints check-script output: ✓ lines green, ✗ red, hints amber.
func colorizeCheck(out string) string {
	var lines []string
	for _, l := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		trimmed := strings.TrimSpace(l)
		switch {
		case strings.HasPrefix(trimmed, "✓"):
			lines = append(lines, okStyle.Render(l))
		case strings.HasPrefix(trimmed, "✗"):
			lines = append(lines, failStyle.Render(l))
		case strings.HasPrefix(trimmed, "hint:"):
			lines = append(lines, warnStyle.Render(l))
		default:
			lines = append(lines, textStyle.Render(l))
		}
	}
	return strings.Join(lines, "\n")
}

func (m Model) allDone() bool {
	for _, t := range m.Topics {
		for _, ex := range t.Exs {
			if !t.St.Completed[ex.ID] {
				return false
			}
		}
	}
	return true
}

func (m Model) detail() string {
	r := m.selRef()
	t, ex := m.exAt(r)
	if ex.ID == "" {
		return "no exercises found"
	}
	mk, mkStyle := m.marker(t, ex)
	var b strings.Builder
	b.WriteString(titleStyle.Render(ex.Title) + "\n\n")
	if ex.Desc != "" {
		b.WriteString(textStyle.Render(ex.Desc) + "\n\n")
	}
	b.WriteString(dimStyle.Render("exercise:  ") + textStyle.Render(ex.ID) + "\n")
	b.WriteString(dimStyle.Render("status:    ") + mkStyle.Render(mk+" "+statusWord(t, ex)) + "\n")
	if ex.Type == exercise.TypeFix {
		b.WriteString(dimStyle.Render("type:      ") + fixStyle.Render("✎ fix — edit the broken config until the check passes") + "\n")
		b.WriteString(dimStyle.Render("edit:      ") + linkStyle.Render("work/"+t.Name+"/"+ex.File) + "\n")
	} else {
		b.WriteString(dimStyle.Render("type:      ") + scenarioStyle.Render("◇ scenario — a live fault you remediate on the running stack") + "\n")
	}
	b.WriteString(dimStyle.Render("hints:     ") + textStyle.Render(fmt.Sprintf("%d available", len(ex.Hints))) +
		dimStyle.Render(" — press ") + keyStyle.Render("h") + "\n\n")

	b.WriteString(headerStyle.Render("Stack") + "\n")
	if m.ActiveTopic == t.Name {
		b.WriteString(dimStyle.Render("  docker:  ") + okStyle.Render("● up · "+t.Name) + "\n")
	} else {
		b.WriteString(dimStyle.Render("  docker:  ") + dimStyle.Render("○ down") + textStyle.Render(" — ") +
			keyStyle.Render("↵") + textStyle.Render(" boots it, persists until ") + keyStyle.Render("d") + "\n")
	}
	b.WriteString(dimStyle.Render("  watch:   ") + textStyle.Render("saving a work file reruns the active exercise's check") + "\n\n")
	b.WriteString(keybar([2]string{"↵", "play (boot + fault + check + watch)"}) + "\n")
	b.WriteString(keybar([2]string{"h", "hint"}, [2]string{"e", "edit"}, [2]string{"r", "reset"}, [2]string{"s", "solution"}))
	return b.String()
}

func statusWord(t TopicEntry, ex exercise.Exercise) string {
	switch {
	case t.St.Completed[ex.ID]:
		return "solved"
	case t.St.Attempted[ex.ID]:
		return "started"
	default:
		return "not started"
	}
}

func (m Model) footer() string {
	if m.ConfirmSol {
		_, ex := m.exAt(m.selRef())
		return warnStyle.Render("apply the shipped solution to \""+ex.Title+"\"? ") +
			keybar([2]string{"y", "yes"}, [2]string{"N", "no"}) + "\n"
	}
	keys := keybar(
		[2]string{"↵", "play"}, [2]string{"h", "hint"}, [2]string{"e", "edit"}, [2]string{"r", "reset"},
		[2]string{"s", "solution"}, [2]string{"n", "next"}, [2]string{"u", "up"}, [2]string{"d", "down"},
		[2]string{"?", "help"}, [2]string{"q", "quit"})
	status := ""
	if t, ex := m.exAt(m.Playing); ex.ID != "" && !m.busy() {
		status = dimStyle.Render("watching ") + textStyle.Render(t.Name+"/"+ex.ID) + dimStyle.Render(" — save to re-check")
	}
	if m.BootErr != nil {
		status = failStyle.Render("stack failed: "+truncate(m.BootErr.Error(), 60)) + dimStyle.Render(" — is Docker running?")
	}
	return keys + "\n" + status
}

func (m Model) splashView() string {
	var b strings.Builder
	b.WriteString("🧪  " + titleStyle.Render("syslings") + "\n")
	b.WriteString(dimStyle.Render("Learn system design the rustlings way — fix broken configs against a live Docker stack") + "\n\n")

	row := func(k, v string) string { return dimStyle.Render(padRight(k, 12)) + v + "\n" }
	b.WriteString(row("Repo", linkStyle.Render("https://github.com/madhank93/learn-system-design")))
	b.WriteString(row("Site", linkStyle.Render("http://localhost:4321 — docs, sims & guides")))
	b.WriteString("\n")

	b.WriteString(headerStyle.Render("How it works") + "\n")
	b.WriteString(dimStyle.Render("  • ") + textStyle.Render("Pick an exercise, press ") + keyStyle.Render("↵") +
		textStyle.Render(" — boots the topic's ") + keyStyle.Render("Docker") + textStyle.Render(" stack with a") + "\n")
	b.WriteString(dimStyle.Render("    ") + textStyle.Render("deliberately broken config (or injects a live fault)") + "\n")
	b.WriteString(dimStyle.Render("  • ") + textStyle.Render("Edit the file in ") + linkStyle.Render("syslings/work/") +
		textStyle.Render(" — every save reruns the check") + "\n")
	b.WriteString(dimStyle.Render("  • ") + textStyle.Render("Go green, advance   ") +
		noneStyle.Render("◌") + dimStyle.Render(" not started · ") + startedStyle.Render("◐") + dimStyle.Render(" started · ") +
		solvedStyle.Render("✓") + dimStyle.Render(" solved") + "\n\n")

	b.WriteString(dimStyle.Render("Keys  ") + keybar(
		[2]string{"↑↓/jk", "move"}, [2]string{"↵", "play"}, [2]string{"h", "hint"}, [2]string{"e", "edit"},
		[2]string{"r", "reset"}, [2]string{"s", "solution"}, [2]string{"d", "stack down"}, [2]string{"q", "quit"}) + "\n\n")
	b.WriteString(warnStyle.Render("press any key to start →"))

	box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("87")).Padding(1, 3).Render(b.String())
	return lipgloss.Place(m.w, m.h, lipgloss.Center, lipgloss.Center, box)
}

func helpText() string {
	row := func(k, d string) string { return "  " + keyStyle.Render(padRight(k, 9)) + textStyle.Render(d) + "\n" }
	return headerStyle.Render("Keys") + "\n\n" +
		row("↵ / space", "play — boot stack (if needed) + inject fault + check + watch") +
		row("h", "reveal the next hint for the selected exercise") +
		row("e", "open the exercise file in $EDITOR") +
		row("r", "reset the file to its broken template (and re-check)") +
		row("s", "apply the shipped solution (asks to confirm)") +
		row("n", "jump to the next unfinished exercise") +
		row("u / d", "stack up / down") +
		row("a", "about / welcome screen") +
		row("esc", "back to exercise detail") +
		row("? / q", "toggle help / quit") + "\n" +
		dimStyle.Render("markers: ") + noneStyle.Render("◌ not started") + dimStyle.Render(" · ") +
		startedStyle.Render("◐ started") + dimStyle.Render(" · ") + solvedStyle.Render("✓ solved") + "\n" +
		dimStyle.Render("types:   ") + fixStyle.Render("✎fix edit the broken config") + dimStyle.Render(" · ") +
		scenarioStyle.Render("◇scenario remediate a live fault")
}
