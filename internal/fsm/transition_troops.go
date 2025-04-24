package fsm

const (
	StateInfantry = "infantry"
	StateLancer   = "lancer"
	StateMarksman = "marksman"
)

var troopsTransitionPaths = map[string]map[string][]TransitionStep{
	StateInfantry: {
		StateMainCity: {},
	},
	StateLancer:   {},
	StateMarksman: {},
}
