package domain

type DailyMissions struct {
	IsNotify bool `yaml:"isNotify"` // Флаг, указывающий, доступна ли награда за ежедневные миссии.

	IsReseachOneTechnologies bool `yaml:"isReseachOneTechnologies"` // Флаг, указывающий, завершено ли исследование одной технологии.
	IsGatherMeat             bool `yaml:"isGatherMeat"`             // Флаг, указывающий, завершено ли задание на сбор мяса.
}
