package model

type SignalType string

const (
	SignalTypeAritectBuys SignalType = "aritect_buys"
	SignalTypeRetransmit  SignalType = "retransmit"
)

func (st SignalType) String() string {
	return string(st)
}

func ParseSignalType(s string) (SignalType, bool) {
	switch s {
	case "aritect_buys":
		return SignalTypeAritectBuys, true
	case "retransmit":
		return SignalTypeRetransmit, true
	default:
		return "", false
	}
}
