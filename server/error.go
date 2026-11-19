package server

type ProtocolError struct {
	Msg string
}

func (p *ProtocolError) Error() string {
	return p.Msg
}
