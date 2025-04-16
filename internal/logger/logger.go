package logger

import (
	"log/slog"
	"os"
)

// InitializeLogger sets up a JSON logger for the given use case name,
// including source code location information.
func InitializeLogger(usecaseName string) (*slog.Logger, error) {
	// Создаём JSON handler, пишущий в консоль
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		// AddSource: true, // можно включить, если хочешь видеть файл:строку
	})

	return slog.New(handler), nil
}
