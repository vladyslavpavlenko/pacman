package ai

import (
	"math/rand"

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

// GhostAI implements the ghost artificial intelligence
func GhostAI(ghost *entities.Entity, distanceMap *DistanceMap, lvl *level.Level) {
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
