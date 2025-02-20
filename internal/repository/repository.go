package repository

import (
	"chat-app/internal/config"
	"chat-app/internal/database"
	"log"
)

type Message struct {
	User    string `json:"user"`
	Time    string `json:"time"`
	Message string `json:"message"`
}

type ConnectionLog struct {
	User  string `json:"user"`
	Time  string `json:"time"`
	Event string `json:"event"`
}

func GetMessages(page, pageSize int) ([]Message, int) {
	db := database.GetDB()
	var messages []Message
	var total int

	err := db.QueryRow("SELECT COUNT(*) FROM messages").Scan(&total)
	if err != nil {
		log.Println("error while counting messages:", err)
		return nil, 0
	}

	offset := (page - 1) * pageSize
	rows, err := db.Query(`
		SELECT users.username, messages.timestamp, messages.message 
		FROM messages 
		JOIN users ON messages.user_id = users.id 
		ORDER BY messages.timestamp DESC 
		LIMIT $1 OFFSET $2`, pageSize, offset)

	if err != nil {
		log.Println("error while retrieving message history:", err)
		return nil, 0
	}
	defer rows.Close()

	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.User, &msg.Time, &msg.Message)
		if err != nil {
			log.Println("error processing message string:", err)
			continue
		}
		messages = append(messages, msg)
	}

	return messages, total
}

func GetConnectionLogs() []ConnectionLog {
	db := database.GetDB()
	var logs []ConnectionLog

	rows, err := db.Query(`
		SELECT users.username, connections.timestamp, connections.event 
		FROM connections 
		JOIN users ON connections.user_id = users.id 
		ORDER BY connections.timestamp DESC`)

	if err != nil {
		log.Println("error when receiving connection logs:", err)
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var logEntry ConnectionLog
		err := rows.Scan(&logEntry.User, &logEntry.Time, &logEntry.Event)
		if err != nil {
			log.Println("error when processing connection log string:", err)
			continue
		}
		logs = append(logs, logEntry)
	}

	return logs
}

func CheckUserExists(username string) bool {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username=$1)`
	err := config.DB.QueryRow(query, username).Scan(&exists)
	if err != nil {
		log.Println("user verification error:", err)
		return false
	}
	return exists
}

func CreateUser(username string) {
	query := `INSERT INTO users (username) VALUES ($1) ON CONFLICT DO NOTHING`
	_, err := config.DB.Exec(query, username)
	if err != nil {
		log.Println("error creating user:", err)
	}
}

func SaveMessage(username, message string) {
	query := `INSERT INTO messages (user_id, message) VALUES ((SELECT id FROM users WHERE username=$1), $2)`
	_, err := config.DB.Exec(query, username, message)
	if err != nil {
		log.Println("error saving message:", err)
	}
}

func LogConnection(username, event string) {
	query := `INSERT INTO connections (user_id, event) VALUES ((SELECT id FROM users WHERE username=$1), $2)`
	_, err := config.DB.Exec(query, username, event)
	if err != nil {
		log.Println("connection logging error:", err)
	}
}
