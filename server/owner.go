package server

type Request struct {
	Cmd     Command
	ReplyCh chan string
}

func RunOwner(requests <-chan Request, store *Store) {
	// requests <-chan Request is a receive only channel
	// that means it receives from channel to variable
	// or receiving from channel

	for req := range requests {
		response := Dispatch(req.Cmd, store)
		req.ReplyCh <- response
	}
}
