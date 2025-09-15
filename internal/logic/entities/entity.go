package entities

import (
	"image/color"

	"github.com/vladyslavpavlenko/pacman/internal/config"
	"github.com/vladyslavpavlenko/pacman/internal/types"
)

// Entity represents a game entity (player or ghost)
type Entity struct {
	Pos        types.Vector      // pixel center position
	Dir        types.Vector      // normalized grid direction (up/down/left/right or zero)
	WantDir    types.Vector      // desired direction from input/AI
	Speed      float64           // movement speed in pixels per frame
	Color      color.RGBA        // entity color
	SpawnTile  types.Tile        // spawn tile coordinates
	SkillLevel config.GhostLevel // AI skill level (only used for ghosts)
	IsPlayer   bool              // true if this is the player entity
}

type PlayerConfig struct {
}

func NewPlayer(spawnX, spawnY int, speed float64, color color.RGBA) *Entity {
	return &Entity{
		Pos:        types.Vector{},
		Dir:        types.Vector{},
		WantDir:    types.Vector{},
		Speed:      speed,
		Color:      color,
		SpawnTile:  types.Tile{spawnX, spawnY},
		SkillLevel: config.GhostSkillLevelNormal, // Not used for player
		IsPlayer:   true,
	}
}

func NewGhost(spawnX, spawnY int, speed float64, color color.RGBA, skillLevel config.GhostLevel) *Entity {
	return &Entity{
		Pos:        types.Vector{},
		Dir:        types.Vector{},
		WantDir:    types.Vector{},
		Speed:      speed,
		Color:      color,
		SpawnTile:  types.Tile{spawnX, spawnY},
		SkillLevel: skillLevel,
		IsPlayer:   false,
	}
}
