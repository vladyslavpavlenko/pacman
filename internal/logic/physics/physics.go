package physics

import (
	"math"

	"github.com/vladyslavpavlenko/pacman/internal/logic/entities"
	"github.com/vladyslavpavlenko/pacman/internal/types"
	"github.com/vladyslavpavlenko/pacman/internal/visual/level"
)

const (
	TileSize = 24
)

// TileCenter returns the pixel center coordinates of a tile
func TileCenter(tileX, tileY int) types.Vector {
	return types.Vector{
		X: float64(tileX*TileSize + TileSize/2),
		Y: float64(tileY*TileSize + TileSize/2),
	}
}

// PosToTile converts pixel coordinates to tile coordinates
func PosToTile(pos types.Vector) (tileX, tileY int) {
	return int(pos.X) / TileSize, int(pos.Y) / TileSize
}

// NearCenter checks if a position is near the center of its tile
func NearCenter(pos types.Vector) bool {
	tileX, tileY := PosToTile(pos)
	center := TileCenter(tileX, tileY)
	return math.Abs(pos.X-center.X) <= 1.0 && math.Abs(pos.Y-center.Y) <= 1.0
}

// TryTurn attempts to turn an entity in the desired direction
func TryTurn(entity *entities.Entity, wantDir types.Vector, lvl *level.Level) {
	if wantDir.Eq(entity.Dir) || (wantDir.X == 0 && wantDir.Y == 0) {
		return
	}

	tileX, tileY := PosToTile(entity.Pos)

	// Only allow turn near tile center
	if !NearCenter(entity.Pos) {
		entity.WantDir = wantDir
		return
	}

	nextX, nextY := tileX+int(wantDir.X), tileY+int(wantDir.Y)
	if lvl.CanWalk(nextX, nextY) {
		entity.Dir = wantDir
		entity.WantDir = wantDir
		// Snap to center to avoid drift
		entity.Pos = TileCenter(tileX, tileY)
	}
}

// StepMove moves an entity one step in its current direction
func StepMove(entity *entities.Entity, lvl *level.Level) {
	// If we have a pending desired direction and can take it at center, try it
	if !entity.WantDir.Eq(entity.Dir) && NearCenter(entity.Pos) {
		TryTurn(entity, entity.WantDir, lvl)
	}

	if entity.Dir.Eq(types.Vector{}) {
		return
	}

	// Calculate next position
	next := entity.Pos.Add(entity.Dir.Mul(entity.Speed))

	// Check collision ahead
	nextTileX, nextTileY := PosToTile(next)

	// When crossing tile boundary, ensure the next tile is walkable
	if !lvl.CanWalk(nextTileX, nextTileY) {
		// Clamp to boundary and stop
		// Snap to current tile center along blocked axis
		currentTileX, currentTileY := PosToTile(entity.Pos)
		center := TileCenter(currentTileX, currentTileY)

		if entity.Dir.X != 0 {
			entity.Pos.X = center.X
		}
		if entity.Dir.Y != 0 {
			entity.Pos.Y = center.Y
		}
		entity.Dir = types.Vector{}
		return
	}

	entity.Pos = next
}

// ResetEntityPosition resets an entity to its spawn position
func ResetEntityPosition(entity *entities.Entity) {
	entity.Pos = TileCenter(entity.SpawnTile.X, entity.SpawnTile.Y)
	entity.Dir = types.Vector{}
	entity.WantDir = types.Vector{}
}

// CheckCollision checks if two entities are colliding within the given radius
func CheckCollision(entity1, entity2 *entities.Entity, radius float64) bool {
	return entity1.Pos.Add(entity2.Pos.Mul(-1)).Len() <= radius
}
