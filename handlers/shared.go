// / this file contains everything that is needed for the handlers to work properly
package handlers

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"
)

var (
	db        *sql.DB
	templates *template.Template
)

func SetDB(database *sql.DB) {
	db = database
}

func SetTemplates(tmpl *template.Template) {
	templates = tmpl
}

type UserProfile struct {
	Username string
	Email    string
	Name     string
	Surname  string
	IsAdmin  bool
	UserID   int
}

type AdminPanelData struct {
	CurrentUser *UserProfile
	AllUsers    []UserProfile
}

type Wallpaper struct {
	ID           int
	UserID       int
	Filename     string
	OriginalName string
	FilePath     string
	UploadedAt   time.Time
	IsPublic     bool
}

type WallpapersPageData struct {
	Wallpapers []Wallpaper
	Username   string
	IsAdmin    bool
}

type PageData struct {
	Username string
	IsAdmin  bool
}

// returns page data with user info
func getPageData(r *http.Request) PageData {
	data := PageData{}
	user := getCurrentUser(r)
	if user != nil {
		data.Username = user.Username
		data.IsAdmin = user.IsAdmin
	}
	return data
}

// returns user ID from session WITH expiry check
func getUserIDFromSession(r *http.Request) (int, error) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return 0, fmt.Errorf("no session cookie")
	}

	var userID int
	var expiresAt time.Time
	err = db.QueryRow("SELECT user_id, expires_at FROM sessions WHERE id = ?", cookie.Value).
		Scan(&userID, &expiresAt)
	if err != nil {
		return 0, fmt.Errorf("session not found: %w", err)
	}

	// Check if session expired
	if time.Now().After(expiresAt) {
		db.Exec("DELETE FROM sessions WHERE id = ?", cookie.Value)
		return 0, fmt.Errorf("session expired")
	}

	return userID, nil
}

// returns the logged-in user or nil if not logged in
func getCurrentUser(r *http.Request) *UserProfile {
	userID, err := getUserIDFromSession(r)
	if err != nil {
		return nil
	}

	var user UserProfile
	err = db.QueryRow("SELECT id, username, email, name, surname, isadmin FROM users WHERE id = ?", userID).
		Scan(&user.UserID, &user.Username, &user.Email, &user.Name, &user.Surname, &user.IsAdmin)
	if err != nil {
		return nil
	}

	return &user
}

// prints all users to console
func printAllUsers() {
	rows, err := db.Query("SELECT id, username, email, name, surname, created_at FROM users")
	if err != nil {
		log.Println("Failed to query users:", err)
		return
	}
	defer rows.Close()

	fmt.Println("📋 Current users in the database:")
	for rows.Next() {
		var id int
		var username, email, name, surname string
		var createdAt time.Time
		if err := rows.Scan(&id, &username, &email, &name, &surname, &createdAt); err != nil {
			log.Println("Row scan error:", err)
			continue
		}
		fmt.Printf("ID: %d | Username: %s | Email: %s | Name: %s %s | Created: %s\n",
			id, username, email, name, surname, createdAt.Format("2006-01-02 15:04:05"))
	}
}
