package adb

import (
	"fmt"
	"log/slog"
	"os/exec"
	"time"
)

const gamePackageName = "com.gof.global"

// RestartApplication перезапускает приложение через adb-команды.
func (a *Controller) RestartApplication() error {
	a.logger.Warn("🔄 Перезапуск приложения", slog.String("package", gamePackageName))

	// Закрываем приложение
	closeCmd := exec.Command("adb", "-s", a.deviceID, "shell", "am", "force-stop", gamePackageName)
	if err := closeCmd.Run(); err != nil {
		a.logger.Error("❌ Ошибка при закрытии приложения", slog.String("package", gamePackageName), slog.Any("error", err))
		return fmt.Errorf("failed to close app %s: %w", gamePackageName, err)
	}

	time.Sleep(2 * time.Second)

	// Запускаем приложение снова
	startCmd := exec.Command("adb", "-s", a.deviceID, "shell", "monkey", "-p", gamePackageName, "-c", "android.intent.category.LAUNCHER", "1")
	if err := startCmd.Run(); err != nil {
		a.logger.Error("❌ Ошибка при запуске приложения", slog.String("package", gamePackageName), slog.Any("error", err))
		return fmt.Errorf("failed to start app %s: %w", gamePackageName, err)
	}

	a.logger.Info("✅ Приложение успешно перезапущено", slog.String("package", gamePackageName))

	return nil
}
