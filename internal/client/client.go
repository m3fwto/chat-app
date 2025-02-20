package client

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gorilla/websocket"
)

func StartClient() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Enter the command to connect (/connect name): ")
	scanner.Scan()
	input := scanner.Text()

	if !strings.HasPrefix(input, "/connect ") {
		log.Println("Error: Use /connect name to login.")
		return
	}

	username := strings.TrimSpace(strings.TrimPrefix(input, "/connect "))
	if username == "" {
		log.Println("Error: Username cannot be empty.")
		return
	}

	// Connect to the WebSocket server
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
	if err != nil {
		log.Fatalf("error connecting to server: %v", err)
	}
	defer conn.Close()

	// Sending the username to the server
	err = conn.WriteJSON(map[string]string{"username": username})
	if err != nil {
		log.Println("error sending username:", err)
		return
	}

	var response map[string]string
	err = conn.ReadJSON(&response)
	if err == nil {
		if errorMsg, exists := response["error"]; exists {
			log.Println("error: ", errorMsg)
			conn.Close()
			return
		}
		if status, exists := response["status"]; exists && status == "ok" {
			fmt.Println(response["message"])
		}
	}

	go func() {
		for {
			var msg map[string]string
			err := conn.ReadJSON(&msg)
			if err != nil {
				log.Println("connection to the server is closed:", err)
				break
			}
			fmt.Printf("%s: %s\n", msg["user"], msg["message"])
		}
	}()

	for {
		fmt.Print("> ")
		scanner.Scan()
		text := scanner.Text()

		if text == "/exit" {
			fmt.Println("Disconnecting...")
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("error while closing connection:", err)
			}
			break
		}

		err := conn.WriteJSON(map[string]string{
			"username": username,
			"message":  text,
		})
		if err != nil {
			log.Println("error sending message:", err)
		}
	}
}
