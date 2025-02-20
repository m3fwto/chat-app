package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"chat-app/internal/repository"
	"chat-app/internal/services"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn     *websocket.Conn
	Username string
}

var clients = make(map[*Client]bool)
var mutex sync.Mutex

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("error WebSocket:", err)
		http.Error(w, "error WebSocket", http.StatusInternalServerError)
		return
	}

	client := &Client{Conn: conn}
	defer func() {
		mutex.Lock()
		delete(clients, client)
		mutex.Unlock()
		conn.Close()
	}()

	clients[client] = true

	var msg struct {
		Username string `json:"username"`
	}
	err = conn.ReadJSON(&msg)
	if err != nil {
		log.Println("error reading username:", err)
		return
	}

	if err := services.ValidateUsername(msg.Username); err != nil {
		conn.WriteJSON(map[string]string{"error": err.Error()})
		conn.Close()
		return
	}

	// add user to the connected list
	client.Username = msg.Username
	services.RegisterUserIfNotExists(client.Username)
	services.RegisterConnection(client.Username)

	conn.WriteJSON(map[string]string{
		"status":  "ok",
		"message": fmt.Sprintf("Connected as %s", client.Username),
	})

	// user connected
	for {
		var chatMsg struct {
			Message string `json:"message"`
		}
		err := conn.ReadJSON(&chatMsg)
		if err != nil {
			log.Printf("connection with %s closed: %v", client.Username, err)
			services.RegisterDisconnection(client.Username)
			break
		}

		if chatMsg.Message == "/exit" {
			log.Printf("user %s exit", client.Username)
			services.RegisterDisconnection(client.Username)
			conn.WriteJSON(map[string]string{"message": "connection closed by user"})
			conn.Close()
			break
		}

		services.HandleMessage(client.Username, chatMsg.Message)
		broadcastMessage(client.Username, chatMsg.Message)
	}
}

func broadcastMessage(username, message string) {
	mutex.Lock()
	defer mutex.Unlock()

	for client := range clients {
		if client.Username == username {
			continue
		}

		client.Conn.WriteJSON(map[string]string{
			"user":    username,
			"message": message,
			"time":    time.Now().Format(time.RFC3339),
		})
	}
}

// REST API
func GetMessagesHandler(w http.ResponseWriter, r *http.Request) {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1 // default value
	}
	pageSize, err := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if err != nil || pageSize < 1 {
		pageSize = 10 // default value
	}

	messages, total := repository.GetMessages(page, pageSize)

	response := struct {
		Total    int                  `json:"total"`
		Messages []repository.Message `json:"messages"`
	}{
		Total:    total,
		Messages: messages,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Println("error while encoding JSON:", err)
		http.Error(w, "error server", http.StatusInternalServerError)
	}
}

func GetConnectionLogsHandler(w http.ResponseWriter, r *http.Request) {
	logs := repository.GetConnectionLogs()
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(logs)
	if err != nil {
		log.Println("error while encoding JSON:", err)
		http.Error(w, "server error", http.StatusInternalServerError)
	}
}

func StartWebSocketServer() {
	http.HandleFunc("/ws", HandleWebSocket)
	fmt.Println("WebSocket server running on port 8080")

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-shutdown
		fmt.Println("\nThe server is shutting down. Disabling all users...")

		DisconnectAllUsers()

		time.Sleep(1 * time.Second)

		fmt.Println("Server shut down")
		os.Exit(0)
	}()

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func StartAPIServer() {
	http.HandleFunc("/messages", GetMessagesHandler)
	http.HandleFunc("/logs", GetConnectionLogsHandler)

	fmt.Println("REST API running on port 8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func DisconnectAllUsers() {
	mutex.Lock()
	defer mutex.Unlock()

	if len(clients) == 0 {
		log.Println("There are no active users to disconnect.")
		return
	}

	for client := range clients {
		log.Printf("Disconnecting a user %s...", client.Username)

		services.RegisterDisconnection(client.Username)

		client.Conn.WriteJSON(map[string]string{"message": "connection closed by server"})

		time.Sleep(500 * time.Millisecond)

		client.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure,
			"Server shutdown"))

		client.Conn.Close()
	}

	clients = make(map[*Client]bool)

	log.Println("All users disconnected.")
}
