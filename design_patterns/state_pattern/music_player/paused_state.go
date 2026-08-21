package music_player

import "fmt"

// PausedState represents the state when playback is paused.
type PausedState struct {
	player *MusicPlayer
}

func (s *PausedState) Play() error {
	fmt.Printf("--> [PausedState] Resuming playback of \"%s\"\n", s.player.GetCurrentTrack())
	s.player.SetState(s.player.GetPlayingState())
	return nil
}

func (s *PausedState) Pause() error {
	fmt.Println("--> [PausedState] Player is already paused.")
	return nil
}

func (s *PausedState) Stop() error {
	fmt.Println("--> [PausedState] Stopping playback from paused state.")
	s.player.SetState(s.player.GetStoppedState())
	return nil
}

func (s *PausedState) NextTrack() error {
	s.player.NextTrackIndex()
	fmt.Printf("--> [PausedState] Switched to next track while paused: \"%s\"\n", s.player.GetCurrentTrack())
	return nil
}

func (s *PausedState) PreviousTrack() error {
	s.player.PreviousTrackIndex()
	fmt.Printf("--> [PausedState] Switched to previous track while paused: \"%s\"\n", s.player.GetCurrentTrack())
	return nil
}
