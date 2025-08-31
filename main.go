package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const DOT = "⊡"
const SQR = "▣"
const HLW = "□"
const CURS = "█" // TODO: Replace with underline style or something

const (
	STATE_PREINIT = iota
	STATE_INIT
	STATE_STOPPED
	STATE_RUNNING
)

const (
	CELL_ALIVE = iota
	CELL_NEW
	CELL_DYING
	CELL_DEAD
)

const HEADER_ROWS = 4

var (
	aliveStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("45"))
	newStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("77"))
	dyingStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("124"))
	deadStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			Padding(0, 0).
			Foreground(lipgloss.Color("8"))
)

type ConwayGame struct {
	board       [][]int8
	writeBoard  [][]int8
	width       int
	height      int
	tickSpeedMS int
	tick        int
	state       int
	cursorX     int
	cursorY     int
	changes     int
}

func New(tickSpeed int) ConwayGame {
	return ConwayGame{
		board:       [][]int8{},
		writeBoard:  [][]int8{},
		width:       0,
		height:      0,
		tickSpeedMS: tickSpeed,
		tick:        0,
		state:       STATE_PREINIT,
		cursorX:     0,
		cursorY:     0,
	}
}

func (c ConwayGame) initializeGame() ([][]int8, [][]int8) {
	// make the board
	createBoard := make([][]int8, 0)
	writeBoard := make([][]int8, 0)
	for range c.height {
		createBoard = append(createBoard, make([]int8, c.width))
		writeBoard = append(writeBoard, make([]int8, c.width))
	}
	for y := range c.height {
		for x := range c.width {
			createBoard[y][x] = CELL_DEAD
			writeBoard[y][x] = CELL_DEAD
		}
	}

	return createBoard, writeBoard
}

type triggerMsg int

func (c ConwayGame) trigger() tea.Msg {
	var d time.Duration = time.Duration(c.tickSpeedMS) * time.Millisecond
	time.Sleep(d)
	return triggerMsg(1)
}

func (c ConwayGame) Init() tea.Cmd {
	return c.trigger
}

func (c ConwayGame) isAlive(y int, x int) bool {
	// assumes y and x are in bounds
	cell := c.board[y][x]
	return cell == CELL_ALIVE || cell == CELL_NEW
}

func (c ConwayGame) getNeighborCount(y int, x int) int {
	// count neighbors
	neighbors := 0
	// row above
	if y > 0 {
		if x > 0 {
			// top left
			if c.isAlive(y-1, x-1) {
				neighbors++
			}
		}
		// top center
		if c.isAlive(y-1, x) {
			neighbors++
		}
		// top right
		if x < len(c.board[y])-1 {
			if c.isAlive(y-1, x+1) {
				neighbors++
			}
		}
	}

	// same row
	if x > 0 {
		// left
		if c.isAlive(y, x-1) {
			neighbors++
		}
	}
	if x < len(c.board[y])-1 {
		// right
		if c.isAlive(y, x+1) {
			neighbors++
		}
	}

	// row below
	if y < len(c.board)-1 {
		if x > 0 {
			// bottom left
			if c.isAlive(y+1, x-1) {
				neighbors++
			}
		}
		// top center
		if c.isAlive(y+1, x) {
			neighbors++
		}
		// top right
		if x < len((c.board)[y])-1 {
			if c.isAlive(y+1, x+1) {
				neighbors++
			}
		}
	}

	return neighbors
}

func (c ConwayGame) checkCell(y int, x int) int8 {

	neighbors := c.getNeighborCount(y, x)
	// determine fate of cell
	currentAlive := c.isAlive(y, x)
	var newCellState int8
	if currentAlive {
		if neighbors < 2 {
			// underpop
			newCellState = CELL_DYING
		} else if neighbors == 2 || neighbors == 3 {
			// good pop
			newCellState = CELL_ALIVE
		} else {
			// overpop
			newCellState = CELL_DYING
		}
	} else {
		// dead cell will live if 3 neighbors
		if neighbors == 3 {
			newCellState = CELL_NEW
		} else {
			newCellState = CELL_DEAD
		}
	}

	return newCellState
}

