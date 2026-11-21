package server

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net"
	"strings"
)

func HandleConnection(conn net.Conn, requests chan<- Request) {

	//request chan<- Request is a send only channel
	// it sends value from variable to channel
	// or sending to channel
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

		if strings.EqualFold(command.Name, "QUIT") {
			conn.Write([]byte("+OK\r\n"))
			return
		}

		replyCh := make(chan string, 1)

		requests <- Request{Cmd: command, ReplyCh: replyCh}
		response := <-replyCh

		if _, err := conn.Write([]byte(response)); err != nil {
			log.Println("write error:", err)
			return
		}
	}

}

func TcpListener() {

	listener, err := net.Listen("tcp", ":8090")
	if err != nil {
		log.Fatal("Error listening:", err)
	}

	defer listener.Close()
	log.Println("Tcp server started at port 8090")

	store := NewStore()
	requests := make(chan Request, 100)

	go RunOwner(requests, store)

	for {

		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting conn:", err)
			continue
		}

		go HandleConnection(conn, requests)
	}
}
