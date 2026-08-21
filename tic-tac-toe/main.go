package main

import (
	"fmt"

	"tic-tac-toe/pkg/game"
	"tic-tac-toe/pkg/models"
	"tic-tac-toe/pkg/strategies/playing"
	"tic-tac-toe/pkg/strategies/winning"
)

func main() {
	fmt.Println("=========================================================")
	fmt.Println("   TIC-TAC-TOE LOW LEVEL DESIGN DEMONSTRATION (Go)")
	fmt.Println("=========================================================")

	runStandardGameWithBot()
	runMultiplayerBigBoardGame()
	runCustomRulesGame()
}

// Scenario 1: Standard 3x3 game (Human vs Smart AI Bot)
func runStandardGameWithBot() {
	fmt.Println("\n--- Scenario 1: Standard 3x3 Game (Human vs AI Bot) ---")

	botStrategy := playing.NewSmartBotStrategy()
	p1 := models.NewHumanPlayer(1, "Alice", models.Symbol("❌"))
	p2 := models.NewBotPlayer(2, "Bot-Alpha", models.Symbol("⭕"), models.EasyDifficulty, botStrategy)

	winStrategy := winning.NewOrderOneWinningStrategy(3)

	g, err := game.NewGameBuilder().
		SetBoardDimensions(3, 3).
		AddPlayer(p1).
		AddPlayer(p2).
		AddWinningStrategy(winStrategy).
		Build()

	if err != nil {
		fmt.Println("Error building game:", err)
		return
	}

	g.Board.Display()

	moves := [][2]int{
		{0, 0}, // Alice plays (0,0)
		{0, 0}, // Bot turn
		{0, 1}, // Alice plays (0,1)
		{0, 0}, // Bot turn
		{1, 0}, // Alice plays (1,0)
		{0, 0}, // Bot turn
		{2, 0}, // Alice plays (2,0) -> Alice completes Col 0 win!
	}

	for _, m := range moves {
		if g.Status != game.InProgress {
			break
		}
		err := g.MakeMove(m[0], m[1])
		if err != nil {
			fmt.Println("Move Error:", err)
		}
		g.Board.Display()
	}
}

// Scenario 2: Multi-Player Big Board (10x10 Board, 3 Players, 5-in-a-row Win)
func runMultiplayerBigBoardGame() {
	fmt.Println("\n--- Scenario 2: Multi-Player Big Board (10x10, 3 Players, K=5 Win Rule) ---")

	p1 := models.NewHumanPlayer(1, "Player-X", models.Symbol("X"))
	p2 := models.NewHumanPlayer(2, "Player-O", models.Symbol("O"))
	p3 := models.NewBotPlayer(3, "Bot-Delta", models.Symbol("#"), models.EasyDifficulty, playing.NewEasyBotStrategy())

	kWinStrategy := winning.NewKInARowWinningStrategy(5)

	g, err := game.NewGameBuilder().
		SetBoardDimensions(10, 10).
		AddPlayer(p1).
		AddPlayer(p2).
		AddPlayer(p3).
		AddWinningStrategy(kWinStrategy).
		Build()

	if err != nil {
		fmt.Println("Error building big board game:", err)
		return
	}

	scriptedMoves := []struct {
		playerID int
		r, c     int
	}{
		{1, 4, 2}, // P1 (X)
		{2, 0, 0}, // P2 (O)
		{3, 0, 0}, // P3 (Bot)
		{1, 4, 3}, // P1 (X)
		{2, 0, 1}, // P2 (O)
		{3, 0, 0}, // P3 (Bot)
		{1, 4, 4}, // P1 (X)
		{2, 0, 2}, // P2 (O)
		{3, 0, 0}, // P3 (Bot)
		{1, 4, 5}, // P1 (X)
		{2, 0, 3}, // P2 (O)
		{3, 0, 0}, // P3 (Bot)
		{1, 4, 6}, // P1 (X) -> Wins with 5-in-a-row!
	}

	for _, sm := range scriptedMoves {
		if g.Status != game.InProgress {
			break
		}
		_ = g.MakeMove(sm.r, sm.c)
	}

	g.Board.Display()
}

// Scenario 3: Custom Winning Rules (Corner Win Rule)
func runCustomRulesGame() {
	fmt.Println("\n--- Scenario 3: Custom Rules (Corner Win Rule) ---")

	p1 := models.NewHumanPlayer(1, "CornerKing", models.Symbol("C"))
	p2 := models.NewHumanPlayer(2, "Opponent", models.Symbol("Z"))

	cornerStrategy := winning.NewCornerWinningStrategy()

	g, err := game.NewGameBuilder().
		SetBoardDimensions(4, 4).
		AddPlayer(p1).
		AddPlayer(p2).
		AddWinningStrategy(cornerStrategy).
		Build()

	if err != nil {
		fmt.Println("Error building custom rules game:", err)
		return
	}

	cornerMoves := [][2]int{
		{0, 0}, // P1 (0,0) - Top Left
		{1, 1}, // P2
		{0, 3}, // P1 (0,3) - Top Right
		{1, 2}, // P2
		{3, 0}, // P1 (3,0) - Bottom Left
		{2, 1}, // P2
		{3, 3}, // P1 (3,3) - Bottom Right -> Wins by Corner Rule!
	}

	for _, m := range cornerMoves {
		if g.Status != game.InProgress {
			break
		}
		_ = g.MakeMove(m[0], m[1])
	}

	g.Board.Display()
}
