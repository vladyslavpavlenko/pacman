package ai

import (
	"math/rand"

	"github.com/vladyslavpavlenko/pacman/internal/config"
	"github.com/vladyslavpavlenko/pacman/internal/entities"
	"github.com/vladyslavpavlenko/pacman/internal/level"
	"github.com/vladyslavpavlenko/pacman/internal/physics"
)

// DistanceMap represents a 2D grid of distances from a target position
type DistanceMap struct {
	distances [][]int
	width     int
	height    int
}

// NewDistanceMap creates a new distance map with the given dimensions
func NewDistanceMap(width, height int) *DistanceMap {
	dm := &DistanceMap{
		width:  width,
		height: height,
	}

	dm.distances = make([][]int, height)
	for y := range dm.distances {
		dm.distances[y] = make([]int, width)
	}

	return dm
}

// BuildBFS builds a breadth-first search distance map from the target position
func (dm *DistanceMap) BuildBFS(targetPos entities.Vec, lvl *level.Level) {
	// Initialize all distances to infinity
	const infinity = 1 << 30
	for y := 0; y < dm.height; y++ {
		for x := 0; x < dm.width; x++ {
			dm.distances[y][x] = infinity
		}
	}

	// Start BFS from target position
	targetX, targetY := physics.PosToTile(targetPos)
	if targetX < 0 || targetY < 0 || targetX >= dm.width || targetY >= dm.height {
		return
	}

	type node struct{ x, y int }
	queue := []node{{targetX, targetY}}
	dm.distances[targetY][targetX] = 0
	head := 0

	directions := []entities.IVec{
		{X: 1, Y: 0},  // right
		{X: -1, Y: 0}, // left
		{X: 0, Y: 1},  // down
		{X: 0, Y: -1}, // up
	}

	for head < len(queue) {
		current := queue[head]
		head++

		for _, dir := range directions {
			nextX, nextY := current.x+dir.X, current.y+dir.Y

			if nextX < 0 || nextY < 0 || nextX >= dm.width || nextY >= dm.height {
				continue
			}

			if !lvl.CanWalk(nextX, nextY) {
				continue
			}

			newDistance := dm.distances[current.y][current.x] + 1
			if dm.distances[nextY][nextX] > newDistance {
				dm.distances[nextY][nextX] = newDistance
				queue = append(queue, node{nextX, nextY})
			}
		}
	}
}

// GetDistance returns the distance at the given tile coordinates
func (dm *DistanceMap) GetDistance(tileX, tileY int) int {
	if tileX < 0 || tileY < 0 || tileX >= dm.width || tileY >= dm.height {
		return 1 << 30 // infinity
	}
	return dm.distances[tileY][tileX]
}

// candidate represents a possible movement direction with its distance
type candidate struct {
	dir      entities.Vec
	distance int
}

// GhostAI implements the ghost artificial intelligence with different skill levels
func GhostAI(ghost *entities.Entity, distanceMap *DistanceMap, lvl *level.Level, skillLevel config.GhostSkillLevel) {
	switch skillLevel {
	case config.SkillLevelDumb:
		dumbGhostAI(ghost, lvl)
	case config.SkillLevelSlow:
		slowGhostAI(ghost, distanceMap, lvl)
	case config.SkillLevelNormal:
		normalGhostAI(ghost, distanceMap, lvl)
	case config.SkillLevelSmart:
		smartGhostAI(ghost, distanceMap, lvl)
	case config.SkillLevelGenius:
		geniusGhostAI(ghost, distanceMap, lvl)
	default:
		normalGhostAI(ghost, distanceMap, lvl)
	}
}

