package domain

// Alliance описывает данные об альянсе, к которому принадлежит персонаж.
type Alliance struct {
	Name    string        `yaml:"name"`    // Название альянса.
	MyLevel int           `yaml:"myLevel"` // R5-R1
	Power   int           `yaml:"power"`   // Мощь альянса.
	Members MembersInfo   `yaml:"members"` // Информация о членах альянса.
	State   AllianceState `yaml:"state"`   // Дополнительное состояние альянса.
	Tech    AllianceTech  `yaml:"tech"`    // Технологии альянса (например, вклад).
}

// MembersInfo содержит сведения о количестве участников альянса.
type MembersInfo struct {
	Count int `yaml:"count"` // Текущее количество участников.
	Max   int `yaml:"max"`   // Максимальное количество участников.
}

// AllianceState описывает состояние альянса.
type AllianceState struct {
	IsNeedSupport              bool `yaml:"isNeedSupport"`              // Признак участия в альянсе.
	IsWar                      int  `yaml:"isWar"`                      // Количество текущих войн.
	IsChests                   int  `yaml:"isChests"`                   // Количество доступных сундуков.
	IsAllianceContributeButton bool `yaml:"isAllianceContributeButton"` // Кнопка вклада в технологию
	IsAllianceTechButton       bool `yaml:"isAllianceTechButton"`       // Кнопка технологий альянса
	PolarTerrorCount           int  `yaml:"polarTerrorCount"`           // Количество успешных присоединений на полярного медведя

	// chests
	IsClaimButton        bool `yaml:"isClaimButton"`        // Кнопка на получение награды альянса
	IsCanClaimAllChests  bool `yaml:"isCanClaimAllChests"`  // Кнопка на получение всех сундуков
	LootCountLimit       int  `yaml:"lootCountLimit"`       // Лимит сундуков
	IsGiftClaimAllButton bool `yaml:"isGiftClaimAllButton"` // Кнопка на получение всех подарков
	IsMainChest          bool `yaml:"isMainChest"`          // Кнопка на получение главного сундука
}

// AllianceTech описывает технологические аспекты альянса.
type AllianceTech struct {
	Favorite bool `yaml:"favorite"` // Признак технологии для вклада.
}
