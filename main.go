package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

var db *sql.DB

func main() {
	// Database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable not set")
	}

	var err error
	db, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	// Test connection
	if err = db.Ping(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	log.Println("Connected to PostgreSQL")

	// Routes
	http.HandleFunc("/api/register", registerHandler)
	http.HandleFunc("/api/login", loginHandler)
	http.HandleFunc("/api/profile", authMiddleware(getProfileHandler))
	http.HandleFunc("/api/profile/update", authMiddleware(updateProfileHandler))
	http.HandleFunc("/api/profile/delete", authMiddleware(deleteProfileHandler))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
