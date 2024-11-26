package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	appNameStyle = lipgloss.NewStyle().Background(lipgloss.Color("99")).Padding(0, 1)

	faintStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Faint(true)

	enumeratorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("99")).MarginRight(1)

	editNoteStyle      = lipgloss.NewStyle().Background(lipgloss.Color("98")).Padding(0, 1)
	editTitleNoteStyle = lipgloss.NewStyle().Background(lipgloss.Color("95")).Padding(0, 1)
)

func (m model) View() string {
	header := appNameStyle.Render("NOTES APP") + "\n\n"

	if m.isLoading {
		return header +
			m.spinner.View() + " Loading..." + "\n\n"
	}

	switch m.state {
	case timeView:
		return header +
			fmt.Sprintf(
				"What’s your time?\n\n%s\n\n%s",
				m.textInputTime.View(),
				faintStyle.Render("enter - next(now save), esc - quit"),
			) + "\n"

	case titleView:
		return header +
			"Note title:\n\n" +
			m.textInput.View() + "\n\n" +
			faintStyle.Render("enter - save, esc - discard")

	case bodyView:
		noteDetails := editNoteStyle.Render("Note:") + "\n\n" +
			editTitleNoteStyle.Render(m.currNote.Title) + "\n\n" +
			m.textArea.View() + "\n\n"

		if m.isEditing {
			noteDetails += faintStyle.Render("Created At: ") + faintStyle.Render(m.currNote.CreatedAt.Format("2006-01-02 15:04:05")) + "\n" +
				faintStyle.Render("Updated At: ") + faintStyle.Render(m.currNote.UpdatedAt.Format("2006-01-02 15:04:05")) + "\n"
		}

		return header + noteDetails + faintStyle.Render("tab - next, esc - discard")

	case listView:
		var notesList string
		for i, n := range m.notes {
			prefix := " "
			if i == m.listIndex {
				prefix = ">"
			}

			shortBody := strings.ReplaceAll(n.Body, "\n", " ")
			if len(shortBody) > 30 {
				shortBody = shortBody[:30] + "..." // Add ellipsis for truncated body
			}

			notesList += enumeratorStyle.Render(prefix) + n.Title + " | " + faintStyle.Render(shortBody) + "\n\n"
		}
		// Conditionally add the "d - delete" option if there is more than one note
		deleteOption := ""
		if len(m.notes) >= 1 {
			deleteOption = faintStyle.Render("d - delete") + ", "
		}

		newNoteOption := faintStyle.Render("n - new note") + ", "
		exitCliOption := faintStyle.Render("q - quit")

		return header + notesList + newNoteOption + deleteOption + exitCliOption
	}

	return header // Fallback to header if no state matches
}
