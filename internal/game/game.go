package game

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/vladyslavpavlenko/pacman/internal/ai"
	"github.com/vladyslavpavlenko/pacman/internal/config"
	"github.com/vladyslavpavlenko/pacman/internal/entities"
	"github.com/vladyslavpavlenko/pacman/internal/level"
	"github.com/vladyslavpavlenko/pacman/internal/physics"
	"github.com/vladyslavpavlenko/pacman/internal/renderer"
)

const (
	PlayerSpeed = 2.2 // pixels per frame
	GhostSpeed  = 1.4 // pixels per frame
	RecalcEvery = 6   // frames between BFS recalcs
	CatchRadius = 8.0 // pixels
	ScreenScale = 1
)

// Game represents the main game state
type Game struct {
	level       *level.Level
	player      *entities.Entity
	ghosts      []*entities.Entity
	score       int
	frame       int
	distMap     *ai.DistanceMap
	renderer    *renderer.Renderer
	difficulty  config.Difficulty
	recalcEvery int
}

// New creates a new game instance
func New() *Game {
	return &Game{
		renderer: renderer.New(),
	}
}

// consumePellet checks if player is on a pellet and consumes it
func (g *Game) consumePellet() {
	tileX, tileY := physics.PosToTile(g.player.Pos)
	if g.level.ConsumePellet(tileX, tileY) {
		g.score++
	}
}

// resetPositions resets all entities to their spawn positions
func (g *Game) resetPositions() {
	physics.ResetEntityPosition(g.player)
	for _, ghost := range g.ghosts {
		physics.ResetEntityPosition(ghost)
	}
}

// checkCaught checks if any ghost has caught the player
func (g *Game) checkCaught() {
	for _, ghost := range g.ghosts {
		if physics.CheckCollision(g.player, ghost, CatchRadius) {
			g.resetPositions()
			return
		}
	}
}

// Update handles game logic updates
func (g *Game) Update() error {
	g.frame++

	// Handle difficulty selection
	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		g.setDifficulty(config.DifficultyEasy)
	}
	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		g.setDifficulty(config.DifficultyMedium)
	}
	if inpututil.IsKeyJustPressed(ebiten.Key3) {
		g.setDifficulty(config.DifficultyHard)
	}
	if inpututil.IsKeyJustPressed(ebiten.Key4) {
		g.setDifficulty(config.DifficultyNightmare)
	}

	// Handle player movement input
	want := entities.Vec{}
	if ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		want = entities.Vec{X: -1, Y: 0}
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		want = entities.Vec{X: 1, Y: 0}
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		want = entities.Vec{X: 0, Y: -1}
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		want = entities.Vec{X: 0, Y: 1}
	}
	if !want.Eq(entities.Vec{}) {
		physics.TryTurn(g.player, want, g.level)
	}

	// Rebuild BFS distance map periodically (frequency depends on difficulty)
	if g.frame%g.recalcEvery == 0 {
		g.distMap.BuildBFS(g.player.Pos, g.level)
	}

	// Ghost AI decisions with individual skill levels
	for _, ghost := range g.ghosts {
		ai.GhostAI(ghost, g.distMap, g.level, ghost.SkillLevel)
	}

	// Movement
	physics.StepMove(g.player, g.level)
	for _, ghost := range g.ghosts {
		physics.StepMove(ghost, g.level)
	}

	// Consume pellets and check win condition
	g.consumePellet()
	if g.score >= g.level.TotalPellets {
		g.initLevel()
	}

	// Check if player was caught
	g.checkCaught()

	// Restart with R key
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.initLevel()
	}

	return nil
}

// Draw renders the game
func (g *Game) Draw(screen *ebiten.Image) {
	g.renderer.DrawLevel(screen, g.level)
	g.renderer.DrawEntity(screen, g.player)
	g.renderer.DrawGhosts(screen, g.ghosts)
	g.drawHUD(screen)
}

// drawHUD draws the game's heads-up display
func (g *Game) drawHUD(screen *ebiten.Image) {
	// Simple text display without external dependencies
	msg := fmt.Sprintf("Score: %d | Difficulty: %s | Keys: 1-4 to change difficulty, R to restart",
		g.score, g.difficulty.String())
	_ = msg // Placeholder - in a real implementation you'd draw this text
}

// setDifficulty changes the game difficulty and reinitializes
func (g *Game) setDifficulty(difficulty config.Difficulty) {
	g.difficulty = difficulty
	g.initLevel()
}

// Layout returns the game's logical screen size
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.level.Width * physics.TileSize, g.level.Height * physics.TileSize
}

// initLevel initializes the game level and entities
func (g *Game) initLevel() {
	// Create level
	g.level = level.New(nil) // Use default level data
	g.score = 0
	g.frame = 0

	// Get difficulty configuration
	diffConfig := config.GetDifficultyConfig(g.difficulty)
	g.recalcEvery = diffConfig.RecalcEvery

	// Create distance map for AI
	g.distMap = ai.NewDistanceMap(g.level.Width, g.level.Height)

	// Get spawn points
	playerSpawn, ghostSpawns := g.level.GetDefaultSpawnPoints()

	// Create player
	g.player = entities.NewPlayer(playerSpawn.X, playerSpawn.Y, PlayerSpeed, renderer.ColorPac)
	g.player.Pos = physics.TileCenter(playerSpawn.X, playerSpawn.Y)

	// Create ghosts with difficulty-based configuration
	g.ghosts = nil
	for i, spawn := range ghostSpawns {
		if i >= len(diffConfig.GhostSpeeds) || i >= len(diffConfig.SkillLevels) {
			break // Don't create more ghosts than configured
		}

		ghostColor := renderer.ColorGhosts[i%len(renderer.ColorGhosts)]
		ghostSpeed := diffConfig.GhostSpeeds[i]
		skillLevel := diffConfig.SkillLevels[i]

		ghost := entities.NewGhost(spawn.X, spawn.Y, ghostSpeed, ghostColor, skillLevel)
		ghost.Pos = physics.TileCenter(spawn.X, spawn.Y)
		g.ghosts = append(g.ghosts, ghost)
	}

	// Build initial distance map
	g.distMap.BuildBFS(g.player.Pos, g.level)
}

// Run initializes and starts the game
func (g *Game) Run() error {
	rand.Seed(time.Now().UnixNano())

	// Set default difficulty
	g.difficulty = config.DifficultyMedium
	g.initLevel()

	ebiten.SetWindowTitle("Pacman - Multi-Difficulty (Keys: 1-4 for difficulty, WASD/Arrows to move, R to restart)")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowSize(g.level.Width*physics.TileSize*ScreenScale, g.level.Height*physics.TileSize*ScreenScale)
	return ebiten.RunGame(g)
}
