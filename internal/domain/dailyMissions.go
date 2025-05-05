package domain

type DailyMissions struct {
	IsNotify bool `yaml:"isNotify"` // Флаг, указывающий, доступна ли награда за ежедневные миссии.

	State DailyMissionsState `yaml:"state"` // Состояние ежедневных миссий.

	Tasks Tasks `yaml:"tasks"` // Задания, которые необходимо выполнить.
}

type Tasks struct {
	IsReseachOneTechnologies bool `yaml:"isReseachOneTechnologies"` // Флаг, указывающий, завершено ли исследование одной технологии.
	IsGatherMeat             bool `yaml:"isGatherMeat"`             // Флаг, указывающий, завершено ли задание на сбор мяса.
}

type DailyMissionsState struct {
	IsClaimAll bool `yaml:"isClaimAll"` // Флаг, указывающий, завершено ли задание на получение всех наград.
}

type GrowthMissions struct {
	IsNotify bool `yaml:"isNotify"` // Флаг, указывающий, доступна ли награда за ежедневные миссии.
}
