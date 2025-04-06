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
}

// AllianceTech описывает технологические аспекты альянса.
type AllianceTech struct {
	Favorite bool `yaml:"favorite"` // Признак технологии для вклада.
}
