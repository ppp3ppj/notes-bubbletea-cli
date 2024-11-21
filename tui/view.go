package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
    appNameStyle = lipgloss.NewStyle().Background(lipgloss.Color("99")).Padding(0, 1)

    faintStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Faint(true)

    enumeratorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("99")).MarginRight(1)

    editNoteStyle = lipgloss.NewStyle().Background(lipgloss.Color("98")).Padding(0, 1)
    editTitleNoteStyle = lipgloss.NewStyle().Background(lipgloss.Color("95")).Padding(0, 1)

)

func (m model) View() string {
    s := appNameStyle.Render("NOTES APP") + "\n\n"

    if m.state == titleView {
        s += "Note title:\n\n"
        s += m.textInput.View() + "\n\n"
        s += faintStyle.Render("enter - save, esc - discard")
    }

    if m.state == bodyView {
        s += editNoteStyle.Render("Note:") + "\n\n"
        s += editTitleNoteStyle.Render(m.currNote.Title) + "\n\n"
        s += m.textArea.View() + "\n\n"
        s += faintStyle.Render("ctrl+s - save, esc - discard")
    }

    if m.state == listView {
        for i, n := range m.notes {
            prefix := " "
            if i == m.listIndex {
                prefix = ">"
            }

            shortBody := strings.ReplaceAll(n.Body, "\n", " ")
            if len(shortBody) > 30 {
                shortBody = shortBody[:30]
            }
            s += enumeratorStyle.Render(prefix) + n.Title + " | " + faintStyle.Render(shortBody) + "\n\n"
        }
        s += faintStyle.Render("n - new note, q - quit")
    }

    return s
}
