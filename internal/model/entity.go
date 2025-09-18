package model

import (
	"image/color"

	"github.com/vladyslavpavlenko/pacman/internal/config"
	"github.com/vladyslavpavlenko/pacman/internal/types"
)

type Entity struct {
	Pos       types.Vector // pixel center position
	Dir       types.Vector // normalized grid direction (up/down/left/right or zero)
	WantDir   types.Vector // desired direction from input/AI
	Speed     float64      // movement speed in pixels per frame
	Color     color.RGBA   // entity color
	SpawnTile types.Tile   // spawn tile coordinates
}

type Player struct {
	Entity
}

type Ghost struct {
	Entity
	SkillLevel config.GhostLevel
}

type Apple struct {
	Entity
}

func NewPlayer(spawnX, spawnY int, speed float64, color color.RGBA) *Player {
	return &Player{
		Entity: Entity{
			Pos:       types.Vector{},
			Dir:       types.Vector{},
			WantDir:   types.Vector{},
			Speed:     speed,
			Color:     color,
			SpawnTile: types.Tile{spawnX, spawnY},
		},
	}
}

// NewGhost creates a new ghost entity
func NewGhost(spawnX, spawnY int, speed float64, color color.RGBA, skillLevel config.GhostLevel) *Ghost {
	return &Ghost{
		Entity: Entity{
			Pos:       types.Vector{},
			Dir:       types.Vector{},
			WantDir:   types.Vector{},
			Speed:     speed,
			Color:     color,
			SpawnTile: types.Tile{spawnX, spawnY},
		},
		SkillLevel: skillLevel,
	}
}

// NewApple creates a new apple entity
func NewApple(spawnX, spawnY int, color color.RGBA) *Apple {
	return &Apple{
		Entity: Entity{
			Pos:       types.Vector{},
			Dir:       types.Vector{},
			WantDir:   types.Vector{},
			Speed:     0, // Apples don't move
			Color:     color,
			SpawnTile: types.Tile{spawnX, spawnY},
		},
	}
}
