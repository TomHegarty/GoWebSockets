package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/net/websocket"
)

type Server struct {
	conns map[*websocket.Conn]bool
}

// Message represents the structure of an incoming message
type Message struct {
	username string `json:"username"`
	color    string `json:"color"`
	content  string `json:"content"`
}

func NewServer() *Server {
	return &Server{
		conns: make(map[*websocket.Conn]bool),
	}
}

func (s *Server) handleWSOrderBook(ws *websocket.Conn) {
	fmt.Println("new incoming connection from client to order book feed:", ws.RemoteAddr())

	for {
		payload := fmt.Sprintf("orderbook data -> %d\n", time.Now().UnixNano())

		ws.Write([]byte(payload))
		time.Sleep(time.Second * 2)
	}
}

func (s *Server) handleWS(ws *websocket.Conn) {
	fmt.Println("new incoming connection from client:", ws.RemoteAddr())

	s.conns[ws] = true

	s.readLoop(ws)
}

func (s *Server) readLoop(ws *websocket.Conn) {
	buf := make([]byte, 1024)

	for {
		n, err := ws.Read(buf)

		if err != nil {

			if err == io.EOF {
				// if client ends connection, close the loop
				break
			}

			fmt.Println("read error:", err)
			continue // if we want to keep the connection open if the client sends malformed request
		}

		rawMsg := buf[:n]

		// Parse the message
		var message Message
		if err := json.Unmarshal(rawMsg, &message); err != nil {
			fmt.Println("error parsing message:", err)
			continue
		}

		// Process the message
		fmt.Printf("Received message from %s (color: %s): %s\n", message.username, message.color, message.content)

		// re-encode and broadcast the message
		encodedMsg, _ := json.Marshal(message)

		s.broadcast(encodedMsg)
	}
}

func (s *Server) broadcast(b []byte) {
	for ws := range s.conns {
		go func(ws *websocket.Conn) {
			if _, err := ws.Write(b); err != nil {
				fmt.Println("Write error:", err)
			}
		}(ws)
	}
}

func main() {
	fmt.Println("sockets main")

	server := NewServer()
	http.Handle("/ws", websocket.Handler(server.handleWS))
	http.Handle("/orderbookfeed", websocket.Handler(server.handleWSOrderBook))
	http.ListenAndServe(":3000", nil)
}
