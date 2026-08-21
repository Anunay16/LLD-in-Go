package game

import (
	"errors"
	"fmt"

	"tic-tac-toe/pkg/models"
	"tic-tac-toe/pkg/strategies/winning"
)

type Game struct {
	Board             *models.Board
	Players           []models.Player
	WinningStrategies []winning.WinningStrategy
	Moves             []models.Move
	Status            GameStatus
	NextPlayerIndex   int
	Winner            models.Player
}

// GameBuilder facilitates safe construction of Game instances with validation.
type GameBuilder struct {
	rows              int
	cols              int
	players           []models.Player
	winningStrategies []winning.WinningStrategy
}

func NewGameBuilder() *GameBuilder {
	return &GameBuilder{
		rows:              3,
		cols:              3,
		players:           make([]models.Player, 0),
		winningStrategies: make([]winning.WinningStrategy, 0),
	}
}

func (gb *GameBuilder) SetBoardDimensions(rows, cols int) *GameBuilder {
	gb.rows = rows
	gb.cols = cols
	return gb
}

func (gb *GameBuilder) AddPlayer(player models.Player) *GameBuilder {
	gb.players = append(gb.players, player)
	return gb
}

func (gb *GameBuilder) AddWinningStrategy(strategy winning.WinningStrategy) *GameBuilder {
	gb.winningStrategies = append(gb.winningStrategies, strategy)
	return gb
}

func (gb *GameBuilder) Build() (*Game, error) {
	if gb.rows <= 0 || gb.cols <= 0 {
		return nil, errors.New("board dimensions must be greater than zero")
	}
	if len(gb.players) < 2 {
		return nil, errors.New("at least 2 players are required to start the game")
	}

	// Validate unique symbols
	symbolMap := make(map[models.Symbol]bool)
	for _, player := range gb.players {
		sym := player.GetSymbol()
		if symbolMap[sym] {
			return nil, fmt.Errorf("duplicate player symbol detected: %s", sym)
		}
		symbolMap[sym] = true
	}

	if len(gb.winningStrategies) == 0 {
		return nil, errors.New("at least one winning strategy must be configured")
	}

	board := models.NewBoard(gb.rows, gb.cols)

	return &Game{
		Board:             board,
		Players:           gb.players,
		WinningStrategies: gb.winningStrategies,
		Moves:             make([]models.Move, 0),
		Status:            InProgress,
		NextPlayerIndex:   0,
		Winner:            nil,
	}, nil
}

func (g *Game) GetCurrentPlayer() models.Player {
	return g.Players[g.NextPlayerIndex]
}

// MakeMove executes a move for the current player at specified (row, col) or auto-generated cell for Bot.
func (g *Game) MakeMove(row, col int) error {
	if g.Status != InProgress {
		return fmt.Errorf("cannot make move: game is already %s", g.Status)
	}

	currentPlayer := g.GetCurrentPlayer()
	var targetCell *models.Cell
	var err error

	if currentPlayer.GetType() == models.BotPlayerType {
		targetCell, err = currentPlayer.DecideMove(g.Board)
		if err != nil {
			return fmt.Errorf("bot player %s failed to decide move: %w", currentPlayer.GetName(), err)
		}
	} else {
		targetCell, err = g.Board.GetCell(row, col)
		if err != nil {
			return fmt.Errorf("invalid move location: %w", err)
		}
	}

	if !targetCell.IsEmpty() {
		return fmt.Errorf("cell (%d, %d) is already occupied by symbol %s", targetCell.Row, targetCell.Col, targetCell.Symbol.String())
	}

	// Apply move
	targetCell.Symbol = currentPlayer.GetSymbol()
	g.Board.EmptyCount--

	move := models.NewMove(currentPlayer, targetCell)
	g.Moves = append(g.Moves, move)

	fmt.Printf("[%s] played %s at position (%d, %d)\n", currentPlayer.GetName(), currentPlayer.GetSymbol().String(), targetCell.Row, targetCell.Col)

	// Check for winning conditions
	for _, strategy := range g.WinningStrategies {
		if strategy.CheckWinner(g.Board, move) {
			g.Status = Ended
			g.Winner = currentPlayer
			fmt.Printf("🎉 Player %s (%s) WON the game!\n", currentPlayer.GetName(), currentPlayer.GetSymbol().String())
			return nil
		}
	}

	// Check for draw
	if g.Board.EmptyCount == 0 {
		g.Status = Draw
		fmt.Println("🤝 Game ended in a DRAW!")
		return nil
	}

	// Advance turn to next player
	g.NextPlayerIndex = (g.NextPlayerIndex + 1) % len(g.Players)
	return nil
}

func (g *Game) Undo() error {
	if len(g.Moves) == 0 {
		return errors.New("no moves to undo")
	}

	lastMove := g.Moves[len(g.Moves)-1]
	g.Moves = g.Moves[:len(g.Moves)-1]

	// Revert cell state
	lastMove.Cell.Clear()
	g.Board.EmptyCount++

	// Revert winning strategies state
	for _, strategy := range g.WinningStrategies {
		strategy.Undo(g.Board, lastMove)
	}

	// Revert player turn
	g.NextPlayerIndex = (g.NextPlayerIndex - 1 + len(g.Players)) % len(g.Players)
	g.Status = InProgress
	g.Winner = nil

	fmt.Printf("↩ Undo move by player %s at (%d, %d)\n", lastMove.Player.GetName(), lastMove.Cell.Row, lastMove.Cell.Col)
	return nil
}
