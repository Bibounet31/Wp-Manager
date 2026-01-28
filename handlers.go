package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserProfile struct {
	Username string
	Email    string
	Name     string
	Surname  string
}

// Register all HTTP routes
func registerRoutes() {
	http.HandleFunc("/", page("index.html"))
	http.HandleFunc("/community", page("community.html"))
	http.HandleFunc("/wallpapers", page("wallpapers.html"))
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/profile", profileHandler)
	http.HandleFunc("/logout", logoutHandler)

}

// page renders GET-only pages
func page(file string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := templates.ExecuteTemplate(w, file, nil); err != nil {
			log.Println("Template error:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

// Handles user registration
func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		username := r.FormValue("username")
		email := r.FormValue("mail")
		name := r.FormValue("name")
		surname := r.FormValue("surname")
		password := r.FormValue("password")

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Failed to hash password", http.StatusInternalServerError)
			return
		}

		_, err = db.Exec(`
            INSERT INTO users (username, email, name, surname, password_hash)
            VALUES (?, ?, ?, ?, ?)`,
			username, email, name, surname, hashedPassword,
		)
		if err != nil {
			http.Error(w, "Failed to save user: "+err.Error(), http.StatusInternalServerError)
			return
		}

		printAllUsers()

		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// GET → show form
	templates.ExecuteTemplate(w, "register.html", nil)
}

// Handles user login
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseForm()
		username := r.FormValue("username")
		password := r.FormValue("password")

		var userID int
		var hashedPassword string

		err := db.QueryRow("SELECT id, password_hash FROM users WHERE username = ?", username).Scan(&userID, &hashedPassword)
		if err != nil {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}

		// Create session
		sessionID := uuid.New().String()
		expiresAt := time.Now().Add(24 * time.Hour)

		_, _ = db.Exec("INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)", sessionID, userID, expiresAt)

		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    sessionID,
			Expires:  expiresAt,
			Path:     "/",
			HttpOnly: true,
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// GET → show login form
	templates.ExecuteTemplate(w, "login.html", nil)
}

// Print all users in console
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

// profileHandler shows user info if logged in
func profileHandler(w http.ResponseWriter, r *http.Request) {
	// Check if session cookie exists
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	sessionID := cookie.Value

	// Get user ID from session
	var userID int
	var expiresAt time.Time
	err = db.QueryRow("SELECT user_id, expires_at FROM sessions WHERE id = ?", sessionID).
		Scan(&userID, &expiresAt)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Check if session expired
	if time.Now().After(expiresAt) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get user info
	var user UserProfile
	err = db.QueryRow("SELECT username, email, name, surname FROM users WHERE id = ?", userID).
		Scan(&user.Username, &user.Email, &user.Name, &user.Surname)
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}

	// Render profile page with struct
	if err := templates.ExecuteTemplate(w, "profile.html", user); err != nil {
		log.Println("Profile template error:", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err == nil {
		db.Exec("DELETE FROM sessions WHERE id = ?", cookie.Value)
		cookie.MaxAge = -1
		http.SetCookie(w, cookie)
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
