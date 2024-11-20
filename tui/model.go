package tui

import (
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
    // store Store
    // textarea.Model
    // ...
}

func NewModel(store *Store) model {
    return model{
        state: listView,
        store: store,
    }
}

func (m model) Init() tea.Cmd {
    return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    return m, nil
}

