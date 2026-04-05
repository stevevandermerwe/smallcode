package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbletea"

	"smallcode/api"
)

func main() {
	p := tea.NewProgram(api.NewModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
