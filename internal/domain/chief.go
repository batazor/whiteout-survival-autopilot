package domain

type Chief struct {
	Contentment int `yaml:"contentment"` // Очки удовлетворенности губернатора

	State ChiefState `yaml:"state"` // Состояние губернатора
}

type ChiefState struct {
	IsNotify bool `yaml:"isNotify"` // Флаг, указывающий, есть ли заказы губернатора.

	IsUrgentMobilization bool `yaml:"isUrgentMobilization"` // Флаг, указывающий, есть ли срочная мобилизация.
	IsComprehensiveCare  bool `yaml:"isComprehensiveCare"`  // Флаг, указывающий, есть ли всеобъемлющий уход.
	IsProductivityDay    bool `yaml:"isProductivityDay"`    // Флаг, указывающий, есть ли день продуктивности.
	IsRushJob            bool `yaml:"isRushJob"`            // Флаг, указывающий, есть ли спешная работа.
	IsDoubleTime         bool `yaml:"isDoubleTime"`         // Флаг, указывающий, есть ли двойное время.
	IsFestivities        bool `yaml:"isFestivities"`        // Флаг, указывающий, есть ли праздники.
}
