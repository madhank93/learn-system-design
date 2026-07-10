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

// marker returns the kubelings-style progress glyph for an exercise.
func (m Model) marker(ex exercise.Exercise) string {
	switch {
	case m.St.Completed[ex.ID]:
		return stylePass.Render("✓")
	case m.St.Attempted[ex.ID]:
		return styleHint.Render("◐")
	default:
		return styleDim.Render("◌")
	}
}

func (m Model) View() string {
	if m.Quitting {
		return "tearing down… run `syslings` to resume where you left off\n"
	}

	var b strings.Builder

	if m.TopicOpen {
		b.WriteString(styleHeader.Render("🧪 syslings — pick a topic") + "\n\n")
		for i, t := range m.Topics {
			doneN := 0
			for _, ex := range t.Exs {
				if t.St.Completed[ex.ID] {
					doneN++
				}
			}
			line := fmt.Sprintf("   %-18s %d/%d", t.Name, doneN, len(t.Exs))
			if i == m.TopicCursor {
				line = styleHeader.Render(fmt.Sprintf("▶  %-18s %d/%d", t.Name, doneN, len(t.Exs)))
			}
			b.WriteString(line + "\n")
			b.WriteString(styleDim.Render("      "+t.Desc) + "\n")
		}
		b.WriteString("\n" + styleDim.Render("↑/↓ move   ⏎ start (boots the Docker stack)   q quit") + "\n")
		return b.String()
	}

	if m.ShowHelp {
		return m.helpView()
	}

	// ── header ──
	done := 0
	for _, ex := range m.Exs {
		if m.St.Completed[ex.ID] {
			done++
		}
	}
	stack := styleDim.Render("stack ○ down (⏎ or u to start)")
	if m.Booting {
		stack = styleHint.Render("stack ◌ starting…")
	} else if m.StackUp {
		stack = stylePass.Render("stack ● up")
	}
	fmt.Fprintf(&b, "%s  %s  %s\n\n",
		styleHeader.Render("🧪 syslings — "+m.Topic),
		styleDim.Render(fmt.Sprintf("%d/%d", done, len(m.Exs))),
		stack)

	if m.BootErr != nil {
		b.WriteString(styleFail.Render("stack failed to start: "+m.BootErr.Error()) + "\n")
		b.WriteString(styleDim.Render("is Docker running?  u retry   q quit") + "\n")
		return b.String()
	}

	// ── exercise list ──
	for i, ex := range m.Exs {
		play := " "
		if i == m.Playing {
			play = "▸"
		}
		row := fmt.Sprintf("  %s %s  %02d  %s", play, m.marker(ex), i+1, ex.Title)
		if i == m.Cursor {
			row = styleHeader.Render(fmt.Sprintf("▶ %s %s  %02d  %s", play, m.markerPlain(ex), i+1, ex.Title))
		}
		b.WriteString(row + "\n")
	}

	// ── detail pane for the selected exercise ──
	b.WriteString("\n")
	sel := m.selected()
	if sel.Desc != "" {
		b.WriteString(styleDim.Render(sel.Desc) + "\n")
	}
	if sel.Type == exercise.TypeFix {
		b.WriteString(styleDim.Render("📁 edit: work/"+m.Topic+"/"+sel.File) + "\n")
	}

	switch {
	case m.Phase == PhaseSettingUp && m.Playing == m.Cursor:
		b.WriteString(styleDim.Render("setting up the scenario…") + "\n")
	case m.Phase == PhaseChecking && m.Playing == m.Cursor:
		b.WriteString(styleDim.Render("checking…") + "\n")
	case m.Phase == PhaseFailed && m.Playing == m.Cursor && m.Output != "":
		b.WriteString(styleFail.Render(strings.TrimRight(m.Output, "\n")) + "\n")
	case m.Phase == PhasePassed && m.Output != "" && m.Playing == -1 && m.St.Completed[sel.ID] && done == len(m.Exs):
		b.WriteString(stylePass.Render(strings.TrimRight(m.Output, "\n")) + "\n\n")
		b.WriteString(stylePass.Render("All exercises complete! 🎉") + "\n")
	case m.Phase == PhasePassed && m.Output != "":
		b.WriteString(stylePass.Render(strings.TrimRight(m.Output, "\n")) + "\n")
		b.WriteString(styleDim.Render("⏎ starts the highlighted exercise") + "\n")
	}

	for i := 0; i < m.Hints[sel.ID] && i < len(sel.Hints); i++ {
		b.WriteString(styleHint.Render("💡 "+sel.Hints[i]) + "\n")
	}

	if m.ConfirmSolItem >= 0 {
		b.WriteString("\n" + styleHint.Render("apply the shipped solution to \""+m.Exs[m.ConfirmSolItem].Title+"\"? [y/N]") + "\n")
	}

	watching := ""
	if m.Playing >= 0 {
		watching = "   watching " + m.playing().ID + "…"
	}
	b.WriteString("\n" + styleDim.Render("⏎ play  h hint  e edit  r reset  s solution  n next  u/d stack  ? help  q quit"+watching) + "\n")
	return b.String()
}

// markerPlain is marker without color styling (for use inside styled rows).
func (m Model) markerPlain(ex exercise.Exercise) string {
	switch {
	case m.St.Completed[ex.ID]:
		return "✓"
	case m.St.Attempted[ex.ID]:
		return "◐"
	default:
		return "◌"
	}
}

func (m Model) helpView() string {
	return styleHeader.Render("🧪 syslings — keys") + "\n\n" +
		`  ↑/↓ j/k   move between exercises
  ⏎ space   play: boot stack if needed → inject fault → check → watch saves
  h         reveal the next hint for the selected exercise
  e         open the exercise file in $EDITOR
  r         reset the exercise file to its broken template (and re-check)
  s         apply the shipped solution (asks first)
  n         jump to the next unfinished exercise
  u / d     start / stop the topic's Docker stack
  q         quit (stack is torn down)

  markers:  ` + styleDim.Render("◌ not started") + `  ` + styleHint.Render("◐ attempted") + `  ` + stylePass.Render("✓ solved") + `
` + "\n" + styleDim.Render("any key to close") + "\n"
}
