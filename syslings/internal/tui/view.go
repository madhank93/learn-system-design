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

	done := 0
	for _, ex := range m.Exs {
		if m.St.Completed[ex.ID] {
			done++
		}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s %s\n\n",
		styleHeader.Render(fmt.Sprintf("🧪 syslings — %s", m.Topic)),
		styleDim.Render(fmt.Sprintf("%d/%d", done, len(m.Exs))))

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

	b.WriteString("\n" + styleDim.Render("[h]int  [r]eset  [s]kip  [q]uit   watching…") + "\n")
	return b.String()
}
