package domain

import (
	"time"
)

type VIP struct {
	Level int           `yaml:"level"` // Уровень VIP-статуса (например, 1, 2, 3 и т.д.).
	Time  time.Duration `yaml:"time"`  // Время, оставшееся до окончания VIP-статуса (например, 30 дней).

	State VIPState `yaml:"state"` // Состояние VIP-статуса (например, активен, истекает и т.д.)
}

type VIPState struct {
	IsNotify           bool `yaml:"isNotify"`           // Флаг, указывающий, есть ли события VIP-статуса.
	IsActive           bool `yaml:"isActive"`           // Флаг, указывающий, активен ли VIP-статус.
	IsAdd              bool `yaml:"isAdd"`              // Флаг, указывающий, можно ли добавить VIP-статус.
	IsAward            bool `yaml:"isAward"`            // Флаг, указывающий, доступна ли награда за VIP-статус.
	IsClaim            bool `yaml:"isClaim"`            // Флаг, указывающий, доступна ли награда за VIP-статус.
	IsVIPAddAvailable  bool `yaml:"isVIPAddAvailable"`  // Флаг, указывающий, доступна ли возможность добавить VIP-статус.
	IsVIPAddAvailableX bool `yaml:"isVIPAddAvailableX"` // Флаг, указывающий, доступна ли возможность добавить VIP-статус (дополнительный флаг).
}
