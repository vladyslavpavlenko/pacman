package config

type Difficulty int

const (
	DifficultyEasy Difficulty = iota
	DifficultyMedium
	DifficultyHard
)

func (d Difficulty) String() string {
	switch d {
	case DifficultyEasy:
		return "Easy"
	case DifficultyMedium:
		return "Medium"
	case DifficultyHard:
		return "Hard"
	default:
		return "Unknown"
	}
}

type GhostLevel int

const (
	GhostSkillLevelDumb   GhostLevel = iota // Random movement, ignores player
	GhostSkillLevelSlow                     // Follows player but makes mistakes
	GhostSkillLevelNormal                   // Standard BFS pathfinding
	GhostSkillLevelSmart                    // Optimized pathfinding with prediction
)

func (s GhostLevel) String() string {
	switch s {
	case GhostSkillLevelDumb:
		return "Dumb"
	case GhostSkillLevelSlow:
		return "Slow"
	case GhostSkillLevelNormal:
		return "Normal"
	case GhostSkillLevelSmart:
		return "Smart"
	default:
		return "Unknown"
	}
}

type DifficultyConfig struct {
	Name        string
	Description string
	GhostSpeeds []float64
	SkillLevels []GhostLevel
	RecalcEvery int // Frames between BFS recalculations
}

func GetDifficultyConfig(difficulty Difficulty) DifficultyConfig {
	switch difficulty {
	case DifficultyEasy:
		return DifficultyConfig{
			Name:        "Easy",
			Description: "Ghosts are slow and not very smart",
			GhostSpeeds: []float64{1.0, 1.1, 1.0, 0.9},
			SkillLevels: []GhostLevel{
				GhostSkillLevelDumb, // Blinky: Random movement
				GhostSkillLevelSlow, // Pinky: Makes mistakes
				GhostSkillLevelDumb, // Inky: Random movement
				GhostSkillLevelSlow, // Clyde: Makes mistakes
			},
			RecalcEvery: 12, // Slower rate
		}
	case DifficultyMedium:
		return DifficultyConfig{
			Name:        "Medium",
			Description: "Balanced gameplay with mixed ghost abilities",
			GhostSpeeds: []float64{1.3, 1.4, 1.2, 1.3}, // Medium speeds
			SkillLevels: []GhostLevel{
				GhostSkillLevelNormal, // Blinky: Standard intelligence
				GhostSkillLevelSlow,   // Pinky: Makes some mistakes
				GhostSkillLevelNormal, // Inky: Standard intelligence
				GhostSkillLevelSlow,   // Clyde: Makes some mistakes
			},
			RecalcEvery: 8, // Medium update rate
		}
	case DifficultyHard:
		return DifficultyConfig{
			Name:        "Hard",
			Description: "Fast and intelligent ghosts",
			GhostSpeeds: []float64{1.5, 1.6, 1.4, 1.5}, // Faster ghosts
			SkillLevels: []GhostLevel{
				GhostSkillLevelSmart,  // Blinky: Smart intelligence
				GhostSkillLevelNormal, // Pinky: Standard intelligence
				GhostSkillLevelSmart,  // Inky: Smart intelligence
				GhostSkillLevelNormal, // Clyde: Standard intelligence
			},
			RecalcEvery: 6, // Standard update rate
		}
	default:
		return GetDifficultyConfig(DifficultyMedium)
	}
}
