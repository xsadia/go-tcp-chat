package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

type Client struct {
	Conn net.Conn
}

type MessageType = int

const (
	ClientConnected MessageType = iota
	ClientDisconnected
	NewMessage
)

type Message struct {
	MessageType MessageType
	Conn        net.Conn
	Body        []byte
}

var userCount = 0

func handleMessages(messages chan Message) {
	clients := map[string]*Client{}
	for {
		message := <-messages
		switch message.MessageType {
		case ClientConnected:
			a := message.Conn.RemoteAddr().String()
			clients[a] = &Client{
				Conn: message.Conn,
			}

			for _, client := range clients {
				if client.Conn.RemoteAddr().String() != a {
					client.Conn.Write([]byte(fmt.Sprintf("[%v] System: %v connected", time.Now().Format("02/01/2006"), a)))
				}
			}

		case ClientDisconnected:
			a := message.Conn.RemoteAddr().String()
			delete(clients, a)
			for _, client := range clients {
				client.Conn.Write([]byte(fmt.Sprintf("[%v] System: %v disconnected", time.Now().Format("02/01/2006"), a)))
			}

		case NewMessage:
			a := message.Conn.RemoteAddr().String()
			for _, client := range clients {
				if client.Conn.RemoteAddr().String() != a {
					client.Conn.Write([]byte(fmt.Sprintf("[%v] %v: %s", time.Now().Format("02/01/2006"), a, message.Body)))
				} else {
					client.Conn.Write([]byte(fmt.Sprintf("[%v] %v: %s", time.Now().Format("02/01/2006"), "you", message.Body)))
				}
			}
		}
	}
}

func handleConnection(conn net.Conn, messages chan Message) {
	buf := make([]byte, 128)
	defer conn.Close()
	for {
		len, err := conn.Read(buf)
		if err != nil {
			messages <- Message{
				MessageType: ClientDisconnected,
				Conn:        conn,
			}

			return
		}

		messages <- Message{
			MessageType: NewMessage,
			Conn:        conn,
			Body:        buf[:len],
		}
	}
}

func main() {
	l, err := net.Listen("tcp", ":6969")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	log.Println("Server listening at port 6969", time.Now().Format("02/01/2006"))
	messages := make(chan Message)
	go handleMessages(messages)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		messages <- Message{
			MessageType: ClientConnected,
			Conn:        conn,
		}
		go handleConnection(conn, messages)
	}
}
