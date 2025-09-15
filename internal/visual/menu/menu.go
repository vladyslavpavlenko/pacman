package menu

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/vladyslavpavlenko/pacman/internal/config"
)

// GameState represents the current state of the game
type GameState int

const (
	StateMenu GameState = iota
	StatePlaying
	StatePaused
)

// Menu represents the main menu system
type Menu struct {
	state          GameState
	selectedOption int
	selectedDiff   config.Difficulty
	options        []string
	difficulties   []config.Difficulty
}

// New creates a new menu instance
func New() *Menu {
	return &Menu{
		state:          StateMenu,
		selectedOption: 0,
		selectedDiff:   config.DifficultyMedium,
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

// Update handles menu input and state transitions
func (m *Menu) Update() (GameState, config.Difficulty, bool) {
	if m.state != StateMenu {
		return m.state, m.selectedDiff, false
	}

	// Handle navigation
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		m.selectedOption = (m.selectedOption - 1 + len(m.options)) % len(m.options)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		m.selectedOption = (m.selectedOption + 1) % len(m.options)
	}

	// Handle selection
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		switch m.selectedOption {
		case 0: // Start Game
			return StatePlaying, m.selectedDiff, true
		case 1: // Difficulty selection
			// Cycle through difficulties
			for i, diff := range m.difficulties {
				if diff == m.selectedDiff {
					m.selectedDiff = m.difficulties[(i+1)%len(m.difficulties)]
					break
				}
			}
		case 2: // Exit
			return StateMenu, m.selectedDiff, true // Signal to exit
		}
	}

	return m.state, m.selectedDiff, false
}

// SetState sets the current game state
func (m *Menu) SetState(state GameState) {
	m.state = state
}

// GetState returns the current game state
func (m *Menu) GetState() GameState {
	return m.state
}

// GetSelectedOption returns the currently selected menu option
func (m *Menu) GetSelectedOption() int {
	return m.selectedOption
}

// GetSelectedDifficulty returns the currently selected difficulty
func (m *Menu) GetSelectedDifficulty() config.Difficulty {
	return m.selectedDiff
}

// GetOptions returns the menu options
func (m *Menu) GetOptions() []string {
	return m.options
}
