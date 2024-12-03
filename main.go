package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// Create a new room
	r := newRoom()

	// Start the room's run method in a separate goroutine
	go r.run()

	// Set up the WebSocket server
	http.Handle("/chat", r) // Handle the WebSocket connection on the /chat endpoint

	// Start the HTTP server
	fmt.Println("Starting WebSocket server on :8080")
	err := http.ListenAndServe("0.0.0.0:8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe error: ", err)
	}
}
