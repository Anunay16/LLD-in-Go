package playing

import (
	"errors"
	"math/rand"
	"tic-tac-toe/pkg/models"
)

// EasyBotStrategy selects a random available cell.
type EasyBotStrategy struct{}

func NewEasyBotStrategy() *EasyBotStrategy {
	return &EasyBotStrategy{}
}

func (s *EasyBotStrategy) DecideMove(board *models.Board) (*models.Cell, error) {
	emptyCells := board.GetEmptyCells()
	if len(emptyCells) == 0 {
		return nil, errors.New("no empty cells available on the board")
	}
	randomIndex := rand.Intn(len(emptyCells))
	return emptyCells[randomIndex], nil
}
