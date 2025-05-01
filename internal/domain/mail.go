package domain

type Mail struct {
	IsHasMail bool `yaml:"isHasMail"` // Флаг, указывающий, что есть непрочтенные письма.

	State MailState `yaml:"state"` // Состояние почты.
}

type MailState struct {
	IsWars     bool `yaml:"isWars"`     // Флаг, указывающий, что почта содержит информацию о войне.
	IsAlliance bool `yaml:"isAlliance"` // Флаг, указывающий, что почта содержит информацию об альянсе.
	IsSystem   bool `yaml:"isSystem"`   // Флаг, указывающий, что почта содержит системные сообщения.
	IsReports  bool `yaml:"isReports"`  // Флаг, указывающий, что почта содержит отчеты.
}
