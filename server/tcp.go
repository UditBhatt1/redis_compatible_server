package server

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func TcpListener() {

	listener, err := net.Listen("tcp", ":8090")
	if err != nil {
		log.Fatal("Error listening:", err)
	}

	defer listener.Close()
	fmt.Println("Tcp server started at port 8090")

	for {

		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting conn:", err)
			continue
		}

		go HandleConnection(conn)
	}
}

func HandleConnection(conn net.Conn) {

	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		command, err := Parser(reader)
		if err != nil {
			var pe *ProtocolError
			switch {
			case errors.Is(err, io.EOF):
				// clean disconnect, nothing to log
			case errors.Is(err, io.ErrUnexpectedEOF):
				log.Println("connection interrupted mid-command:", err)
			case errors.As(err, &pe):
				conn.Write([]byte("-ERR " + pe.Error() + "\r\n"))
			default:
				log.Println("read error:", err)
			}
			return
		}
		log.Printf("cmd %q args %q", command.Name, command.Args)

		if strings.EqualFold(command.Name, "QUIT") {
			conn.Write([]byte("+OK\r\n"))
			return
		}

		reply := Dispatch(command)
		if _, err := conn.Write([]byte(reply)); err != nil {
			log.Println("write error:", err)
			return
		}
	}

}
