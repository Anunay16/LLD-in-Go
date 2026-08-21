package game

type GameStatus string

const (
	InProgress GameStatus = "IN_PROGRESS"
	Ended      GameStatus = "ENDED"
	Draw       GameStatus = "DRAW"
)
