package device

import (
	"context"

	"github.com/batazor/whiteout-survival-autopilot/internal/config"
)

func (d *Device) loadUseCases(ctx context.Context, path string) {
	usecaseLoader := config.NewUseCaseLoader(path)

	usecases, err := usecaseLoader.LoadAll(ctx)
	if err != nil {
		d.Logger.Error("❌ Не удалось загрузить usecases", "error", err)
		return
	}

	for _, uc := range usecases {
		select {
		case <-ctx.Done():
			d.Logger.Warn("🛑 Загрузка usecases отменена")
			return
		default:
			if err := d.Queue.Push(ctx, uc); err != nil {
				d.Logger.Error("❌ Не удалось добавить usecase в очередь", "usecase", uc.Name, "error", err)
			} else {
				d.Logger.Info("📥 Usecase добавлен в очередь", "usecase", uc.Name)
			}
		}
	}
}
