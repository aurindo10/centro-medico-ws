package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
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
	log.Println("Servidor iniciado na porta :8080")
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
