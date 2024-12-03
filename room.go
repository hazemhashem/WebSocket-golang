package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type room struct {
	clients   map[*client]bool
	join      chan *client
	leave     chan *client
	forward   chan []byte
	usernames map[*client]string // To store simple usernames for each client
}

func newRoom() *room {
	return &room{
		forward:   make(chan []byte),
		join:      make(chan *client),
		leave:     make(chan *client),
		clients:   make(map[*client]bool),
		usernames: make(map[*client]string),
	}
}

func (r *room) run() {
	userID := 1 // Simple user ID counter

	for {
		select {
		case client := <-r.join:
			r.clients[client] = true
			// Assign a user ID
			r.usernames[client] = fmt.Sprintf("User %d", userID)
			userID++
			r.broadcastActiveUsers()

		case client := <-r.leave:
			delete(r.clients, client)
			delete(r.usernames, client)
			close(client.receive)
			r.broadcastActiveUsers()

		case msg := <-r.forward:
			for client := range r.clients {
				client.receive <- msg
			}
		}
	}
}

func (r *room) broadcastActiveUsers() {
	// Create a list of active users with their status
	var activeUsers []string
	for client := range r.clients {
		activeUsers = append(activeUsers, fmt.Sprintf("%s (active)", r.usernames[client]))
	}

	// Send the active user list to all connected clients
	for client := range r.clients {
		activeUsersMsg := fmt.Sprintf("Active users: %v", activeUsers)
		client.receive <- []byte(activeUsersMsg)
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  socketBufferSize,
	WriteBufferSize: socketBufferSize,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}
	client := &client{
		socket:  socket,
		receive: make(chan []byte, messageBufferSize),
		room:    r,
	}
	r.join <- client
	defer func() { r.leave <- client }()
	go client.write()
	client.read()
}
