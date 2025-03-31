package domain

type Config struct {
	Devices []Device `yaml:"devices"`
}

type Device struct {
	Name     string    `yaml:"name"`
	Profiles []Profile `yaml:"profiles"`
}

type Profile struct {
	Email string           `yaml:"email"`
	Gamer []GamerOfProfile `yaml:"gamer"`
}

type GamerOfProfile struct {
	ID       int    `yaml:"id"`
	Nickname string `yaml:"nickname"`
}
