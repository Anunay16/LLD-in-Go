package game_test

import (
	"testing"

	"tic-tac-toe/pkg/game"
	"tic-tac-toe/pkg/models"
	"tic-tac-toe/pkg/strategies/playing"
	"tic-tac-toe/pkg/strategies/winning"
)

func TestGameBuilderValidations(t *testing.T) {
	p1 := models.NewHumanPlayer(1, "Player 1", models.Symbol("X"))
	p2 := models.NewHumanPlayer(2, "Player 2", models.Symbol("O"))

	t.Run("Fails with duplicate player symbols", func(t *testing.T) {
		p2Duplicate := models.NewHumanPlayer(2, "Player 2", models.Symbol("X"))
		_, err := game.NewGameBuilder().
			SetBoardDimensions(3, 3).
			AddPlayer(p1).
			AddPlayer(p2Duplicate).
			AddWinningStrategy(winning.NewOrderOneWinningStrategy(3)).
			Build()

		if err == nil {
			t.Errorf("Expected error for duplicate symbols, got nil")
		}
	})

	t.Run("Fails with less than 2 players", func(t *testing.T) {
		_, err := game.NewGameBuilder().
			SetBoardDimensions(3, 3).
			AddPlayer(p1).
			AddWinningStrategy(winning.NewOrderOneWinningStrategy(3)).
			Build()

		if err == nil {
			t.Errorf("Expected error for insufficient players, got nil")
		}
	})

	t.Run("Successfully builds valid game", func(t *testing.T) {
		g, err := game.NewGameBuilder().
			SetBoardDimensions(3, 3).
			AddPlayer(p1).
			AddPlayer(p2).
			AddWinningStrategy(winning.NewOrderOneWinningStrategy(3)).
			Build()

		if err != nil {
			t.Fatalf("Unexpected build error: %v", err)
		}
		if g.Status != game.InProgress {
			t.Errorf("Expected status IN_PROGRESS, got %s", g.Status)
		}
	})
}

func TestGamePlayAndUndo(t *testing.T) {
	p1 := models.NewHumanPlayer(1, "Alice", models.Symbol("X"))
	p2 := models.NewBotPlayer(2, "Bot", models.Symbol("O"), models.EasyDifficulty, playing.NewEasyBotStrategy())

	g, err := game.NewGameBuilder().
		SetBoardDimensions(3, 3).
		AddPlayer(p1).
		AddPlayer(p2).
		AddWinningStrategy(winning.NewOrderOneWinningStrategy(3)).
		Build()

	if err != nil {
		t.Fatalf("Failed to build game: %v", err)
	}

	// Make move
	err = g.MakeMove(0, 0)
	if err != nil {
		t.Fatalf("Failed to make move: %v", err)
	}

	if len(g.Moves) != 1 {
		t.Errorf("Expected 1 move in history, got %d", len(g.Moves))
	}

	// Test Undo
	err = g.Undo()
	if err != nil {
		t.Fatalf("Failed to undo move: %v", err)
	}

	if len(g.Moves) != 0 {
		t.Errorf("Expected 0 moves in history after undo, got %d", len(g.Moves))
	}
	cell, _ := g.Board.GetCell(0, 0)
	if !cell.IsEmpty() {
		t.Errorf("Cell (0,0) should be empty after undo")
	}
}

func TestBigBoardMultiplayerKInARow(t *testing.T) {
	p1 := models.NewHumanPlayer(1, "P1", models.Symbol("X"))
	p2 := models.NewHumanPlayer(2, "P2", models.Symbol("O"))
	p3 := models.NewHumanPlayer(3, "P3", models.Symbol("A"))

	kWinStrategy := winning.NewKInARowWinningStrategy(4)

	g, err := game.NewGameBuilder().
		SetBoardDimensions(8, 8).
		AddPlayer(p1).
		AddPlayer(p2).
		AddPlayer(p3).
		AddWinningStrategy(kWinStrategy).
		Build()

	if err != nil {
		t.Fatalf("Failed to create big board game: %v", err)
	}

	// P1: (0,0)
	_ = g.MakeMove(0, 0)
	// P2: (0,1)
	_ = g.MakeMove(0, 1)
	// P3: (0,2)
	_ = g.MakeMove(0, 2)

	// P1: (1,1)
	_ = g.MakeMove(1, 1)
	// P2: (1,0)
	_ = g.MakeMove(1, 0)
	// P3: (1,2)
	_ = g.MakeMove(1, 2)

	// P1: (2,2)
	_ = g.MakeMove(2, 2)
	// P2: (2,0)
	_ = g.MakeMove(2, 0)
	// P3: (2,1)
	_ = g.MakeMove(2, 1)

	// P1: (3,3) -> Completes 4 in a row diagonal win!
	_ = g.MakeMove(3, 3)

	if g.Status != game.Ended {
		t.Errorf("Expected game status ENDED, got %s", g.Status)
	}
	if g.Winner != p1 {
		t.Errorf("Expected winner P1, got %v", g.Winner)
	}
}
