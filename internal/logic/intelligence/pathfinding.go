package intelligence

import (
	"math"
	"math/rand"

	"github.com/vladyslavpavlenko/pacman/internal/config"
	"github.com/vladyslavpavlenko/pacman/internal/logic/physics"
	"github.com/vladyslavpavlenko/pacman/internal/model"
	"github.com/vladyslavpavlenko/pacman/internal/types"
)

type DistanceMap struct {
	distances [][]int
	width     int
	height    int
}

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

func (dm *DistanceMap) BuildBFS(targetPos types.Vector, lvl *model.Level) {
	const infinity = 1 << 30
	for y := 0; y < dm.height; y++ {
		for x := 0; x < dm.width; x++ {
			dm.distances[y][x] = infinity
		}
	}

	targetX, targetY := physics.PosToTile(targetPos)
	if targetX < 0 || targetY < 0 || targetX >= dm.width || targetY >= dm.height {
		return
	}

	type node struct{ x, y int }
	queue := []node{{targetX, targetY}}
	dm.distances[targetY][targetX] = 0
	head := 0

	directions := []types.Tile{
		{X: 1, Y: 0},
		{X: -1, Y: 0},
		{X: 0, Y: 1},
		{X: 0, Y: -1},
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
		return 1 << 30
	}
	return dm.distances[tileY][tileX]
}

type candidate struct {
	dir      types.Vector
	distance int
}

func GhostAI(ghost *model.Entity, distanceMap *DistanceMap, lvl *model.Level, difficulty config.Difficulty) {
	availableLevels := getAvailableSkillLevels(difficulty)

	behaviorIndex := rand.Intn(len(availableLevels))
	chosenBehavior := availableLevels[behaviorIndex]

	switch chosenBehavior {
	case config.GhostSkillLevelDumb:
		dumbGhostAI(ghost, lvl)
	case config.GhostSkillLevelSlow:
		slowGhostAI(ghost, distanceMap, lvl)
	case config.GhostSkillLevelNormal:
		normalGhostAI(ghost, distanceMap, lvl)
	case config.GhostSkillLevelSmart:
		smartGhostAI(ghost, distanceMap, lvl)
	default:
		normalGhostAI(ghost, distanceMap, lvl)
	}
}

// getAvailableSkillLevels returns skill levels available for a given difficulty
func getAvailableSkillLevels(difficulty config.Difficulty) []config.GhostLevel {
	switch difficulty {
	case config.DifficultyEasy:
		return []config.GhostLevel{
			config.GhostSkillLevelDumb,
			config.GhostSkillLevelDumb,
			config.GhostSkillLevelSlow,
			config.GhostSkillLevelSlow,
			config.GhostSkillLevelNormal,
		}
	case config.DifficultyMedium:
		return []config.GhostLevel{
			config.GhostSkillLevelSlow,
			config.GhostSkillLevelSlow,
			config.GhostSkillLevelNormal,
			config.GhostSkillLevelNormal,
			config.GhostSkillLevelSmart,
		}
	case config.DifficultyHard:
		return []config.GhostLevel{
			config.GhostSkillLevelNormal,
			config.GhostSkillLevelNormal,
			config.GhostSkillLevelSmart,
			config.GhostSkillLevelSmart,
			config.GhostSkillLevelSlow,
		}
	default:
		return []config.GhostLevel{config.GhostSkillLevelNormal}
	}
}

// dumbGhostAI implements random movement (ignores player)
func dumbGhostAI(ghost *model.Entity, lvl *model.Level) {
	if !physics.AtCenter(ghost.Pos) && !ghost.Dir.Eq(types.Vector{}) {
		return
	}

	if ghost.Dir.Eq(types.Vector{}) {
		tileX, tileY := physics.PosToTile(ghost.Pos)
		for _, dir := range []types.Vector{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			nextX, nextY := tileX+int(dir.X), tileY+int(dir.Y)
			if lvl.CanWalk(nextX, nextY) {
				ghost.Dir = dir
				ghost.WantDir = dir
				ghost.Pos = physics.TileCenter(tileX, tileY)
				return
			}
		}
	}

	tileX, tileY := physics.PosToTile(ghost.Pos)

	var directions []types.Vector
	checkDirection := func(dx, dy float64) {
		nextX, nextY := tileX+int(dx), tileY+int(dy)
		if lvl.CanWalk(nextX, nextY) {
			directions = append(directions, types.Vector{X: dx, Y: dy})
		}
	}

	checkDirection(1, 0)
	checkDirection(-1, 0)
	checkDirection(0, 1)
	checkDirection(0, -1)

	if len(directions) == 0 {
		if !ghost.Dir.Eq(types.Vector{}) {
			nextX, nextY := tileX+int(ghost.Dir.X), tileY+int(ghost.Dir.Y)
			if lvl.CanWalk(nextX, nextY) {
				ghost.Pos = physics.TileCenter(tileX, tileY)
				return
			}
		}
		for _, dir := range []types.Vector{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			nextX, nextY := tileX+int(dir.X), tileY+int(dir.Y)
			if lvl.CanWalk(nextX, nextY) {
				ghost.Dir = dir
				ghost.WantDir = dir
				ghost.Pos = physics.TileCenter(tileX, tileY)
				return
			}
		}
		ghost.Dir = types.Vector{}
		ghost.WantDir = types.Vector{}
		return
	}

	chosen := directions[rand.Intn(len(directions))]
	ghost.WantDir = chosen
}

