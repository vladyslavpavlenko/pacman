package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/vladyslavpavlenko/pacman/internal/config"
	"github.com/vladyslavpavlenko/pacman/internal/view"
)

type UI struct {
	state          view.State
	selectedOption int
	selectedDiff   config.Difficulty
	options        []string
	difficulties   []config.Difficulty
}

func New() *UI {
	return &UI{
		state:          view.StateMenu,
		selectedOption: 0,
		selectedDiff:   config.DifficultyEasy,
		options: []string{
			"Start Game",
			"Difficulty: ",
			"Exit",
		},
		difficulties: []config.Difficulty{
			config.DifficultyEasy,
			config.DifficultyMedium,
			config.DifficultyHard,
		},
	}
}

func (m *UI) Update() (view.State, config.Difficulty, bool) {
	if m.state != view.StateMenu {
		return m.state, m.selectedDiff, false
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		m.selectedOption = (m.selectedOption - 1 + len(m.options)) % len(m.options)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		m.selectedOption = (m.selectedOption + 1) % len(m.options)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		switch m.selectedOption {
		case 0:
			return view.StatePlaying, m.selectedDiff, true
		case 1:
			for i, diff := range m.difficulties {
				if diff == m.selectedDiff {
					m.selectedDiff = m.difficulties[(i+1)%len(m.difficulties)]
					break
				}
			}
		case 2:
			return view.StateMenu, m.selectedDiff, true
		}
	}

	return m.state, m.selectedDiff, false
}

func (m *UI) SetState(state view.State) {
	m.state = state
}

func (m *UI) GetState() view.State {
	return m.state
}

func (m *UI) GetSelectedOption() int {
	return m.selectedOption
}

func (m *UI) GetSelectedDifficulty() config.Difficulty {
	return m.selectedDiff
}

func (m *UI) GetOptions() []string {
	return m.options
}
