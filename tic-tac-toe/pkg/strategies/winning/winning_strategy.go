package winning

import "tic-tac-toe/pkg/models"

// WinningStrategy defines the interface for evaluating win conditions.
type WinningStrategy interface {
	CheckWinner(board *models.Board, lastMove models.Move) bool
	Undo(board *models.Board, lastMove models.Move)
}
