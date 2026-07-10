package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/madhank93/learn-system-design/syslings/internal/exercise"
)

var (
	styleHeader = lipgloss.NewStyle().Bold(true)
	styleDim    = lipgloss.NewStyle().Faint(true)
	styleFail   = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	stylePass   = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	styleHint   = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
)

func (m Model) View() string {
	if m.Quitting {
		return "tearing down… run `syslings " + m.Topic + "` to resume where you left off\n"
	}

	var b strings.Builder

	if m.TopicOpen {
		b.WriteString(styleHeader.Render("🧪 syslings — pick a topic") + "\n\n")
		for i, t := range m.Topics {
			n := len(t.Exs)
			doneN := 0
			for _, ex := range t.Exs {
				if t.St.Completed[ex.ID] {
					doneN++
				}
			}
			line := fmt.Sprintf("   %-18s %d/%d", t.Name, doneN, n)
			if i == m.TopicCursor {
				line = styleHeader.Render(fmt.Sprintf("▶  %-18s %d/%d", t.Name, doneN, n))
			}
			b.WriteString(line + "\n")
			b.WriteString(styleDim.Render("      "+t.Desc) + "\n")
		}
		b.WriteString("\n" + styleDim.Render("↑/↓ move   enter start (boots the Docker stack)   q quit") + "\n")
		return b.String()
	}

	if m.BootErr != nil {
		b.WriteString(styleFail.Render("stack failed to start: "+m.BootErr.Error()) + "\n\n")
		b.WriteString(styleDim.Render("is Docker running?   [q]uit") + "\n")
		return b.String()
	}
	if m.Booting {
		b.WriteString(styleHeader.Render("🧪 syslings — "+m.Topic) + "\n\n")
		b.WriteString(styleDim.Render("🐳 starting the stack (first run builds images)…") + "\n")
		return b.String()
	}

	done := 0
	for _, ex := range m.Exs {
		if m.St.Completed[ex.ID] {
			done++
		}
	}
	fmt.Fprintf(&b, "%s %s\n\n",
		styleHeader.Render(fmt.Sprintf("🧪 syslings — %s", m.Topic)),
		styleDim.Render(fmt.Sprintf("%d/%d", done, len(m.Exs))))

	if m.ListOpen {
		b.WriteString(styleHeader.Render("Pick an exercise") + "\n\n")
		for i, ex := range m.Exs {
			mark := "○"
			if m.St.Completed[ex.ID] {
				mark = "✅"
			}
			line := fmt.Sprintf("   %s  %02d  %s", mark, i+1, ex.Title)
			if i == m.ListCursor {
				line = styleHeader.Render(fmt.Sprintf("▶ %s  %02d  %s", mark, i+1, ex.Title))
			}
			b.WriteString(line + "\n")
			if i == m.ListCursor && ex.Desc != "" {
				b.WriteString(styleDim.Render("        "+ex.Desc) + "\n")
			}
		}
		b.WriteString("\n" + styleDim.Render("↑/↓ move   enter start   q quit") + "\n")
		return b.String()
	}

	if m.Phase == PhaseAllDone {
		b.WriteString(stylePass.Render("All exercises complete! 🎉") + "\n\n")
		b.WriteString(styleDim.Render("[q]uit") + "\n")
		return b.String()
	}

	ex := m.current()
	fmt.Fprintf(&b, "%s %s\n", styleHeader.Render(fmt.Sprintf("Exercise %02d:", m.Cur+1)), ex.Title)
	if ex.Type == exercise.TypeFix {
		fmt.Fprintf(&b, "%s\n", styleDim.Render("📁 edit: work/"+m.Topic+"/"+ex.File))
	}
	b.WriteString("\n")

	switch m.Phase {
	case PhaseSettingUp:
		b.WriteString(styleDim.Render("setting up the scenario…") + "\n")
	case PhaseChecking:
		b.WriteString(styleDim.Render("checking…") + "\n")
	case PhaseFailed:
		b.WriteString(styleFail.Render(strings.TrimRight(m.Output, "\n")) + "\n")
	}

	for i := 0; i < m.HintsShown && i < len(ex.Hints); i++ {
		fmt.Fprintf(&b, "%s\n", styleHint.Render("💡 "+ex.Hints[i]))
	}

	b.WriteString("\n" + styleDim.Render("[h]int  [l]ist  [r]eset  [s]kip  [q]uit   watching…") + "\n")
	return b.String()
}
