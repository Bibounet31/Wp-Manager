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
	"wp-manager/handlers"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var db *sql.DB
var templates *template.Template

func main() {
	var err error

	//get data from .env
	godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	dsn := convertScalingoDSN(dbURL)

	log.Println("🔍 Using DSN:", dsn)

	// error handling
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}

	// check if we can reach the db
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to db >.<")

	//logs if error from initDatabase()
	if err := initDatabase(); err != nil {
		log.Fatal(err)
	}

	// create template> so we can inject code directly into the html
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

	registerRoutes()

	// Static files
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("web/css"))))
	http.Handle("/scripts/", http.StripPrefix("/scripts/", http.FileServer(http.Dir("web/scripts"))))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// Convert Scalingo DSN
func convertScalingoDSN(dbURL string) string {
	u, err := url.Parse(dbURL)
	if err != nil {
		log.Fatal("invalid DATABASE_URL:", err)
	}

	credentials := u.User.String()
	host := u.Host
	dbName := strings.TrimPrefix(u.Path, "/")

	return fmt.Sprintf("%s@tcp(%s)/%s?parseTime=true", credentials, host, dbName)
}

// init all dbs:
func initDatabase() error {

	//user table
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

	//session table
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

	// comment all drops if /uploads are not deleted when container boot..
	//--------------------------------------------------
	//comments (since wallpaper are removed)
	_, err = db.Exec(`
		DROP TABLE IF EXISTS comments;`)
	//wallpapers
	_, err = db.Exec(`
		DROP TABLE IF EXISTS wallpapers;`)
	///-----------------------------------------------------------

	// table wallpapers
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS wallpapers (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			filename VARCHAR(255) NOT NULL,
			original_name VARCHAR(255) NOT NULL,
			file_path VARCHAR(500) NOT NULL,
			category VARCHAR(50) DEFAULT 'other',
			uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			toreview bool NOT NULL DEFAULT false,
		    ispublic bool NOT NULL DEFAULT false,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);
	`)
	if err != nil {
		return fmt.Errorf("wallpapers table: %w", err)
	}

	// table comments
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS comments (
			id INT AUTO_INCREMENT PRIMARY KEY,
			wallpaper_id INT NOT NULL,
			user_id INT NOT NULL,
			text TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (wallpaper_id) REFERENCES wallpapers(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			INDEX idx_wallpaper (wallpaper_id),
			INDEX idx_created (created_at)
		)
	`)
	if err != nil {
		return fmt.Errorf("comments table: %w", err)
	}

	log.Println("Database tables initialized~")
	return nil
}

// Register routes.. obviously
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
	http.HandleFunc("/adminpanel", handlers.AdminpannelHandler)
	http.HandleFunc("/admin/promote", handlers.PromoteUserHandler)
	http.HandleFunc("/admin/demote", handlers.DemoteUserHandler)
	http.HandleFunc("/admin/deleteacc", handlers.DeleteAccHandler)
	http.HandleFunc("/forgot-password", handlers.ForgotpasswordHandler)
	http.HandleFunc("/publish", handlers.PublishHandler)
	http.HandleFunc("/toreview", handlers.ReviewHandler)
	http.HandleFunc("/denypublish", handlers.DenyHandler)
	http.HandleFunc("/unpublish", handlers.UnpublishHandler)
	http.HandleFunc("/deletewp", handlers.DeletewpHandler)
	http.HandleFunc("/addfavorite", handlers.AddfavoriteHandler)
	http.HandleFunc("/rate", handlers.RateHandler)

	// API routes, get/post a comment without reloading the page ^^
	http.HandleFunc("/api/comments/", handlers.GetCommentsHandler)
	http.HandleFunc("/api/comments", handlers.PostCommentHandler)

	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("web/uploads"))))
}
