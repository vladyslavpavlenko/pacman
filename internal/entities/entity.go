package entities

import (
	"image/color"
	"math"
)

// Vec represents a 2D vector with float64 coordinates
type Vec struct {
	X, Y float64
}

// Add returns the sum of two vectors
func (v Vec) Add(u Vec) Vec {
	return Vec{v.X + u.X, v.Y + u.Y}
}

// Mul returns the vector multiplied by a scalar
func (v Vec) Mul(s float64) Vec {
	return Vec{v.X * s, v.Y * s}
}

// Len returns the length (magnitude) of the vector
func (v Vec) Len() float64 {
	return math.Hypot(v.X, v.Y)
}

// Norm returns the normalized (unit) vector
func (v Vec) Norm() Vec {
	l := v.Len()
	if l == 0 {
		return Vec{}
	}
	return Vec{v.X / l, v.Y / l}
}

// Eq checks if two vectors are equal
func (v Vec) Eq(u Vec) bool {
	return v.X == u.X && v.Y == u.Y
}

// Approx checks if two vectors are approximately equal
func (v Vec) Approx(u Vec) bool {
	return math.Abs(v.X-u.X) < 0.001 && math.Abs(v.Y-u.Y) < 0.001
}

// IVec represents a 2D vector with integer coordinates
type IVec struct {
	X, Y int
}

// Entity represents a game entity (player or ghost)
type Entity struct {
	Pos       Vec        // pixel center position
	Dir       Vec        // normalized grid direction (up/down/left/right or zero)
	WantDir   Vec        // desired direction from input/AI
	Speed     float64    // movement speed in pixels per frame
	Color     color.RGBA // entity color
	SpawnTile IVec       // spawn tile coordinates
}

// NewPlayer creates a new player entity
func NewPlayer(spawnX, spawnY int, speed float64, color color.RGBA) *Entity {
	return &Entity{
		Pos:       Vec{},
		Dir:       Vec{},
		WantDir:   Vec{},
		Speed:     speed,
		Color:     color,
		SpawnTile: IVec{spawnX, spawnY},
	}
}

// NewGhost creates a new ghost entity
func NewGhost(spawnX, spawnY int, speed float64, color color.RGBA) *Entity {
	return &Entity{
		Pos:       Vec{},
		Dir:       Vec{},
		WantDir:   Vec{},
		Speed:     speed,
		Color:     color,
		SpawnTile: IVec{spawnX, spawnY},
	}
}
