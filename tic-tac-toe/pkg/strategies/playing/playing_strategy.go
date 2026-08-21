package playing

import "tic-tac-toe/pkg/models"

// PlayingStrategy defines the contract for bot move generation.
type PlayingStrategy interface {
	DecideMove(board *models.Board) (*models.Cell, error)
}
