package music_player

import "fmt"

// PlayingState represents the state when audio is currently playing.
type PlayingState struct {
	player *MusicPlayer
}

func (s *PlayingState) Play() error {
	fmt.Println("--> [PlayingState] Player is already playing.")
	return nil
}

func (s *PlayingState) Pause() error {
	fmt.Printf("--> [PlayingState] Pausing playback of \"%s\"\n", s.player.GetCurrentTrack())
	s.player.SetState(s.player.GetPausedState())
	return nil
}

func (s *PlayingState) Stop() error {
	fmt.Println("--> [PlayingState] Stopping playback.")
	s.player.SetState(s.player.GetStoppedState())
	return nil
}

func (s *PlayingState) NextTrack() error {
	s.player.NextTrackIndex()
	fmt.Printf("--> [PlayingState] Playing next track: \"%s\"\n", s.player.GetCurrentTrack())
	return nil
}

func (s *PlayingState) PreviousTrack() error {
	s.player.PreviousTrackIndex()
	fmt.Printf("--> [PlayingState] Playing previous track: \"%s\"\n", s.player.GetCurrentTrack())
	return nil
}
