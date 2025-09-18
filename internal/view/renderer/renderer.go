package renderer

import (
	"fmt"
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
	ColorApple  = color.RGBA{R: 255, G: 0, B: 0, A: 255}
	ColorGhosts = []color.RGBA{
		{R: 255, G: 64, B: 64, A: 255},
		{R: 255, G: 128, B: 255, A: 255},
		{R: 64, G: 255, B: 255, A: 255},
		{R: 255, G: 128, B: 0, A: 255},
	}
	ColorMenuBackground = color.RGBA{R: 0, G: 0, B: 0, A: 255}
	ColorMenuText       = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	ColorMenuSelected   = color.RGBA{R: 255, G: 215, B: 0, A: 255}
	ColorMenuTitle      = color.RGBA{R: 255, G: 215, B: 0, A: 255}
	ColorSpeedBoost     = color.RGBA{R: 255, G: 255, B: 0, A: 255}
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

func (r *Renderer) DrawEntity(screen *ebiten.Image, entity *model.Ghost) {
	r.DrawGhost(screen, entity)
}

func (r *Renderer) DrawPlayer(screen *ebiten.Image, player *model.Player) {
	if player.Dir.X != 0 || player.Dir.Y != 0 {
		r.LastPlayerDir = r.AnimationManager.GetDirectionFromVector(player.Dir)
	}

	direction := r.LastPlayerDir

	if player.Dir.X != 0 || player.Dir.Y != 0 {
		r.AnimationEngine.Update()
		frameCount := r.AnimationManager.GetFrameCount(direction)
		animationFrame := r.AnimationEngine.GetCurrentFrame(frameCount)
		sprite := r.AnimationManager.GetSprite(direction, animationFrame)

		if sprite != nil {
			op := &ebiten.DrawImageOptions{}

			spriteW, spriteH := sprite.Size()
			op.GeoM.Translate(
				float64(player.Pos.X)-float64(spriteW)/2,
				float64(player.Pos.Y)-float64(spriteH)/2,
			)

			screen.DrawImage(sprite, op)
		}
	} else {
		sprite := r.AnimationManager.GetSprite(direction, 0)

		if sprite != nil {
			op := &ebiten.DrawImageOptions{}

			spriteW, spriteH := sprite.Size()
			op.GeoM.Translate(
				float64(player.Pos.X)-float64(spriteW)/2,
				float64(player.Pos.Y)-float64(spriteH)/2,
			)

			screen.DrawImage(sprite, op)
		}
	}
}

func (r *Renderer) DrawGhost(screen *ebiten.Image, ghost *model.Ghost) {
	sprite := r.AnimationManager.GetGhostSprite(ghost.Color)

	if sprite != nil {
		op := &ebiten.DrawImageOptions{}

		spriteW, spriteH := sprite.Size()
		op.GeoM.Translate(
			float64(ghost.Pos.X)-float64(spriteW)/2,
			float64(ghost.Pos.Y)-float64(spriteH)/2,
		)

		screen.DrawImage(sprite, op)
	} else {
		radius := float32(physics.TileSize/2 - 3)
		vector.DrawFilledCircle(
			screen,
			float32(ghost.Pos.X),
			float32(ghost.Pos.Y),
			radius,
			ghost.Color,
			false,
		)
	}
}

func (r *Renderer) DrawGhosts(screen *ebiten.Image, ghosts []*model.Ghost, debugMode bool, ghostAlgorithms []string) {
	for i, ghost := range ghosts {
		r.DrawGhost(screen, ghost)

		if debugMode && i < len(ghostAlgorithms) {
			algorithmName := ghostAlgorithms[i]
			textX := int(ghost.Pos.X)
			textY := int(ghost.Pos.Y) - 20

			textWidth := len(algorithmName) * 6
			textX -= textWidth / 2

			r.TextRenderer.DrawText(screen, algorithmName, textX, textY, ColorSpeedBoost, 8)
		}
	}
}

func (r *Renderer) DrawApple(screen *ebiten.Image, apple *model.Apple) {
	sprite := r.AnimationManager.GetAppleSprite()
	if sprite != nil {
		op := &ebiten.DrawImageOptions{}

		spriteW, spriteH := sprite.Size()
		op.GeoM.Translate(
			float64(apple.Pos.X)-float64(spriteW)/2,
			float64(apple.Pos.Y)-float64(spriteH)/2,
		)

		screen.DrawImage(sprite, op)
	} else {
		radius := float32(physics.TileSize/2 - 6)
		vector.DrawFilledCircle(
			screen,
			float32(apple.Pos.X),
			float32(apple.Pos.Y),
			radius,
			apple.Color,
			false,
		)
	}
}

func (r *Renderer) DrawApples(screen *ebiten.Image, apples []*model.Apple) {
	for _, apple := range apples {
		r.DrawApple(screen, apple)
	}
}

func (r *Renderer) DrawMenu(screen *ebiten.Image, menu *ui.UI, screenWidth, screenHeight int) {
	screen.Fill(ColorMenuBackground)
	r.drawMenu(screen, menu, screenWidth, screenHeight)
}

func (r *Renderer) DrawWinScreen(screen *ebiten.Image, score int, screenWidth, screenHeight int) {
	screen.Fill(ColorMenuBackground)

	winMsg := "YOU WIN!"
	titleY := screenHeight / 3
	leftMargin := screenWidth / 4
	r.TextRenderer.DrawText(screen, winMsg, leftMargin, titleY, ColorMenuTitle, 32)

	scoreMsg := fmt.Sprintf("Final Score: %d", score)
	scoreY := screenHeight / 2
	r.TextRenderer.DrawText(screen, scoreMsg, leftMargin, scoreY, ColorMenuText, 16)

	instructions := "Press R to restart or ESC to return to menu"
	instructionsY := screenHeight * 2 / 3
	r.TextRenderer.DrawText(screen, instructions, leftMargin, instructionsY, ColorMenuText, 12)
}

func (r *Renderer) drawMenu(screen *ebiten.Image, menu *ui.UI, screenWidth, screenHeight int) {
	titleY := screenHeight / 3
	leftMargin := screenWidth / 4
	r.TextRenderer.DrawText(screen, "PACMAN", leftMargin, titleY, ColorMenuTitle, 32)

	options := menu.GetOptions()
	startY := screenHeight/2 - 30
	lineHeight := 40

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
			r.TextRenderer.DrawText(screen, fullText, leftMargin, y, textColor, 16)
		} else {
			r.TextRenderer.DrawText(screen, displayText, leftMargin, y, textColor, 16)
		}
	}

	// Draw copyright notice at the bottom
	copyrightText := "(c) Vladyslav Pavlenko, TTP-41"
	copyrightY := screenHeight - 30
	r.TextRenderer.DrawText(screen, copyrightText, leftMargin, copyrightY, ColorMenuText, 10)
}
