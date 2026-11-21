package test

import (
	"testing"
	"time"

	"github.com/UditBhatt1/redis_compatible_server/server"
)

func TestOwner(t *testing.T) {
	store := server.NewStore()
	cmd := server.Command{Name: "ping"}
	requests := make(chan server.Request, 1)
	defer close(requests)

	reply := make(chan string, 1)
	go server.RunOwner(requests, store)

	requests <- server.Request{Cmd: cmd, ReplyCh: reply}
	select {
	case got := <-reply:
		want := "+PONG\r\n"

		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	case <-time.After(time.Second):
		t.Fatal("time out waiting for the owner's reply")
	}
}
