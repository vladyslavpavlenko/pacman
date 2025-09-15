package game

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/vladyslavpavlenko/pacman/internal/config"
	"github.com/vladyslavpavlenko/pacman/internal/logic/intelligence"
	"github.com/vladyslavpavlenko/pacman/internal/logic/physics"
	"github.com/vladyslavpavlenko/pacman/internal/model"
	"github.com/vladyslavpavlenko/pacman/internal/types"
	"github.com/vladyslavpavlenko/pacman/internal/view"
	"github.com/vladyslavpavlenko/pacman/internal/view/renderer"
	"github.com/vladyslavpavlenko/pacman/internal/view/ui"
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
	level       *model.Level
	player      *model.Entity
	ghosts      []*model.Entity
	score       int
	frame       int
	distMap     *intelligence.DistanceMap
	renderer    *renderer.Renderer
	difficulty  config.Difficulty
	recalcEvery int
	menu        *ui.UI
	gameState   view.State
	shouldExit  bool
}

// New creates a new game instance
func New() *Game {
	return &Game{
		renderer:   renderer.New(),
		menu:       ui.New(),
		gameState:  view.StateMenu,
		shouldExit: false,
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
	// Handle menu state
	if g.gameState == view.StateMenu {
		newState, selectedDiff, shouldExit := g.menu.Update()
		if shouldExit && newState == view.StateMenu {
			// Exit game
			g.shouldExit = true
			return nil
		}
		if newState == view.StatePlaying {
			g.gameState = view.StatePlaying
			g.difficulty = selectedDiff
			g.initLevel()
		}
		return nil
	}

	// Handle ESC key to return to menu
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.gameState = view.StateMenu
		return nil
	}

	// Game logic only runs when playing
	if g.gameState != view.StatePlaying {
		return nil
	}

	g.frame++

	// Handle player movement input
	want := types.Vector{}
	if ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		want = types.Vector{X: -1, Y: 0}
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		want = types.Vector{X: 1, Y: 0}
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		want = types.Vector{X: 0, Y: -1}
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		want = types.Vector{X: 0, Y: 1}
	}
	if !want.Eq(types.Vector{}) {
		physics.TryTurn(g.player, want, g.level)
	}

	// Rebuild BFS distance map periodically (frequency depends on difficulty)
	if g.frame%g.recalcEvery == 0 {
		g.distMap.BuildBFS(g.player.Pos, g.level)
	}

	// Ghost AI decisions with individual skill levels
	for _, ghost := range g.ghosts {
		intelligence.GhostAI(ghost, g.distMap, g.level, ghost.SkillLevel)
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

func (g *Game) Draw(screen *ebiten.Image) {
	if g.gameState == view.StateMenu {
		screenWidth, screenHeight := screen.Bounds().Dx(), screen.Bounds().Dy()
		g.renderer.DrawMenu(screen, g.menu, screenWidth, screenHeight)
	} else if g.gameState == view.StatePlaying {
		g.renderer.DrawLevel(screen, g.level)
		g.renderer.DrawEntity(screen, g.player, g.frame)
		g.renderer.DrawGhosts(screen, g.ghosts, g.frame)
		g.drawHUD(screen)
	}
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	msg := fmt.Sprintf("Score: %d", g.score)
	screenWidth := screen.Bounds().Dx()
	g.renderer.TextRenderer.DrawTextCentered(screen, msg, screenWidth/2, 5, renderer.ColorMenuText, 10)
}

func (g *Game) setDifficulty(difficulty config.Difficulty) {
	g.difficulty = difficulty
	g.initLevel()
}

// Layout returns the game's logical screen size
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	if g.gameState == view.StateMenu {
		// For menu, use the actual window size
		return outsideWidth, outsideHeight
	}
	// For game, use level-based size
	if g.level != nil {
		return g.level.Width * physics.TileSize, g.level.Height * physics.TileSize
	}
	// Fallback
	return outsideWidth, outsideHeight
}

// initLevel initializes the game level and entities
func (g *Game) initLevel() {
	// Create level
	g.level = model.New(nil) // Use default level data
	g.score = 0
	g.frame = 0

	// Get difficulty configuration
	diffConfig := config.GetDifficultyConfig(g.difficulty)
	g.recalcEvery = diffConfig.RecalcEvery

	// Create distance map for AI
	g.distMap = intelligence.NewDistanceMap(g.level.Width, g.level.Height)

	// Get spawn points
	playerSpawn, ghostSpawns := g.level.GetDefaultSpawnPoints()

	// Create player
	g.player = model.NewPlayer(playerSpawn.X, playerSpawn.Y, PlayerSpeed, renderer.ColorPac)
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

		ghost := model.NewGhost(spawn.X, spawn.Y, ghostSpeed, ghostColor, skillLevel)
		ghost.Pos = physics.TileCenter(spawn.X, spawn.Y)
		g.ghosts = append(g.ghosts, ghost)
	}

	// Build initial distance map
	g.distMap.BuildBFS(g.player.Pos, g.level)
}

func (g *Game) Run() error {
	g.difficulty = config.DifficultyMedium

	ebiten.SetWindowTitle("Pacman")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	ebiten.SetWindowSize(800, 600)

	err := ebiten.RunGame(g)

	if g.shouldExit {
		return nil
	}

	return err
}
