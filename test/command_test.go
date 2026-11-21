package test

import (
	"github.com/UditBhatt1/redis_compatible_server/server"
	"testing"
)

func TestCommand(t *testing.T) {

	test := []struct {
		name string
		cmd  server.Command
		want string
	}{
		{"ping", server.Command{Name: "ping"}, "+PONG\r\n"},
		{"ping_uppercase", server.Command{Name: "PING"}, "+PONG\r\n"},
		{"unknown", server.Command{Name: "nope"}, "-ERR unknown command 'nope'\r\n"},
		{"ping extra arg", server.Command{Name: "ping", Args: []string{"x"}},
			"-ERR wrong number of arguments for 'ping' command\r\n"},
	}

	for _, tc := range test {
		store := server.NewStore()
		t.Run(tc.name, func(t *testing.T) {
			got := server.Dispatch(tc.cmd, store)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}

}
