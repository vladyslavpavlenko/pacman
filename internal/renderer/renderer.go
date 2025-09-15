package renderer

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/vladyslavpavlenko/pacman/internal/entities"
	"github.com/vladyslavpavlenko/pacman/internal/level"
	"github.com/vladyslavpavlenko/pacman/internal/physics"
)

var (
	ColorWall   = color.RGBA{R: 40, G: 60, B: 200, A: 255}
	ColorFloor  = color.RGBA{R: 10, G: 10, B: 10, A: 255}
	ColorPellet = color.RGBA{R: 230, G: 230, B: 230, A: 255}
	ColorPac    = color.RGBA{R: 255, G: 215, A: 255}
	ColorGhosts = []color.RGBA{
		{R: 255, G: 64, B: 64, A: 255},   // Blinky-ish
		{R: 255, G: 128, B: 255, A: 255}, // Pinky-ish
		{R: 64, G: 255, B: 255, A: 255},  // Inky-ish
		{R: 255, G: 128, B: 0, A: 255},   // Clyde-ish
	}
)

// Renderer handles all drawing operations for the game.
type Renderer struct{}

func New() *Renderer {
	return &Renderer{}
}

// DrawLevel renders the game level (walls, floor, pellets).
func (r *Renderer) DrawLevel(screen *ebiten.Image, lvl *level.Level) {
	screen.Fill(color.Black)

	// Draw tiles
	for y := 0; y < lvl.Height; y++ {
		for x := 0; x < lvl.Width; x++ {
			px, py := float32(x*physics.TileSize), float32(y*physics.TileSize)

			switch lvl.GetTile(x, y) {
			case level.TileWall:
				vector.DrawFilledRect(screen, px, py, float32(physics.TileSize), float32(physics.TileSize), ColorWall, false)
			default:
				vector.DrawFilledRect(screen, px, py, float32(physics.TileSize), float32(physics.TileSize), ColorFloor, false)
			}

			if lvl.GetTile(x, y) == level.TilePel {
				// Draw pellet in center of tile
				cx, cy := px+float32(physics.TileSize)/2, py+float32(physics.TileSize)/2
				vector.DrawFilledCircle(screen, cx, cy, 3, ColorPellet, false)
			}
		}
	}
}

// DrawEntity renders a single entity.
func (r *Renderer) DrawEntity(screen *ebiten.Image, entity *entities.Entity) {
	radius := float32(physics.TileSize/2 - 3)
	vector.DrawFilledCircle(
		screen,
		float32(entity.Pos.X),
		float32(entity.Pos.Y),
		radius,
		entity.Color,
		false,
	)
}

// DrawGhosts renders all ghost entities
func (r *Renderer) DrawGhosts(screen *ebiten.Image, ghosts []*entities.Entity) {
	for _, ghost := range ghosts {
		r.DrawEntity(screen, ghost)
	}
}

//// DrawHUD renders the game's heads-up display
//func (r *Renderer) DrawHUD(screen *ebiten.Image, score int) {
//	msg := fmt.Sprintf("Score: %d  (R to restart)", score)
//	r.drawDebugText(screen, msg)
//}
//
//// drawDebugText draws simple debug text without external font dependencies
//func (r *Renderer) drawDebugText(dst *ebiten.Image, str string) {
//	// Draw a translucent banner for text background
//	banner := color.RGBA{0, 0, 0, 160}
//	vector.DrawFilledRect(dst, 0, 0, float32(200), 16, banner, false)
//
//	// Note: For actual text rendering, you would typically use:
//	// ebitenutil.DebugPrint(dst, str)
//	// from "github.com/hajimehoshi/ebiten/v2/ebitenutil"
//	// This is kept minimal to avoid additional dependencies
//	_ = str
//}
