package models

type PlayerType string

const (
	HumanPlayerType PlayerType = "HUMAN"
	BotPlayerType   PlayerType = "BOT"
)

type BotDifficulty string

const (
	EasyDifficulty   BotDifficulty = "EASY"
	MediumDifficulty BotDifficulty = "MEDIUM"
	HardDifficulty   BotDifficulty = "HARD"
)

type PlayingStrategy interface {
	DecideMove(board *Board) (*Cell, error)
}

// Player interface abstracts common behaviors for both Human and Bot players.
type Player interface {
	GetID() int
	GetName() string
	GetSymbol() Symbol
	GetType() PlayerType
	DecideMove(board *Board) (*Cell, error)
}

// BasePlayer contains attributes common to all players.
type BasePlayer struct {
	ID     int
	Name   string
	Symbol Symbol
}

func (b *BasePlayer) GetID() int       { return b.ID }
func (b *BasePlayer) GetName() string   { return b.Name }
func (b *BasePlayer) GetSymbol() Symbol { return b.Symbol }

// HumanPlayer represents a human participant with no AI/Bot overhead.
type HumanPlayer struct {
	BasePlayer
}

func NewHumanPlayer(id int, name string, symbol Symbol) *HumanPlayer {
	return &HumanPlayer{
		BasePlayer: BasePlayer{
			ID:     id,
			Name:   name,
			Symbol: symbol,
		},
	}
}

func (h *HumanPlayer) GetType() PlayerType {
	return HumanPlayerType
}

func (h *HumanPlayer) DecideMove(board *Board) (*Cell, error) {
	// Human moves are provided via UI/CLI input to Game.MakeMove(r, c)
	return nil, nil
}

// BotPlayer represents an AI participant holding bot difficulty and playing strategy.
type BotPlayer struct {
	BasePlayer
	Difficulty  BotDifficulty
	BotStrategy PlayingStrategy
}

func NewBotPlayer(id int, name string, symbol Symbol, difficulty BotDifficulty, strategy PlayingStrategy) *BotPlayer {
	return &BotPlayer{
		BasePlayer: BasePlayer{
			ID:     id,
			Name:   name,
			Symbol: symbol,
		},
		Difficulty:  difficulty,
		BotStrategy: strategy,
	}
}

func (b *BotPlayer) GetType() PlayerType {
	return BotPlayerType
}

func (b *BotPlayer) DecideMove(board *Board) (*Cell, error) {
	if b.BotStrategy == nil {
		return nil, nil
	}
	return b.BotStrategy.DecideMove(board)
}
