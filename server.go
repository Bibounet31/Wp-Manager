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
	"time"
	"wp-manager/handlers"
	"wp-manager/security"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB
var templates *template.Template

func main() {
	var err error
	var dsn string

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Println("⚠️  Using local database")
		dsn = "wpmanager:secret123@tcp(127.0.0.1:3306)/wallpaper_manager?parseTime=true"
	} else {
		dsn = parseScalingoDSN(dbURL)
	}

	log.Println("🔍 Using DSN:", dsn)

	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}

	// Connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	log.Println("✅ Connected to database!")

	if err := initDatabase(); err != nil {
		log.Fatal(err)
	}

	// Parse templates
	templates = template.Must(
		template.New("").
			Funcs(template.FuncMap{
				"add":        func(a, b int) int { return a + b },
				"pathEscape": url.PathEscape,
			}).
			ParseGlob("web/html/*.html"),
	)

	// Initialize handlers with database and templates
	handlers.SetDB(db)
	handlers.SetTemplates(templates)

	// Register routes
	registerRoutes()

	// Static files
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("web/css"))))
	http.Handle("/scripts/", http.StripPrefix("/scripts/", http.FileServer(http.Dir("web/scripts"))))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	go cleanupExpiredCSRFTokens()
	log.Printf("🚀 Server running on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// Convert Scalingo DSN
func parseScalingoDSN(dbURL string) string {
	dbURL = strings.TrimPrefix(dbURL, "mysql://")
	dbURL = strings.Split(dbURL, "?")[0]

	parts := strings.SplitN(dbURL, "@", 2)
	credentials := parts[0]
	hostAndDb := parts[1]

	hostParts := strings.SplitN(hostAndDb, "/", 2)
	host := hostParts[0]
	dbName := hostParts[1]

	return fmt.Sprintf("%s@tcp(%s)/%s?parseTime=true", credentials, host, dbName)
}

// Initialize database tables
func initDatabase() error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(50) UNIQUE NOT NULL,
			email VARCHAR(100) UNIQUE NOT NULL,
			name VARCHAR(50),
			surname VARCHAR(50),
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			isadmin bool NOT NULL DEFAULT false
		)
	`)
	if err != nil {
		return fmt.Errorf("users table: %w", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			id VARCHAR(36) PRIMARY KEY,
			user_id INT NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("sessions table: %w", err)
	}

	_, err = db.Exec(`
	DROP TABLE IF EXISTS wallpapers;
`)

	_, err = db.Exec(`
		CREATE TABLE wallpapers (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			filename VARCHAR(255) NOT NULL,
			original_name VARCHAR(255) NOT NULL,
			file_path VARCHAR(500) NOT NULL,
			uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			toreview bool NOT NULL DEFAULT false,
		    ispublic bool NOT NULL DEFAULT false,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);
	`)
	if err != nil {
		return fmt.Errorf("wallpapers table: %w", err)
	}

	log.Println("✅ Database tables initialized!")
	return nil
}

func cleanupExpiredCSRFTokens() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		log.Println("🧹 Cleaning up expired CSRF tokens...")
	}
}

// HTTP routes
func registerRoutes() {
	http.HandleFunc("/", handlers.IndexHandler)
	http.HandleFunc("/community", handlers.CommunityHandler)

	// Auth routes
	http.HandleFunc("/login", security.Middleware(handlers.LoginHandler))
	http.HandleFunc("/register", security.Middleware(handlers.RegisterHandler))
	http.HandleFunc("/logout", security.Middleware(handlers.LogoutHandler))

	// Protected routes
	http.HandleFunc("/wallpapers", security.Middleware(handlers.WallpapersHandler))
	http.HandleFunc("/profile", security.Middleware(handlers.ProfileHandler))
	http.HandleFunc("/upload", security.Middleware(handlers.UploadHandler))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("web/uploads"))))
	http.HandleFunc("/rename", security.Middleware(handlers.RenameHandler))
	http.HandleFunc("/publish", security.Middleware(handlers.PublishHandler))
	http.HandleFunc("/toreview", security.Middleware(handlers.ReviewHandler))

	// Admin routes
	http.HandleFunc("/adminpannel", security.Middleware(handlers.AdminpannelHandler))
	http.HandleFunc("/admin/promote", security.Middleware(handlers.PromoteUserHandler))
	http.HandleFunc("/admin/demote", security.Middleware(handlers.DemoteUserHandler))
	http.HandleFunc("/admin/deleteacc", security.Middleware(handlers.DeleteAccHandler))

}
