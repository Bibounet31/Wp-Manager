package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL != "" {
		// Parse Scalingo's DATABASE_URL format
		// From: mysql://user:pass@host:port/db
		// To: user:pass@tcp(host:port)/db

		dbURL = strings.Replace(dbURL, "mysql://", "", 1)

		// Split into credentials and host/db parts
		parts := strings.Split(dbURL, "@")
		if len(parts) == 2 {
			credentials := parts[0]
			hostAndDb := parts[1]

			// Rebuild in the correct format
			dbURL = credentials + "@tcp(" + hostAndDb + ")"
		}

		db, err := sql.Open("mysql", dbURL)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		err = db.Ping()
		if err != nil {
			log.Fatal(err)
		}

		log.Println("✅ Connected to database!")
	}

	// Routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Wallpaper Manager is running!")
	})

	// Get port from Scalingo
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
