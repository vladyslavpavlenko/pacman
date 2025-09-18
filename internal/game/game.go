package game

import (
	"fmt"
	"math/rand"
	"time"

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
	PlayerSpeed          = 2.2 // pixels per frame
	GhostSpeed           = 1.4 // pixels per frame
	RecalcEvery          = 6   // frames between BFS recalcs
	CatchRadius          = 8.0 // pixels
	ScreenScale          = 1
	AppleRadius          = 6.0 // pixels
	SpeedBoostTime       = 300 // frames (5 seconds at 60fps)
	SpeedBoostMultiplier = 1.8
)

// Game represents the main game state
type Game struct {
	level            *model.Level
	player           *model.Player
	ghosts           []*model.Ghost
	score            int
	pelletsCollected int
	finalScore       int
	frame            int
	distMap          *intelligence.DistanceMap
	renderer         *renderer.Renderer
	difficulty       config.Difficulty
	recalcEvery      int
	menu             *ui.UI
	gameState        view.State
	shouldExit       bool
	speedBoostFrames int
	basePlayerSpeed  float64
	debugMode        bool
	ghostAlgorithms  []string
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
		g.pelletsCollected++
	}
}

// checkAppleCollection checks if player collected any apples
func (g *Game) checkAppleCollection() {
	for i := len(g.level.Apples) - 1; i >= 0; i-- {
		apple := g.level.Apples[i]
		if physics.CheckCollision(&g.player.Entity, &apple.Entity, AppleRadius) {
			// Remove apple from level
			g.level.RemoveApple(apple)
			// Add score
			g.score++
			// Apply speed boost
			g.applySpeedBoost()
		}
	}
}

// applySpeedBoost applies a temporary speed boost to the player
func (g *Game) applySpeedBoost() {
	g.speedBoostFrames = SpeedBoostTime
	g.player.Speed = g.basePlayerSpeed * SpeedBoostMultiplier
}

// updateSpeedBoost updates the speed boost timer
func (g *Game) updateSpeedBoost() {
	if g.speedBoostFrames > 0 {
		g.speedBoostFrames--
		if g.speedBoostFrames == 0 {
			g.player.Speed = g.basePlayerSpeed
		}
	}
}

// resetPositions resets all entities to their spawn positions
func (g *Game) resetPositions() {
	physics.ResetEntityPosition(&g.player.Entity)
	for _, ghost := range g.ghosts {
		physics.ResetEntityPosition(&ghost.Entity)
	}
}

// resetLevel resets the level to its original state (restores pellets)
func (g *Game) resetLevel() {
	// Reset the level to original state
	g.level = model.New(nil)

	// Reset counters
	g.score = 0
	g.pelletsCollected = 0
	g.speedBoostFrames = 0
	g.basePlayerSpeed = PlayerSpeed

	// Reset player speed
	g.player.Speed = g.basePlayerSpeed

	// Respawn apples
	g.spawnApples()

	// Reset positions
	g.resetPositions()
}

// updateGhostAI updates a ghost's AI based on the algorithm name
func (g *Game) updateGhostAI(ghost *model.Ghost, algorithmName string) {
	// Define corner positions for scatter behavior
	corners := []types.Vector{
		{X: 1, Y: 1},                                                    // Top-left
		{X: float64(g.level.Width - 2), Y: 1},                           // Top-right
		{X: 1, Y: float64(g.level.Height - 2)},                          // Bottom-left
		{X: float64(g.level.Width - 2), Y: float64(g.level.Height - 2)}, // Bottom-right
	}

	// Define patrol points
	patrolPoints := []types.Vector{
		{X: float64(g.level.Width / 4), Y: float64(g.level.Height / 4)},
		{X: float64(3 * g.level.Width / 4), Y: float64(3 * g.level.Height / 4)},
	}

	switch algorithmName {
	case "Chase":
		intelligence.ChaseAI(&ghost.Entity, g.distMap, g.level, g.player.Pos)
	case "Scatter":
		// Use different corners for different ghosts
		cornerIndex := len(g.ghosts) % len(corners)
		intelligence.ScatterAI(&ghost.Entity, g.distMap, g.level, corners[cornerIndex])
	case "Frightened":
		intelligence.FrightenedAI(&ghost.Entity, g.distMap, g.level)
	case "Patrol":
		intelligence.PatrolAI(&ghost.Entity, g.distMap, g.level, patrolPoints)
	case "Ambush":
		intelligence.AmbushAI(&ghost.Entity, g.distMap, g.level, g.player.Pos, g.player.Dir)
	case "Random":
		intelligence.FrightenedAI(&ghost.Entity, g.distMap, g.level) // Use random movement
	default:
		// Fallback to old AI
		intelligence.GhostAI(&ghost.Entity, g.distMap, g.level, g.difficulty)
	}
}

