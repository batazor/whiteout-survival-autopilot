package domain

type Mail struct {
	IsWars     bool `yaml:"is_wars"`     // Флаг, указывающий, что почта содержит информацию о войне.
	IsAlliance bool `yaml:"is_alliance"` // Флаг, указывающий, что почта содержит информацию о альянсе.
	IsSystem   bool `yaml:"is_system"`   // Флаг, указывающий, что почта содержит системные сообщения.
	IsReports  bool `yaml:"is_reports"`  // Флаг, указывающий, что почта содержит отчеты.
}