// dumbGhostAI implements random movement (ignores player)
func dumbGhostAI(ghost *entities.Entity, lvl *level.Level) {
	// Only make decisions at intersections or when stopped
	if !physics.NearCenter(ghost.Pos) && !ghost.Dir.Eq(entities.Vec{}) {
		return
	}

	tileX, tileY := physics.PosToTile(ghost.Pos)

	// Get all walkable directions
	var directions []entities.Vec
	checkDirection := func(dx, dy float64) {
		nextX, nextY := tileX+int(dx), tileY+int(dy)
		if lvl.CanWalk(nextX, nextY) {
			directions = append(directions, entities.Vec{X: dx, Y: dy})
		}
	}

	checkDirection(1, 0)  // right
	checkDirection(-1, 0) // left
	checkDirection(0, 1)  // down
	checkDirection(0, -1) // up

	if len(directions) == 0 {
		ghost.Dir = entities.Vec{}
		return
	}

	// Choose random direction
	chosen := directions[rand.Intn(len(directions))]
	ghost.WantDir = chosen
	ghost.Dir = chosen
	ghost.Pos = physics.TileCenter(tileX, tileY)
}

// slowGhostAI implements AI that follows player but makes mistakes
func slowGhostAI(ghost *entities.Entity, distanceMap *DistanceMap, lvl *level.Level) {
	// Only make decisions at intersections or when stopped
	if !physics.NearCenter(ghost.Pos) && !ghost.Dir.Eq(entities.Vec{}) {
		return
	}

	// 30% chance to make a random move instead of optimal
	if rand.Float32() < 0.3 {
		dumbGhostAI(ghost, lvl)
		return
	}

	// Otherwise use normal AI
	normalGhostAI(ghost, distanceMap, lvl)
}

// smartGhostAI implements optimized pathfinding with some prediction
func smartGhostAI(ghost *entities.Entity, distanceMap *DistanceMap, lvl *level.Level) {
	// Only make decisions at intersections or when stopped
	if !physics.NearCenter(ghost.Pos) && !ghost.Dir.Eq(entities.Vec{}) {
		return
	}

	tileX, tileY := physics.PosToTile(ghost.Pos)

	// Check all possible directions
	var options []candidate

	checkDirection := func(dx, dy float64) {
		nextX, nextY := tileX+int(dx), tileY+int(dy)
		if lvl.CanWalk(nextX, nextY) {
			distance := distanceMap.GetDistance(nextX, nextY)
			// Smart ghosts prefer directions that don't lead to dead ends
			// Check if the next position has multiple exits
			exitCount := 0
			for _, checkDir := range []entities.IVec{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
				if lvl.CanWalk(nextX+checkDir.X, nextY+checkDir.Y) {
					exitCount++
				}
			}
			// Prefer positions with more exits (avoid dead ends)
			if exitCount == 1 {
				distance += 5 // Penalty for dead ends
			}
			options = append(options, candidate{
				dir:      entities.Vec{X: dx, Y: dy},
				distance: distance,
			})
		}
	}

	checkDirection(1, 0)  // right
	checkDirection(-1, 0) // left
	checkDirection(0, 1)  // down
	checkDirection(0, -1) // up

	if len(options) == 0 {
		ghost.Dir = entities.Vec{}
		return
	}

	// Find the minimum distance
	minDistance := 1 << 30
	for _, option := range options {
		if option.distance < minDistance {
			minDistance = option.distance
		}
	}

	// Collect all options with minimum distance
	var bestOptions []candidate
	for _, option := range options {
		if option.distance == minDistance {
			bestOptions = append(bestOptions, option)
		}
	}

	// Choose randomly among the best options
	chosen := bestOptions[rand.Intn(len(bestOptions))]
	ghost.WantDir = chosen.dir
	ghost.Dir = chosen.dir
	ghost.Pos = physics.TileCenter(tileX, tileY)
}