// slowGhostAI implements AI that follows player but makes mistakes
func slowGhostAI(ghost *model.Entity, distanceMap *DistanceMap, lvl *model.Level) {
	if !physics.AtCenter(ghost.Pos) && !ghost.Dir.Eq(types.Vector{}) {
		return
	}

	if rand.Float32() < 0.3 {
		dumbGhostAI(ghost, lvl)
		return
	}

	normalGhostAI(ghost, distanceMap, lvl)
}

// smartGhostAI implements optimized pathfinding with some prediction
func smartGhostAI(ghost *model.Entity, distanceMap *DistanceMap, lvl *model.Level) {
	if !physics.AtCenter(ghost.Pos) && !ghost.Dir.Eq(types.Vector{}) {
		return
	}

	tileX, tileY := physics.PosToTile(ghost.Pos)

	var options []candidate

	checkDirection := func(dx, dy float64) {
		nextX, nextY := tileX+int(dx), tileY+int(dy)
		if lvl.CanWalk(nextX, nextY) {
			distance := distanceMap.GetDistance(nextX, nextY)
			exitCount := 0
			for _, checkDir := range []types.Tile{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
				if lvl.CanWalk(nextX+checkDir.X, nextY+checkDir.Y) {
					exitCount++
				}
			}
			if exitCount == 1 {
				distance += 5
			}
			options = append(options, candidate{
				dir:      types.Vector{X: dx, Y: dy},
				distance: distance,
			})
		}
	}

	checkDirection(1, 0)  // right
	checkDirection(-1, 0) // left
	checkDirection(0, 1)  // down
	checkDirection(0, -1) // up

	if len(options) == 0 {
		ghost.Dir = types.Vector{}
		return
	}

	minDistance := 1 << 30
	for _, option := range options {
		if option.distance < minDistance {
			minDistance = option.distance
		}
	}

	var bestOptions []candidate
	for _, option := range options {
		if option.distance == minDistance {
			bestOptions = append(bestOptions, option)
		}
	}

	chosen := bestOptions[rand.Intn(len(bestOptions))]
	ghost.WantDir = chosen.dir
}

// geniusGhostAI implements advanced AI with player movement prediction
func geniusGhostAI(ghost *model.Entity, distanceMap *DistanceMap, lvl *model.Level) {
	if !physics.AtCenter(ghost.Pos) && !ghost.Dir.Eq(types.Vector{}) {
		return
	}

	tileX, tileY := physics.PosToTile(ghost.Pos)

	var options []candidate

	checkDirection := func(dx, dy float64) {
		nextX, nextY := tileX+int(dx), tileY+int(dy)
		if lvl.CanWalk(nextX, nextY) {
			distance := distanceMap.GetDistance(nextX, nextY)

			if distance <= 3 {
				distance -= 2
			}

			exitCount := 0
			for _, checkDir := range []types.Tile{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
				if lvl.CanWalk(nextX+checkDir.X, nextY+checkDir.Y) {
					exitCount++
				}
			}
			if exitCount == 1 {
				distance += 10
			} else if exitCount >= 3 {
				distance -= 1
			}

			options = append(options, candidate{
				dir:      types.Vector{X: dx, Y: dy},
				distance: distance,
			})
		}
	}

	checkDirection(1, 0)  // right
	checkDirection(-1, 0) // left
	checkDirection(0, 1)  // down
	checkDirection(0, -1) // up

	if len(options) == 0 {
		ghost.Dir = types.Vector{}
		return
	}

	minDistance := 1 << 30
	for _, option := range options {
		if option.distance < minDistance {
			minDistance = option.distance
		}
	}

	var bestOptions []candidate
	for _, option := range options {
		if option.distance == minDistance {
			bestOptions = append(bestOptions, option)
		}
	}

	if len(bestOptions) > 1 {
		for _, option := range bestOptions {
			if option.dir.Eq(ghost.Dir) {
				ghost.WantDir = option.dir
				return
			}
		}
	}

	chosen := bestOptions[rand.Intn(len(bestOptions))]
	ghost.WantDir = chosen.dir
}

