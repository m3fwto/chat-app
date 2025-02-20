package main

import (
	"fmt"
	"sync"

	"chat-app/internal/config"
	"chat-app/internal/database"
	"chat-app/internal/server"
)

func main() {
	fmt.Println("starting the server...")

	config.ConnectDB()
	database.Migrate()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		server.StartWebSocketServer()
	}()

	go func() {
		defer wg.Done()
		server.StartAPIServer()
	}()

	wg.Wait()
}
