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

// Register all HTTP routes
func registerRoutes() {
	http.HandleFunc("/", handlers.IndexHandler)
	http.HandleFunc("/community", handlers.CommunityHandler)
	http.HandleFunc("/wallpapers", handlers.WallpapersHandler)
	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/profile", handlers.ProfileHandler)
	http.HandleFunc("/logout", handlers.LogoutHandler)
	http.HandleFunc("/upload", handlers.UploadHandler)
	http.HandleFunc("/rename", handlers.RenameHandler)
	http.HandleFunc("/adminpannel", handlers.AdminpannelHandler)
	http.HandleFunc("/admin/promote", handlers.PromoteUserHandler)
	http.HandleFunc("/admin/demote", handlers.DemoteUserHandler)
	http.HandleFunc("/admin/deleteacc", handlers.DeleteAccHandler)
	http.HandleFunc("/forgot-password", handlers.ForgotpasswordHandler)
	http.HandleFunc("/publish", handlers.PublishHandler)
	http.HandleFunc("/toreview", handlers.ReviewHandler)
	http.HandleFunc("/denypublish", handlers.DenyHandler)
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("web/uploads"))))
}