// normalGhostAI implements standard BFS pathfinding
func normalGhostAI(ghost *model.Entity, distanceMap *DistanceMap, lvl *model.Level) {
	if !physics.AtCenter(ghost.Pos) && !ghost.Dir.Eq(types.Vector{}) {
		return
	}

	tileX, tileY := physics.PosToTile(ghost.Pos)

	var options []candidate

	checkDirection := func(dx, dy float64) {
		nextX, nextY := tileX+int(dx), tileY+int(dy)
		if lvl.CanWalk(nextX, nextY) {
			distance := distanceMap.GetDistance(nextX, nextY)
			options = append(options, candidate{
				dir:      types.Vector{X: dx, Y: dy},
				distance: distance,
			})
		}
	}

	checkDirection(1, 0)
	checkDirection(-1, 0)
	checkDirection(0, 1)
	checkDirection(0, -1)

	if len(options) == 0 {
		if !ghost.Dir.Eq(types.Vector{}) {
			nextX, nextY := tileX+int(ghost.Dir.X), tileY+int(ghost.Dir.Y)
			if lvl.CanWalk(nextX, nextY) {
				ghost.Pos = physics.TileCenter(tileX, tileY)
				return
			}
		}
		ghost.Dir = types.Vector{}
		ghost.WantDir = types.Vector{}
		return
	}

	minDistance := 1 << 30
	for _, option := range options {
		if option.distance < minDistance {
			minDistance = option.distance
		}
	}

	var bestOptions []candidate
	for _, option := range options {
		if option.distance == minDistance {
			bestOptions = append(bestOptions, option)
		}
	}

	chosen := bestOptions[rand.Intn(len(bestOptions))]
	ghost.WantDir = chosen.dir
}

// GhostAlgorithmType represents different AI algorithms for ghosts
type GhostAlgorithmType int

const (
	ChaseAlgorithm GhostAlgorithmType = iota
	ScatterAlgorithm
	FrightenedAlgorithm
	PatrolAlgorithm
	AmbushAlgorithm
	RandomAlgorithm
)

// GetAlgorithmName returns a human-readable name for the algorithm
func GetAlgorithmName(algorithm GhostAlgorithmType) string {
	switch algorithm {
	case ChaseAlgorithm:
		return "Chase"
	case ScatterAlgorithm:
		return "Scatter"
	case FrightenedAlgorithm:
		return "Frightened"
	case PatrolAlgorithm:
		return "Patrol"
	case AmbushAlgorithm:
		return "Ambush"
	case RandomAlgorithm:
		return "Random"
	default:
		return "Unknown"
	}
}

// ChaseAI implements direct pursuit of the player
func ChaseAI(ghost *model.Entity, distanceMap *DistanceMap, lvl *model.Level, playerPos types.Vector) {
	if !physics.AtCenter(ghost.Pos) && !ghost.Dir.Eq(types.Vector{}) {
		return
	}

	tileX, tileY := physics.PosToTile(ghost.Pos)
	playerTileX, playerTileY := physics.PosToTile(playerPos)

	var bestDir types.Vector
	minDistance := math.MaxInt32

	for _, dir := range []types.Vector{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
		nextX, nextY := tileX+int(dir.X), tileY+int(dir.Y)
		if lvl.CanWalk(nextX, nextY) {
			// Manhattan distance: sum of absolute differences in x and y coordinates
			// This gives the shortest path distance in a grid where movement is only horizontal/vertical
			dist := int(math.Abs(float64(nextX-playerTileX)) + math.Abs(float64(nextY-playerTileY)))
			if dist < minDistance {
				minDistance = dist
				bestDir = dir
			}
		}
	}

	if !bestDir.Eq(types.Vector{}) {
		ghost.WantDir = bestDir
	}
}