// geniusGhostAI implements advanced AI with player movement prediction
func geniusGhostAI(ghost *entities.Entity, distanceMap *DistanceMap, lvl *level.Level) {
	// Only make decisions at intersections or when stopped
	if !physics.NearCenter(ghost.Pos) && !ghost.Dir.Eq(entities.Vec{}) {
		return
	}

	tileX, tileY := physics.PosToTile(ghost.Pos)

	// Check all possible directions
	var options []candidate

	checkDirection := func(dx, dy float64) {
		nextX, nextY := tileX+int(dx), tileY+int(dy)
		if lvl.CanWalk(nextX, nextY) {
			distance := distanceMap.GetDistance(nextX, nextY)

			// Genius ghosts try to predict player movement and cut them off
			// They prefer positions that would intercept the player's likely path
			if distance <= 3 {
				// When close, try to cut off escape routes
				distance -= 2
			}

			// Avoid dead ends even more aggressively
			exitCount := 0
			for _, checkDir := range []entities.IVec{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
				if lvl.CanWalk(nextX+checkDir.X, nextY+checkDir.Y) {
					exitCount++
				}
			}
			if exitCount == 1 {
				distance += 10 // Higher penalty for dead ends
			} else if exitCount >= 3 {
				distance -= 1 // Prefer intersections for better positioning
			}

			options = append(options, candidate{
				dir:      entities.Vec{X: dx, Y: dy},
				distance: distance,
			})
		}
	}

	checkDirection(1, 0)  // right
	checkDirection(-1, 0) // left
	checkDirection(0, 1)  // down
	checkDirection(0, -1) // up

	if len(options) == 0 {
		ghost.Dir = entities.Vec{}
		return
	}

	// Find the minimum distance
	minDistance := 1 << 30
	for _, option := range options {
		if option.distance < minDistance {
			minDistance = option.distance
		}
	}

	// Collect all options with minimum distance
	var bestOptions []candidate
	for _, option := range options {
		if option.distance == minDistance {
			bestOptions = append(bestOptions, option)
		}
	}

	// Genius ghosts are more deterministic - prefer consistent directions when possible
	if len(bestOptions) > 1 {
		// Prefer continuing in the same direction if it's still optimal
		for _, option := range bestOptions {
			if option.dir.Eq(ghost.Dir) {
				ghost.WantDir = option.dir
				ghost.Dir = option.dir
				ghost.Pos = physics.TileCenter(tileX, tileY)
				return
			}
		}
	}

	// Choose randomly among the best options if no continuation preference
	chosen := bestOptions[rand.Intn(len(bestOptions))]
	ghost.WantDir = chosen.dir
	ghost.Dir = chosen.dir
	ghost.Pos = physics.TileCenter(tileX, tileY)
}

// normalGhostAI implements standard BFS pathfinding
func normalGhostAI(ghost *entities.Entity, distanceMap *DistanceMap, lvl *level.Level) {
	// Only make decisions at intersections or when stopped
	if !physics.NearCenter(ghost.Pos) && !ghost.Dir.Eq(entities.Vec{}) {
		return
	}

	tileX, tileY := physics.PosToTile(ghost.Pos)

	// Check all possible directions
	var options []candidate

	checkDirection := func(dx, dy float64) {
		nextX, nextY := tileX+int(dx), tileY+int(dy)
		if lvl.CanWalk(nextX, nextY) {
			distance := distanceMap.GetDistance(nextX, nextY)
			options = append(options, candidate{
				dir:      entities.Vec{X: dx, Y: dy},
				distance: distance,
			})
		}
	}

	checkDirection(1, 0)  // right
	checkDirection(-1, 0) // left
	checkDirection(0, 1)  // down
	checkDirection(0, -1) // up

	if len(options) == 0 {
		ghost.Dir = entities.Vec{}
		return
	}

	// Find the minimum distance
	minDistance := 1 << 30
	for _, option := range options {
		if option.distance < minDistance {
			minDistance = option.distance
		}
	}

	// Collect all options with minimum distance
	var bestOptions []candidate
	for _, option := range options {
		if option.distance == minDistance {
			bestOptions = append(bestOptions, option)
		}
	}

	// Choose randomly among the best options
	chosen := bestOptions[rand.Intn(len(bestOptions))]
	ghost.WantDir = chosen.dir
	ghost.Dir = chosen.dir

	// Snap to center
	ghost.Pos = physics.TileCenter(tileX, tileY)
}
