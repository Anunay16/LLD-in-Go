package playing

import (
	"errors"
	"tic-tac-toe/pkg/models"
)

// SmartBotStrategy attempts to pick the center cell if available, otherwise first empty cell.
type SmartBotStrategy struct{}

func NewSmartBotStrategy() *SmartBotStrategy {
	return &SmartBotStrategy{}
}

func (s *SmartBotStrategy) DecideMove(board *models.Board) (*models.Cell, error) {
	emptyCells := board.GetEmptyCells()
	if len(emptyCells) == 0 {
		return nil, errors.New("no empty cells available")
	}

	// Prefer center cell if empty
	centerRow := board.Rows / 2
	centerCol := board.Cols / 2
	centerCell, err := board.GetCell(centerRow, centerCol)
	if err == nil && centerCell.IsEmpty() {
		return centerCell, nil
	}

	// Fallback to first available empty cell
	return emptyCells[0], nil
}