func (c ConwayGame) UpdateBoard(writeBoard *[][]int8) {
	for y, row := range c.board {
		for x := range row {
			(*writeBoard)[y][x] = c.checkCell(y, x)
		}
	}
}

func (c ConwayGame) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case triggerMsg:
		if c.state == STATE_RUNNING {
			c.UpdateBoard(&c.writeBoard)
			tmp := c.board
			c.board = c.writeBoard
			c.writeBoard = tmp
			c.tick++
		}
		return c, c.trigger
	case tea.KeyMsg:
		switch msg.String() {
		// test the key
		case tea.KeyCtrlC.String(), "q":
			return c, tea.Quit
		case "left", "h":
			if c.cursorX > 0 {
				c.cursorX--
			}
		case "right", "l":
			if c.cursorX < len(c.board[0])-1 {
				c.cursorX++
			}
		case "up", "k":
			if c.cursorY > 0 {
				c.cursorY--
			}
		case "down", "j":
			if c.cursorY < len(c.board)-1 {
				c.cursorY++
			}
		case "enter", "m":
			c.board[c.cursorY][c.cursorX] = CELL_NEW
		case "backspace", "u":
			c.board[c.cursorY][c.cursorX] = CELL_DEAD
		case "s":
			c.state = STATE_STOPPED

		case "r":
			c.state = STATE_RUNNING
		}
	case tea.WindowSizeMsg:
		if c.width < 1 {
			// remove boarder rows
			c.width = msg.Width - 2
			// remove header rows and boarder rows
			c.height = msg.Height - 2 - HEADER_ROWS
			c.board, c.writeBoard = c.initializeGame()
		}
	}
	return c, nil
}

func (c ConwayGame) getState() string {
	switch c.state {
	case STATE_PREINIT:
		return "PREINIT"
	case STATE_INIT:
		return "INIT"
	case STATE_STOPPED:
		return "STOPPED"
	case STATE_RUNNING:
		return "RUNNING"
	default:
		return "UNKNOWN"
	}
}

func (c ConwayGame) cellState(y int, x int) string {
	switch c.board[y][x] {
	case CELL_ALIVE:
		return "ALIVE"
	case CELL_DEAD:
		return "DEAD"
	case CELL_DYING:
		return "DYING"
	case CELL_NEW:
		return "NEW"
	}
	return "UNKNOWN"
}

func (c ConwayGame) View() string {
	if c.width < 1 || c.height < 1 {
		return "Loading..."
	}

	s := strings.Builder{}
	header := strings.Builder{}

	header.WriteString(fmt.Sprintf("Cursor: (%d, %d)\n", c.cursorY, c.cursorX))
	header.WriteString(fmt.Sprintf("Cursor Cell: %s\n", c.cellState(c.cursorY, c.cursorX)))
	// s.WriteString(fmt.Sprintf("Neighbor Count: %d\n", c.getNeighborCount(c.cursorY, c.cursorX)))
	header.WriteString(fmt.Sprintf("Tick: %d\n", c.tick))
	header.WriteString(fmt.Sprintf("State: %s", c.getState()))

	for y, row := range c.board {
		for x, item := range row {
			if y == c.cursorY && x == c.cursorX {
				// cursor
				s.WriteString(cursorStyle.Render(SQR))
				continue
			}
			// dead cell
			if item == CELL_DEAD {
				s.WriteString(deadStyle.Render(HLW))
				continue
			}
			if item == CELL_DYING {
				s.WriteString(dyingStyle.Render(HLW))
				continue
			}
			// live cell
			if item == CELL_ALIVE {
				s.WriteString(aliveStyle.Render(DOT))
				continue
			}
			if item == CELL_NEW {
				s.WriteString(newStyle.Render(DOT))
				continue
			}
		}
		if y != len(c.board)-1 {
			s.WriteRune('\n')
		}
	}

	return borderStyle.Render(header.String() + "\n" + s.String())
}

func main() {
	sModel := NewSelectorModel()
	s := tea.NewProgram(&sModel)

	if _, err := s.Run(); err != nil {
		fmt.Printf("There has been an error: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(New(sModel.TickSpeedMS))
	if _, err := p.Run(); err != nil {
		fmt.Printf("There has been an error: %v\n", err)
		os.Exit(1)
	}
}
