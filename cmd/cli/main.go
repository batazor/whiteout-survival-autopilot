package main

import (
	"log"

	teaapp "github.com/batazor/whiteout-survival-autopilot/internal/tea"
)

// Minimal entry point: we just create the App and run it.
func main() {
	app, err := teaapp.NewApp()
	if err != nil {
		log.Fatalf("init failed: %v", err)
	}

	if err := app.Run(); err != nil {
		log.Fatalf("runtime failed: %v", err)
	}
}
