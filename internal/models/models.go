package models

import "time"

type User struct {
	ID        int
	Username  string
	CreatedAt time.Time
}

type Message struct {
	ID        int
	UserID    int
	Message   string
	Timestamp time.Time
}

type ConnectionLog struct {
	ID        int
	UserID    int
	Event     string
	Timestamp time.Time
}