// ScatterAI makes ghosts move to corners and patrol
func ScatterAI(ghost *model.Entity, distanceMap *DistanceMap, lvl *model.Level, cornerPos types.Vector) {
	if !physics.AtCenter(ghost.Pos) && !ghost.Dir.Eq(types.Vector{}) {
		return
	}

	tileX, tileY := physics.PosToTile(ghost.Pos)
	cornerTileX, cornerTileY := physics.PosToTile(cornerPos)

	var bestDir types.Vector
	minDistance := math.MaxInt32

	for _, dir := range []types.Vector{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
		nextX, nextY := tileX+int(dir.X), tileY+int(dir.Y)
		if lvl.CanWalk(nextX, nextY) {
			dist := int(math.Abs(float64(nextX-cornerTileX)) + math.Abs(float64(nextY-cornerTileY)))
			if dist < minDistance {
				minDistance = dist
				bestDir = dir
			}
		}
	}

	if !bestDir.Eq(types.Vector{}) {
		ghost.WantDir = bestDir
	}
}

// FrightenedAI makes ghosts move randomly when player has power-up
func FrightenedAI(ghost *model.Entity, distanceMap *DistanceMap, lvl *model.Level) {
	if !physics.AtCenter(ghost.Pos) && !ghost.Dir.Eq(types.Vector{}) {
		return
	}

	tileX, tileY := physics.PosToTile(ghost.Pos)
	var validDirs []types.Vector

	for _, dir := range []types.Vector{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
		nextX, nextY := tileX+int(dir.X), tileY+int(dir.Y)
		if lvl.CanWalk(nextX, nextY) {
			validDirs = append(validDirs, dir)
		}
	}

	if len(validDirs) > 0 {
		chosen := validDirs[rand.Intn(len(validDirs))]
		ghost.WantDir = chosen
	}
}

// PatrolAI makes ghosts patrol between two points
func PatrolAI(ghost *model.Entity, distanceMap *DistanceMap, lvl *model.Level, patrolPoints []types.Vector) {
	if !physics.AtCenter(ghost.Pos) && !ghost.Dir.Eq(types.Vector{}) {
		return
	}

	if len(patrolPoints) < 2 {
		FrightenedAI(ghost, distanceMap, lvl)
		return
	}

	tileX, tileY := physics.PosToTile(ghost.Pos)

	distToFirst := math.Abs(float64(tileX-int(patrolPoints[0].X))) + math.Abs(float64(tileY-int(patrolPoints[0].Y)))
	distToSecond := math.Abs(float64(tileX-int(patrolPoints[1].X))) + math.Abs(float64(tileY-int(patrolPoints[1].Y)))

	var targetPos types.Vector
	if distToFirst < distToSecond {
		targetPos = patrolPoints[1]
	} else {
		targetPos = patrolPoints[0]
	}

	targetTileX, targetTileY := physics.PosToTile(targetPos)
	var bestDir types.Vector
	minDistance := math.MaxInt32

	for _, dir := range []types.Vector{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
		nextX, nextY := tileX+int(dir.X), tileY+int(dir.Y)
		if lvl.CanWalk(nextX, nextY) {
			dist := int(math.Abs(float64(nextX-targetTileX)) + math.Abs(float64(nextY-targetTileY)))
			if dist < minDistance {
				minDistance = dist
				bestDir = dir
			}
		}
	}

	if !bestDir.Eq(types.Vector{}) {
		ghost.WantDir = bestDir
	}
}

// AmbushAI tries to intercept the player by predicting their movement
func AmbushAI(ghost *model.Entity, distanceMap *DistanceMap, lvl *model.Level, playerPos types.Vector, playerDir types.Vector) {
	if !physics.AtCenter(ghost.Pos) && !ghost.Dir.Eq(types.Vector{}) {
		return
	}

	predictedPos := playerPos.Add(playerDir.Mul(3))

	tileX, tileY := physics.PosToTile(ghost.Pos)
	predTileX, predTileY := physics.PosToTile(predictedPos)

	var bestDir types.Vector
	minDistance := math.MaxInt32

	for _, dir := range []types.Vector{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
		nextX, nextY := tileX+int(dir.X), tileY+int(dir.Y)
		if lvl.CanWalk(nextX, nextY) {
			dist := int(math.Abs(float64(nextX-predTileX)) + math.Abs(float64(nextY-predTileY)))
			if dist < minDistance {
				minDistance = dist
				bestDir = dir
			}
		}
	}

	if !bestDir.Eq(types.Vector{}) {
		ghost.WantDir = bestDir
	}
}
