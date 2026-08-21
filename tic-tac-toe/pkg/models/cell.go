package models

// Cell represents a single square on the Tic-Tac-Toe board.
type Cell struct {
	Row    int
	Col    int
	Symbol Symbol
}

func NewCell(row, col int) *Cell {
	return &Cell{
		Row:    row,
		Col:    col,
		Symbol: EmptySymbol,
	}
}

func (c *Cell) IsEmpty() bool {
	return c.Symbol.IsEmpty()
}

func (c *Cell) Clear() {
	c.Symbol = EmptySymbol
}
