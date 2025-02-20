package database

import (
	"chat-app/internal/config"
	"log"
)

func Migrate() {
	query := `
	CREATE TABLE IF NOT EXISTS users (
	    id SERIAL PRIMARY KEY,
	    username VARCHAR(10) UNIQUE NOT NULL,
	    created_at TIMESTAMP DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS messages (
	    id SERIAL PRIMARY KEY,
	    user_id INT REFERENCES users(id),
	    message TEXT NOT NULL,
	    timestamp TIMESTAMP DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS connections (
	    id SERIAL PRIMARY KEY,
	    user_id INT REFERENCES users(id),
	    event VARCHAR(20) NOT NULL,
	    timestamp TIMESTAMP DEFAULT NOW()
	);`

	_, err := config.DB.Exec(query)
	if err != nil {
		log.Fatalf("database migration failed: %v", err)
	}

	log.Println("Database migration is complete")
}
