package game

import (
	"fmt"
	"sync"
)

const INITIAL_LIVES = 6

type Cell struct {
	Value    string `json:"value"`
	Lives    int    `json:"lives"`
	WinState bool   `json:"winState"`
}
type Row [3]Cell
type Board struct {
	Cells [3]Row
	Mux   *sync.RWMutex
}

func NewBoard() *Board {
	return &Board{
		Cells: [3]Row{
			{Cell{"-", INITIAL_LIVES, false}, Cell{"-", INITIAL_LIVES, false}, Cell{"-", INITIAL_LIVES, false}},
			{Cell{"-", INITIAL_LIVES, false}, Cell{"-", INITIAL_LIVES, false}, Cell{"-", INITIAL_LIVES, false}},
			{Cell{"-", INITIAL_LIVES, false}, Cell{"-", INITIAL_LIVES, false}, Cell{"-", INITIAL_LIVES, false}},
		},
		Mux: &sync.RWMutex{},
	}
}

func CheckValidIndex(i int) bool {
	return i >= 0 && i < 3
}

func (b *Board) CheckValidInsertion(row, col int) bool {
	value, err := b.GetPoint(row, col)
	return err == nil && value == "-"
}

func (b *Board) GetPoint(row, col int) (string, error) {
	if !CheckValidIndex(row) {
		return "", fmt.Errorf("row index %d is invalid", row)
	}
	if !CheckValidIndex(col) {
		return "", fmt.Errorf("column index %d is invalid", col)
	}

	b.Mux.RLock()
	defer b.Mux.RUnlock()
	return b.Cells[row][col].Value, nil
}

func (b *Board) SetCell(row, col int, mark string) error {
	if !b.CheckValidInsertion(row, col) {
		return fmt.Errorf("invalid insertion at row %d, col %d", row, col)
	}

	b.Mux.Lock()
	defer b.Mux.Unlock()
	cell := Cell{Value: mark, Lives: INITIAL_LIVES, WinState: false}
	b.Cells[row][col] = cell
	return nil
}

func (b *Board) CheckWin() bool {
	b.Mux.RLock()
	defer b.Mux.RUnlock()

	// Check rows and columns
	for i := range 3 {
		if b.Cells[i][0].Value != "-" && b.Cells[i][0].Value == b.Cells[i][1].Value && b.Cells[i][1].Value == b.Cells[i][2].Value {
			b.Cells[i][0].WinState = true
			b.Cells[i][1].WinState = true
			b.Cells[i][2].WinState = true
			return true
		}
		if b.Cells[0][i].Value != "-" && b.Cells[0][i].Value == b.Cells[1][i].Value && b.Cells[1][i].Value == b.Cells[2][i].Value {
			b.Cells[0][i].WinState = true
			b.Cells[1][i].WinState = true
			b.Cells[2][i].WinState = true
			return true
		}
	}

	// Check diagonals
	if b.Cells[0][0].Value != "-" && b.Cells[0][0].Value == b.Cells[1][1].Value && b.Cells[1][1].Value == b.Cells[2][2].Value {
		b.Cells[0][0].WinState = true
		b.Cells[1][1].WinState = true
		b.Cells[2][2].WinState = true
		return true
	}
	if b.Cells[0][2].Value != "-" && b.Cells[0][2].Value == b.Cells[1][1].Value && b.Cells[1][1].Value == b.Cells[2][0].Value {
		b.Cells[0][2].WinState = true
		b.Cells[1][1].WinState = true
		b.Cells[2][0].WinState = true
		return true
	}

	return false
}

func (b *Board) CheckDraw() bool {
	b.Mux.RLock()
	defer b.Mux.RUnlock()

	for i := range 3 {
		for j := range 3 {
			if b.Cells[i][j].Value == "-" {
				return false
			}
		}
	}
	return true
}

func (b *Board) UpdateLives() {
	b.Mux.Lock()
	defer b.Mux.Unlock()

	for i := range 3 {
		for j := range 3 {
			if b.Cells[i][j].Value == "-" {
				continue
			}
			if b.Cells[i][j].Lives == 0 {
				b.Cells[i][j].Value = "-"
				b.Cells[i][j].WinState = false
				b.Cells[i][j].Lives = INITIAL_LIVES
				continue
			}
			b.Cells[i][j].Lives--
		}
	}
}
