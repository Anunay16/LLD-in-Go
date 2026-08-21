package winning

import (
	"tic-tac-toe/pkg/models"
)

// KInARowWinningStrategy checks if K consecutive cells in any direction belong to the player.
// Ideal for big board multi-player games (e.g. 10x10 board with 5 in a row).
type KInARowWinningStrategy struct {
	K int
}

func NewKInARowWinningStrategy(k int) *KInARowWinningStrategy {
	return &KInARowWinningStrategy{K: k}
}

func (s *KInARowWinningStrategy) CheckWinner(board *models.Board, lastMove models.Move) bool {
	row := lastMove.Cell.Row
	col := lastMove.Cell.Col
	targetSymbol := lastMove.Player.GetSymbol()

	// 4 directions: Horizontal, Vertical, Diagonal (\), Anti-Diagonal (/)
	directions := [][2]int{
		{0, 1},  // Horizontal
		{1, 0},  // Vertical
		{1, 1},  // Main Diagonal
		{1, -1}, // Anti-Diagonal
	}

	for _, dir := range directions {
		count := 1 // Count the lastMove cell itself

		// Count forward along direction
		r, c := row+dir[0], col+dir[1]
		for board.IsValidCell(r, c) && board.Grid[r][c].Symbol == targetSymbol {
			count++
			r += dir[0]
			c += dir[1]
		}

		// Count backward along opposite direction
		r, c = row-dir[0], col-dir[1]
		for board.IsValidCell(r, c) && board.Grid[r][c].Symbol == targetSymbol {
			count++
			r -= dir[0]
			c -= dir[1]
		}

		if count >= s.K {
			return true
		}
	}

	return false
}

func (s *KInARowWinningStrategy) Undo(board *models.Board, lastMove models.Move) {
	// Stateless strategy - board cell symbol resetting in Game.Undo is sufficient.
}
