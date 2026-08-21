package music_player

import "fmt"

// StoppedState represents the state when playback is stopped.
type StoppedState struct {
	player *MusicPlayer
}

func (s *StoppedState) Play() error {
	fmt.Printf("--> [StoppedState] Starting playback of \"%s\"\n", s.player.GetCurrentTrack())
	s.player.SetState(s.player.GetPlayingState())
	return nil
}

func (s *StoppedState) Pause() error {
	return fmt.Errorf("cannot pause: player is currently stopped")
}

func (s *StoppedState) Stop() error {
	fmt.Println("--> [StoppedState] Player is already stopped.")
	return nil
}

func (s *StoppedState) NextTrack() error {
	s.player.NextTrackIndex()
	fmt.Printf("--> [StoppedState] Selected next track: \"%s\"\n", s.player.GetCurrentTrack())
	return nil
}

func (s *StoppedState) PreviousTrack() error {
	s.player.PreviousTrackIndex()
	fmt.Printf("--> [StoppedState] Selected previous track: \"%s\"\n", s.player.GetCurrentTrack())
	return nil
}
