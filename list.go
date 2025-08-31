package main

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	itemStyle     = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("#FFFFFF"))
	selectedStyle = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("#0000FF"))
)

type SelectorModel struct {
	TickSpeedMS int
	cursor      int
	selected    int
}

type Speed struct {
	display string
	speed   int
}

func NewSelectorModel() SelectorModel {
	return SelectorModel{
		TickSpeedMS: -1,
		cursor:      0,
		selected:    -1,
	}
}

var speeds = []Speed{
	{"1s", 1000},
	{".5s", 500},
	{".25s", 250},
	{".125s", 125},
	{".1s", 100},
	{".05s", 50},
}

func (s *SelectorModel) Init() tea.Cmd {
	return nil
}

func (s *SelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			s.TickSpeedMS = -1
			return s, tea.Quit
		case "up", "k":
			// move up
			if s.cursor > 0 {
				s.cursor--
			}
		case "down", "j":
			//move down
			if s.cursor < len(speeds)-1 {
				s.cursor++
			}
		case "enter", "space":
			// select
			s.TickSpeedMS = speeds[s.cursor].speed
			s.selected = s.cursor
			return s, tea.Quit
		}
	}

	return s, nil
}

func (s *SelectorModel) View() string {

	b := strings.Builder{}

	b.WriteString("What Speed Would You Like the Game to run at?\n")

	for i, speed := range speeds {
		if i == s.cursor {
			b.WriteString(selectedStyle.Render(speed.display))
			b.WriteString("\n")
			continue
		}
		b.WriteString(itemStyle.Render(speed.display))
		b.WriteString("\n")
	}

	return b.String()
}
