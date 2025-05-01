package domain

import "time"

// BuildingDetails описывает параметры здания, которые не зависят от уровня и меняются редко.
// Эти данные могут быть использованы для расчёта времени постройки, стоимости и бонусов здания.
type BuildingDetails struct {
	// ConstructionTime – время, необходимое для постройки здания (например, "2h30m").
	ConstructionTime time.Duration `yaml:"construction_time"`

	// Cost – затраты ресурсов для постройки здания.
	Cost Resources `yaml:"cost"`

	// Benefits – описание бонусов, которые дает здание (например, увеличение производства продовольствия).
	Benefits string `yaml:"benefits"`
}
