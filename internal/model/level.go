package model

import (
	"github.com/vladyslavpavlenko/pacman/internal/types"
	"image/color"
)

type Tile byte

const (
	TileEmpty Tile = ' '
	TileWall  Tile = '#'
	TilePel   Tile = '.'
	TileApple Tile = 'a'
)

type Level struct {
	Grid         [][]Tile
	Width        int
	Height       int
	TotalPellets int
	Apples       []*Apple
}

var DefaultLevelData = []string{
	"#####################",
	"#...................#",
	"#.###.#.###.#.###.###",
	"#.#...#...#...#...#.#",
	"#.#.#####.#.#####.#.#",
	"#...................#",
	"#.###.#.###.#.###.###",
	"#.#...#...#...#...#.#",
	"#.#.#####.#.#####.#.#",
	"#...................#",
	"#####################",
}

func New(levelData []string) *Level {
	if len(levelData) == 0 {
		levelData = DefaultLevelData
	}

	level := &Level{
		Width:  len(levelData[0]),
		Height: len(levelData),
		Apples: make([]*Apple, 0),
	}

	level.Grid = make([][]Tile, level.Height)
	level.TotalPellets = 0

	for y := 0; y < level.Height; y++ {
		level.Grid[y] = make([]Tile, level.Width)
		for x := 0; x < level.Width; x++ {
			ch := levelData[y][x]
			switch ch {
			case '#':
				level.Grid[y][x] = TileWall
			case '.':
				level.Grid[y][x] = TilePel
				level.TotalPellets++
			default:
				level.Grid[y][x] = TileEmpty
			}
		}
	}

	return level
}

// CanWalk checks if the given tile coordinates are walkable
func (l *Level) CanWalk(x, y int) bool {
	if x < 0 || y < 0 || x >= l.Width || y >= l.Height {
		return false
	}
	return l.Grid[y][x] != TileWall
}

// GetTile returns the tile at the given coordinates
func (l *Level) GetTile(x, y int) Tile {
	if x < 0 || y < 0 || x >= l.Width || y >= l.Height {
		return TileWall
	}
	return l.Grid[y][x]
}

// SetTile sets the tile at the given coordinates
func (l *Level) SetTile(x, y int, tile Tile) {
	if x < 0 || y < 0 || x >= l.Width || y >= l.Height {
		return
	}
	l.Grid[y][x] = tile
}

// ConsumePellet removes a pellet at the given coordinates and returns true if consumed
func (l *Level) ConsumePellet(x, y int) bool {
	if l.GetTile(x, y) == TilePel {
		l.SetTile(x, y, TileEmpty)
		return true
	}
	return false
}

// GetDefaultSpawnPoints returns the default spawn points for player and ghosts
func (l *Level) GetDefaultSpawnPoints() (playerSpawn types.Tile, ghostSpawns []types.Tile) {
	playerSpawn = types.Tile{X: 1, Y: 1}
	ghostSpawns = []types.Tile{
		{X: l.Width - 2, Y: 1},
		{X: l.Width - 2, Y: l.Height - 2},
		{X: 1, Y: l.Height - 2},
		{X: l.Width / 2, Y: l.Height / 2},
	}
	return
}

// AddApple adds an apple to the level at the specified position
func (l *Level) AddApple(x, y int, color color.RGBA) {
	apple := NewApple(x, y, color)
	// Position will be set by the caller using physics.TileCenter
	l.Apples = append(l.Apples, apple)
}

// RemoveApple removes an apple from the level
func (l *Level) RemoveApple(apple *Apple) {
	for i, a := range l.Apples {
		if a == apple {
			l.Apples = append(l.Apples[:i], l.Apples[i+1:]...)
			break
		}
	}
}

// GetWalkableTiles returns all walkable tile positions
func (l *Level) GetWalkableTiles() []types.Tile {
	var tiles []types.Tile
	for y := 0; y < l.Height; y++ {
		for x := 0; x < l.Width; x++ {
			if l.CanWalk(x, y) {
				tiles = append(tiles, types.Tile{X: x, Y: y})
			}
		}
	}
	return tiles
}
