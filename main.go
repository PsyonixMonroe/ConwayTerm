package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const DOT = '⊡'
const SQR = '▣'
const HLW = '□'
const CURS = '█' // TODO: Replace with underline style or something

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

type ConwayGame struct {
	board       [][]int8
	writeBoard  [][]int8
	tickSpeedMS int
	tick        int
	state       int
	cursorX     int
	cursorY     int
	changes     int
}

func initializeGame() ConwayGame {
	// make the board
	sizeY := 30
	sizeX := 100
	createBoard := make([][]int8, 0)
	writeBoard := make([][]int8, 0)
	for range sizeY {
		createBoard = append(createBoard, make([]int8, sizeX))
		writeBoard = append(writeBoard, make([]int8, sizeX))
	}
	for y := range sizeY {
		for x := range sizeX {
			createBoard[y][x] = CELL_DEAD
			writeBoard[y][x] = CELL_DEAD
		}
	}
	return ConwayGame{
		board:       createBoard,
		writeBoard:  writeBoard,
		tickSpeedMS: 250,
		tick:        0,
		state:       STATE_PREINIT,
		cursorX:     0,
		cursorY:     0,
		changes:     0,
	}
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
		case "ctrl_c", "q":
			return c, tea.Quit
		case "left", "h":
			if c.cursorX > 0 {
				c.cursorX--
			}
		case "right", "l":
			if c.cursorX < len(c.board[0]) {
				c.cursorX++
			}
		case "up", "k":
			if c.cursorY > 0 {
				c.cursorY--
			}
		case "down", "j":
			if c.cursorY < len(c.board) {
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
	s := strings.Builder{}

	s.WriteString(fmt.Sprintf("Cursor: (%d, %d)\n", c.cursorY, c.cursorX))
	s.WriteString(fmt.Sprintf("Cursor Cell: %s\n", c.cellState(c.cursorY, c.cursorX)))
	s.WriteString(fmt.Sprintf("Neighbor Count: %d\n", c.getNeighborCount(c.cursorY, c.cursorX)))
	s.WriteString(fmt.Sprintf("Tick: %d\n", c.tick))
	s.WriteString(fmt.Sprintf("State: %s\n", c.getState()))

	// drop top boarder
	s.WriteRune('-') // left hand column header
	for range len(c.board[0]) {
		s.WriteRune('-')
	}
	s.WriteRune('-') // right hand column header
	s.WriteRune('\n')

	for y, row := range c.board {
		s.WriteRune('|')
		for x, item := range row {
			if y == c.cursorY && x == c.cursorX {
				// cursor
				s.WriteRune(SQR)
				continue
			}
			if item == CELL_DEAD || item == CELL_DYING {
				// dead cell
				s.WriteRune(HLW)
				continue
			}
			// live cell
			s.WriteRune(DOT)
		}
		s.WriteRune('|')
		s.WriteRune('\n')
	}

	// drop bottom boarder
	s.WriteRune('-') // left hand column footer
	for range len(c.board[0]) {
		s.WriteRune('-')
	}
	s.WriteRune('-') // right hand column footer
	return s.String()
}

func main() {
	p := tea.NewProgram(initializeGame())
	if _, err := p.Run(); err != nil {
		fmt.Printf("There has been an error: %v\n", err)
		os.Exit(1)
	}
}
