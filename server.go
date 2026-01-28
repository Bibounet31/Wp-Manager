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

		dbURL = strings.TrimPrefix(dbURL, "mysql://")

		// Find the @ that separates credentials from host
		atIndex := strings.Index(dbURL, "@")
		if atIndex != -1 {
			credentials := dbURL[:atIndex]
			hostAndDb := dbURL[atIndex+1:]

			// Find the / that separates host:port from database
			slashIndex := strings.Index(hostAndDb, "/")
			if slashIndex != -1 {
				hostPort := hostAndDb[:slashIndex]
				database := hostAndDb[slashIndex+1:]

				// Rebuild in correct format
				dbURL = fmt.Sprintf("%s@tcp(%s)/%s", credentials, hostPort, database)
			}
		}

		log.Printf("Connecting to database...")
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
