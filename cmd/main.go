package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Intervalo de envio da mensagem ping.
	pingPeriod = (pongWait * 9) / 10
	// Tempo que o servidor aguardará por uma mensagem pong antes de encerrar a conexão.
	pongWait = 60 * time.Second
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	conn *websocket.Conn
	group string
}

var clients = make(map[*Client]bool)
var broadcast = make(chan Message)

type Message struct {
	Group string `json:"group"`
	Data  string `json:"data"`
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	group := r.URL.Query().Get("group")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("Erro ao fazer o upgrade:", err)
		return
	}
	defer conn.Close()
	client := &Client{conn: conn, group: group}
	clients[client] = true

	// Define o tempo que o servidor aguardará por uma mensagem pong antes de encerrar a conexão.
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Goroutine para enviar pings periodicamente.
	go func() {
		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()
		for range ticker.C {
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}()

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			delete(clients, client)
			break
		}
		broadcast <- msg
	}
}

func handleMessages() {
	for {
		msg := <-broadcast
		for client := range clients {
			if client.group == msg.Group {
				err := client.conn.WriteJSON(msg)
				if err != nil {
					log.Printf("Erro ao enviar mensagem: %v", err)
					client.conn.Close()
					delete(clients, client)
				}
			}
		}
	}
}

func main() {
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)
	http.HandleFunc("/ws", handleConnections)
	go handleMessages()
	port := fmt.Sprintf(":%s", os.Getenv("PORT"))
	log.Println("Servidor iniciado na porta", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
