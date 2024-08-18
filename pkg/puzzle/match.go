package puzzle

type Match uint

// inspiration see: https://forum.golangbridge.org/t/can-i-use-enum-in-template/25296
func (m Match) Is(mIn Match) bool { return m == mIn }

const (
	MatchNone Match = iota + 1
	MatchVague
	MatchExact
)
