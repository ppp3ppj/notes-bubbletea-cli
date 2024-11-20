package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ppp3ppj/notes-bubbletea-cli/tui"
)

func main() {
    m := tui.NewModel()
    p := tea.NewProgram(m)
    if _, err := p.Run(); err != nil {
        log.Fatalf("unable to run tui: %v", err)
    }
}
