package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func parseScalingoDSN(dbURL string) string {
	// Convert mysql://user:pass@host:port/db to user:pass@tcp(host:port)/db
	dbURL = strings.TrimPrefix(dbURL, "mysql://")
	dbURL = strings.Split(dbURL, "?")[0] // Remove query params

	parts := strings.SplitN(dbURL, "@", 2)
	if len(parts) != 2 {
		return dbURL
	}

	credentials := parts[0]
	hostAndDb := parts[1]

	hostParts := strings.SplitN(hostAndDb, "/", 2)
	if len(hostParts) != 2 {
		return dbURL
	}

	return fmt.Sprintf("%s@tcp(%s)/%s?tls=skip-verify", credentials, hostParts[0], hostParts[1])
}

func main() {
	// Database connection
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		db, err := sql.Open("mysql", parseScalingoDSN(dbURL))
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		if err := db.Ping(); err != nil {
			log.Fatal(err)
		}

		log.Println("✅ Connected to database!")
	}

	// Routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		render(w, "index.html", nil)
	})
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		render(w, "login.html", nil)
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func render(w http.ResponseWriter, file string, data any) {
	funcMap := template.FuncMap{
		"add":        func(a, b int) int { return a + b },
		"pathEscape": url.PathEscape,
	}

	t, err := template.New(file).Funcs(funcMap).ParseFiles("web/html/" + file)
	if err != nil {
		http.Error(w, "Template not found", 500)
		return
	}
	t.Execute(w, data)
}
