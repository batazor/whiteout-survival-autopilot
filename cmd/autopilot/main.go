package main

import (
	"log"

	"github.com/batazor/whiteout-survival-autopilot/internal/app"
)

// Minimal entry point: we just create the App and run it.
func main() {
	application, err := app.NewApp()
	if err != nil {
		log.Fatalf("failed to create application: %v", err)
	}
	if err := application.Run(); err != nil {
		log.Fatalf("application error: %v", err)
	}
}
