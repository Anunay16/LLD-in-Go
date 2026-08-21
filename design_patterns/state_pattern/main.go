package main

import (
	"fmt"
	"state_pattern/music_player"
	"state_pattern/vending_machine"
)

func main() {
	runVendingMachineDemo()
	fmt.Println()
	runMusicPlayerDemo()
}

func runVendingMachineDemo() {
	fmt.Println("==========================================")
	fmt.Println("     State Pattern Demo 1: Vending Machine")
	fmt.Println("==========================================")

	vm := vending_machine.NewVendingMachine(2)
	vm.DisplayCurrentState()

	fmt.Println("\n[Scenario 1] Normal Purchase:")
	if err := vm.InsertCoin(); err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}
	if err := vm.SelectProduct(); err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}
	vm.DisplayCurrentState()

	fmt.Println("\n[Scenario 2] Insert Coin and Eject:")
	if err := vm.InsertCoin(); err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}
	if err := vm.EjectCoin(); err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}
	vm.DisplayCurrentState()

	fmt.Println("\n[Scenario 3] Purchase Final Item:")
	if err := vm.InsertCoin(); err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}
	if err := vm.SelectProduct(); err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}
	vm.DisplayCurrentState()
}

func runMusicPlayerDemo() {
	fmt.Println("==========================================")
	fmt.Println("     State Pattern Demo 2: Music Player   ")
	fmt.Println("==========================================")

	playlist := []string{
		"Bohemian Rhapsody - Queen",
		"Hotel California - Eagles",
		"Stairway to Heaven - Led Zeppelin",
	}

	player := music_player.NewMusicPlayer(playlist)
	player.DisplayStatus()

	fmt.Println("\n[Action] Press Play:")
	if err := player.Play(); err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}
	player.DisplayStatus()

	fmt.Println("\n[Action] Next Track:")
	if err := player.NextTrack(); err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}
	player.DisplayStatus()

	fmt.Println("\n[Action] Press Pause:")
	if err := player.Pause(); err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}
	player.DisplayStatus()

	fmt.Println("\n[Action] Try Pausing Again:")
	if err := player.Pause(); err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}

	fmt.Println("\n[Action] Resume Playback:")
	if err := player.Play(); err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}
	player.DisplayStatus()

	fmt.Println("\n[Action] Press Stop:")
	if err := player.Stop(); err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}
	player.DisplayStatus()

	fmt.Println("\n[Action] Try Pausing While Stopped:")
	if err := player.Pause(); err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}
}
