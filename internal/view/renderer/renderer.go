package renderer

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/vladyslavpavlenko/pacman/internal/logic/physics"
	"github.com/vladyslavpavlenko/pacman/internal/model"
	"github.com/vladyslavpavlenko/pacman/internal/view/ui"
)

var (
	ColorWall   = color.RGBA{R: 40, G: 60, B: 200, A: 255}
	ColorFloor  = color.RGBA{R: 10, G: 10, B: 10, A: 255}
	ColorPellet = color.RGBA{R: 230, G: 230, B: 230, A: 255}
	ColorPac    = color.RGBA{R: 255, G: 215, A: 255}
	ColorGhosts = []color.RGBA{
		{R: 255, G: 64, B: 64, A: 255},
		{R: 255, G: 128, B: 255, A: 255},
		{R: 64, G: 255, B: 255, A: 255},
		{R: 255, G: 128, B: 0, A: 255},
	}
	ColorMenuBackground = color.RGBA{R: 20, G: 20, B: 40, A: 255}
	ColorMenuText       = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	ColorMenuSelected   = color.RGBA{R: 255, G: 215, B: 0, A: 255}
	ColorMenuTitle      = color.RGBA{R: 255, G: 100, B: 100, A: 255}
)

type Renderer struct {
	TextRenderer     *TextRenderer
	AnimationManager *AnimationManager
	AnimationEngine  *AnimationEngine
	LastPlayerDir    string // Track last player direction for when stopped
}

func New() *Renderer {
	textRenderer, err := NewTextRenderer()
	if err != nil {
		panic("initialize text renderer: " + err.Error())
	}

	animationManager := NewAnimationManager()
	animationEngine := NewAnimationEngine(16) // Change frame every 16 game frames (much slower animation)

	return &Renderer{
		TextRenderer:     textRenderer,
		AnimationManager: animationManager,
		AnimationEngine:  animationEngine,
		LastPlayerDir:    "right", // Default direction
	}
}

func (r *Renderer) DrawLevel(screen *ebiten.Image, lvl *model.Level) {
	screen.Fill(color.Black)

	for y := 0; y < lvl.Height; y++ {
		for x := 0; x < lvl.Width; x++ {
			px, py := float32(x*physics.TileSize), float32(y*physics.TileSize)

			switch lvl.GetTile(x, y) {
			case model.TileWall:
				vector.DrawFilledRect(screen, px, py, float32(physics.TileSize), float32(physics.TileSize), ColorWall, false)
			default:
				vector.DrawFilledRect(screen, px, py, float32(physics.TileSize), float32(physics.TileSize), ColorFloor, false)
			}

			if lvl.GetTile(x, y) == model.TilePel {
				cx, cy := px+float32(physics.TileSize)/2, py+float32(physics.TileSize)/2
				vector.DrawFilledCircle(screen, cx, cy, 3, ColorPellet, false)
			}
		}
	}
}

func (r *Renderer) DrawEntity(screen *ebiten.Image, entity *model.Entity, frame int) {
	if entity.IsPlayer {
		r.DrawPlayer(screen, entity, frame)
	} else {
		r.DrawGhost(screen, entity)
	}
}

func (r *Renderer) DrawPlayer(screen *ebiten.Image, entity *model.Entity, frame int) {
	// Update last direction if player is moving
	if entity.Dir.X != 0 || entity.Dir.Y != 0 {
		r.LastPlayerDir = r.AnimationManager.GetDirectionFromVector(entity.Dir)
	}

	// Use last direction for rendering (whether moving or stopped)
	direction := r.LastPlayerDir

	// Only animate when the player is actually moving
	if entity.Dir.X != 0 || entity.Dir.Y != 0 {
		r.AnimationEngine.Update()
		frameCount := r.AnimationManager.GetFrameCount(direction)
		animationFrame := r.AnimationEngine.GetCurrentFrame(frameCount)
		sprite := r.AnimationManager.GetSprite(direction, animationFrame)

		if sprite != nil {
			op := &ebiten.DrawImageOptions{}

			// Center the sprite on the entity position
			spriteW, spriteH := sprite.Size()
			op.GeoM.Translate(
				float64(entity.Pos.X)-float64(spriteW)/2,
				float64(entity.Pos.Y)-float64(spriteH)/2,
			)

			screen.DrawImage(sprite, op)
		}
	} else {
		// When not moving, show the first frame (mouth closed) in the last direction
		sprite := r.AnimationManager.GetSprite(direction, 0)

		if sprite != nil {
			op := &ebiten.DrawImageOptions{}

			// Center the sprite on the entity position
			spriteW, spriteH := sprite.Size()
			op.GeoM.Translate(
				float64(entity.Pos.X)-float64(spriteW)/2,
				float64(entity.Pos.Y)-float64(spriteH)/2,
			)

			screen.DrawImage(sprite, op)
		}
	}
}

func (r *Renderer) DrawGhost(screen *ebiten.Image, entity *model.Entity) {
	// Get the appropriate ghost sprite based on color
	sprite := r.AnimationManager.GetGhostSprite(entity.Color)

	if sprite != nil {
		op := &ebiten.DrawImageOptions{}

		// Center the sprite on the entity position
		spriteW, spriteH := sprite.Size()
		op.GeoM.Translate(
			float64(entity.Pos.X)-float64(spriteW)/2,
			float64(entity.Pos.Y)-float64(spriteH)/2,
		)

		screen.DrawImage(sprite, op)
	} else {
		// Fallback to circle if sprite not found
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
}

func (r *Renderer) DrawGhosts(screen *ebiten.Image, ghosts []*model.Entity, frame int) {
	for _, ghost := range ghosts {
		r.DrawEntity(screen, ghost, frame)
	}
}

func (r *Renderer) DrawMenu(screen *ebiten.Image, menu *ui.UI, screenWidth, screenHeight int) {
	screen.Fill(ColorMenuBackground)
	r.drawMenu(screen, menu, screenWidth, screenHeight)
}

func (r *Renderer) drawMenu(screen *ebiten.Image, menu *ui.UI, screenWidth, screenHeight int) {
	titleY := screenHeight / 3
	leftMargin := screenWidth / 4
	r.TextRenderer.DrawText(screen, "PACMAN", leftMargin, titleY, ColorMenuTitle, 32)

	options := menu.GetOptions()
	startY := screenHeight/2 - 30
	lineHeight := 60

	for i, option := range options {
		y := startY + i*lineHeight
		textColor := ColorMenuText

		var displayText string
		if i == 1 {
			displayText = option + menu.GetSelectedDifficulty().String()
		} else {
			displayText = option
		}

		if i == menu.GetSelectedOption() {
			textColor = ColorMenuSelected
			fullText := "> " + displayText
			r.TextRenderer.DrawText(screen, fullText, leftMargin, y, textColor, 18)
		} else {
			r.TextRenderer.DrawText(screen, displayText, leftMargin, y, textColor, 18)
		}
	}
}
