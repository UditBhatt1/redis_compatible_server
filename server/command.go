package server

import (
	"fmt"
	"strings"
)

type Command struct {
	Name string
	Args []string
}

type Handler func(args []string) string

type CommandEntry struct {
	handler Handler
	arity   int // exact number of args after the command name
}

var commands = map[string]CommandEntry{
	"PING": {handler: ping, arity: 0},
}

func ping(args []string) string {
	return "+PONG\r\n"
}

func Dispatch(cmd Command) string {
	entry, ok := commands[strings.ToUpper(cmd.Name)]
	if !ok {
		return fmt.Sprintf("-ERR unknown command '%s'\r\n", cmd.Name)
	}
	if len(cmd.Args) != entry.arity {
		return fmt.Sprintf("-ERR wrong number of arguments for '%s' command\r\n",
			strings.ToLower(cmd.Name))
	}
	return entry.handler(cmd.Args)
}
