package models

// Symbol is a domain type representing a player's mark on the board.
// Using `type Symbol string` provides strong type safety without struct boilerplate,
// while supporting full Unicode/Emoji symbols (e.g. "X", "O", "❌", "⭕", "P1").
type Symbol string

const EmptySymbol Symbol = ""

func (s Symbol) IsEmpty() bool {
	return s == EmptySymbol
}

func (s Symbol) String() string {
	if s.IsEmpty() {
		return "."
	}
	return string(s)
}
