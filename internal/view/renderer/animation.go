package renderer

import (
	"bytes"
	_ "embed"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/vladyslavpavlenko/pacman/internal/types"
)

//go:embed assets/pacman/down/1.png
var pacmanDown1 []byte

//go:embed assets/pacman/down/2.png
var pacmanDown2 []byte

//go:embed assets/pacman/down/3.png
var pacmanDown3 []byte

//go:embed assets/pacman/left/1.png
var pacmanLeft1 []byte

//go:embed assets/pacman/left/2.png
var pacmanLeft2 []byte

//go:embed assets/pacman/left/3.png
var pacmanLeft3 []byte

//go:embed assets/pacman/right/1.png
var pacmanRight1 []byte

//go:embed assets/pacman/right/2.png
var pacmanRight2 []byte

//go:embed assets/pacman/right/3.png
var pacmanRight3 []byte

//go:embed assets/pacman/up/1.png
var pacmanUp1 []byte

//go:embed assets/pacman/up/2.png
var pacmanUp2 []byte

//go:embed assets/pacman/up/3.png
var pacmanUp3 []byte

// Ghost sprites
//
//go:embed assets/ghosts/blinky.png
var ghostBlinky []byte

//go:embed assets/ghosts/pinky.png
var ghostPinky []byte

//go:embed assets/ghosts/inky.png
var ghostInky []byte

//go:embed assets/ghosts/clyde.png
var ghostClyde []byte

//go:embed assets/ghosts/blue.png
var ghostBlue []byte

type AnimationManager struct {
	sprites      map[string][]*ebiten.Image
	ghostSprites map[string]*ebiten.Image
}

type AnimationEngine struct {
	frameCount    int
	framesPerStep int
	currentStep   int
}

func NewAnimationManager() *AnimationManager {
	am := &AnimationManager{
		sprites:      make(map[string][]*ebiten.Image),
		ghostSprites: make(map[string]*ebiten.Image),
	}

	am.loadSprites()
	am.loadGhostSprites()
	return am
}

func (am *AnimationManager) loadSprites() {
	// Load down sprites
	am.sprites["down"] = []*ebiten.Image{
		am.loadImageFromBytes(pacmanDown1),
		am.loadImageFromBytes(pacmanDown2),
		am.loadImageFromBytes(pacmanDown3),
	}

	// Load left sprites
	am.sprites["left"] = []*ebiten.Image{
		am.loadImageFromBytes(pacmanLeft1),
		am.loadImageFromBytes(pacmanLeft2),
		am.loadImageFromBytes(pacmanLeft3),
	}

	// Load right sprites
	am.sprites["right"] = []*ebiten.Image{
		am.loadImageFromBytes(pacmanRight1),
		am.loadImageFromBytes(pacmanRight2),
		am.loadImageFromBytes(pacmanRight3),
	}

	// Load up sprites
	am.sprites["up"] = []*ebiten.Image{
		am.loadImageFromBytes(pacmanUp1),
		am.loadImageFromBytes(pacmanUp2),
		am.loadImageFromBytes(pacmanUp3),
	}
}

func (am *AnimationManager) loadGhostSprites() {
	// Load ghost sprites
	am.ghostSprites["blinky"] = am.loadImageFromBytes(ghostBlinky)
	am.ghostSprites["pinky"] = am.loadImageFromBytes(ghostPinky)
	am.ghostSprites["inky"] = am.loadImageFromBytes(ghostInky)
	am.ghostSprites["clyde"] = am.loadImageFromBytes(ghostClyde)
	am.ghostSprites["blue"] = am.loadImageFromBytes(ghostBlue)
}

func (am *AnimationManager) loadImageFromBytes(data []byte) *ebiten.Image {
	img, _, err := ebitenutil.NewImageFromReader(bytes.NewReader(data))
	if err != nil {
		log.Fatal("Failed to load image:", err)
	}
	return img
}

func (am *AnimationManager) GetSprite(direction string, frame int) *ebiten.Image {
	if sprites, exists := am.sprites[direction]; exists && frame < len(sprites) {
		return sprites[frame]
	}
	// Fallback to first frame if direction not found
	if sprites, exists := am.sprites["right"]; exists {
		return sprites[0]
	}
	return nil
}

func (am *AnimationManager) GetDirectionFromVector(dir types.Vector) string {
	if dir.X > 0 {
		return "right"
	} else if dir.X < 0 {
		return "left"
	} else if dir.Y > 0 {
		return "down"
	} else if dir.Y < 0 {
		return "up"
	}
	return "right" // Default direction
}

func (am *AnimationManager) GetFrameCount(direction string) int {
	if sprites, exists := am.sprites[direction]; exists {
		return len(sprites)
	}
	return 3 // Default frame count
}

// GetGhostSprite returns the appropriate ghost sprite based on color
func (am *AnimationManager) GetGhostSprite(ghostColor color.RGBA) *ebiten.Image {
	// Map ghost colors to sprite names
	// Red -> Blinky, Pink -> Pinky, Cyan -> Inky, Orange -> Clyde
	if ghostColor.R == 255 && ghostColor.G == 64 && ghostColor.B == 64 {
		return am.ghostSprites["blinky"]
	} else if ghostColor.R == 255 && ghostColor.G == 128 && ghostColor.B == 255 {
		return am.ghostSprites["pinky"]
	} else if ghostColor.R == 64 && ghostColor.G == 255 && ghostColor.B == 255 {
		return am.ghostSprites["inky"]
	} else if ghostColor.R == 255 && ghostColor.G == 128 && ghostColor.B == 0 {
		return am.ghostSprites["clyde"]
	}

	// Default to blinky if color doesn't match
	return am.ghostSprites["blinky"]
}

func NewAnimationEngine(framesPerStep int) *AnimationEngine {
	return &AnimationEngine{
		frameCount:    0,
		framesPerStep: framesPerStep,
		currentStep:   0,
	}
}

func (ae *AnimationEngine) Update() {
	ae.frameCount++
	if ae.frameCount >= ae.framesPerStep {
		ae.frameCount = 0
		ae.currentStep++
	}
}

func (ae *AnimationEngine) GetCurrentFrame(maxFrames int) int {
	return ae.currentStep % maxFrames
}

func (ae *AnimationEngine) Reset() {
	ae.frameCount = 0
	ae.currentStep = 0
}
