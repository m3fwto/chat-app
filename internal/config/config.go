package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func ConnectDB() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error loading .env file: %v", err)
	}

	connStr := os.Getenv("DATABASE_URL")

	for i := 0; i < 5; i++ {
		DB, err = sql.Open("postgres", connStr)
		if err == nil {
			err = DB.Ping()
			if err == nil {
				fmt.Println("Connected to database")
				return
			}
		}
		fmt.Println("Waiting to connect to the database...")
		time.Sleep(5 * time.Second)
	}

	log.Fatalf("error connecting to database: %v", err)
}
