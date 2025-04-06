package domain

type Config struct {
	Devices []Device `yaml:"devices"`
}

type Device struct {
	Name     string    `yaml:"name"`
	Profiles []Profile `yaml:"profiles"`
}

// AllProfiles возвращает плоский список всех профилей из всех девайсов
func (c *Config) AllProfiles() []Profile {
	var result []Profile
	for _, device := range c.Devices {
		result = append(result, device.Profiles...)
	}
	return result
}
