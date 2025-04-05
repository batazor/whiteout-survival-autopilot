package device

import (
	"context"
	"time"
)

func (d *Device) Play() {
	ctx := context.Background()

	for {
		uc, err := d.Queue.Pop(ctx)
		if err != nil {
			d.Logger.Warn("⚠️ Не удалось извлечь usecase из очереди", "error", err)
			break
		}
		if uc == nil {
			d.Logger.Info("📭 Очередь пуста, переход к следующему игроку")
			break
		}

		d.Logger.Info("🚀 Выполняем usecase", "name", uc.Name, "priority", uc.Priority)

		d.Executor.ExecuteUseCase(ctx, uc, &d.Profiles[d.activeProfileIdx].Gamer[d.activeGamerIdx])
		time.Sleep(2 * time.Second)
	}

	d.Logger.Info("⏭️ Очередь завершена. Готов к переключению.")
}
