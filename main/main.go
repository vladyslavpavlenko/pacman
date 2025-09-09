package main

import (
	"log"

	"github.com/vladyslavpavlenko/pacman/internal/game"
)

func main() {
	if err := game.New().Run(); err != nil {
		log.Fatal(err)
	}
}
