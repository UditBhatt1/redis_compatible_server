package test

import (
	"bufio"
	"errors"
	"io"
	"slices"
	"strings"
	"testing"

	"github.com/UditBhatt1/redis_compatible_server/server"
)

type errKind int

const (
	noErr        errKind = iota // zero value: success rows can omit the field
	eofErr                      // clean disconnect
	truncatedErr                // cut off mid-command
	protocolErr                 // client sent bad bytes
)

func TestParser(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedName string
		expectedArgs []string
		wantErr      errKind
	}{
		// valid
		{name: "PING", input: "*1\r\n$4\r\nPING\r\n", expectedName: "PING"},
		{name: "SET", input: "*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$4\r\nudit\r\n", expectedName: "SET", expectedArgs: []string{"name", "udit"}},
		{name: "lowercase name", input: "*1\r\n$4\r\nping\r\n", expectedName: "ping"},
		{name: "DEL many keys", input: "*4\r\n$3\r\nDEL\r\n$1\r\na\r\n$1\r\nb\r\n$1\r\nc\r\n", expectedName: "DEL", expectedArgs: []string{"a", "b", "c"}},
		{name: "empty bulk string", input: "*2\r\n$3\r\nGET\r\n$0\r\n\r\n", expectedName: "GET", expectedArgs: []string{""}},
		{name: "value contains CRLF", input: "*3\r\n$3\r\nSET\r\n$1\r\nk\r\n$5\r\na\r\nbc\r\n", expectedName: "SET", expectedArgs: []string{"k", "a\r\nbc"}},

		// clean EOF and truncation
		{name: "empty input", input: "", wantErr: eofErr},
		{name: "partial header", input: "*3", wantErr: truncatedErr},
		{name: "complete first header", input: "*3\r\n", wantErr: truncatedErr},
		{name: "stops mid data", input: "*1\r\n$4\r\nPI", wantErr: truncatedErr},
		{name: "missing terminator", input: "*1\r\n$4\r\nPING", wantErr: truncatedErr},
		{name: "second item missing", input: "*2\r\n$3\r\nGET\r\n", wantErr: truncatedErr},

		// protocol errors
		{name: "wrong first byte", input: "+PING\r\n", wantErr: protocolErr},
		{name: "bare newline in header", input: "*1\n$4\r\nPING\r\n", wantErr: protocolErr},
		{name: "array count not a number", input: "*abc\r\n", wantErr: protocolErr},
		{name: "array count zero", input: "*0\r\n", wantErr: protocolErr},
		{name: "array count negative", input: "*-1\r\n", wantErr: protocolErr},
		{name: "array count over cap", input: "*1025\r\n", wantErr: protocolErr},
		{name: "item without dollar", input: "*1\r\nPING\r\n", wantErr: protocolErr},
		{name: "bulk length not a number", input: "*1\r\n$abc\r\n", wantErr: protocolErr},
		{name: "bulk length negative", input: "*1\r\n$-1\r\n", wantErr: protocolErr},
		{name: "bulk length over cap", input: "*1\r\n$99999999\r\n", wantErr: protocolErr},
		{name: "wrong bytes after data", input: "*1\r\n$4\r\nPINGxx", wantErr: protocolErr},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tc.input))
			command, err := server.Parser(reader)

			switch tc.wantErr {
			case noErr:
				if err != nil {
					t.Fatalf("input %q: unexpected error: %v", tc.input, err)
				}
			case eofErr:
				if !errors.Is(err, io.EOF) {
					t.Fatalf("input %q: got error %v, want io.EOF", tc.input, err)
				}
				return
			case truncatedErr:
				if !errors.Is(err, io.ErrUnexpectedEOF) {
					t.Fatalf("input %q: got error %v, want io.ErrUnexpectedEOF", tc.input, err)
				}
				return
			case protocolErr:
				var pe *server.ProtocolError
				if !errors.As(err, &pe) {
					t.Fatalf("input %q: got error %v, want *ProtocolError", tc.input, err)
				}
				return
			default:
				t.Fatalf("unknown error kind %d", tc.wantErr)
			}

			// only reached when no error was expected
			if command.Name != tc.expectedName {
				t.Errorf("input %q: got name %q, want %q", tc.input, command.Name, tc.expectedName)
			}
			if !slices.Equal(command.Args, tc.expectedArgs) {
				t.Errorf("input %q: got args %q, want %q", tc.input, command.Args, tc.expectedArgs)
			}
		})
	}
}
