package winning

import (
	"tic-tac-toe/pkg/models"
)

// CornerWinningStrategy grants victory if a player captures all 4 corners of the board.
type CornerWinningStrategy struct{}

func NewCornerWinningStrategy() *CornerWinningStrategy {
	return &CornerWinningStrategy{}
}

func (s *CornerWinningStrategy) CheckWinner(board *models.Board, lastMove models.Move) bool {
	targetSymbol := lastMove.Player.GetSymbol()
	corners := [][2]int{
		{0, 0},
		{0, board.Cols - 1},
		{board.Rows - 1, 0},
		{board.Rows - 1, board.Cols - 1},
	}

	for _, c := range corners {
		if board.Grid[c[0]][c[1]].Symbol != targetSymbol {
			return false
		}
	}
	return true
}

func (s *CornerWinningStrategy) Undo(board *models.Board, lastMove models.Move) {
	// Stateless evaluation
}
