package models

import (
	"errors"
	"fmt"
	"strings"
)

// Board represents the N x M Tic-Tac-Toe grid.
type Board struct {
	Rows       int
	Cols       int
	Grid       [][]*Cell
	EmptyCount int
}

func NewBoard(rows, cols int) *Board {
	grid := make([][]*Cell, rows)
	for i := 0; i < rows; i++ {
		grid[i] = make([]*Cell, cols)
		for j := 0; j < cols; j++ {
			grid[i][j] = NewCell(i, j)
		}
	}
	return &Board{
		Rows:       rows,
		Cols:       cols,
		Grid:       grid,
		EmptyCount: rows * cols,
	}
}

func (b *Board) IsValidCell(row, col int) bool {
	return row >= 0 && row < b.Rows && col >= 0 && col < b.Cols
}

func (b *Board) GetCell(row, col int) (*Cell, error) {
	if !b.IsValidCell(row, col) {
		return nil, errors.New("cell out of board bounds")
	}
	return b.Grid[row][col], nil
}

func (b *Board) GetEmptyCells() []*Cell {
	emptyCells := make([]*Cell, 0, b.EmptyCount)
	for i := 0; i < b.Rows; i++ {
		for j := 0; j < b.Cols; j++ {
			if b.Grid[i][j].IsEmpty() {
				emptyCells = append(emptyCells, b.Grid[i][j])
			}
		}
	}
	return emptyCells
}

func (b *Board) Display() {
	fmt.Println()
	for i := 0; i < b.Rows; i++ {
		rowStr := make([]string, b.Cols)
		for j := 0; j < b.Cols; j++ {
			if b.Grid[i][j].IsEmpty() {
				rowStr[j] = " . "
			} else {
				rowStr[j] = fmt.Sprintf(" %s ", b.Grid[i][j].Symbol.String())
			}
		}
		fmt.Println(strings.Join(rowStr, "|"))
		if i < b.Rows-1 {
			border := strings.Repeat("---+", b.Cols-1) + "---"
			fmt.Println(border)
		}
	}
	fmt.Println()
}
