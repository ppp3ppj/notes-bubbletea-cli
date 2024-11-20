package tui

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

const (
    listView uint = iota
    titleView
    bodyView
)

type model struct {
    state uint
    store *Store
    notes []Note
    currNote Note
    listIndex int
    // store Store
    // textarea.Model
    // ...
}

func NewModel(store *Store) model {
    notes, err := store.GetNotes()
    if err != nil {
        log.Fatalf("unable to get notes: %v", err)
    }

    return model{
        state: listView,
        store: store,
        notes: notes,
    }
}

func (m model) Init() tea.Cmd {
    return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    return m, nil
}

