package domain

type Mail struct {
	IsHasMail int `yaml:"isHasMail"` // Флаг, указывающий, что почта есть.

	State MailState `yaml:"state"` // Состояние почты.
}

type MailState struct {
	IsWars     int `yaml:"isWars"`     // Флаг, указывающий, что почта содержит информацию о войне.
	IsAlliance int `yaml:"isAlliance"` // Флаг, указывающий, что почта содержит информацию об альянсе.
	IsSystem   int `yaml:"isSystem"`   // Флаг, указывающий, что почта содержит системные сообщения.
	IsReports  int `yaml:"isReports"`  // Флаг, указывающий, что почта содержит отчеты.
}