// assignGhostAlgorithms assigns different algorithms to ghosts based on difficulty
func (g *Game) assignGhostAlgorithms() {
	g.ghostAlgorithms = make([]string, len(g.ghosts))

	switch g.difficulty {
	case config.DifficultyEasy:
		// Easy: Mostly random and patrol, one chase
		algorithms := []string{"Random", "Patrol", "Chase", "Frightened"}
		for i := range g.ghosts {
			g.ghostAlgorithms[i] = algorithms[i%len(algorithms)]
		}
	case config.DifficultyMedium:
		// Medium: Mix of chase, scatter, and patrol
		algorithms := []string{"Chase", "Scatter", "Patrol", "Ambush"}
		for i := range g.ghosts {
			g.ghostAlgorithms[i] = algorithms[i%len(algorithms)]
		}
	case config.DifficultyHard:
		// Hard: Mostly chase and ambush, one scatter
		algorithms := []string{"Chase", "Ambush", "Chase", "Scatter"}
		for i := range g.ghosts {
			g.ghostAlgorithms[i] = algorithms[i%len(algorithms)]
		}
	default:
		// Default: Random assignment
		algorithms := []string{"Chase", "Scatter", "Patrol", "Ambush"}
		for i := range g.ghosts {
			g.ghostAlgorithms[i] = algorithms[i%len(algorithms)]
		}
	}
}

// checkCaught checks if any ghost has caught the player
func (g *Game) checkCaught() {
	for _, ghost := range g.ghosts {
		if physics.CheckCollision(&g.player.Entity, &ghost.Entity, CatchRadius) {
			g.resetLevel() // Reset everything including pellets
			return
		}
	}
}

// Update handles game logic updates
func (g *Game) Update() error {
	if g.gameState == view.StateMenu {
		newState, selectedDiff, shouldExit := g.menu.Update()
		if shouldExit && newState == view.StateMenu {
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

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if g.gameState == view.StateWon {
			g.gameState = view.StateMenu
		} else {
			g.gameState = view.StateMenu
		}
		return nil
	}

	if g.gameState == view.StateWon {
		if inpututil.IsKeyJustPressed(ebiten.KeyR) {
			g.initLevel()
			g.gameState = view.StatePlaying
		}
		return nil
	}

	if g.gameState != view.StatePlaying {
		return nil
	}

	g.frame++

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
		physics.TryTurn(&g.player.Entity, want, g.level)
	}

	if g.frame%g.recalcEvery == 0 {
		g.distMap.BuildBFS(g.player.Pos, g.level)
	}

	for i, ghost := range g.ghosts {
		if i < len(g.ghostAlgorithms) {
			g.updateGhostAI(ghost, g.ghostAlgorithms[i])
		}
	}

	physics.StepMove(&g.player.Entity, g.level)
	for _, ghost := range g.ghosts {
		physics.StepMove(&ghost.Entity, g.level)
	}

	g.consumePellet()
	g.checkAppleCollection()
	g.updateSpeedBoost()

	// Check win condition - only when all pellets are collected
	if g.pelletsCollected >= g.level.TotalPellets {
		g.finalScore = g.score
		g.gameState = view.StateWon
		return nil
	}

	g.checkCaught()

	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.initLevel()
	}

	// Toggle debug mode
	if inpututil.IsKeyJustPressed(ebiten.KeyD) {
		g.debugMode = !g.debugMode
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screenWidth, screenHeight := screen.Bounds().Dx(), screen.Bounds().Dy()

	if g.gameState == view.StateMenu {
		g.renderer.DrawMenu(screen, g.menu, screenWidth, screenHeight)
	} else if g.gameState == view.StatePlaying {
		g.renderer.DrawLevel(screen, g.level)
		g.renderer.DrawPlayer(screen, g.player)
		g.renderer.DrawGhosts(screen, g.ghosts, g.debugMode, g.ghostAlgorithms)
		g.renderer.DrawApples(screen, g.level.Apples)
		g.drawHUD(screen)
	} else if g.gameState == view.StateWon {
		g.renderer.DrawWinScreen(screen, g.finalScore, screenWidth, screenHeight)
	}
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	screenWidth := screen.Bounds().Dx()

	scoreMsg := fmt.Sprintf("Score: %d", g.score)
	g.renderer.TextRenderer.DrawText(screen, scoreMsg, 10, 5, renderer.ColorMenuText, 8)

	difficultyMsg := fmt.Sprintf("Difficulty: %s", g.difficulty.String())
	g.renderer.TextRenderer.DrawText(screen, difficultyMsg, screenWidth-len(difficultyMsg)*9+5, 5, renderer.ColorMenuText, 8)

	if g.speedBoostFrames > 0 {
		boostMsg := fmt.Sprintf("SPEED BOOST! (%d)", g.speedBoostFrames/60+1)
		g.renderer.TextRenderer.DrawText(screen, boostMsg, 10, 25, renderer.ColorSpeedBoost, 8)
	}
}

