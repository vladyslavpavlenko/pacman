package physics

import (
	"math"

	"github.com/vladyslavpavlenko/pacman/internal/model"
	"github.com/vladyslavpavlenko/pacman/internal/types"
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

// NearCenter checks if a position is near the center of its tile (for player turning)
func NearCenter(pos types.Vector) bool {
	tileX, tileY := PosToTile(pos)
	center := TileCenter(tileX, tileY)
	return math.Abs(pos.X-center.X) <= 4.0 && math.Abs(pos.Y-center.Y) <= 4.0
}

// AtCenter checks if a position is at the center of its tile (for AI decisions)
func AtCenter(pos types.Vector) bool {
	tileX, tileY := PosToTile(pos)
	center := TileCenter(tileX, tileY)
	return math.Abs(pos.X-center.X) <= 1.0 && math.Abs(pos.Y-center.Y) <= 1.0
}

// VeryNearCenter checks if a position is very close to the center (for immediate turns)
func VeryNearCenter(pos types.Vector) bool {
	tileX, tileY := PosToTile(pos)
	center := TileCenter(tileX, tileY)
	return math.Abs(pos.X-center.X) <= 2.0 && math.Abs(pos.Y-center.Y) <= 2.0
}

// CanTurn checks if an entity can turn in the desired direction
func CanTurn(pos types.Vector, wantDir types.Vector, lvl *model.Level) bool {
	tileX, tileY := PosToTile(pos)
	nextX, nextY := tileX+int(wantDir.X), tileY+int(wantDir.Y)
	return lvl.CanWalk(nextX, nextY)
}

// TryTurn attempts to turn an entity in the desired direction
func TryTurn(entity *model.Entity, wantDir types.Vector, lvl *model.Level) {
	if wantDir.X == 0 && wantDir.Y == 0 {
		return
	}

	entity.WantDir = wantDir

	if wantDir.Eq(entity.Dir) {
		return
	}

	if !CanTurn(entity.Pos, wantDir, lvl) {
		return
	}

	if VeryNearCenter(entity.Pos) {
		entity.Dir = wantDir
		tileX, tileY := PosToTile(entity.Pos)
		entity.Pos = TileCenter(tileX, tileY)
	}
}

// StepMove moves an entity one step in its current direction
func StepMove(entity *model.Entity, lvl *model.Level) {
	if !entity.WantDir.Eq(entity.Dir) && CanTurn(entity.Pos, entity.WantDir, lvl) {
		// If we're at center (for AI) or near center (for player), turn
		if AtCenter(entity.Pos) || NearCenter(entity.Pos) {
			entity.Dir = entity.WantDir
			tileX, tileY := PosToTile(entity.Pos)
			entity.Pos = TileCenter(tileX, tileY)
		}
	}

	if entity.Dir.Eq(types.Vector{}) {
		return
	}

	next := entity.Pos.Add(entity.Dir.Mul(entity.Speed))

	// Check collision with proper hitbox
	if !CanMoveTo(next, lvl) {
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

// CanMoveTo checks if an entity can move to a position considering hitbox
func CanMoveTo(pos types.Vector, lvl *model.Level) bool {
	hitboxSize := float64(TileSize) * 0.8 // 80% of tile size

	corners := []types.Vector{
		{X: pos.X - hitboxSize/2, Y: pos.Y - hitboxSize/2}, // top-left
		{X: pos.X + hitboxSize/2, Y: pos.Y - hitboxSize/2}, // top-right
		{X: pos.X - hitboxSize/2, Y: pos.Y + hitboxSize/2}, // bottom-left
		{X: pos.X + hitboxSize/2, Y: pos.Y + hitboxSize/2}, // bottom-right
	}

	for _, corner := range corners {
		tileX, tileY := PosToTile(corner)
		if !lvl.CanWalk(tileX, tileY) {
			return false
		}
	}

	return true
}

// ResetEntityPosition resets an entity to its spawn position
func ResetEntityPosition(entity *model.Entity) {
	entity.Pos = TileCenter(entity.SpawnTile.X, entity.SpawnTile.Y)
	entity.Dir = types.Vector{}
	entity.WantDir = types.Vector{}
}

// CheckCollision checks if two entities are colliding within the given radius
func CheckCollision(entity1, entity2 *model.Entity, radius float64) bool {
	return entity1.Pos.Add(entity2.Pos.Mul(-1)).Len() <= radius
}
