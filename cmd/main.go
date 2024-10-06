package main

import (
	"fmt"
	"song-library/internal/config"
)

func main() {
	// Config
	cfg := config.MustLoad()

	fmt.Printf("Config : %v", cfg)
}
