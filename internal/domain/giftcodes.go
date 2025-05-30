package domain

// GiftCodes — корень YAML-файла db/giftCodes.yaml
type GiftCodes struct {
	Codes []GiftCode `yaml:"codes"`
}

// GiftCode — один промокод
type GiftCode struct {
	Name    string            `yaml:"name"`
	Expires string            `yaml:"expires,omitempty"` // RFC 3339 UTC
	UserFor map[string]string `yaml:"userFor"`           // uid → статус
}
