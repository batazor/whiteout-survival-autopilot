package domain

type Config struct {
	Devices []Device `yaml:"devices"`
}

type Device struct {
	Name     string    `yaml:"name"`
	Profiles []Profile `yaml:"profiles"`
}

type Profile struct {
	Email string  `yaml:"email"`
	Gamer []Gamer `yaml:"gamer"`
}
