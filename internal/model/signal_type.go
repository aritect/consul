package model

type SignalType string

const (
	SignalTypeBuys       SignalType = "buys"
	SignalTypeRetransmit SignalType = "retransmit"
)

func (st SignalType) String() string {
	return string(st)
}

func ParseSignalType(s string) (SignalType, bool) {
	switch s {
	case "buys":
		return SignalTypeBuys, true
	case "retransmit":
		return SignalTypeRetransmit, true
	default:
		return "", false
	}
}
