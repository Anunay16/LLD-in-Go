package music_player

import "fmt"

// State defines actions available on the Music Player.
type State interface {
	Play() error
	Pause() error
	Stop() error
	NextTrack() error
	PreviousTrack() error
}

// MusicPlayer acts as the Context in the State Pattern.
type MusicPlayer struct {
	stoppedState State
	playingState State
	pausedState  State

	currentState      State
	playlist          []string
	currentTrackIndex int
}

// NewMusicPlayer initializes the context with a playlist and concrete state objects.
func NewMusicPlayer(playlist []string) *MusicPlayer {
	if len(playlist) == 0 {
		playlist = []string{"Default Track"}
	}
	p := &MusicPlayer{
		playlist:          playlist,
		currentTrackIndex: 0,
	}

	p.stoppedState = &StoppedState{player: p}
	p.playingState = &PlayingState{player: p}
	p.pausedState = &PausedState{player: p}

	p.currentState = p.stoppedState
	return p
}

func (p *MusicPlayer) Play() error {
	return p.currentState.Play()
}

func (p *MusicPlayer) Pause() error {
	return p.currentState.Pause()
}

func (p *MusicPlayer) Stop() error {
	return p.currentState.Stop()
}

func (p *MusicPlayer) NextTrack() error {
	return p.currentState.NextTrack()
}

func (p *MusicPlayer) PreviousTrack() error {
	return p.currentState.PreviousTrack()
}

func (p *MusicPlayer) SetState(s State) {
	p.currentState = s
}

func (p *MusicPlayer) GetStoppedState() State {
	return p.stoppedState
}

func (p *MusicPlayer) GetPlayingState() State {
	return p.playingState
}

func (p *MusicPlayer) GetPausedState() State {
	return p.pausedState
}

func (p *MusicPlayer) GetCurrentTrack() string {
	return p.playlist[p.currentTrackIndex]
}

func (p *MusicPlayer) NextTrackIndex() {
	p.currentTrackIndex = (p.currentTrackIndex + 1) % len(p.playlist)
}

func (p *MusicPlayer) PreviousTrackIndex() {
	p.currentTrackIndex = (p.currentTrackIndex - 1 + len(p.playlist)) % len(p.playlist)
}

func (p *MusicPlayer) DisplayStatus() {
	fmt.Printf("State: %T | Current Track: \"%s\"\n", p.currentState, p.GetCurrentTrack())
}
