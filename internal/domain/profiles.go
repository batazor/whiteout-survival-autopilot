package domain

type Profiles []Profile

type Profile struct {
	Email string  `yaml:"email"`
	Gamer []Gamer `yaml:"gamer"`
}

// Len returns the number of profiles.
func (p Profiles) Len() int {
	return len(p)
}

// Swap swaps the profiles at indices i and j.
func (p Profiles) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

// Less reports whether the profile at index i should sort before the one at index j.
// Here, we sort by the Email field.
func (p Profiles) Less(i, j int) bool {
	return p[i].Email < p[j].Email
}
