package renderer

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/vladyslavpavlenko/pacman/internal/logic/entities"
	"github.com/vladyslavpavlenko/pacman/internal/logic/physics"
	"github.com/vladyslavpavlenko/pacman/internal/visual/level"
	"github.com/vladyslavpavlenko/pacman/internal/visual/menu"
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
	TextRenderer *TextRenderer
}

func New() *Renderer {
	textRenderer, err := NewTextRenderer()
	if err != nil {
		panic("initialize text renderer: " + err.Error())
	}

	return &Renderer{
		TextRenderer: textRenderer,
	}
}

func (r *Renderer) DrawLevel(screen *ebiten.Image, lvl *level.Level) {
	screen.Fill(color.Black)

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
				cx, cy := px+float32(physics.TileSize)/2, py+float32(physics.TileSize)/2
				vector.DrawFilledCircle(screen, cx, cy, 3, ColorPellet, false)
			}
		}
	}
}

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

func (r *Renderer) DrawGhosts(screen *ebiten.Image, ghosts []*entities.Entity) {
	for _, ghost := range ghosts {
		r.DrawEntity(screen, ghost)
	}
}

func (r *Renderer) DrawMenu(screen *ebiten.Image, menu *menu.Menu, screenWidth, screenHeight int) {
	screen.Fill(ColorMenuBackground)
	r.drawMenu(screen, menu, screenWidth, screenHeight)
}

func (r *Renderer) drawMenu(screen *ebiten.Image, menu *menu.Menu, screenWidth, screenHeight int) {
	titleY := screenHeight / 3
	r.TextRenderer.DrawTextCentered(screen, "PACMAN", screenWidth/2, titleY, ColorMenuTitle, 48)

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
			r.TextRenderer.DrawTextCentered(screen, fullText, screenWidth/2, y, textColor, 24)
		} else {
			r.TextRenderer.DrawTextCentered(screen, displayText, screenWidth/2, y, textColor, 24)
		}
	}
}