func (g *Game) setDifficulty(difficulty config.Difficulty) {
	g.difficulty = difficulty
	g.initLevel()
}

// Layout returns the game's logical screen size
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	if g.gameState == view.StateMenu || g.gameState == view.StateWon {
		return outsideWidth, outsideHeight
	}
	if g.level != nil {
		return g.level.Width * physics.TileSize, g.level.Height * physics.TileSize
	}
	return outsideWidth, outsideHeight
}

// spawnApples randomly spawns 2-3 apples on the level
func (g *Game) spawnApples() {
	g.level.Apples = make([]*model.Apple, 0)

	walkableTiles := g.level.GetWalkableTiles()
	if len(walkableTiles) == 0 {
		return
	}

	// Spawn 2-3 apples
	numApples := 2 + rand.Intn(2) // 2 or 3 apples
	appleColor := renderer.ColorApple

	usedTiles := make(map[types.Tile]bool)

	for i := 0; i < numApples && i < len(walkableTiles); i++ {
		// Find a random unused tile
		var tile types.Tile
		attempts := 0
		for {
			tile = walkableTiles[rand.Intn(len(walkableTiles))]
			if !usedTiles[tile] {
				// Make sure it's not too close to player spawn
				playerSpawn, _ := g.level.GetDefaultSpawnPoints()
				if tile.X != playerSpawn.X || tile.Y != playerSpawn.Y {
					break
				}
			}
			attempts++
			if attempts > 100 { // Prevent infinite loop
				break
			}
		}

		usedTiles[tile] = true
		g.level.AddApple(tile.X, tile.Y, appleColor)
		// Set the position of the last added apple
		if len(g.level.Apples) > 0 {
			lastApple := g.level.Apples[len(g.level.Apples)-1]
			lastApple.Pos = physics.TileCenter(tile.X, tile.Y)
		}
	}
}

// initLevel initializes the game level and entities
func (g *Game) initLevel() {
	g.level = model.New(nil) // Use default level data
	g.score = 0
	g.pelletsCollected = 0
	g.frame = 0
	g.speedBoostFrames = 0
	g.basePlayerSpeed = PlayerSpeed

	diffConfig := config.GetDifficultyConfig(g.difficulty)
	g.recalcEvery = diffConfig.RecalcEvery

	g.distMap = intelligence.NewDistanceMap(g.level.Width, g.level.Height)

	playerSpawn, ghostSpawns := g.level.GetDefaultSpawnPoints()

	g.player = model.NewPlayer(playerSpawn.X, playerSpawn.Y, PlayerSpeed, renderer.ColorPac)
	g.player.Pos = physics.TileCenter(playerSpawn.X, playerSpawn.Y)

	g.ghosts = nil
	for i, spawn := range ghostSpawns {
		if i >= len(diffConfig.GhostSpeeds) {
			break
		}

		ghostColor := renderer.ColorGhosts[i%len(renderer.ColorGhosts)]
		ghostSpeed := diffConfig.GhostSpeeds[i]
		skillLevel := config.GhostSkillLevelNormal

		ghost := model.NewGhost(spawn.X, spawn.Y, ghostSpeed, ghostColor, skillLevel)
		ghost.Pos = physics.TileCenter(spawn.X, spawn.Y)
		g.ghosts = append(g.ghosts, ghost)
	}

	// Spawn apples
	g.spawnApples()

	// Assign ghost algorithms based on difficulty
	g.assignGhostAlgorithms()

	g.distMap.BuildBFS(g.player.Pos, g.level)
}

func (g *Game) Run() error {
	g.difficulty = config.DifficultyEasy

	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	ebiten.SetWindowTitle("Pacman")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	ebiten.SetWindowSize(800, 600)

	err := ebiten.RunGame(g)

	if g.shouldExit {
		return nil
	}

	return err
}
