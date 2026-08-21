package winning

import (
	"tic-tac-toe/pkg/models"
)

// OrderOneWinningStrategy checks win condition in O(1) time per move for N x N boards.
type OrderOneWinningStrategy struct {
	rowCounts      []map[models.Symbol]int
	colCounts      []map[models.Symbol]int
	diagCounts     map[models.Symbol]int
	antiDiagCounts map[models.Symbol]int
}

func NewOrderOneWinningStrategy(boardSize int) *OrderOneWinningStrategy {
	rowCounts := make([]map[models.Symbol]int, boardSize)
	colCounts := make([]map[models.Symbol]int, boardSize)
	for i := 0; i < boardSize; i++ {
		rowCounts[i] = make(map[models.Symbol]int)
		colCounts[i] = make(map[models.Symbol]int)
	}

	return &OrderOneWinningStrategy{
		rowCounts:      rowCounts,
		colCounts:      colCounts,
		diagCounts:     make(map[models.Symbol]int),
		antiDiagCounts: make(map[models.Symbol]int),
	}
}

func (s *OrderOneWinningStrategy) CheckWinner(board *models.Board, lastMove models.Move) bool {
	row := lastMove.Cell.Row
	col := lastMove.Cell.Col
	symbol := lastMove.Player.GetSymbol()

	// Update row count
	s.rowCounts[row][symbol]++
	if s.rowCounts[row][symbol] == board.Cols {
		return true
	}

	// Update col count
	s.colCounts[col][symbol]++
	if s.colCounts[col][symbol] == board.Rows {
		return true
	}

	// Update diagonal count if cell is on main diagonal
	if row == col {
		s.diagCounts[symbol]++
		if s.diagCounts[symbol] == board.Rows {
			return true
		}
	}

	// Update anti-diagonal count if cell is on anti-diagonal
	if row+col == board.Rows-1 {
		s.antiDiagCounts[symbol]++
		if s.antiDiagCounts[symbol] == board.Rows {
			return true
		}
	}

	return false
}

func (s *OrderOneWinningStrategy) Undo(board *models.Board, lastMove models.Move) {
	row := lastMove.Cell.Row
	col := lastMove.Cell.Col
	symbol := lastMove.Player.GetSymbol()

	s.rowCounts[row][symbol]--
	s.colCounts[col][symbol]--

	if row == col {
		s.diagCounts[symbol]--
	}
	if row+col == board.Rows-1 {
		s.antiDiagCounts[symbol]--
	}
}
