package tui

import (
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
)

const (
	listView uint = iota
	titleView
	bodyView
)

type model struct {
	state     uint
	store     *Store
	notes     []Note
	currNote  Note
	listIndex int
	textArea  textarea.Model
	textInput textinput.Model
	isEditing bool

	spinner   spinner.Model
	isLoading bool
}

// Custom message for loading notes
type notesLoadedMsg struct {
	notes []Note
}

type saveCompleteMsg struct {
	notes []Note
}

func NewModel(store *Store) model {
	notes, err := store.GetNotes()
	if err != nil {
		log.Fatalf("unable to get notes: %v", err)
	}

	spin := spinner.New()
	spin.Spinner = spinner.Dot

	return model{
		state:     listView,
		store:     store,
		notes:     notes,
		textArea:  textarea.New(),
		textInput: textinput.New(),
		spinner:   spin,
		isLoading: false,
	}
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds []tea.Cmd
		cmd  tea.Cmd
	)

	m.textInput, cmd = m.textInput.Update(msg)
	cmds = append(cmds, cmd)

	m.textArea, cmd = m.textArea.Update(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case spinner.TickMsg:
		if m.isLoading {
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case notesLoadedMsg:
		// Update notes after loading completes
		m.notes = msg.notes
		m.isLoading = false

	case saveCompleteMsg:
		// Handle save completion
		m.notes = msg.notes
		m.isLoading = false
		m.isEditing = false
		m.currNote = Note{}
		m.state = listView

	case tea.KeyMsg:
		key := msg.String() //up, down, etc ...
		switch m.state {
		case listView:
			if m.isLoading {
               // Don't allow interaction with the list when loading
			   return m, tea.Batch(cmds...)
			}
			switch key {
			case "q":
				return m, tea.Quit
			case "n":
				m.textInput.SetValue("")
				m.textInput.Focus()
				m.currNote = Note{}
				m.state = titleView
			case "up", "k":
				if m.listIndex > 0 {
					m.listIndex--
				}
			case "down", "j":
				if m.listIndex < len(m.notes)-1 {
					m.listIndex++
				}
			case "enter":
				m.currNote = m.notes[m.listIndex]
				m.textArea.SetValue(m.currNote.Body)
				m.textArea.Focus()
				m.textArea.CursorEnd()

				m.isEditing = m.currNote.Id != "" // Set if editing

				m.state = bodyView

			case "r":
				m.isLoading = true
				//return m, m.spinner.Tick
				return m, tea.Batch(
					m.spinner.Tick,
					func() tea.Msg {
						// Simulate a delay (e.g., fetching notes)
						time.Sleep(1 * time.Second)
						newNotes, err := m.store.GetNotes()
						if err != nil {
							// Handle error (for simplicity, quit)
							return tea.Quit
						}
						return notesLoadedMsg{notes: newNotes}
					},
				)
			}
		case titleView:
			switch key {
			case "enter":
				title := m.textInput.Value()
				if title != "" {
					m.currNote.Title = title
					m.textArea.SetValue("")
					m.textArea.Focus()
					m.textArea.CursorEnd()

					m.state = bodyView
				}
			case "esc":
				m.state = listView
			}
		case bodyView:
			switch key {
			case "ctrl+s":
				body := m.textArea.Value()
				m.currNote.Body = body

				// Start loading spinner
				m.isLoading = true

				return m, tea.Batch(
					m.spinner.Tick,
					func() tea.Msg {
						err := m.store.SaveNote(m.currNote)
						if err != nil {
							// Handle save error (simplified for example)
							return tea.Quit
						}
						newNotes, err := m.store.GetNotes()
						if err != nil {
							// Handle fetch error (simplified for example)
							return tea.Quit
						}

						// Simulate load operation with a delay
						time.Sleep(400 * time.Millisecond) // Simulated delay
						return saveCompleteMsg{notes: newNotes}
					},
				)
			case "esc":
				m.isEditing = false // Reset editing case
				m.state = listView
			}
		}
	}
	return m, tea.Batch(cmds...)
}


