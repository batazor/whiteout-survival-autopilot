package device

import (
	"fmt"
	"time"

	"github.com/batazor/whiteout-survival-autopilot/internal/domain"
)

type player struct {
	Email string
	Gamer domain.GamerOfProfile
}

func Start(deviceName string, profiles []domain.Profile) {
	fmt.Printf("📱 Запуск устройства: %s\n", deviceName)

	var queue []player
	for _, profile := range profiles {
		for _, gamer := range profile.Gamer {
			queue = append(queue, player{
				Email: profile.Email,
				Gamer: gamer,
			})
		}
	}

	var prevEmail string
	var prevGamerID int

	for {
		for _, p := range queue {
			if p.Email != prevEmail {
				fmt.Printf("🔄 Смена профиля на устройстве %s: %s → %s\n",
					deviceName, prevEmail, p.Email)
				fmt.Println("   ⚙️  Выполняем набор команд 1 (смена профиля)")
				// TODO: команда при смене профиля
			} else if p.Gamer.ID != prevGamerID {
				fmt.Printf("👤 Смена игрока на устройстве %s: %d → %d\n",
					deviceName, prevGamerID, p.Gamer.ID)
				fmt.Println("   ⚙️  Выполняем набор команд 2 (смена игрока)")
				// TODO: команда при смене игрока
			}

			fmt.Printf("▶️  Активный игрок на %s: Никнейм: %s | ID: %d | Профиль: %s\n",
				deviceName, p.Gamer.Nickname, p.Gamer.ID, p.Email)

			prevEmail = p.Email
			prevGamerID = p.Gamer.ID

			time.Sleep(5 * time.Second)
		}
	}
}
