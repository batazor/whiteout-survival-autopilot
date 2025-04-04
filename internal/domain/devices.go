package domain

type Config struct {
	Devices []Device `yaml:"devices"`
}

type Device struct {
	Name     string    `yaml:"name"`
	Profiles []Profile `yaml:"profiles"`
}
