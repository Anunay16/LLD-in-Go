package models

// Move records a move played by a player on a specific cell.
type Move struct {
	Player Player
	Cell   *Cell
}

func NewMove(player Player, cell *Cell) Move {
	return Move{
		Player: player,
		Cell:   cell,
	}
}
