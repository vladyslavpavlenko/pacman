package renderer

import (
	"bytes"
	_ "embed"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

//go:embed assets/emulogic_font.ttf
var font []byte

type TextRenderer struct {
	faceSource *text.GoTextFaceSource
}

func NewTextRenderer() (*TextRenderer, error) {
	faceSource, err := text.NewGoTextFaceSource(bytes.NewReader(font))
	if err != nil {
		return nil, err
	}

	return &TextRenderer{
		faceSource: faceSource,
	}, nil
}

func (tr *TextRenderer) DrawText(screen *ebiten.Image, textStr string, x, y int, textColor color.RGBA, fontSize float64) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.ScaleWithColor(textColor)

	text.Draw(screen, textStr, &text.GoTextFace{
		Source: tr.faceSource,
		Size:   fontSize,
	}, op)
}

func (tr *TextRenderer) DrawTextCentered(screen *ebiten.Image, textStr string, x, y int, textColor color.RGBA, fontSize float64) {
	textWidth := len(textStr) * int(fontSize*0.6)
	startX := x - textWidth/2
	tr.DrawText(screen, textStr, startX, y, textColor, fontSize)
}

func (tr *TextRenderer) DrawTextWithOptions(screen *ebiten.Image, textStr string, x, y int, textColor color.RGBA, fontSize float64, lineSpacing float64) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.ScaleWithColor(textColor)
	op.LineSpacing = lineSpacing

	text.Draw(screen, textStr, &text.GoTextFace{
		Source: tr.faceSource,
		Size:   fontSize,
	}, op)
}
