package server

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// example string "*1\r\n$4\r\nPING\r\n"
const maxBulkLength = 1024 * 1024 * 5 // 5 mb
const maxArrayLength = 1024

// badPrefix builds the "expected '*', got 'x'" error. A \r or \n in the
// offending byte is shown as a space, so it can't break the reply's line.
func badPrefix(want byte, body string) *ProtocolError {
	got := " "
	if len(body) > 0 && body[0] != '\r' && body[0] != '\n' {
		got = body[:1]
	}
	return &ProtocolError{fmt.Sprintf("Protocol error: expected '%c', got '%s'", want, got)}
}

func Parser(reader *bufio.Reader) (Command, error) {
	var items []string

	line, err := reader.ReadString('\n')
	if err != nil {
		// no bytes received: peer closed between commands, normal end of session
		if errors.Is(err, io.EOF) && line == "" {
			return Command{}, io.EOF
		}
		if errors.Is(err, io.EOF) && line != "" {
			return Command{}, io.ErrUnexpectedEOF
		}
		return Command{}, err
	}

	body, ok := strings.CutSuffix(line, "\r\n")
	if !ok {
		return Command{}, &ProtocolError{"Protocol error: expected CRLF line ending"}
	}

	body, ok = strings.CutPrefix(body, "*")
	if !ok {
		return Command{}, badPrefix('*', body)
	}
	// now length is *1 and * is removed so only 1
	arrayLength, err := strconv.Atoi(body)
	if err != nil || arrayLength <= 0 || arrayLength > maxArrayLength {
		return Command{}, &ProtocolError{"Protocol error: invalid multibulk length"}
	}

	for range arrayLength {
		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				return Command{}, io.ErrUnexpectedEOF
			}
			return Command{}, err
		}

		body, ok := strings.CutSuffix(line, "\r\n")
		if !ok {
			return Command{}, &ProtocolError{"Protocol error: expected CRLF line ending"}
		}

		body, ok = strings.CutPrefix(body, "$")
		if !ok {
			return Command{}, badPrefix('$', body)
		}

		bulkLength, err := strconv.Atoi(body)
		if err != nil || bulkLength < 0 || bulkLength > maxBulkLength {
			return Command{}, &ProtocolError{"Protocol error: invalid bulk length"}
		}

		buffer := make([]byte, bulkLength)

		if _, err := io.ReadFull(reader, buffer); err != nil {
			if errors.Is(err, io.EOF) {
				// zero bytes read so we promote eof to EUEOF
				return Command{}, io.ErrUnexpectedEOF
			}
			// if some bytes are read it already returns EUEOF, else other errors
			return Command{}, err
		}

		var crlf [2]byte
		if _, err := io.ReadFull(reader, crlf[:]); err != nil {
			if errors.Is(err, io.EOF) {
				return Command{}, io.ErrUnexpectedEOF
			}
			return Command{}, err
		}
		if crlf[0] != '\r' || crlf[1] != '\n' {
			return Command{}, &ProtocolError{"Protocol error: expected CRLF after bulk data"}
		}

		items = append(items, string(buffer))
	}

	// just a recheck
	if len(items) == 0 {
		return Command{}, &ProtocolError{"Protocol error: invalid multibulk length"}
	}

	return Command{items[0], items[1:]}, nil
}
